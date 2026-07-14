package claudecli

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
	"github.com/ndscm/theseed/seed/newtype/amadeus/playpen"
	"github.com/ndscm/theseed/seed/newtype/gajetto/payload/claudepayload"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	stdoutBufferInitial = 64 * 1024
	stdoutBufferMax     = 16 * 1024 * 1024
	stderrBufferInitial = 64 * 1024
	stderrBufferMax     = 1 * 1024 * 1024
)

type ongoingTracker struct {
	mutex sync.Mutex
	cond  *sync.Cond

	taskUuid string

	cmdExited bool
}

func newOngoingTracker() *ongoingTracker {
	tracker := &ongoingTracker{}
	tracker.cond = sync.NewCond(&tracker.mutex)
	return tracker
}

// waitAdmit blocks until no request is in flight, then reserves the single
// in-flight slot and returns true, enforcing at most one simultaneous input
// per topic. A second input therefore waits for the first to complete
// instead of being rejected. It returns false without reserving when the
// subprocess has already exited, so the caller can fail the request rather
// than write to a dead subprocess.
func (t *ongoingTracker) waitAdmit(taskUuid string) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for t.taskUuid != "" && !t.cmdExited {
		t.cond.Wait()
	}
	if t.cmdExited {
		return false
	}
	t.taskUuid = taskUuid
	t.cond.Broadcast()
	return true
}

func (t *ongoingTracker) release() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.taskUuid != "" {
		t.taskUuid = ""
	} else {
		seedlog.Warnf("received result with no ongoing request")
	}
	t.cond.Broadcast()
}

func (t *ongoingTracker) onCmdExit() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.cmdExited = true
	t.cond.Broadcast()
}

func (t *ongoingTracker) waitIdle() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for t.taskUuid != "" && !t.cmdExited {
		t.cond.Wait()
	}
}

type thinkRequest struct {
	input *brainpb.BrainInput
	err   chan error
}

type topicRunner struct {
	topic    string
	topicDir string

	runnerCtx    context.Context
	runnerCancel context.CancelFunc

	stdin io.WriteCloser

	// wait blocks until the claude process exits, returning its exit error.
	// It abstracts over the two launch modes so waitCmd need not know whether
	// claude runs as a host subprocess or inside the playpen container.
	wait func() error

	// thinkQueue serializes stdin writes through thinkLoop. Using a channel
	// instead of a mutex lets Input select on ctx.Done() when the
	// subprocess stalls reading stdin, rather than queueing behind a
	// stuck Write call that holds the lock. It is buffered so that a burst
	// of Input callers can enqueue without blocking on thinkLoop draining
	// the previous request; once the buffer fills, Input blocks until a
	// slot frees or its context is cancelled.
	thinkQueue chan thinkRequest

	handlerMutex sync.RWMutex
	handler      brain.BrainStepHandler

	// ongoing tracks the in-flight request: thinkLoop admits at most one at
	// a time via waitAdmit (a second input blocks in thinkLoop until the
	// first completes rather than being rejected), and releases the slot
	// when a stream output line of type "result" is received. The slot must
	// be reserved in thinkLoop (not after the caller reads req.err) so a fast
	// "result" line can't release a slot that was never reserved.
	// Hibernate(wait=true) uses it to block until the subprocess has drained
	// or exited.
	ongoing *ongoingTracker

	done chan struct{}
}

// brainProc is the launched claude process, either a subprocess on the host or
// a process running inside the playpen container. Both expose the same stdio
// pipes plus a wait, so the runner's loops don't care which is in use.
type brainProc struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	wait   func() error
	pid    int
}

