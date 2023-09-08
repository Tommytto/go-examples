package common

import "time"

//go:generate mockgen -source=clock.go -destination=../mocks/mock_clock.go -package=mocks
type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	After(time.Duration) <-chan time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

func (RealClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
