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
