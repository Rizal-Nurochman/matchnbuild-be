# Match and Build — Backend API

Backend service for **Match and Build**, a platform that connects clients with designers. Clients post project requests, designers send quotations, payments are processed through Midtrans, and both sides communicate over a real-time chat.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.23-blue.svg)](https://golang.org/) [![PostgreSQL](https://img.shields.io/badge/PostgreSQL-%3E%3D%2015.0-blue.svg)](https://www.postgresql.org/) [![Gin](https://img.shields.io/badge/Gin-Web%20Framework-red.svg)](https://gin-gonic.com/) [![GORM](https://img.shields.io/badge/GORM-ORM-green.svg)](https://gorm.io/)

## Overview

The project follows **Clean Architecture** with a Controller–Service–Repository layering. Each feature lives in its own module under `modules/`, wired together through dependency injection (`samber/do`).

### Tech Stack
- **Language:** Go 1.23 (toolchain 1.24)
- **Web framework:** Gin
- **ORM:** GORM (PostgreSQL, with SQLite support for tests)
- **Dependency injection:** `samber/do`
- **Auth:** JWT (`golang-jwt/jwt`)
- **Real-time:** WebSocket (`gorilla/websocket`)
- **Payments:** Midtrans Snap
- **Media:** ImageKit
- **Validation:** `go-playground/validator`

## Modules

| Module | Responsibility |
| --- | --- |
| `auth` | Registration, login, JWT issuing, refresh tokens |
| `user` | User accounts and profiles |
| `user_preferences` | Per-user preference data used for recommendations |
| `designer` | Designer profiles |
| `design_item` | Design catalog items |
| `recommendation` | Design item recommendations |
| `project_request` | Client project requests and conversations |
| `quotation` | Designer quotations and resulting orders |
| `payment` | Midtrans Snap tokens, payment status, webhook handling |
| `chat` | Conversations, message history, and real-time chat over WebSocket |
| `upload` | File/image uploads via ImageKit |

## Project Structure

```text
cmd/            Application entrypoint (main.go)
config/         Database and third-party configuration
database/       Entities, migrations, seeders, migration manager
middlewares/    Auth, CORS, and other Gin middleware
modules/        Feature modules (controller / service / repository / dto)
pkg/            Shared helpers, constants, utils
providers/      Dependency injection wiring
script/         CLI scripts runnable via flags
```

## Quick Start

### Prerequisites
- Go `>= 1.23`
- PostgreSQL `>= 15.0`
- The `uuid-ossp` Postgres extension (entities use `uuid_generate_v4()`)

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/Rizal-Nurochman/matchnbuild.git
   cd matchnbuild
   ```
2. Copy the example environment file and configure it:
   ```bash
   cp .env.example .env
   ```
3. Install dependencies:
   ```bash
   make dep
   ```

## Running the Application

### Option 1: Without Docker
1. Configure `.env` with your PostgreSQL credentials (see [Environment Variables](#environment-variables)).
2. Run migrations and start the server:
   ```bash
   make migrate       # Run migrations
   make seed          # Run seeders (optional)
   make migrate-seed  # Migrations + seeders in one step
   make run           # Start the application
   ```
   The server listens on `GOLANG_PORT` (default `8888`).

### Option 2: With Docker
1. Configure `.env` with your PostgreSQL credentials.
2. Build and start the containers:
   ```bash
   make init-docker
   ```
3. Run migrations and seeders inside the container:
   ```bash
   make migrate-seed-docker
   ```

### Graceful Shutdown
The server runs an HTTP server alongside the WebSocket hub. On `SIGINT`/`SIGTERM` it stops the chat hub (closing client connections and flushing presence) and then shuts the HTTP server down with a 10-second timeout.

## Environment Variables

```env
APP_NAME=Match.And.Build
IS_LOGGER=true

DB_HOST=postgres
DB_USER=postgres
DB_PASS=<your password>
DB_NAME=<your database name>
DB_PORT=5432

NGINX_PORT=80
GOLANG_PORT=8888
APP_ENV=localhost          # use "production" to enable strict WS origin checks
JWT_SECRET=<your secret key>

# Comma-separated allowlist of browser origins permitted to open the WebSocket.
# Required in production: when empty, browser origins are denied while APP_ENV=production.
WS_ALLOWED_ORIGIN=https://your-frontend.example.com,http://localhost:3000

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_SENDER_NAME="Match and Build <no-reply@testing.com>"
SMTP_AUTH_EMAIL=<your email>
SMTP_AUTH_PASSWORD=<your password>

IMAGEKIT_PUBLIC_KEY=<your public key>
IMAGEKIT_PRIVATE_KEY=<your private key>
IMAGEKIT_URL_ENDPOINT=<your url endpoint>

MIDTRANS_SERVER_KEY=<your Midtrans server key>
MIDTRANS_CLIENT_KEY=<your Midtrans client key>
MIDTRANS_ENVIRONMENT=sandbox
MIDTRANS_NOTIFICATION_URL=https://your-public-api.example.com/api/v1/payment/notification
MIDTRANS_FINISH_URL=https://your-frontend.example.com/payment/finish
MIDTRANS_EXPIRY_MINUTES=60
```

## Real-Time Chat (WebSocket)

The chat module exposes both REST endpoints and a WebSocket connection. Messages sent through either path are persisted and broadcast to the other participants through the same hub, so REST and WebSocket stay in sync.

### REST Endpoints

All routes are prefixed with `/api/v1` and require a `Bearer` token.

```text
GET  /conversations                      List the caller's conversations
GET  /conversations/:id/messages         Paginated message history (?before=, ?after=, ?limit=)
POST /conversations/:id/messages         Send a message
GET  /conversations/:id/unread-count     Unread count for one conversation
GET  /conversations/unread-count         Total unread count across conversations
```

### WebSocket Connection

```text
GET /api/v1/ws
```

Authentication token is read in the following order:
1. `Authorization: Bearer <jwt>` header
2. `Sec-WebSocket-Protocol: bearer, <jwt>` subprotocol
3. `?token=<jwt>` query parameter (fallback for browser clients)

On connect, the server resolves the user's conversations and joins them automatically, so incoming messages are delivered without an explicit subscribe step. Origin is validated against `WS_ALLOWED_ORIGIN` (exact match); with `APP_ENV=production` and no allowlist configured, browser origins are denied.

### Client → Server Events

| `type` | Payload (`data`) | Description |
| --- | --- | --- |
| `message.send` | `conversation_id`, `client_message_id`, `text`, `message_type`, `attachment_url?` | Send a message. `client_message_id` makes the send idempotent on retry. |
| `message.read` | `conversation_id`, `message_id` | Mark messages as read up to `message_id`. |
| `typing.start` / `typing.stop` | `conversation_id` | Typing indicator. |
| `conversation.join` | `conversation_id` | Join a conversation created after connecting. |

### Server → Client Events

| `type` | Description |
| --- | --- |
| `message.created` | A new message (also an ACK to the sender, carrying `request_id`). |
| `message.read` | A participant read up to a message. |
| `typing.start` / `typing.stop` | A participant's typing state. |
| `presence.changed` | A peer went online/offline (scoped to shared conversations). |
| `error` | Validation/authorization failure, with `code` and `message`. |

Each connection is rate limited (50 events/second) and pinged periodically to detect dead connections.

## Make Commands

### Development
```bash
make dep          # Install and tidy dependencies
make run          # Run the application locally
make build        # Build the binary (./main)
make run-build    # Build and run
make module name=<module_name>  # Scaffold a new module
```

### Migrations
```bash
make migrate                                # Run all pending migrations
make migrate-status                         # Show migration status
make migrate-rollback                       # Rollback the last batch
make migrate-rollback-batch batch=<number>  # Rollback a specific batch
make migrate-rollback-all                   # Rollback all migrations
make migrate-create name=<migration_name>   # Create a new migration file
make seed                                   # Run seeders
make migrate-seed                           # Migrations + seeders
```

Docker equivalents are available with a `-docker` suffix (e.g. `make migrate-docker`).

**Migration system features:**
- Batch-based migrations with status tracking and rollback by batch or all.
- When creating a migration named `create_*_table`, the tooling also scaffolds the entity in `database/entities/` and registers it in `database/migration.go`.

## CLI Flags

Migrations, seeding, and scripts can also be run directly, optionally keeping the server up with `--run`:

```bash
go run cmd/main.go --migrate:run --seed --run
```

| Flag | Description |
| --- | --- |
| `--migrate` / `--migrate:run` | Apply pending migrations |
| `--migrate:status` | Show migration status |
| `--migrate:rollback` | Rollback the last batch |
| `--migrate:rollback <batch>` | Rollback a specific batch |
| `--migrate:rollback:all` | Rollback all migrations |
| `--migrate:create:<name>` | Create a migration file |
| `--seed` | Seed the database |
| `--script:<name>` | Run a script registered in `script/` |
| `--run` | Keep the server running after the above |

## Testing

```bash
go test ./...                       # Run all tests
go test ./modules/chat/...          # Run chat module tests
```

> The race detector (`go test -race`) requires CGO and a C compiler installed.

## Payments (Midtrans)

The payment module uses Midtrans Snap. Endpoints:

```text
POST /api/v1/payment/:order_id/snap-token  Bearer token required
GET  /api/v1/payment/:order_id/status      Bearer token required
POST /api/v1/payment/notification          Public Midtrans webhook
```

Set the Payment Notification URL in the Midtrans dashboard to the same HTTPS URL as `MIDTRANS_NOTIFICATION_URL`. The webhook verifies the Midtrans SHA-512 signature and confirms the transaction via the Get Status API before updating the local payment and order.

Frontend integration:

```html
<script
  src="https://app.sandbox.midtrans.com/snap/snap.js"
  data-client-key="SB-Mid-client-...">
</script>
<script>
  snap.pay(response.data.snap_token);
</script>
```

For production, set `MIDTRANS_ENVIRONMENT=production` and load `https://app.midtrans.com/snap/snap.js`.

## Logs

A built-in logging interface is available while the application is running:

```text
http://your-domain/logs
```

It supports monthly filtering, refresh, and expandable entries.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Gin Web Framework](https://gin-gonic.com/)
- [GORM](https://gorm.io/)
- [samber/do](https://github.com/samber/do) for dependency injection
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Go Playground Validator](https://github.com/go-playground/validator)
- [Testify](https://github.com/stretchr/testify) for testing
