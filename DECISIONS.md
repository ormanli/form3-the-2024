## Decisions

`DummyService` simulates the required behavior.
`ValidationService` validates the amount before calling `DummyService`.
This separation of concerns allows for modularity, making it easier to implement different transports (e.g., HTTP, gRPC) without affecting the business logic.

`net.Listener` is initialized when `tcp.Transport.Start` is called.
This triggers a goroutine to accept incoming connections.
For each connection, a new goroutine is spawned to handle the request.
These goroutines continue until the client closes the connection or the grace period expires.

When the provided `context.Context` is canceled, `tcp.Transport.Start` closes the `net.Listener` and waits for the grace period to finish.
This ensures that ongoing requests are processed before the server shuts down.
If the grace period elapses, remaining requests are rejected.

To test graceful shutdown scenarios, I used a clock mocking library.
This approach eliminates the need for tests to wait for actual time to pass.

This implementation does not set KeepAlive or Deadline values.
In a production service, these should be configured with appropriate values.

### Potential Performance Optimization:

The current approach of creating a new goroutine for each connection can lead to resource exhaustion, especially under heavy load.
Although goroutines are lightweight, having many active goroutines won't necessarily improve performance.
A more efficient solution would involve using a goroutine pool.
This pool could handle incoming requests, allowing for better resource utilization and improved performance.

For example, when handling a connection, a separate is spawned to handle graceful shutdown.
This task could be delegated to a worker pool that consumes requests from a channel and returns responses to another channel.
This approach reduces the number of active goroutines and improves overall performance.