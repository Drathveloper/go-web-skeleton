# go-web-skeleton

A runnable template for a server-rendered Go web application, plus a generator
that turns a field list into a complete CRUD module.

A new project starts with authentication, sessions, i18n, a component-based UI
and the whole build toolchain already working. Adding an entity is one command.

## Stack

Gin · GORM + pgx · PostgreSQL · Redis-backed sessions · HTMX · Tailwind ·
go-i18n · go-playground/validator · uber/mock

## Quick start

```bash
docker compose up -d
cp cmd/server/config/application.example.yaml cmd/server/config/application.yaml
make build && make test
SEED_ADMIN_USERNAME=admin SEED_ADMIN_PASSWORD='choose-one' make run
```

Then open <http://localhost:8000>. `application.yaml` is gitignored on purpose:
startup fails loudly when it is missing rather than falling back to a committed
default, and the example carries no credentials worth stealing.

The seed variables create the first administrator on an empty database and are
ignored afterwards. There is no built-in default password — every user-management
route sits behind `Authorize(AdminRole)`, so without a seeded admin there is no
way in, and a password baked into a template would ship to every project made
from it.

## Configuration

The configuration file is `cmd/server/config/application.yaml`; when it does
not exist the loader falls back to `application.json` (copy
`application.example.json` instead of the YAML example if you prefer that
format). Only a *missing* YAML file falls back: a malformed one aborts startup
instead of silently switching to the JSON next to it.

Any file value can be overridden per deployment with an environment variable:
prefix `APP_`, double underscore for nesting, since the keys themselves
contain single underscores.

```bash
APP_SERVER__READ_TIMEOUT=30s        # server.read_timeout
APP_DATABASES__POSTGRES__PORT=15432 # databases.postgres.port
```

Loading is strict on purpose: an unknown key in the file — or an `APP_`
variable with a typo, since it maps to a key the model does not have — aborts
startup with an error naming it. A misspelled setting that is silently ignored
already sat in a production config once; that failure mode is opted out of.

Process-level settings (`PORT`, `GIN_MODE`, `ENVIRONMENT`, `SERVICE_NAME`,
`ENABLE_TLS`, `TLS_CERT_FILE`, `TLS_KEY_FILE`, `SEED_ADMIN_USERNAME`,
`SEED_ADMIN_PASSWORD`) keep their unprefixed names — they configure the
process around the application, not values inside the file.

## Starting a project from this template

```bash
go run ./cmd/scaffold new \
  --name my-erp \
  --module github.com/acme/my-erp \
  --out ../my-erp \
  --roles admin,operator
```

This copies the tree, rewrites the module path everywhere, regenerates
`common/domain/roles.go` with the roles you named, and leaves
`application.yaml` behind. `cmd/scaffold` is copied too, so the generator keeps
working inside the new project.

Add `--no-example` to drop the demonstration module and everything that
referenced it.

## Adding a CRUD module

```bash
go run ./cmd/scaffold module \
  --context billing \
  --name invoice --plural invoices \
  --roles admin \
  --field number:string:required \
  --field amount:money:required \
  --field notes:text \
  --field issued_at:date:required \
  --field paid:bool \
  --field customer_id:ref=customer
```

That writes 12 files — 15 the first time a bounded context appears, since it
also gets its shared mapper helper and its own catalog — and edits 6 shared
ones. The result builds and passes `make lint` with no manual step. Restart the application and `/invoice` is a
working listing with a modal form, HTMX row swaps, validation and translations.

### Field types

| type | Go | column | control |
|---|---|---|---|
| `string` | `string` | `text` (255) | `text` |
| `text` | `string` | `text` | `textarea` |
| `int` / `uint` | `int` / `uint` | `integer` | `number` |
| `bool` | `bool` | `boolean` | `checkbox` |
| `date` | `time.Time` | `date` | `date` |
| `datetime` | `time.Time` | `timestamptz` | `datetime-local` |
| `money` | `uint` (cents) | `bigint` | `text` + `decimal2` |
| `email` | `string` | `text` (255) | `email` |
| `ref=<entity>` | `uint` + relation | FK, `ON DELETE RESTRICT` | `select` from lookups |

