package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
)

type Date struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type Time struct {
	Second int `json:"second"`
	Minute int `json:"minute"`
	Hour   int `json:"hour"`
}

func CreateDate() *Date {
	return &Date{}
}

func CreateTime() *Time {
	return &Time{}
}

func (dt *Date) SetDate(d int, m int, y int) error {
	if err := ValidateDate(d, m, y); err != nil {
		return err
	}

	dt.Day, dt.Month, dt.Year = d, m, y
	return nil
}

func ValidateDate(d int, m int, y int) error {
	var errString string
	if d < 0 || d > 31 {
		errString = fmt.Sprintf("incorrect 'day'=[%d] value. ", d)
	}

	if m < 1 || m > 12 {
		errString += fmt.Sprintf("incorrect 'month'=[%d] value. ", m)
	}

	if y < 0 {
		errString += fmt.Sprintf("incorrect 'year'=[%d] value", y)
	}

	if errString != "" {
		return errors.New(errString)
	}
	return nil
}

func (t *Time) SetTime(s int, m int, h int) error {
	if err := ValidateTime(s, m, h); err != nil {
		return err
	}

	t.Second, t.Minute, t.Hour = s, m, h
	return nil
}

func ValidateTime(s int, m int, h int) error {
	var errString string
	if s < 0 || s > 59 {
		errString = fmt.Sprintf("incorrect 'second'=[%d] value. ", s)
	}

	if m < 0 || m > 59 {
		errString += fmt.Sprintf("incorrect 'minute'=[%d] value. ", m)
	}

	if h < 0 || h > 23 {
		errString += fmt.Sprintf("incorrect 'hour'=[%d] value", h)
	}
	if errString != "" {
		return errors.New(errString)
	}
	return nil
}

//Timestamp TODO: Change type of the "type" field of the TimeStamp structures to UUID
type Timestamp interface {
	GetType() string
}

func Validate(ts Timestamp) error {
	switch ts.GetType() {
	case "time":
		_ts, ok := ts.(TimeStamp)
		if !ok {
			return errors.New("invalid timestamp type")
		}
		return ValidateTime(_ts.Time.Second, _ts.Time.Minute, _ts.Time.Hour)

	case "date":
		_ts, ok := ts.(DateStamp)
		if !ok {
			return errors.New("invalid timestamp type")
		}
		return ValidateDate(_ts.Date.Day, _ts.Date.Month, _ts.Date.Year)

	case "datetime":
		_ts, ok := ts.(*DateTimeStamp)
		if !ok {
			return errors.New("invalid timestamp type")
		}
		VDateErr := ValidateDate(_ts.Date.Day, _ts.Date.Month, _ts.Date.Year)
		VTimeErr := ValidateTime(_ts.Time.Second, _ts.Time.Minute, _ts.Time.Hour)
		var ErrText string
		if VDateErr != nil {
			ErrText = VDateErr.Error()
		}
		if VTimeErr != nil {
			ErrText += VTimeErr.Error()
		}
		if ErrText != "" {
			return errors.New(ErrText)
		}
		return nil

	case "event":

		_ts, ok := ts.(EventTimeStamp)
		if !ok {
			return errors.New("invalid timestamp (event). The type doesn't match the content")
		}

		if _ts.Event.GetTimestamp() == nil {
			return errors.New("")
		}
		if _ts.Event.GetTimestamp().GetType() == "event" {
			return errors.New("invalid. recursion timestamp")
		}
		return Validate(_ts.Event.GetTimestamp()) //TODO: warning! May cause infinite recursion
	default:
		return fmt.Errorf("unexpected timestamp type: %s", ts.GetType())
	}
}

func MarshalToJSON(ts Timestamp) (string, error) {
	if NotValidError := Validate(ts); NotValidError != nil {
		return "", NotValidError
	}

	body, err := json.Marshal(ts)
	if err != nil {
		return "", nil
	}

	return string(body), nil
}

