package claudecli

import (
	"bufio"
	"context"
	"encoding/json/v2"
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

type inputPromise struct {
	input *brainpb.BrainInput
	err   chan error
}

// inputScheduler groups pending inputs by threadUuid so that inputs for the
// thread currently occupying stdin (the "ongoing" thread) stream straight through,
// while inputs for other threads wait their turn. At most one thread is ongoing at
// a time. The claude CLI emits a single "result" line for however many inputs
// were streamed to it as one turn, so a "result" frees the whole ongoing thread
// rather than a single input; the freed slot is then handed to the next thread.
// Inputs for a non-ongoing thread are buffered in their thread's group, and the
// groups are promoted in first-arrival order once the ongoing thread frees the
// slot.
type inputScheduler struct {
	mutex sync.Mutex
	cond  *sync.Cond

	// ongoingThreadUuid is the thread currently allowed to write to stdin, or ""
	// when idle.
	ongoingThreadUuid string

	// order holds the threadUuids of buffered groups in first-arrival order;
	// pending maps each to its buffered inputs. A thread appears in order at most
	// once, added when its first input is buffered.
	order   []string
	pending map[string][]inputPromise

	cmdExited bool
}

// ongoing returns the ongoing threadUuid so stdout lines can be tagged with
// the thread that produced them.
func (s *inputScheduler) ongoing() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.ongoingThreadUuid
}

// submit records req in its thread's group so the scheduler loop can hand it to
// stdin once the thread is (or becomes) ongoing. It never writes to stdin itself;
// the scheduler loop pulls writable inputs out via popWritable. It returns an
// error without recording anything when the input has no thread uuid, or once the
// subprocess has exited, so the caller can fail the request rather than queue it
// for a dead subprocess.
func (s *inputScheduler) submit(req inputPromise) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.cmdExited {
		return seederr.WrapErrorf("subprocess exited")
	}
	threadUuid := req.input.GetThreadUuid()
	if threadUuid == "" {
		return seederr.WrapErrorf("thread uuid is empty")
	}
	// A thread is listed in order exactly while it has buffered inputs and is not
	// the ongoing one; add it the first time such an input appears.
	if threadUuid != s.ongoingThreadUuid && len(s.pending[threadUuid]) == 0 {
		s.order = append(s.order, threadUuid)
	}
	s.pending[threadUuid] = append(s.pending[threadUuid], req)
	return nil
}

// popWritable removes and returns the next input to write. When the slot is
// idle it promotes the first buffered thread group to ongoing, so a whole group's
// inputs are drained (and same-thread inputs keep streaming) before any other
// thread takes over. It returns nil when nothing is writable right now: either a
// thread is ongoing but its group is momentarily empty because its inputs are in
// flight awaiting a "result", or nothing is buffered at all.
func (s *inputScheduler) popWritable() *inputPromise {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.ongoingThreadUuid == "" {
		if len(s.order) == 0 {
			return nil
		}
		s.ongoingThreadUuid = s.order[0]
		s.order = s.order[1:]
		s.cond.Broadcast()
	}
	group := s.pending[s.ongoingThreadUuid]
	if len(group) == 0 {
		return nil
	}
	req := group[0]
	rest := group[1:]
	if len(rest) == 0 {
		delete(s.pending, s.ongoingThreadUuid)
	} else {
		s.pending[s.ongoingThreadUuid] = rest
	}
	return &req
}

