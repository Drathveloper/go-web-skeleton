# go-web-skeleton

Runnable template for a server-rendered Go web application, plus a scaffold CLI
that generates a full CRUD module from a field list.

Extracted from a modular monolith; see `docs/` for the extraction plan. This
README is written for real in phase 8.

## Stack

Gin, GORM + pgx, PostgreSQL, Redis-backed sessions, HTMX, Tailwind,
go-i18n, go-playground/validator, uber/mock.

## Quick start

```bash
docker compose up -d
cp cmd/server/config/application.example.yaml cmd/server/config/application.yaml
make build && make test && make run
```

`application.yaml` is gitignored on purpose: startup must fail loudly when
configuration is absent rather than fall back to a committed default.
