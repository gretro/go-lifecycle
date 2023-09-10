package lifecycle_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gretro/go-lifecycle"
	assert2 "github.com/stretchr/testify/assert"
)

func Test_WhenStartingPoll_ShouldPollAtIntervals(t *testing.T) {
	readyCheck := lifecycle.NewReadyCheck()
	calls := atomic.Int32{}

	pollCheck := readyCheck.RegisterPollComponent("my-poll-component", func() bool {
		calls.Add(1)

		return true
	}, 50*time.Millisecond)

	assert := assert2.New(t)
	assert.Equal("my-poll-component", pollCheck.Name())
	assert.False(pollCheck.Ready(), "by default, component should not be ready")

	go pollCheck.Start()

	time.Sleep(275 * time.Millisecond)
	pollCheck.Stop()

	nbCalls := calls.Load()
	assert.GreaterOrEqual(calls.Load(), int32(5), "should have called the poll check at least 5 times")

	// We wait a bit to see if anymore calls will be made later down
	time.Sleep(125 * time.Millisecond)

	assert.LessOrEqual(calls.Load(), nbCalls+1, "should have call check no more than 1 extra time")
}
