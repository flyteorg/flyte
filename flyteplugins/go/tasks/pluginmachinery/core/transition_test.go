package core

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestTransitionType_String(t *testing.T) {
	assert.Equal(t, TransitionTypeBarrier.String(), "TransitionTypeBarrier")
	assert.Equal(t, TransitionTypeEphemeral.String(), "TransitionTypeEphemeral")
}

func ExampleTransition_String() {
	trns := DoTransitionType(TransitionTypeBarrier, PhaseInfoUndefined)
	fmt.Println(trns.String())
	// Output: TransitionTypeBarrier,Phase<PhaseUndefined:0 <nil> Reason:>
}

func TestDoTransition(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		trns := DoTransition(PhaseInfoUndefined)
		assert.Equal(t, TransitionTypeEphemeral, trns.Type())
		assert.Equal(t, PhaseInfoUndefined, trns.Info())
		assert.Equal(t, PhaseUndefined, trns.Info().Phase())
	})

	t.Run("someInfo", func(t *testing.T) {
		pInfo := PhaseInfoSuccess(nil)
		trns := DoTransition(pInfo)
		assert.Equal(t, TransitionTypeEphemeral, trns.Type())
		assert.Equal(t, pInfo, trns.Info())
		assert.Equal(t, PhaseSuccess, trns.Info().Phase())
	})
}

func TestDoTransitionType(t *testing.T) {
	t.Run("unknown", func(t *testing.T) {
		trns := DoTransitionType(TransitionTypeBarrier, PhaseInfoUndefined)
		assert.Equal(t, TransitionTypeBarrier, trns.Type())
		assert.Equal(t, PhaseInfoUndefined, trns.Info())
		assert.Equal(t, PhaseUndefined, trns.Info().Phase())
	})

	t.Run("someInfo", func(t *testing.T) {
		pInfo := PhaseInfoSuccess(nil)
		trns := DoTransitionType(TransitionTypeBarrier, pInfo)
		assert.Equal(t, TransitionTypeBarrier, trns.Type())
		assert.Equal(t, pInfo, trns.Info())
		assert.Equal(t, PhaseSuccess, trns.Info().Phase())
	})
}
