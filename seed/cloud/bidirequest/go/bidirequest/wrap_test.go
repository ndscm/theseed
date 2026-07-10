package bidirequest

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
)

// pipeStream is one end of an in-memory PayloadStream pair, standing in for the
// WebSocket the two peers really share.
type pipeStream struct {
	in  <-chan *bidirequestpb.Payload
	out chan<- *bidirequestpb.Payload

	closeOnce sync.Once
	done      chan struct{}
}

func (s *pipeStream) Send(payload *bidirequestpb.Payload) error {
	select {
	case s.out <- payload:
		return nil
	case <-s.done:
		return io.ErrClosedPipe
	}
}

func (s *pipeStream) Receive() (*bidirequestpb.Payload, error) {
	select {
	case payload := <-s.in:
		return payload, nil
	case <-s.done:
		return nil, io.EOF
	}
}

func (s *pipeStream) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
	})
	return nil
}

func newPipeStreamPair() (*pipeStream, *pipeStream) {
	toServer := make(chan *bidirequestpb.Payload, 256)
	toClient := make(chan *bidirequestpb.Payload, 256)
	done := make(chan struct{})
	server := &pipeStream{in: toServer, out: toClient, done: done}
	client := &pipeStream{in: toClient, out: toServer, done: done}
	return server, client
}

// connectPeers wires a client-side and a server-side transport together and
// returns the http.Client each uses to reach the other's handler.
func connectPeers(
	t *testing.T, clientHandler http.Handler, serverHandler http.Handler,
) (towardServer *http.Client, towardClient *http.Client) {
	t.Helper()
	serverStream, clientStream := newPipeStreamPair()
	t.Cleanup(func() {
		serverStream.Close()
	})
	// The client side reaches the server side's handler, and vice versa.
	towardServer = WrapClientSide(clientStream, clientHandler)
	towardClient = WrapServerSide(serverStream, serverHandler)
	return towardServer, towardClient
}

func TestUnaryRoundTrip(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("handler read body: %v", err)
			return
		}
		w.Header().Set("X-Echo-Path", r.URL.Path)
		w.Header().Set("X-Echo-Header", r.Header.Get("X-Send-Header"))
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("got:" + string(body)))
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("body, headers, and status survive the tunnel", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "http://peer/pkg.Service/Method",
			strings.NewReader("hello"))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		req.Header.Set("X-Send-Header", "sent")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusTeapot {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusTeapot)
		}
		if got := resp.Header.Get("X-Echo-Path"); got != "/pkg.Service/Method" {
			t.Errorf("echoed path = %q", got)
		}
		if got := resp.Header.Get("X-Echo-Header"); got != "sent" {
			t.Errorf("echoed header = %q", got)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if string(body) != "got:hello" {
			t.Errorf("body = %q, want %q", body, "got:hello")
		}
	})

	t.Run("a request with no body is served", func(t *testing.T) {
		resp, err := client.Get("http://peer/pkg.Service/Method")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if string(body) != "got:" {
			t.Errorf("body = %q, want %q", body, "got:")
		}
	})
}

func TestRequestClaimsHttp2(t *testing.T) {
	protoMajor := make(chan int, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		protoMajor <- r.ProtoMajor
		w.WriteHeader(http.StatusOK)
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("both ends see HTTP/2 so connect permits bidi", func(t *testing.T) {
		resp, err := client.Get("http://peer/x")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer resp.Body.Close()

		select {
		case got := <-protoMajor:
			if got != 2 {
				t.Errorf("request ProtoMajor = %d, want 2", got)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("handler never ran")
		}
		if resp.ProtoMajor != 2 {
			t.Errorf("response ProtoMajor = %d, want 2", resp.ProtoMajor)
		}
	})
}

func TestServerStreamsIncrementally(t *testing.T) {
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("first\n"))
		w.(http.Flusher).Flush()
		<-release
		w.Write([]byte("second\n"))
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("a flushed write arrives before the handler returns", func(t *testing.T) {
		resp, err := client.Get("http://peer/stream")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)
		// The handler is parked on release, so this can only be read if the
		// response body streams rather than being buffered until it returns.
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read first line: %v", err)
		}
		if line != "first\n" {
			t.Fatalf("first line = %q", line)
		}

		close(release)
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read second line: %v", err)
		}
		if line != "second\n" {
			t.Errorf("second line = %q", line)
		}
	})
}

