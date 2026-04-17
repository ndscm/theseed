package eventstream

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

const (
	maxDispatchWorkers = 16
	maxDispatchQueue   = 256
)

func Subscribe(ctx context.Context, url string, lastEventId string, handler func(*Event)) error {
	backoff := time.Second
	maxBackoff := 30 * time.Second
	retryMs := 0
	lastId := []byte(lastEventId)

	evQueue := make(chan *Event, maxDispatchQueue)
	evWaitGroup := sync.WaitGroup{}
	for range maxDispatchWorkers {
		evWaitGroup.Add(1)
		go func() {
			defer evWaitGroup.Done()
			for ev := range evQueue {
				func() {
					defer func() {
						r := recover()
						if r != nil {
							seedlog.Errorf("dispatch handler panic: %v", r)
						}
					}()
					handler(ev)
				}()
			}
		}()
	}
	defer func() {
		close(evQueue)
		evWaitGroup.Wait()
	}()

	for ctx.Err() == nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return seederr.Wrap(err)
		}
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		if len(lastId) > 0 {
			req.Header.Set("Last-Event-ID", string(lastId))
		}

		wait := backoff
		if retryMs > 0 {
			wait = time.Duration(retryMs) * time.Millisecond
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			seedlog.Warnf("connect failed: %v; retrying in %v", err, wait)
			select {
			case <-ctx.Done():
			case <-time.After(wait):
			}
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			seedlog.Warnf("unexpected status: %v; retrying in %v", resp.Status, wait)
			select {
			case <-ctx.Done():
			case <-time.After(wait):
			}
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		backoff = time.Second

		eventType := []byte{}
		eventId := []byte{}
		data := []byte{}
		reader := bufio.NewReader(resp.Body)
		readErr := error(nil)
		for {
			line := []byte{}
			line, readErr = reader.ReadBytes('\n')
			if readErr != nil && len(line) == 0 {
				break
			}
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			if len(line) == 0 {
				if len(data) > 0 {
					ev := &Event{
						id:    append([]byte(nil), eventId...),
						event: append([]byte(nil), eventType...),
						data:  append([]byte(nil), data...),
					}
					select {
					case evQueue <- ev:
					case <-ctx.Done():
					}
				}
				eventType = eventType[:0]
				eventId = eventId[:0]
				data = data[:0]
			} else if line[0] != ':' {
				name, value, _ := bytes.Cut(line, []byte{':'})
				if len(value) > 0 && value[0] == ' ' {
					value = value[1:]
				}
				switch string(name) {
				case "event":
					eventType = append(eventType[:0], value...)
				case "data":
					if len(data) > 0 {
						data = append(data, '\n')
					}
					data = append(data, value...)
				case "id":
					eventId = append(eventId[:0], value...)
					lastId = append(lastId[:0], value...)
				case "retry":
					ms, convErr := strconv.Atoi(string(value))
					if convErr == nil && ms > 0 {
						retryMs = ms
					}
				}
			}
			if readErr != nil {
				break
			}
		}
		if readErr == io.EOF {
			readErr = nil
		}
		resp.Body.Close()

		if ctx.Err() != nil {
			break
		}

		wait = backoff
		if retryMs > 0 {
			wait = time.Duration(retryMs) * time.Millisecond
		}
		if readErr != nil {
			seedlog.Warnf("stream error: %v; reconnecting in %v", readErr, wait)
		} else {
			seedlog.Infof("stream closed; reconnecting in %v", wait)
		}
		select {
		case <-ctx.Done():
		case <-time.After(wait):
		}
	}

	return nil
}
