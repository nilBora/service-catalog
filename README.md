# Service Catalog

A self-hosted internal service catalog for teams. Keep track of all your microservices, tools, and internal apps in one place — with URLs, categories, status, and logos.

## Features

- **Service management** — add, edit, delete services with name, URL, description, logo, and color
- **Categories** — create and manage categories; drag-and-drop to reorder; order is reflected on the dashboard
- **Dashboard** — services grouped by category with stat cards; clicking a category card scrolls to that group
- **Status tracking** — mark services as active or paused
- **Themes** — light, dark, dark-electric, dark-cyber
- **Single binary** — no external dependencies, SQLite database, embeds all static assets

## Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.24, [chi](https://github.com/go-chi/chi) router |
| Frontend | [HTMX](https://htmx.org), vanilla JS |
| Database | SQLite via [sqlx](https://github.com/jmoiron/sqlx) + [modernc/sqlite](https://gitlab.com/cznic/sqlite) |
| Auth | Session-based with bcrypt passwords |

## Getting Started

### Build and run

```bash
go build -o service-catalog ./app
./service-catalog
```

Server starts on `http://localhost:8090`. On first visit, a setup wizard creates the admin account.

### Options

| Flag | Env | Default | Description |
|---|---|---|---|
| `--db` | `DB` | `services.db` | SQLite database file path |
| `--address` | `ADDRESS` | `:8090` | Listen address |
| `--debug` | `DEBUG` | `false` | Verbose logging |

```bash
./service-catalog --db /data/services.db --address :8080
```

### Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o service-catalog ./app

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/service-catalog .
VOLUME ["/data"]
EXPOSE 8090
CMD ["./service-catalog", "--db", "/data/services.db"]
```

## Project Structure

```
app/
├── main.go               # Entry point, CLI flags
├── server/
│   ├── server.go         # HTTP server, routes setup
│   └── web/
│       ├── handler.go    # Route registration, template parsing
│       ├── pages.go      # Page handlers (dashboard, services)
│       ├── services.go   # Services CRUD handlers
│       ├── categories.go # Categories CRUD + reorder handlers
│       ├── auth.go       # Login, logout, setup handlers
│       ├── static/       # CSS, JS, HTMX (embedded)
│       └── templates/    # HTML templates (embedded)
└── store/
    ├── store.go          # Store interfaces
    ├── db.go             # SQLite connection, schema, seed data
    ├── models.go         # Data models
    ├── services.go       # Service queries
    ├── categories.go     # Category queries
    └── users.go          # User & session queries
```

## Default Categories

On first run the following categories are created automatically:

Monitoring · Infrastructure · Analytics · CRM · Documentation · Internal Tools · Communication · Security · CI/CD · Other

Categories can be renamed, reordered via drag-and-drop, or deleted from the **Manage → Categories** tab.
