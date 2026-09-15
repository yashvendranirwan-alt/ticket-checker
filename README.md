# Ticket System API

A small Gin REST API with in-memory users and tickets, bcrypt password hashing, and JWT authentication.

## Local run

Requirements: Go 1.22 or newer.

```bash
go mod tidy
go run ./cmd
```

The service listens on `http://localhost:8080`. Set `JWT_SECRET` in production; the development fallback is intentionally not suitable for deployment.

The web dashboard is embedded in the Go binary and is available at `http://localhost:8080/`. It provides registration, login, ticket creation, filtering, status progression, and sign out.

## Docker

```bash
docker build -t ticket-system .
docker run --rm -p 8080:8080 -e JWT_SECRET=replace-with-a-long-secret ticket-system
```

Deployed URL: `https://your-deployed-url.example.com`

Health check:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

## Endpoints

- `POST /auth/register` with `{ "username": "alice", "password": "secret" }`
- `POST /auth/login` with the same credentials; returns a JWT
- `POST /tickets` with `{ "title": "Fix login" }`
- `GET /tickets`
- `GET /tickets/:id`
- `PATCH /tickets/:id/status` with `{ "status": "in_progress" }`, then `closed`

Send the token on protected endpoints as `Authorization: Bearer <token>`. Users can only access their own tickets. Statuses may only move `open -> in_progress -> closed`.