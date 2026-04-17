package eventstream

// See: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
type Event struct {
	id    []byte
	data  []byte
	event []byte
}

func (ev *Event) Id() string {
	return string(ev.id)
}

func (ev *Event) Event() string {
	return string(ev.event)
}

func (ev *Event) Data() []byte {
	return ev.data
}
