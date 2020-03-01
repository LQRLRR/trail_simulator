package helpers

import (
	"sync"
	"time"
)

type timeBomb struct {
	running bool
	wg      sync.WaitGroup
}

// TimeBomb terminates program, if can't call Clear() in set time.
type TimeBomb interface {
	Start(seconds int, message string)
	Clear()
}

// CreateTimeBomb is initialization of timeBomb
func CreateTimeBomb() TimeBomb {
	return &timeBomb{false, sync.WaitGroup{}}
}

func (tb *timeBomb) Start(seconds int, message string) {
	if tb.running {
		tb.Clear() // restart
	}
	tb.running = true
	tb.wg.Add(1)
	go func() {
		defer func() { tb.wg.Done() }()

		t1 := time.Now()
		for t2 := time.Now(); t2.Sub(t1).Seconds() < float64(seconds); {
			t2 = time.Now()
			if !tb.running {
				return
			}
		}
		panic(message)
	}()
}

func (tb *timeBomb) Clear() {
	tb.running = false
	tb.wg.Wait()
}
