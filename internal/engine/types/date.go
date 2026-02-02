package types

import (
	"encoding/json"
	"time"
)

//nolint:recvcheck // see MarshalJSON comment
type Date struct {
	time time.Time
}

func NewDate(t time.Time) Date {
	return Date{t}
}

// MarshalJSON turn Date into JSON
// Value instead of pointer receiver because only that way it can be used for both.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.time.IsZero() {
		return json.Marshal(nil)
	}

	return json.Marshal(d.time.Format(time.DateOnly))
}

// UnmarshalJSON turn JSON into Date.
func (d *Date) UnmarshalJSON(text []byte) (err error) {
	var value string
	err = json.Unmarshal(text, &value)
	if err != nil {
		return err
	}
	if value == "" {
		return nil
	}
	d.time, err = time.Parse(time.DateOnly, value)

	return err
}

func (d Date) String() string {
	if d.time.IsZero() {
		return ""
	}

	return d.time.Format(time.DateOnly)
}
