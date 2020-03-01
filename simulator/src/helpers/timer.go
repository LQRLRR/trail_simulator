package helpers

import (
	"fmt"
	"time"
)

type timer struct {
	start    time.Time
	location string
}

// Timer prints time elapsed since Start().
// location identifies the location where timer is recordiing.
type Timer interface {
	Start(location string)
	RecordLap()
}

// CreateTimer is initialization of timer
func CreateTimer() Timer {
	return &timer{}
}

func (t *timer) Start(location string) {
	t.start = time.Now()
	t.location = location
}

func (t *timer) RecordLap() {
	now := time.Now()
	fmt.Println("timer."+t.location, ": ", now.Sub(t.start).Milliseconds(), "ms")
}