// startBrainProc launches the claude CLI in stream-json mode for a topic. With
// a nil playpen controller it runs as a subprocess on the host with topicDir as
// its working directory; otherwise it runs inside the playpen container with
// topicDir as the container-side working directory, confining the model to the
// container's filesystem and process namespace.
//
// Security note: claude is started with --permission-mode=bypassPermissions,
// which disables every interactive permission prompt the CLI would otherwise
// raise. This is the only practical mode for an unattended agent — there is no
// human available to approve file edits, shell commands, or network calls at
// the prompt — but it also means the `claude` process has the full file-system
// and shell access of the user it runs as. Treat the container itself as the
// security boundary: anything reachable from the process is reachable by the
// model, which is why a playpen confines it to the container.
func startBrainProc(
	ctx context.Context, topicDir string, pc *playpen.PlaypenController,
) (*brainProc, error) {
	claudeArgs := []string{
		"--continue",
		"--thinking-display", "summarized",
		"--input-format", "stream-json",
		"--model", "opus",
		"--output-format", "stream-json",
		"--permission-mode", "bypassPermissions",
		"--print",
		"--verbose",
	}
	if pc != nil {
		// Run claude under the playpen user's login shell so ~/.zshrc is
		// sourced (PATH and environment) before claude starts, then delegate
		// the session to claude so it owns the streams for continuous stdin.
		tty, err := pc.StartShell(ctx, "/usr/bin/zsh", []string{"-i"})
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = tty.Delegate(topicDir, "claude", claudeArgs)
		if err != nil {
			// StartShell already started the podman exec client; reap it before
			// returning so it (and its os/exec kill goroutine) isn't leaked.
			closeErr := tty.Close()
			if closeErr != nil {
				seedlog.Warnf("topic %q: error closing tty after delegate failure: %v",
					topicDir, closeErr)
			}
			return nil, seederr.Wrap(err)
		}
		return &brainProc{
			stdin:  tty.Stdin,
			stdout: tty.Stdout,
			stderr: tty.Stderr,
			wait:   tty.Wait,
			pid:    tty.Pid(),
		}, nil
	}

	cmd := exec.CommandContext(ctx, "claude", claudeArgs...)
	cmd.Dir = topicDir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = cmd.Start()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &brainProc{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		wait:   cmd.Wait,
		pid:    cmd.Process.Pid,
	}, nil
}

// newTopicRunner spawns the per-topic `claude` process in stream-json mode
// under topicDir and returns a runner that serializes stdin writes and
// dispatches stdout JSON lines as BrainSteps. When pc is non-nil the process
// runs inside the playpen container instead of directly on the host.
func newTopicRunner(topic string, topicDir string, pc *playpen.PlaypenController) (*topicRunner, error) {
	ctx, cancel := context.WithCancel(context.Background())

	proc, err := startBrainProc(ctx, topicDir, pc)
	if err != nil {
		cancel()
		return nil, seederr.Wrap(err)
	}

	tr := &topicRunner{
		topic:        topic,
		topicDir:     topicDir,
		runnerCtx:    ctx,
		runnerCancel: cancel,
		stdin:        proc.stdin,
		wait:         proc.wait,
		thinkQueue:   make(chan thinkRequest, 32),
		ongoing:      newOngoingTracker(),
		done:         make(chan struct{}),
	}

	go tr.thinkLoop()
	go tr.readStdout(proc.stdout)
	go tr.readStderr(proc.stderr)
	go tr.waitCmd()

	seedlog.Infof("topic %q: claude started (pid=%d, dir=%s)",
		topic, proc.pid, topicDir)

	return tr, nil
}

// thinkLoop serializes writes to tr.stdin so that Input callers never
// share a live Write call. It exits when runnerCtx is cancelled; a Write
// that is stuck on a stalled subprocess is unblocked by Close() closing
// stdin.
func (tr *topicRunner) thinkLoop() {
	for {
		select {
		case req := <-tr.thinkQueue:
			// Block until any prior request has drained so at most one is
			// ever in flight. The request stays held here rather than being
			// rejected, preserving it for delivery once the slot frees.
			taskUuid := req.input.GetTaskUuid()
			if taskUuid == "" {
				req.err <- seederr.WrapErrorf("task uuid is empty")
				continue
			}
			if !tr.ongoing.waitAdmit(taskUuid) {
				req.err <- seederr.WrapErrorf(
					"topic %q: runner closed", tr.topic)
				continue
			}
			payload := claudepayload.StreamInputUser{
				StreamInputEnvelope: claudepayload.StreamInputEnvelope{Type: "user"},
				Message: &claudepayload.StreamInputMessage{
					Role:    "user",
					Content: req.input.GetText(),
				},
			}
			line, err := json.Marshal(payload)
			if err != nil {
				tr.ongoing.release()
				req.err <- seederr.Wrap(err)
				continue
			}
			line = append(line, '\n')
			_, err = tr.stdin.Write(line)
			if err != nil {
				tr.ongoing.release()
			}
			req.err <- err
		case <-tr.runnerCtx.Done():
			return
		}
	}
}

