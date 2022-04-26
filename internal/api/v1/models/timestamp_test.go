package models

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestTimeStamp_MarshalToJSON(t *testing.T) {
	testCases := []struct {
		name    string
		ts      func() Timestamp
		isValid bool
	}{
		{
			name: "TimeStamp valid",
			ts: func() Timestamp {
				return TimeStamp{
					Type: "time",
					Time: Time{20, 20, 20},
				}
			},
			isValid: true,
		},
		{
			name: "TimeStamp invalid",
			ts: func() Timestamp {
				return TimeStamp{
					Type: "time",
					Time: Time{77, 77, 77},
				}
			},
			isValid: false,
		},
		{
			name: "DateStamp valid",
			ts: func() Timestamp {
				return DateStamp{
					Type: "date",
					Date: Date{12, 3, 4567},
				}
			},
			isValid: true,
		},
		{
			name: "DateStamp invalid",
			ts: func() Timestamp {
				return DateStamp{
					Type: "date",
					Date: Date{77, 77, -2222},
				}
			},
			isValid: false,
		},
		{
			name: "DateTimeStamp valid",
			ts: func() Timestamp {
				return DateTimeStamp{
					Type: "datetime",
					Date: Date{12, 3, 4567},
					Time: Time{20, 20, 20},
				}
			},
			isValid: true,
		},
		{
			name: "DateTimeStamp invalid",
			ts: func() Timestamp {
				return DateTimeStamp{
					Type: "datetime",
					Date: Date{77, 77, -2222},
					Time: Time{77, 77, 77},
				}
			},
			isValid: false,
		},
		{
			name: "EventTimeStamp valid",
			ts: func() Timestamp {
				return EventTimeStamp{
					Type: "event",
					Event: EventInstance{
						Id:   uuid.Nil,
						Type: TypeEvent,
						Timestamp: DateTimeStamp{
							Type: "datetime",
							Date: Date{12, 3, 4567},
							Time: Time{20, 20, 20},
						},
					},
				}
			},
			isValid: true,
		},
		{
			name: "EventTimeStamp invalid",
			ts: func() Timestamp {
				return EventTimeStamp{
					Type: "event",
					Event: EventInstance{
						Id:   uuid.Nil,
						Type: TypeTodo,
						Timestamp: DateTimeStamp{
							Type: "datetime",
							Date: Date{77, 77, -2222},
							Time: Time{77, 77, 77},
						},
					},
				}
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			result, err := MarshalToJSON(tc.ts())
			strings.Compare(result, result)

			fmt.Println(result)

			if tc.isValid {
				assert.NoError(t, err)
			} else {
				if !assert.Error(t, err) {
					//t.Errorf
				}

			}
		})
	}
}

func TestUnmarshalFromJSON(t *testing.T) {
	testCases := []struct {
		name    string
		ts      string
		isValid bool
	}{
		{
			name:    "TimeStamp valid",
			ts:      "{\"type\":\"time\",\"time\":{\"second\":20,\"minute\":20,\"hour\":20}}",
			isValid: true,
		},
		{
			name:    "TimeStamp invalid",
			ts:      "{\"type\":\"time\",\"time\":{\"second\":77,\"minute\":77,\"hour\":77}}",
			isValid: false,
		},
		{
			name:    "DateStamp valid",
			ts:      "{\"type\":\"date\",\"date\":{\"day\":12,\"month\":3,\"year\":4567}}",
			isValid: true,
		},
		{
			name:    "DateStamp invalid",
			ts:      "{\"type\":\"date\",\"date\":{\"day\":77,\"month\":77,\"year\":-2222}}",
			isValid: false,
		},
		{
			name:    "DateTimeStamp valid",
			ts:      `{"type":"datetime","date":{"day":12,"month":3,"year":4567},"time":{"second":20,"minute":20,"hour":20}}`,
			isValid: true,
		},
		{
			name:    "DateTimeStamp invalid",
			ts:      `{"type":"datetime","date":{"day":77,"month":77,"year":-2222},"time":{"second":77,"minute":77,"hour":77}}`,
			isValid: false,
		},
		{
			name: "EventTimeStamp valid",
			ts: `
{
  "type": "event",
  "event": {
    "id": 12312,
    "type": "todo",
    "timestamp": {
      "type": "datetime",
      "date": {
        "day": 12,
        "month": 3,
        "year": 4567
      },
      "time": {
        "second": 20,
        "minute": 20,
        "hour": 20
      }
    },
    "subscribed": false
  }
}
`,
			isValid: true,
		},
		{
			name: "EventTimeStamp invalid",
			ts: `
{
  "type": "event",
  "event": {
    "id": 12312,
    "type": "todo",
    "timest_ts = {models.DateTimeStamp} amp": {
      "type": "datetime",
      "date": {
        "day": 77,
        "month": 77,
        "year": -2222
      },
      "time": {
        "second": 77,
        "minute": 77,
        "hour": 77
      }
    },
    "subscribed": false
  }
}
				`,
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			var ts Timestamp
			err = UnmarshalFromJSON(tc.ts, &ts)
			//strings.Compare(result, result)

			//fmt.Println(result)

			if tc.isValid {
				assert.NoError(t, err)
			} else {
				if assert.Error(t, err) {
				}

			}
		})
	}
}
