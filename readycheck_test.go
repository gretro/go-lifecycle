package lifecycle_test

import (
	"testing"
	"time"

	"github.com/gretro/go-lifecycle"
	assert2 "github.com/stretchr/testify/assert"
)

func Test_WhenReadyChecking_ShouldGetReadinessFromComponents(t *testing.T) {
	assert := assert2.New(t)

	readycheck := lifecycle.NewReadyCheck()

	readycheck.RegisterPollComponent("component-1", func() bool {
		return false
	}, 25*time.Millisecond)

	pushCheck := readycheck.RegisterPushComponent("component-2")
	pushCheck.SetReady(false)

	readycheck.RegisterPollComponent("component-3", func() bool {
		return true
	}, 25*time.Millisecond)

	readycheck.StartPolling()

	time.Sleep(100 * time.Millisecond)

	assert.False(readycheck.Ready(), "should not be ready")

	explanation := readycheck.Explain()

	componentReady, ok := explanation["component-1"]
	if !assert.True(ok, "unknown component-1") {
		assert.False(componentReady, "component-1 should not be ready")
	}

	componentReady, ok = explanation["component-2"]
	if !assert.True(ok, "unknown component-2") {
		assert.True(componentReady)
	}

	componentReady, ok = explanation["component-3"]
	if !assert.True(ok, "unknown component-3") {
		assert.True(componentReady)
	}
}