func (tr *topicRunner) readStdout(stdout io.ReadCloser) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, stdoutBufferInitial), stdoutBufferMax)
	for scanner.Scan() {
		tr.dispatchLine(scanner.Bytes())
	}
	err := scanner.Err()
	if err != nil {
		seedlog.Errorf("topic %q: stdout scan: %v", tr.topic, err)
	}
}

func (tr *topicRunner) readStderr(stderr io.ReadCloser) {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, stderrBufferInitial), stderrBufferMax)
	for scanner.Scan() {
		seedlog.Warnf("topic %q: claude stderr: %s", tr.topic, scanner.Text())
	}
}

func (tr *topicRunner) waitCmd() {
	defer func() {
		// Wake any Hibernate(wait=true) callers blocked on ongoing so they
		// don't deadlock when the subprocess exits without emitting the
		// expected "result" lines.
		tr.ongoing.onCmdExit()
		close(tr.done)
	}()
	err := tr.wait()
	if err != nil {
		seedlog.Warnf("topic %q: claude exited: %v", tr.topic, err)
		return
	}
	seedlog.Infof("topic %q: claude exited cleanly", tr.topic)
}

func (tr *topicRunner) dispatchLine(line []byte) {
	stepType, data, err := claudepayload.DecodeStreamOutputData(line)
	if err != nil {
		seedlog.Warnf("topic %q: unparsable stdout line: %v: %s",
			tr.topic, err, string(line))
		return
	}

	step := &brainpb.BrainStep{
		Uuid:      uuid.NewString(),
		Timestamp: timestamppb.Now(),
		Type:      stepType,
		Topic:     tr.topic,
		TaskUuid:  tr.ongoing.taskUuid,
		Data:      data,
	}

	if step.Type == "result" {
		tr.ongoing.release()
	}

	tr.handlerMutex.RLock()
	handler := tr.handler
	tr.handlerMutex.RUnlock()

	if handler == nil {
		return
	}
	handler.HandleBrainStep(tr.runnerCtx, tr.topic, step)
}

func (tr *topicRunner) Input(ctx context.Context, input *brainpb.BrainInput) error {
	req := thinkRequest{input: input, err: make(chan error, 1)}
	select {
	case tr.thinkQueue <- req:
	case <-ctx.Done():
		return seederr.Wrap(ctx.Err())
	case <-tr.runnerCtx.Done():
		return seederr.WrapErrorf("topic %q: runner closed", tr.topic)
	}

	select {
	case err := <-req.err:
		if err != nil {
			return seederr.Wrap(err)
		}
		return nil
	case <-ctx.Done():
		return seederr.Wrap(ctx.Err())
	}
}

func (tr *topicRunner) RegisterStepHandler(handler brain.BrainStepHandler) error {
	tr.handlerMutex.Lock()
	defer tr.handlerMutex.Unlock()
	if tr.handler != nil {
		return seederr.WrapErrorf("topic %q: step handler already registered", tr.topic)
	}
	tr.handler = handler
	return nil
}

func (tr *topicRunner) Hibernate(wait bool) error {
	if wait {
		tr.ongoing.waitIdle()
	}
	err := tr.stdin.Close()
	if err != nil {
		seedlog.Warnf("topic %q: error closing stdin: %v", tr.topic, err)
	}
	tr.runnerCancel()
	<-tr.done
	return seederr.Wrap(err)
}
