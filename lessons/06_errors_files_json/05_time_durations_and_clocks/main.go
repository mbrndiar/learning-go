// This lesson distinguishes elapsed durations from wall-clock timestamps and
// shows how explicit clock dependencies keep time-based code deterministic.
package main

import (
	"fmt"
	"time"
)

// Clock is the smallest capability needed by code that asks for the current
// instant. Production code can use time.Now; tests can provide a fixed clock.
type Clock interface {
	Now() time.Time
}

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func expiresAt(clock Clock, lifetime time.Duration) time.Time {
	return clock.Now().Add(lifetime)
}

func main() {
	fmt.Println("--- durations ---")
	lifetime, err := time.ParseDuration("1h30m")
	if err != nil {
		fmt.Println("parse duration:", err)
		return
	}
	fmt.Printf("lifetime=%s minutes=%.0f\n", lifetime, lifetime.Minutes())

	fmt.Println("--- instants, locations, and UTC ---")
	raw := "2026-07-18T17:45:00+02:00"
	instant, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		fmt.Println("parse timestamp:", err)
		return
	}
	fmt.Println("as supplied:", instant.Format(time.RFC3339))
	fmt.Println("same instant in UTC:", instant.UTC().Format(time.RFC3339))

	otherLocation := instant.In(time.FixedZone("course-zone", 2*60*60))
	fmt.Println("Equal compares instants:", instant.Equal(otherLocation))
	fmt.Println("== also compares location metadata:", instant == otherLocation)

	fmt.Println("--- arithmetic and custom layouts ---")
	start := time.Date(2026, time.July, 18, 15, 0, 0, 0, time.UTC)
	deadline := start.Add(lifetime)
	fmt.Println("deadline:", deadline.Format(time.RFC3339))
	fmt.Println("elapsed:", deadline.Sub(start))
	// Custom layouts format Go's reference instant, 2006-01-02 15:04:05
	// -0700 MST, the way the desired output should look.
	fmt.Println("display:", deadline.Format("2006-01-02 15:04 MST"))

	fmt.Println("--- testable clocks ---")
	clock := fixedClock{now: start}
	fmt.Println("fixed expiry:", expiresAt(clock, 30*time.Second).Format(time.RFC3339))

	// Values returned by time.Now may carry a monotonic-clock reading. Sub and
	// time.Since use it for reliable elapsed time when both values retain it.
	// Formatting, JSON, and database storage preserve wall time, not that
	// process-local monotonic reading.
}
