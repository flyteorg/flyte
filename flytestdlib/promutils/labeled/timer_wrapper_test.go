package labeled

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeTimer struct {
	stopCount int
}

func (f *fakeTimer) Stop() float64 {
	f.stopCount++
	return 0
}

func TestTimerStop(t *testing.T) {
	ft := &fakeTimer{}
	tim := timer{
		Timers: []Timer{
			ft, ft, ft,
		},
	}

	tim.Stop()
	assert.Equal(t, 3, ft.stopCount)
}
