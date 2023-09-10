# Go Lifecycle

Go Lifecycle offers utilities to help you manage your application's lifecycle. This is specially useful in the context of a WebApp, where many components may need to gracefully shutdown upon receive a signal, or may need to report their readiness status.

## Installation

You may install Go Lifecycle by running the following command:

```sh
go get github.com/gretro/go-lifecycle
```

## Usage

### GracefulShutdown

The `GracefulShutdown` component allows you to register components for graceful shutdown.

```go
package main

import "github.com/gretro/go-lifecycle"

func main() {
  gs := lifecycle.NewGracefulShutdown(context.Background())

  BootstrapMyComponent(gs)

  err := gs.WaitForShutdown()
  if err != nil {
    panic("unable to gracefully shutdown " + err.Error())
  }
}

func BootstrapMyComponent() {
  // Bootstrap your component here

  err := gs.RegisterComponentWithFn("MyComponent", func() error {
    // Perform the shutdown logic for the component here
    return nil
  })

  if err != nil {
    panic("unable to bootstrap component")
  }
}
```

For more advanced use cases, you may use the `NewGracefulShutdownWithOptions` function instead.

### Ready check
The `ReadyCheck` component allows you to register checks with 3rd party components. This is useful when dealing with
readiness check in platforms such as Kubernetes.

```go
package main

import "github.com/gretro/go-lifecycle"

func main() {
  readycheck := lifecycle.NewReadyCheck()

  // Poll Checks will poll a given component at regular intervals
  readycheck.RegisterPollComponent("db", func() bool {
    // Implement DB check
  }, 5 * time.Second)

  // Pulse Checks must record a pulse before the timeout happens
  pulseCheck := readycheck.RegisterPulseComponent("pulse")

  // Whatever component is here...
  component := struct{}
  component.On("alive", func() {
    pulseCheck.RecordPulse()
  })

  // Push Checks manually record readiness checks without an expiration mechanism
  pushCheck := readycheck.RegisterPushComponent("push")
  pushCheck.SetReady(true)

  // Starts executing poll checks
  readycheck.StartPolling()

  // Returns `true` if all components are ready. Useful when wrapped in a HTTP endpoint
  isReady := readycheck.Ready()

  // Returns a map explaining which components are ready and which are not
  explanation := readycheck.Explain()

  // Stops poll checks
  readycheck.StopPolling()
}
```

You can implement your own health check mechanism by implementing the `ComponentCheck` interface and calling `RegisterComponent` on your Ready check.