// claimWrite reports whether req may still be written. It is called just before
// the write, with req already removed from its group by popWritable but its thread
// no longer guaranteed to be ongoing: a "result" for that thread may have arrived,
// or the slot may have been promoted to another thread, in the window between
// popWritable and here. When req's thread is still ongoing the write proceeds
// (true). Otherwise req is an un-written input for a thread that is no longer
// ongoing, so it is put back at the front of its group — restoring the state
// popWritable left — to be promoted and written as a fresh turn, and false is
// returned so the caller skips the write and leaves req's caller waiting for
// that later write. It returns an error only once the subprocess has exited, so
// the caller fails req rather than re-buffering it for a dead subprocess.
func (s *inputScheduler) claimWrite(req *inputPromise) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	threadUuid := req.input.GetThreadUuid()
	if s.ongoingThreadUuid == threadUuid {
		return true, nil
	}
	if s.cmdExited {
		return false, seederr.WrapErrorf("subprocess exited")
	}
	// A thread is listed in order exactly while it has buffered inputs and is not
	// ongoing; re-add it only if popWritable's delete emptied the group.
	if len(s.pending[threadUuid]) == 0 {
		s.order = append(s.order, threadUuid)
	}
	s.pending[threadUuid] = append([]inputPromise{*req}, s.pending[threadUuid]...)
	s.cond.Broadcast()
	return false, nil
}

// completeOngoing frees the ongoing thread's slot when its "result" line arrives.
// Inputs for that thread that were submitted but not yet written form a fresh
// turn, so they are re-queued rather than orphaned. It returns nil when a thread
// was actually ongoing, so the caller can nudge the scheduler loop to promote
// the next group; it returns an error when the slot was already free, which
// means a "result" arrived that no in-flight input can account for.
func (s *inputScheduler) completeOngoing() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.ongoingThreadUuid == "" {
		return seederr.WrapErrorf("result with no ongoing thread")
	}
	old := s.ongoingThreadUuid
	s.ongoingThreadUuid = ""
	if len(s.pending[old]) > 0 {
		s.order = append(s.order, old)
	}
	s.cond.Broadcast()
	return nil
}

// abandonOngoing frees the ongoing thread's slot after a write to it failed and
// returns its remaining un-written inputs so the caller can fail them too. It
// does not re-queue them, since the failed write means stdin is broken and
// retrying would only spin. It returns nil when no thread was ongoing.
func (s *inputScheduler) abandonOngoing() []inputPromise {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.ongoingThreadUuid == "" {
		return nil
	}
	old := s.ongoingThreadUuid
	s.ongoingThreadUuid = ""
	leftover := s.pending[old]
	delete(s.pending, old)
	s.cond.Broadcast()
	return leftover
}

// drain removes and returns every buffered input so callers blocked on
// them can be failed when the runner is shutting down or the subprocess has
// exited. Inputs already written are not returned; their callers have already
// been answered.
func (s *inputScheduler) drain() []inputPromise {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	remains := []inputPromise{}
	for _, group := range s.pending {
		remains = append(remains, group...)
	}
	s.order = nil
	s.pending = make(map[string][]inputPromise)
	s.ongoingThreadUuid = ""
	s.cond.Broadcast()
	return remains
}

func (s *inputScheduler) onCmdExit() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.cmdExited = true
	s.cond.Broadcast()
}

// waitIdle blocks until no thread is ongoing and nothing is buffered, or the
// subprocess has exited.
func (s *inputScheduler) waitIdle() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for (s.ongoingThreadUuid != "" || len(s.order) > 0) && !s.cmdExited {
		s.cond.Wait()
	}
}