func UnmarshalFromJSON(from string, to *Timestamp) error {
	var typeStruct struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(from), &typeStruct); err != nil {
		return err
	}

	switch typeStruct.Type {
	case "time":
		var ts TimeStamp
		err := json.Unmarshal([]byte(from), &ts)
		if err != nil {
			return err
		}
		*to = ts

	case "date":
		var ts DateStamp
		err := json.Unmarshal([]byte(from), &ts)
		if err != nil {
			return err
		}
		*to = ts

	case "datetime":
		var ts DateTimeStamp
		err := json.Unmarshal([]byte(from), &ts)
		if err != nil {
			return err
		}
		*to = ts

	case "event":
		// type field of event timestamp
		var tsTypeEx struct {
			Ev struct {
				Ts struct {
					Type string `json:"type"`
				} `json:"timestamp"`
			} `json:"event"`
		}

		// Parse the type of event timestamp
		err := json.Unmarshal([]byte(from), &tsTypeEx)
		if err != nil {
			return err
		}

		ts, err := GetNewEmptyTimestamp(tsTypeEx.Ev.Ts.Type)
		if err != nil {
			return nil
		}

		event, err := NewEventWithID(uuid.Nil, TypeEvent, ts, false)
		if err != nil {
			return nil
		}

		if err := json.Unmarshal([]byte(from), event); err != nil {
			return err
		}
		//tsEvent := GetNewEmptyTimestamp("event")

		*to = EventTimeStamp{
			Type:  tsTypeEx.Ev.Ts.Type,
			Event: event,
		}
		/*
			var eventTs struct {
				Type  string
				Event struct {
					Id         int       `json:"id"`   //TODO: Change type to UUID
					Type       string    `json:"type"` //Prevent, ...
					Timestamp  Timestamp `json:"timestamp"`
					Subscribed bool      `json:"subscribed"`
				}
			}

			// Getting instance of event timestamp
			var ev struct {
				EventInstance `json:"event"`
				Type          string `json:"type"`
			}
			ev.EventInstance.Timestamp, err = GetNewEmptyTimestamp(tsTypeEx.Ev.Ts.Type)

			var tstedEvent EventTimeStamp

			typeOfTs, err := GetNewEmptyTimestamp(tsTypeEx.Ev.Ts.Type)
			tstedEvent.Event = EventInstance{
				Timestamp: typeOfTs,
			}
			if err := json.Unmarshal([]byte(from), &tstedEvent); err != nil {
				return err
			} else {
				*to = tstedEvent

			}*/

		//if err := json.Unmarshal([]byte(from), &ts); err == nil {
		//	*to = ts
		//}
		/*		if err := json.Unmarshal([]byte(from), &ev); err != nil {
					return err
				} else {
					ts = EventTimeStamp{
						Type:  ev.Type,
						Event: ev.EventInstance,
					}
					*to = ts

				}*/

	default:
		return fmt.Errorf("invalid timestamp type: %s", typeStruct.Type)
	}

	err := Validate(*to)
	return err
}

type TimeStamp struct {
	Type string `json:"type"`
	Time Time   `json:"time"`
}

type DateStamp struct {
	Type string `json:"type"`
	Date Date   `json:"date"`
}

type DateTimeStamp struct {
	Type string `json:"type"`
	Date Date   `json:"date"`
	Time Time   `json:"time"`
}

type EventTimeStamp struct {
	Type  string `json:"type"`
	Event Event  `json:"event"`
}

func GetNewEmptyTimestamp(tsType string) (Timestamp, error) {
	switch tsType {
	case "time":
		return &TimeStamp{Type: "time"}, nil
	case "date":
		return &DateStamp{Type: "date"}, nil
	case "datetime":
		return &DateTimeStamp{Type: "datetime"}, nil
	case "event":
		return &EventTimeStamp{Type: "event"}, nil
	}

	return nil, fmt.Errorf("unsupported type of timestamp")
}

func (s TimeStamp) GetType() string      { return s.Type }
func (s DateStamp) GetType() string      { return s.Type }
func (s DateTimeStamp) GetType() string  { return s.Type }
func (s EventTimeStamp) GetType() string { return s.Type }
