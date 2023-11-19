package recorder

import (
	"fmt"
	"io"
	"os/exec"
	"time"
)

type RecordedCmd struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	cmd           *exec.Cmd
	stdoutMonitor WriterMonitor
	stderrMonitor WriterMonitor
	startTime     int64
	timeline      Timeline
}

func Command(name string, arg ...string) *RecordedCmd {
	cr := &RecordedCmd{
		cmd:           exec.Command(name, arg...),
		stdoutMonitor: WriterMonitor{},
		stderrMonitor: WriterMonitor{},
		timeline:      Timeline{},
	}

	cr.stdoutMonitor.OnWrite = func(in []byte) { cr.onEvent(Stdout, in) }
	cr.stderrMonitor.OnWrite = func(in []byte) { cr.onEvent(Stderr, in) }

	cr.cmd.Stdout = &cr.stdoutMonitor
	cr.cmd.Stderr = &cr.stderrMonitor

	return cr
}

func (rc *RecordedCmd) Run() (*Recording, error) {
	rc.startTime = time.Now().UnixNano()
	rc.timeline = Timeline{}

	rc.stdoutMonitor.Writer = rc.Stdout
	rc.stderrMonitor.Writer = rc.Stderr
	rc.cmd.Stdin = rc.Stdin

	err := rc.cmd.Run()
	if err != nil {
		_, isExitError := err.(*exec.ExitError)

		if !isExitError {
			return nil, fmt.Errorf("failed to run the command: %v", err)
		}
	}

	rc.recordEvent(ProcessExitEvent{
		ExitCode: rc.cmd.ProcessState.ExitCode(),
	})

	return rc.Recording(), nil
}

func (rc *RecordedCmd) Recording() *Recording {
	path := rc.cmd.Path
	args := []string{}

	if len(rc.cmd.Args) != 0 {
		args = rc.cmd.Args[1:]
	}

	return &Recording{
		Path:     path,
		Args:     args,
		Timeline: rc.timeline,
	}
}

func (rc *RecordedCmd) onEvent(stream Stream, in []byte) {
	if rc.cmd.Process == nil {
		// not started yet, so don't record anything
		return
	}

	data := make([]byte, len(in))
	copy(data, in)

	rc.recordEvent(StreamWriteEvent{
		Stream: stream,
		Data:   data,
	})
}

func (rc *RecordedCmd) recordEvent(event TimelineEvent) {
	currentTime := time.Now().UnixNano()

	rc.timeline = append(rc.timeline, TimelineEntry{
		Type:       event.Type(),
		TimeOffset: time.Duration(currentTime - rc.startTime),
		Event:      event,
	})
}