func newInputScheduler() *inputScheduler {
	s := &inputScheduler{pending: make(map[string][]inputPromise)}
	s.cond = sync.NewCond(&s.mutex)
	return s
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

type topicRunner struct {
	topic    string
	topicDir string

	// inbox delivers inputs to startSchedulerLoop, which records them in scheduler.
	// Using a channel instead of a mutex lets Input select on ctx.Done() when
	// the subprocess stalls reading stdin, rather than queueing behind a stuck
	// Write call that holds a lock. It is buffered so that a burst of Input
	// callers can enqueue without blocking on the scheduler loop; once the
	// buffer fills, Input blocks until a slot frees or its context is cancelled.
	inbox chan inputPromise

	// promote nudges startSchedulerLoop to re-evaluate after the ongoing thread
	// frees its slot. It is signalled non-blockingly from the stdout goroutine
	// (on a "result" line) and from a failed write; a buffer of one coalesces
	// redundant nudges since at most one promotion is pending at a time.
	promote chan struct{}

	// next carries the single next input to write from startSchedulerLoop to
	// startWriteLoop. It is unbuffered so that at most one input is ever in
	// flight outside the scheduler, keeping shutdown accounting exact.
	next chan *inputPromise

	// scheduler groups pending inputs by threadUuid. Inputs for the ongoing thread
	// stream straight to stdin; inputs for other threads are buffered and
	// promoted a whole group at a time once the ongoing thread frees its slot
	// (signalled by a stream output line of type "result"). Hibernate(wait=true)
	// uses it to block until the runner is idle or the subprocess has exited.
	scheduler *inputScheduler

	runnerCtx    context.Context
	runnerCancel context.CancelFunc

	stdin io.WriteCloser

	// wait blocks until the claude process exits, returning its exit error.
	// It abstracts over the two launch modes so waitCmd need not know whether
	// claude runs as a host subprocess or inside the playpen container.
	wait func() error

	handlerMutex sync.RWMutex
	handler      brain.BrainStepHandler

	done chan struct{}
}

// signalPromote nudges the scheduler loop to re-evaluate now that the ongoing
// thread has freed its slot. The send is non-blocking because promote is
// buffered and coalescing: a nudge already queued is enough.
func (tr *topicRunner) signalPromote() {
	select {
	case tr.promote <- struct{}{}:
	default:
	}
}

// signalFail answers every still-buffered input with err, so callers blocked
// in Input do not hang once the runner is shutting down or the subprocess has
// exited.
func (tr *topicRunner) signalFail(err error) {
	for _, req := range tr.scheduler.drain() {
		req.err <- err
	}
}

// startSchedulerLoop owns the scheduler state. It records incoming inputs,
// promotes thread groups as the ongoing one frees the slot, and hands the next
// writable input to startWriteLoop over tr.next. It never writes to stdin
// itself. Sending to tr.next is one case of a select alongside tr.inbox, so a
// startWriteLoop stuck on a stalled Write cannot stop this loop from continuing
// to buffer arriving inputs. It exits when runnerCtx is cancelled, failing the
// input it was about to hand over along with everything still buffered.
func (tr *topicRunner) startSchedulerLoop() {
	head := (*inputPromise)(nil)
	for {
		if head == nil {
			head = tr.scheduler.popWritable()
		}
		next := (chan *inputPromise)(nil)
		if head != nil {
			next = tr.next
		}
		select {
		case next <- head:
			head = nil
		case req := <-tr.inbox:
			err := tr.scheduler.submit(req)
			if err != nil {
				req.err <- seederr.WrapErrorf("topic %q: %w", tr.topic, err)
			}
		case <-tr.promote:
			// The ongoing thread freed its slot; loop to re-evaluate popWritable.
		case <-tr.runnerCtx.Done():
			err := seederr.WrapErrorf("topic %q: runner closed", tr.topic)
			if head != nil {
				head.err <- err
			}
			tr.signalFail(err)
			return
		}
	}
}

// startWriteLoop is the single goroutine that writes to tr.stdin, so Input
// callers never share a live Write call. It writes each input the scheduler
// loop hands to it over tr.next and answers that input's caller. It exits when
// runnerCtx is cancelled; a Write stuck on a stalled subprocess is unblocked by
// Close() closing stdin.
func (tr *topicRunner) startWriteLoop() {
	for {
		select {
		case req := <-tr.next:
			tr.writeInput(req)
		case <-tr.runnerCtx.Done():
			return
		}
	}
}

// writeInput marshals an input and writes it to stdin, answering req.err with
// the outcome. A marshal failure fails only that input: stdin is still healthy,
// so the ongoing thread keeps its slot and its other inputs still flow. A write
// failure means stdin is broken, so it abandons the ongoing thread — failing its
// remaining un-written inputs and nudging the scheduler loop to move on rather
// than stall waiting for a "result" that will never come.
//
// Just before the write it re-checks, via claimWrite, that req's thread is still
// ongoing. popWritable removes req from its group but leaves the write to happen
// outside the scheduler lock; a "result" for that thread can arrive in that window
// and free the slot. Without the re-check req would be written into whatever
// thread next holds the slot, mixing one thread's input into another's turn. When
// claimWrite re-buffers req, this returns without answering req.err — its caller
// stays blocked until req is promoted and written for real — and nudges the
// scheduler loop to re-evaluate. The marshal happens first so the re-check sits
// as close to the Write as possible, keeping the residual window minimal (it
// cannot be closed entirely without holding the lock across the Write, which
// would let a stalled subprocess wedge scheduling and stdout tagging).
func (tr *topicRunner) writeInput(req *inputPromise) {
	payload := claudepayload.StreamInputUser{
		StreamInputEnvelope: claudepayload.StreamInputEnvelope{Type: "user"},
		Message: &claudepayload.StreamInputMessage{
			Role:    "user",
			Content: req.input.GetText(),
		},
	}
	line, err := json.Marshal(payload)
	if err != nil {
		req.err <- seederr.Wrap(err)
		return
	}
	line = append(line, '\n')

	proceed, err := tr.scheduler.claimWrite(req)
	if err != nil {
		req.err <- seederr.Wrap(err)
		return
	}
	if !proceed {
		tr.signalPromote()
		return
	}

	_, err = tr.stdin.Write(line)
	if err != nil {
		for _, sibling := range tr.scheduler.abandonOngoing() {
			sibling.err <- seederr.Wrap(err)
		}
		tr.signalPromote()
	}
	req.err <- err
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
		// Wake any Hibernate(wait=true) caller blocked in the scheduler and
		// fail any buffered inputs so their callers don't deadlock when the
		// subprocess exits without emitting the expected "result" lines.
		tr.scheduler.onCmdExit()
		tr.signalFail(seederr.WrapErrorf("topic %q: subprocess exited", tr.topic))
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
		Uuid:       uuid.NewString(),
		Timestamp:  timestamppb.Now(),
		Type:       stepType,
		Topic:      tr.topic,
		ThreadUuid: tr.scheduler.ongoing(),
		Data:       data,
	}

	if step.Type == "result" {
		err := tr.scheduler.completeOngoing()
		if err != nil {
			seedlog.Warnf("topic %q: %v", tr.topic, err)
		} else {
			tr.signalPromote()
		}
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
	req := inputPromise{input: input, err: make(chan error, 1)}
	select {
	case tr.inbox <- req:
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
	case <-tr.runnerCtx.Done():
		return seederr.WrapErrorf("topic %q: runner closed", tr.topic)
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
		tr.scheduler.waitIdle()
	}
	err := tr.stdin.Close()
	if err != nil {
		seedlog.Warnf("topic %q: error closing stdin: %v", tr.topic, err)
	}
	tr.runnerCancel()
	<-tr.done
	return seederr.Wrap(err)
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
		topic:    topic,
		topicDir: topicDir,

		inbox:     make(chan inputPromise, 32),
		promote:   make(chan struct{}, 1),
		next:      make(chan *inputPromise),
		scheduler: newInputScheduler(),

		runnerCtx:    ctx,
		runnerCancel: cancel,
		stdin:        proc.stdin,
		wait:         proc.wait,

		done: make(chan struct{}),
	}

	go tr.startSchedulerLoop()
	go tr.startWriteLoop()
	go tr.readStdout(proc.stdout)
	go tr.readStderr(proc.stderr)
	go tr.waitCmd()

	seedlog.Infof("topic %q: claude started (pid=%d, dir=%s)",
		topic, proc.pid, topicDir)

	return tr, nil
}