Append `:required` to any of them to add `binding:"required"` and `not null`.

Money is stored in cents as an integer, never a float. `ref` adds a preloaded
relation so a listing shows the related name without a query per row.

## How it fits together

**Layering.** `domain` depends on nothing. The consumer declares the interface
it needs — `XService` in `http/handler`, `XRepository` in `service` — so a
handler never imports a service package and only `common/wire` sees both sides.
Two mapper layers are never skipped: entity↔domain in `repository/rdbms/mapper`,
domain↔DTO in `http/mapper`. The domain type is the only thing that crosses a
boundary.

**Templates register themselves.** `common/http/templates/templates.go` walks
the embedded filesystem: `pages/x/y.gohtml` becomes template `"x/y"` composed
with the base and layouts, `fragments/x/y.gohtml` becomes `"fragments/x/y"`
standalone. No template is ever named by hand, which is what makes generation
work — the generator drops a file in a directory and it is registered.

**Components hold the markup.** Everything shared lives under
`templates/files/components/`, driven by `dto.TableView` and `dto.FormView`. A
generated list page is five lines and ships no markup of its own, so a change to
the table design is one edit rather than one per module. Components live there
rather than under `fragments/` because the engine parses a fragment alone: a
`{{ define }}` inside `fragments/` is unreachable from any page.

**Shared files carry markers.** The six files a module has to be mentioned in
each contain a `scaffold:` comment, and the generator inserts above it. The seam
is visible to whoever opens the file next, and removing a generated line is
removing a line.

**Translations are per module.** `common/i18n/files/<module>.<lang>.json`, with
the language taken from the filename. A generated module writes its own catalog
instead of the generator editing a shared one — Go loses key order when it
reserialises a map, so every generation would otherwise produce a huge spurious
diff. The generator emits readable English placeholders in every locale;
translating them is your job.

## Make targets

| target | |
|---|---|
| `make build` | binary into `bin/` |
| `make test` | the whole module, `./...` |
| `make lint` | golangci-lint with `govet: enable-all` |
| `make run` | run from source |
| `make css` | recompile `styles.css` from `tailwind.css` |
| `make generate-mocks` | regenerate uber/mock mocks |

`styles.css` is a build artifact. Change `tailwind.css` and run `make css`;
never edit the compiled file.

## Conventions worth knowing before you edit

- Validation tags are `binding:`, never `validate:`. Gin's binder only reads the
  former, so a `validate:` tag looks like validation while never running.
- Errors wrap as `fmt.Errorf(constants.DefaultWrappedErrorTemplate, xxxErrMsg, err)`
  with a per-layer message constant.
- A raw `err.Error()` never reaches a page. `helper.RenderErrorPage` and
  `helper.FlashError` take an i18n key for the user and log the cause; the alert
  constructors do not accept an `error` at all.
- Error pages render with the real status code. A 500 rendered as 200 makes
  browsers cache it and HTMX treat it as a good swap.
- Struct field order has to satisfy `fieldalignment`; govet runs with
  `enable-all`.
- `gorm.ErrRecordNotFound` becomes `common/database.ErrRecordNotFound` in the
  repository and a domain error in the service. A missing record is a 404, not a
  500, which is why the sentinels live in `domain` where a handler can see them.

## Testing the generator

The generator is validated by regenerating the example module and diffing:

```bash
go run ./cmd/scaffold new --name go-web-skeleton \
  --module github.com/Drathveloper/go-web-skeleton --out /tmp/check --no-example
go run ./cmd/scaffold module --root /tmp/check --context example \
  --name item_category --field name:string:required
go run ./cmd/scaffold module --root /tmp/check --context example --name item \
  --field name:string:required --field notes:text --field stock:uint \
  --field price:money:required --field contact:email \
  --field released_at:date:required --field starts_at:datetime:required \
  --field category_id:ref=item_category:required --field active:bool
diff -r example /tmp/check/example
```

The diff must be empty. If it is not, the generator has drifted from the
reference module and one of them is wrong.
