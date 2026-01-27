package labeled

// Defines a common interface for timers.
type Timer interface {
	// Stops the timer and reports observation.
	Stop() float64
}

type timer struct {
	Timers []Timer
}

func (t timer) Stop() float64 {
	var res float64
	for _, timer := range t.Timers {
		res = timer.Stop()
	}

	return res
}
