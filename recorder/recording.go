package recorder

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
)

type TimelineEvent interface {
	Type() string

	// TODO: will probably need some context
	Execute() error
}

type StreamWriteEvent struct {
	Stream Stream
	Data   []byte
}

func (StreamWriteEvent) Type() string { return "stream-write" }
func (e StreamWriteEvent) Execute() (err error) {
	switch e.Stream {
	case Stdout:
		_, err = os.Stdout.Write(e.Data)
	case Stderr:
		_, err = os.Stderr.Write(e.Data)
	}

	return
}

type StreamCloseEvent struct{}

func (StreamCloseEvent) Type() string   { return "stream-close" }
func (StreamCloseEvent) Execute() error { return nil }

type ProcessExitEvent struct {
	ExitCode int
}

func (ProcessExitEvent) Type() string     { return "process-exit" }
func (e ProcessExitEvent) Execute() error { os.Exit(e.ExitCode); return nil }

type Stream int

const (
	Stdin Stream = iota
	Stdout
	Stderr
)

type TimelineEntry struct {
	Type       string
	TimeOffset time.Duration
	Event      TimelineEvent
}

type Timeline []TimelineEntry

type Recording struct {
	Path     string
	Args     []string
	Timeline Timeline
}

type rawTimelineEntry struct {
	Type       string
	TimeOffset time.Duration
	Event      json.RawMessage
}

func (e *TimelineEntry) UnmarshalJSON(b []byte) error {
	var rawEvent rawTimelineEntry
	err := json.Unmarshal(b, &rawEvent)
	if err != nil {
		return err
	}

	e.Type = rawEvent.Type
	e.TimeOffset = rawEvent.TimeOffset

	matchedEventType, ok := validEventTypes[e.Type]

	if !ok {
		return fmt.Errorf("unexpected event type '%s'", e.Type)
	}

	t := reflect.TypeOf(matchedEventType)
	event := reflect.New(t).Interface().(TimelineEvent)

	err = json.Unmarshal(rawEvent.Event, &event)
	if err != nil {
		return err
	}

	e.Event = event

	return nil
}

var validEventTypes map[string]TimelineEvent

func init() {
	eventTypesList := []TimelineEvent{
		StreamWriteEvent{},
		StreamCloseEvent{},
		ProcessExitEvent{},
	}

	validEventTypes = make(map[string]TimelineEvent, len(eventTypesList))
	for _, e := range eventTypesList {
		validEventTypes[e.Type()] = e
	}
}
