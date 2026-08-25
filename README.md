# Camagru

*take a picture, put something on it*

Camagru is an Instagram-like photo app: capture a picture from your webcam (or upload one), composite it server-side with a picture-frame/sticker overlay, and share it to a public gallery where other users can like and comment. It also handles the less glamorous side of a real web app — email confirmation, password reset, GitHub OAuth login/linking, avatar & profile management — all built on a plain Go backend and zero-build-step vanilla JS frontend.

<!-- screenshot: gallery feed -->

## Features

**Auth**
- Register / log in (by username) / refresh / log out, with short-lived JWT access tokens and rotating, hashed refresh tokens
- Email confirmation on signup, forgot/reset password
- Sign in with GitHub, link or unlink a GitHub account from settings (with a safeguard against locking yourself out if you have no password set)

**Gallery & social**
- Public, paginated feed of every photo, newest first
- Like / unlike, comment (with an email notification to the post's owner, opt-out in settings)
- Dedicated post page for the full comment thread
- Click any username to view that person's public profile and pictures

**Photo capture & editing**
- Live webcam preview with a selectable overlay, or upload a file if you don't have a camera
- The actual image composite (webcam frame + overlay) happens server-side, never in the browser
- Manage your own posts — delete anything you've published

**Account**
- Avatar upload, profile page with stats (pictures / likes received / comments)
- Edit username or email (email changes require confirming the new address) from settings
- Delete your account entirely, including your uploaded images

## Tech stack

| | |
|---|---|
| **Backend** | Go, `net/http` (stdlib), routing/types generated from an OpenAPI spec via `oapi-codegen`, DB access generated via `sqlc`, PostgreSQL, migrations via `goose`, JWT auth, `bcrypt` password hashing |
| **Frontend** | Plain HTML + CSS + ES modules — no framework, no build step, no bundler |
| **Infra** | Docker Compose (separate prod/dev/test stacks), Caddy as reverse proxy + static file server |

## Getting started

**Prerequisites:** Docker and Docker Compose.

```bash
git clone git@github.com:grysha11/camagru.git
cd camagru
cp .env.example .env
```

Fill in `.env` — the app validates its config at startup and refuses to boot if any of these are missing:

- `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` / `DB_URL` — database credentials
- `JWT_SECRET` — any random string; e.g. `openssl rand -base64 32`
- `APP_BASE_URL` — `http://localhost:3000` for local use
- `SMTP_HOST` / `SMTP_PORT` / `SMTP_FROM` — for local dev, use `SMTP_HOST=mailpit` and `SMTP_PORT=1025`; `make dev` runs a [Mailpit](https://mailpit.axllent.org/) container for you automatically, so emails are caught locally with nothing else to configure. For `make prod` you'll need a real SMTP relay (Mailgun, SendGrid, Brevo, etc.)
- `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` / `GITHUB_REDIRECT_URL` — register an OAuth App at [github.com/settings/developers](https://github.com/settings/developers) with callback URL `http://localhost:3000/api/oauth/github/callback`
- `UPLOAD_PATH` — where uploaded/composited images are stored, e.g. `./uploads`

Then run it:

```bash
make dev    # dev stack: hot reload (air) + Mailpit, on http://localhost:3000
# or
make prod   # production build, on http://localhost:3000
```

`make dev-down` / `make prod-down` to stop. Migrations run automatically on container start.

## API docs

With the stack running, interactive API documentation (Swagger UI, generated from the OpenAPI spec) is available at [`/docs`](http://localhost:3000/docs).

## Running tests

```bash
make test        # spins up an isolated test database, runs the full backend suite
make test-down    # tears it down
```

## Project structure

```
.
├── .env.example                 template for local secrets/config
├── .gitignore
├── backend
│   ├── .air.toml                 hot-reload config for dev
│   ├── api                       OpenAPI spec + generated Swagger docs page
│   ├── assets                    overlay PNGs used for photo compositing
│   ├── cmd                       entrypoint (main.go)
│   ├── Dockerfile                 production image
│   ├── Dockerfile.dev              dev image (air hot reload)
│   ├── entrypoint.sh               runs migrations, then starts the server
│   ├── go.mod
│   ├── go.sum
│   ├── internal                   handlers, auth, db, imaging, mailer, oauth, middleware...
│   ├── php_compliance.MD           notes on the "PHP stdlib only" constraint
│   ├── scripts                    test runner script
│   ├── sql                        goose migrations + sqlc source queries
│   ├── sqlc.yaml
│   └── uploads                    captured/composited images (Docker volume)
├── docker-compose.dev.yaml        dev stack — hot reload + Mailpit
├── docker-compose.test.yaml       isolated test database
├── docker-compose.yaml            prod stack
├── frontend
│   ├── 404.html
│   ├── capture.html               webcam capture + overlay picker
│   ├── confirm-email.html
│   ├── css                        single stylesheet, no framework
│   ├── forgot-password.html
│   ├── gallery.html               public feed, main landing page
│   ├── index.html                 login / register
│   ├── js                         ES modules — api client + one controller per page
│   ├── post.html                  single post + full comment thread
│   ├── profile.html               own profile, or ?username= for a public one
│   ├── reset-password.html
│   └── settings.html              account, avatar, GitHub link, password
├── infra
│   └── Caddyfile                  reverse proxy + static file serving
├── Makefile                       make prod / dev / test
└── README.md
```

## Security highlights

- Passwords hashed with `bcrypt`; refresh tokens stored hashed (SHA-256), never in plaintext
- Refresh tokens rotate on every use, with reuse detection — a replayed token revokes the whole session family
- Per-IP rate limiting on login/register and other sensitive endpoints
- All database access goes through parameterized queries (`sqlc`) — no string-built SQL
- Frontend renders everything via `textContent`, never `innerHTML` — no DOM-based XSS
- Every mutating endpoint checks resource ownership against the authenticated user, never a client-supplied ID

## Screenshots

<!-- screenshot: photo capture with overlay selection -->

<!-- screenshot: profile page -->

## Credits

Built by [@grysha11](https://github.com/grysha11).
