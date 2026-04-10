# Repository Rules

## Commits
- Use conventional commits in the format `operation (service/scope): English message`.
- `service/scope` is mandatory.
- Split commits by semantic groups.
- Commit changes from different microservices separately, even for one feature.
- If a change spans shared infrastructure and multiple use-cases, commit shared infrastructure first, then each use-case separately.

## Service Boundaries
- Each microservice may know only what exists inside its own directory.
- Do not couple a service to sibling directories during build or runtime.

## Go Modules
- Repository submodules are independent Go modules.
- Use path-prefixed tags such as `lib-common/v0.0.1`.

## Application And Domain
- Implement use-cases as separate packages, one package per use-case.
- A use-case handler should default to a struct with dependencies and a `Handle` method.
- Return simple values when a dedicated result type adds no real value.
- Do not place the happy path above error handling.
- Wrap errors with context at the current layer instead of returning bare `err`.

## HTTP Transport
- Keep HTTP a thin adapter layer: decode requests, perform structural validation, call use-cases, and map results to responses.
- Business validation belongs to application/domain, not HTTP.
- Apply middleware in `main`, not inside route handlers.
- Internal HTTP handler signature: `func(request *http.Request) (status int, payload any)`.
- HTTP request example files must be named after the route pattern, keeping path parameter braces.

## WebSocket Transport
- Use WebSocket for server-to-client realtime delivery and connection-scoped runtime interactions.
- Runtime WebSocket subscription state is not business state unless the product explicitly requires persisted subscriptions.
- Expected chat flow: fetch history over HTTP, then subscribe over WebSocket; when switching chats, unsubscribe from the old chat, subscribe to the new chat, and fetch required history over HTTP.
- Keep connection management and chat subscription management separate when these states can be separated.
- Centralize WebSocket connection cleanup in one idempotent method, for example `ConnectionDropper.Drop`.

## Event Consumers
- If a Redis consumer handles one concrete event type, name it concretely, for example `MessageCreatedConsumer`.

## Validation And Errors
- Validation failures must use `validation.Error`.
- Transport validation is structural and safety-oriented; business validation belongs to application/domain.
- The same field may be validated in multiple layers if the purpose differs by layer.
- Error messages must be as specific as possible and include concrete entity identifiers when applicable.

## Repositories
- For repository query methods, use the contract `nil, nil` for "not found" when absence is not exceptional.

## Runtime
- Each microservice is encapsulated in Docker.
- If a service uses a database, it must also have its own database container and one-shot migrator image.
- Use a shared Docker network for inter-service communication and a dedicated local Docker network inside one service.
- All external entry goes through the `proxy` service.
- `proxy` is the single ingress point and forwards traffic to services over the internal network.
- When `main.go` wiring becomes long, group it with section comments.

## Migrations
- One migration file may operate on only one table.

## Frontend
- CSS class names must follow BEM.
- Use 4 spaces for indentation in `.tsx` files.

## IDE Files
- Do not commit local IDE user files such as `dataSources.xml`.
