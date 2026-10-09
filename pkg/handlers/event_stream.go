package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	maxEventLineBytes = 1 << 20
	maxEventBytes     = 1 << 20
)

// Event is a decoded Server-Sent Event.
type Event[T any] struct {
	ID   string
	Name string
	Data T
}

// DecodeEventStream reads Server-Sent Events whose data fields contain JSON values.
func DecodeEventStream[T any](ctx context.Context, source io.Reader, handle func(Event[T]) error) error {
	if ctx == nil {
		return errors.New("frontal: event stream context is required")
	}
	if source == nil {
		return errors.New("frontal: event stream reader is required")
	}
	if handle == nil {
		return errors.New("frontal: event handler is required")
	}
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 4096), maxEventLineBytes)
	var event Event[T]
	var data []string
	var lastID string
	eventBytes := 0
	dispatch := func() error {
		if len(data) == 0 {
			event = Event[T]{ID: lastID}
			eventBytes = 0
			return nil
		}
		var value T
		if err := json.Unmarshal([]byte(strings.Join(data, "\n")), &value); err != nil {
			return fmt.Errorf("frontal: decode event data: %w", err)
		}
		event.Data = value
		if err := handle(event); err != nil {
			return fmt.Errorf("frontal: handle event: %w", err)
		}
		event = Event[T]{ID: lastID}
		data = data[:0]
		eventBytes = 0
		return nil
	}

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := scanner.Text()
		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
			continue
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		field, value, found := strings.Cut(line, ":")
		if !found {
			value = ""
		}
		value = strings.TrimPrefix(value, " ")
		switch field {
		case "id":
			if !strings.ContainsRune(value, '\x00') {
				event.ID = value
				lastID = value
			}
		case "event":
			event.Name = value
		case "data":
			lineBytes := len(value)
			if len(data) > 0 {
				lineBytes++
			}
			if lineBytes > maxEventBytes-eventBytes {
				return fmt.Errorf("frontal: event data exceeds %d bytes", maxEventBytes)
			}
			eventBytes += lineBytes
			data = append(data, value)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("frontal: read event stream: %w", err)
	}
	return dispatch()
}
