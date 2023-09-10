package lifecycle_test

import (
	"testing"
	"time"

	"github.com/gretro/go-lifecycle"
	assert2 "github.com/stretchr/testify/assert"
)

func Test_WhenPulseCheckHasNoPulse_ShouldNotBeReady(t *testing.T) {
	assert := assert2.New(t)

	readyCheck := lifecycle.NewReadyCheck()
	pulse := readyCheck.RegisterPulseComponent("pulse-component", 2*time.Second)

	assert.False(pulse.Ready(), "when no pulse is recorded, should not be ready")
}

func Test_WhenPulseCheckHasNonExpiredPulse_ShouldBeReady(t *testing.T) {
	assert := assert2.New(t)

	readyCheck := lifecycle.NewReadyCheck()
	pulse := readyCheck.RegisterPulseComponent("pulse", 2*time.Second)

	pulse.RecordPulse()

	time.Sleep(100 * time.Millisecond)

	assert.True(pulse.Ready(), "pulse was recorded less than 2 seconds ago, should be ready")
}

func Test_WhenPulseCheckExpired_ShouldNotBeReady(t *testing.T) {
	assert := assert2.New(t)

	readyCheck := lifecycle.NewReadyCheck()
	pulse := readyCheck.RegisterPulseComponent("pulse", 50*time.Millisecond)

	pulse.RecordPulse()

	time.Sleep(100 * time.Millisecond)

	assert.False(pulse.Ready(), "pulse was recorded more than 50ms ago, should NOT be ready")
}
