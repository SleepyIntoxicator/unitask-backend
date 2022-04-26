package models

import (
	"fmt"
	"github.com/google/uuid"
)

type Event interface {
	GetTimestamp() Timestamp
	SetTimestamp(ts Timestamp)
	//TODO: body
}

type EventType byte

const (
	TypeEvent = EventType(iota)
	TypeTodo
)

type EventInstance struct {
	Id         uuid.UUID `json:"id"`   //TODO: Change type to UUID
	Type       EventType `json:"type"` //Prevent, ...
	Timestamp  Timestamp `json:"timestamp"`
	Subscribed bool      `json:"subscribed"`
}

func NewEvent(eventType EventType, ts Timestamp, isSubscribed bool) (Event, error) {
	switch eventType {
	case TypeEvent:
		newID, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}

		return &EventInstance{
			Id:         newID,
			Type:       TypeEvent,
			Timestamp:  ts,
			Subscribed: isSubscribed,
		}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %d", eventType)
	}
}

func NewEventWithID(id uuid.UUID, eventType EventType, ts Timestamp, isSubscribed bool) (Event, error) {
	switch eventType {
	case TypeEvent:
		return &EventInstance{
			Id:         id,
			Type:       TypeEvent,
			Timestamp:  ts,
			Subscribed: isSubscribed,
		}, nil
	default:
		return nil, fmt.Errorf("unknown event type: %d", eventType)
	}
}

func (e EventInstance) GetTimestamp() Timestamp {
	return e.Timestamp
}

func (e EventInstance) SetTimestamp(ts Timestamp) {
	e.Timestamp = ts
}
