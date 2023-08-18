package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// GracefulShutdownOptions are options used in conjunction with the [GracefulShutdown] type
type GracefulShutdownOptions struct {
	// Timeout duration after which the remaining components are considered non-responsive
	//
	// Default: 5s
	Timeout time.Duration
	// PollDuration is the delay between each poll on the shutdown channels
	//
	// Default: 100ms
	PollDuration time.Duration

	// Signals is the array of OS Signal to listen for in WaitForShutdown function
	//
	// Default: SIGINT, SIGTERM
	Signals []os.Signal
}

// GracefulShutdown is an utility that allows you to perform graceful shutdowns on different components of your application.
type GracefulShutdown struct {
	componentMutex *sync.RWMutex
	waitMutex      *sync.Mutex

	options      GracefulShutdownOptions
	appContext   context.Context
	shutdownFunc func()

	components map[string]<-chan error

	disposed bool
}

// ShutdownChan is a Producer channel used to report an error in the Shutdown process
type ShutdownChan = chan<- error

var (
	DefaultTimeout      = 5 * time.Second
	DefaultPollDuration = 100 * time.Millisecond
	DefaultSignals      = []os.Signal{
		os.Interrupt,
		syscall.SIGTERM,
	}

	ErrComponentAlreadyRegistered = errors.New("component was already registered")
	ErrAlreadyShutdown            = errors.New("shutdown has already occurred")
	ErrAlreadyWaitingForShutdown  = errors.New("shutdown is being awaited for")
	ErrShutdownTimeout            = errors.New("shutdown took too long to complete")
)

// NewGracefulShutdownWithOptions creates a new instance of [*GracefulShutdown]. You may provide a [context.Context] to enable Context Cancellation, as well as behaviour options.
func NewGracefulShutdownWithOptions(ctx context.Context, options GracefulShutdownOptions) *GracefulShutdown {
	appCtx, cancel := context.WithCancel(ctx)

	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
	}

	if options.PollDuration == 0 {
		options.PollDuration = DefaultPollDuration
	}

	if len(options.Signals) == 0 {
		options.Signals = DefaultSignals
	}

	return &GracefulShutdown{
		componentMutex: &sync.RWMutex{},
		waitMutex:      &sync.Mutex{},

		options:      options,
		appContext:   appCtx,
		shutdownFunc: cancel,

		components: make(map[string]<-chan error),
	}
}

// NewGracefulShutdown creates a new instance of [*GracefulShutdown]. You may provide a [context.Context] to enable Context Cancellation. Default options will be used.
func NewGracefulShutdown(ctx context.Context) *GracefulShutdown {
	gs := NewGracefulShutdownWithOptions(ctx, GracefulShutdownOptions{
		Timeout:      DefaultTimeout,
		PollDuration: DefaultPollDuration,
		Signals:      DefaultSignals,
	})

	return gs
}

// AppContext is the GracefulShutdown's context. Use its Done method to determine if the shutdown was requested or not.
func (gs *GracefulShutdown) AppContext() context.Context {
	return gs.appContext
}

// RegisteredComponents returns the list of registered components
func (gs *GracefulShutdown) RegisteredComponents() []string {
	gs.componentMutex.RLock()
	defer gs.componentMutex.RUnlock()

	components := make([]string, len(gs.components))
	i := 0
	for componentName := range gs.components {
		components[i] = componentName
		i++
	}

	return components
}

// RegisterComponent registers a component and return a [ShutdownChan]. Used in conjucture with `*GracefulShutdown.AppContext().Done()`,
// it allows you to report when the shutdown is done and report an optional error if the component failed to gracefully shutdown.
func (gs *GracefulShutdown) RegisterComponent(name string) (ShutdownChan, error) {
	gs.componentMutex.Lock()
	defer gs.componentMutex.Unlock()

	shutdownChan := make(chan error)

	if _, ok := gs.components[name]; ok {
		return nil, ErrComponentAlreadyRegistered
	}

	gs.components[name] = shutdownChan

	return shutdownChan, nil
}

