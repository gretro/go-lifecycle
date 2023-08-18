package lifecycle_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gretro/go-lifecycle"
	assert2 "github.com/stretchr/testify/assert"
)

func CreateSuccessComponent(gs *lifecycle.GracefulShutdown, name string, delay time.Duration) {
	gs.RegisterComponentWithFn(name, func() error {
		time.Sleep(delay)

		// Shutdown is successful
		return nil
	})
}

func CreateErrorComponent(gs *lifecycle.GracefulShutdown, name string, err error) {
	gs.RegisterComponentWithFn(name, func() error {
		return err
	})
}

func Test_GracefulShutdown_AllSuccess(t *testing.T) {
	assert := assert2.New(t)
	gs := lifecycle.NewGracefulShutdown(context.Background())

	CreateSuccessComponent(gs, "ComponentA", 100*time.Millisecond)
	CreateSuccessComponent(gs, "ComponentB", 500*time.Millisecond)

	err := gs.Shutdown()
	assert.NoError(err)
}

func Test_GracefulShutdown_SomeFailure(t *testing.T) {
	assert := assert2.New(t)
	gs := lifecycle.NewGracefulShutdown(context.Background())

	expectedErr := errors.New("test error")

	CreateSuccessComponent(gs, "ComponentA", 100*time.Millisecond)
	CreateErrorComponent(gs, "ComponentB", expectedErr)

	err := gs.Shutdown()
	assert.Error(err)

	shutdownErr := lifecycle.ShutdownError{}
	if !assert.ErrorAs(err, &shutdownErr, "error should be a ShutdownError") {
		return
	}

	componentBErr, ok := shutdownErr.ComponentErrors["ComponentB"]
	assert.True(ok, "ComponentB should report an error")

	assert.ErrorIs(componentBErr, expectedErr)
}

func Test_GracefulShutdown_TimeoutError(t *testing.T) {
	assert := assert2.New(t)

	gs := lifecycle.NewGracefulShutdownWithOptions(context.Background(), lifecycle.GracefulShutdownOptions{
		Timeout:      250 * time.Millisecond,
		PollDuration: 50 * time.Millisecond,
	})

	CreateSuccessComponent(gs, "ComponentA", 100*time.Millisecond)

	// ComponentB will timeout
	CreateSuccessComponent(gs, "ComponentB", 500*time.Millisecond)

	err := gs.Shutdown()
	assert.Error(err)

	shutdownErr := lifecycle.ShutdownError{}
	if !assert.ErrorAs(err, &shutdownErr, "error should be a ShutdownError") {
		return
	}

	assert.True(shutdownErr.IsTimeoutErr(), "ShutdownError should only return timeout errors")
}