func TestFullDuplex(t *testing.T) {
	// An echo handler that answers each line as it arrives, while the client is
	// still sending. Nothing here can complete unless both directions of the
	// stream are live at once.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		reader := bufio.NewReader(r.Body)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				w.Write([]byte("echo:" + line))
				w.(http.Flusher).Flush()
			}
			if err != nil {
				return
			}
		}
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("responses come back while the request body is still open", func(t *testing.T) {
		requestReader, requestWriter := io.Pipe()
		req, err := http.NewRequest(http.MethodPost, "http://peer/duplex", requestReader)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		_, err = io.WriteString(requestWriter, "one\n")
		if err != nil {
			t.Fatalf("write first line: %v", err)
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read first echo: %v", err)
		}
		if line != "echo:one\n" {
			t.Fatalf("first echo = %q", line)
		}

		// The second line is only reachable because the request body was never
		// closed: a half-duplex transport would have needed it drained first.
		_, err = io.WriteString(requestWriter, "two\n")
		if err != nil {
			t.Fatalf("write second line: %v", err)
		}
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read second echo: %v", err)
		}
		if line != "echo:two\n" {
			t.Errorf("second echo = %q", line)
		}

		requestWriter.Close()
	})
}

func TestReverseDirection(t *testing.T) {
	clientSideHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from the client side"))
	})
	_, towardClient := connectPeers(t, clientSideHandler, http.NotFoundHandler())

	t.Run("the server side can call the client side's handler", func(t *testing.T) {
		resp, err := towardClient.Get("http://peer/reverse")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("read response body: %v", err)
		}
		if string(body) != "from the client side" {
			t.Errorf("body = %q", body)
		}
	})
}

func TestConcurrentStreams(t *testing.T) {
	// The slow stream is held open while the fast one runs end to end, so a
	// transport that served one stream at a time would deadlock here.
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/slow" {
			<-release
		}
		w.Write([]byte("done:" + r.URL.Path))
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("a parked stream does not block another", func(t *testing.T) {
		slowDone := make(chan error, 1)
		go func() {
			resp, err := client.Get("http://peer/slow")
			if err != nil {
				slowDone <- err
				return
			}
			defer resp.Body.Close()
			_, err = io.ReadAll(resp.Body)
			slowDone <- err
		}()

		resp, err := client.Get("http://peer/fast")
		if err != nil {
			t.Fatalf("fast Get: %v", err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatalf("fast read: %v", err)
		}
		if string(body) != "done:/fast" {
			t.Errorf("fast body = %q", body)
		}

		close(release)
		select {
		case err := <-slowDone:
			if err != nil {
				t.Errorf("slow stream: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("slow stream never finished")
		}
	})
}

func TestRequestCancellation(t *testing.T) {
	handlerDone := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(handlerDone)
		// Never responds; the caller must give up on its own.
		<-r.Context().Done()
	})
	client, _ := connectPeers(t, http.NotFoundHandler(), handler)

	t.Run("a cancelled request returns rather than hanging", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://peer/hang", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}

		done := make(chan error, 1)
		go func() {
			resp, err := client.Do(req)
			if resp != nil {
				resp.Body.Close()
			}
			done <- err
		}()

		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected the cancelled request to fail")
			}
		case <-time.After(5 * time.Second):
			t.Fatal("cancelled request hung")
		}
	})

	t.Run("the reset releases the handler on the other side", func(t *testing.T) {
		select {
		case <-handlerDone:
		case <-time.After(5 * time.Second):
			t.Fatal("handler was never cancelled; it would leak for the life of the connection")
		}
	})
}