// RegisterComponentWithFn registers a component using a function in parameter. This is a simplified way of using the registration, especially for
// simpler cases.
func (gs *GracefulShutdown) RegisterComponentWithFn(name string, shutdownFn func() error) error {
	shutdownChan, err := gs.RegisterComponent(name)
	if err != nil {
		return err
	}

	go func() {
		// Waiting for the Graceful shutdown to be requested
		<-gs.appContext.Done()

		err := shutdownFn()
		shutdownChan <- err
	}()

	return nil
}

// Shutdown will trigger the graceful shutdown process. The AppContext will be considered done, and each component will be expected to shutdown
// within the allocated time period. If any component fails to do so, the error will be reported as a return value.
//
// Invoking Shutdown multiple times will return a [ErrAlreadyShutdown] error.
func (gs *GracefulShutdown) Shutdown() error {
	gs.shutdownFunc()

	ctx, cancel := context.WithTimeout(context.Background(), gs.options.Timeout)
	defer cancel()

	err := gs.waitForComponents(ctx)
	return err
}

// WaitForShutdown blocks until the configured OS Signal is received. Once it is received, the graceful shutdown process will be triggered.
// Each component will be expected to shutdown within the allocated time period. If any component fails to do so, the error will be reported as a return value.
//
// Invoking this method multiple times will return a [ErrAlreadyWaitingForShutdown] error to be returned.
//
// Invoking this method after the application was shutdown already will cause a [ErrAlreadyShutdown] error.
func (gs *GracefulShutdown) WaitForShutdown() error {
	success := gs.waitMutex.TryLock()
	defer gs.waitMutex.Unlock()
	if !success {
		return ErrAlreadyWaitingForShutdown
	}

	if gs.disposed {
		return ErrAlreadyShutdown
	}

	ctx, cancel := signal.NotifyContext(context.Background(), gs.options.Signals...)
	defer cancel()

	<-ctx.Done()

	err := gs.Shutdown()
	return err
}

func (gs *GracefulShutdown) waitForComponents(ctx context.Context) error {
	gs.componentMutex.Lock()
	defer gs.componentMutex.Unlock()

	// Disposing of the GracefulShutdown instance
	defer func() {
		gs.disposed = true
	}()

	if gs.disposed {
		return ErrAlreadyShutdown
	}

	componentErrors := make(map[string]error)

	remainingComponents := make(map[string]<-chan error, len(gs.components))
	for componentName, channel := range gs.components {
		remainingComponents[componentName] = channel
	}

	for {
		// Check for timeout
		select {
		case <-ctx.Done():
			for componentName := range remainingComponents {
				componentErrors[componentName] = ErrShutdownTimeout
			}

			return ShutdownError{
				ComponentErrors: componentErrors,
			}
		default:
		}

		futureRemComponents := make(map[string]<-chan error, len(remainingComponents))

		for componentName, shutdownChan := range remainingComponents {
			select {
			case err := <-shutdownChan:
				if err != nil {
					componentErrors[componentName] = err
				}
			default:
				futureRemComponents[componentName] = shutdownChan
			}
		}

		// All components were shutdown with success
		if len(futureRemComponents) == 0 {
			if len(componentErrors) == 0 {
				return nil
			}

			return ShutdownError{
				ComponentErrors: componentErrors,
			}
		}

		remainingComponents = futureRemComponents
		time.Sleep(gs.options.PollDuration)
	}
}

// ShutdownError details errors by component
type ShutdownError struct {
	ComponentErrors map[string]error
}

func (err ShutdownError) Error() string {
	return fmt.Sprintf("error while shutting down (%+v)", err.ComponentErrors)
}

// IsTimeout returns true if all component errors are of type [ErrShutdownTimeout]
func (err ShutdownError) IsTimeoutErr() bool {
	for _, error := range err.ComponentErrors {
		isTimeout := errors.Is(error, ErrShutdownTimeout)

		if !isTimeout {
			return false
		}
	}

	return true
}
