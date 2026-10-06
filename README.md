# Social Feed Backend Service

[![Tests](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml/badge.svg)](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml)


An independent Go backend project with authentication and feed services, an API gateway, PostgreSQL, and Redis. The local setup runs entirely on your computer. Historical GKE configuration is retained for reference.

---

## Demo status

The previous public frontend demo is no longer available; its domain and cloud deployment are no longer maintained. This repository currently provides local API testing through Swagger UI and an automated smoke test. `make local-up` starts the backend and its database/Redis dependencies; it does not start a frontend application.

To try the project, follow the local setup below, then open the authentication and feed Swagger pages on localhost. No paid domain or cloud account is needed.

## Local development (recommended)

No domain, Supabase, Upstash, GKE, or paid account is required. Start Docker Desktop, then run from the repository root:

```bash
make local-up       # build, start PostgreSQL/Redis, initialize tables, wait for APIs
make local-test     # API flow, author permissions, concurrent refresh, health and metrics
make local-status   # show container status
make local-logs     # follow logs; Ctrl-C stops log viewing only
make local-down     # stop containers; keep database data
```

The first startup downloads images and Go dependencies and takes longer. Later startups use cached layers. Local ports bind to `127.0.0.1` only. Local credentials in `compose.local.yml` are disposable development defaults.

### Test in the browser

- Gateway: http://localhost:2000/
- Authentication Swagger UI: http://localhost:4000/swagger/index.html
- Feed Swagger UI: http://localhost:3000/swagger/index.html

In Swagger, use **Try it out** to sign up and log in. The login response contains `token` and `refreshToken`. For protected feed/comment endpoints, click **Authorize** in the feed Swagger UI and enter `Bearer YOUR_ACCESS_TOKEN`. The smoke test exercises the full flow automatically.

```bash
curl -X POST http://localhost:2000/api/user/signup \
  -H 'Content-Type: application/json' \
  -d '{"username":"mike_local","password":"LocalDemo123!"}'

curl -X POST http://localhost:2000/api/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"mike_local","password":"LocalDemo123!"}'

# Replace YOUR_ACCESS_TOKEN with the token returned by login.
curl -X POST http://localhost:2000/api/feed/create \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  -d '{"title":"My first local feed","content":"No domain needed"}'
```

### Fast tests without starting Docker

With Go installed:

```bash
make test
```

Repository tests use SQLite and miniredis; these tests do not need paid services or a running PostgreSQL/Redis instance. macOS command line tools are needed for SQLite and the race detector.

### Run Go directly while editing

Start only local infrastructure:

```bash
docker compose -f compose.local.yml up -d --wait postgres redis
```

In each terminal, export these variables before running Go:

```bash
export APP_ENV=local
export JWT_SECRET=local-development-only-change-for-deployment
export DB_URL='postgres://social:social_local@localhost:5433/social_feed?sslmode=disable'
```
Existing `.env` files can supply missing variables, so use the fresh checkout or verify those files before starting.

1. In `user-service`, run `go run ./migrate`, then `PORT=4000 REDIS_URL=redis://localhost:6380/0 go run .`.
2. In `feed-service`, run `go run ./migrate`, then `PORT=3000 go run .`.
3. In `api-gateway`, run `PORT=2000 FRONTEND_ORIGIN=http://localhost:5173 USER_SERVICE_URL=http://localhost:4000 FEED_SERVICE_URL=http://localhost:3000 go run .`.
4. Run `make local-test` from the root.

Stop the full container stack with `make local-down` before this mode to avoid conflicting API ports. PostgreSQL is available on localhost:5433 and Redis on localhost:6380 for optional database inspection tools.

### Data and troubleshooting

- `make local-down` preserves the named PostgreSQL volume. Local Redis is ephemeral; after recreating it, log in again to obtain new refresh tokens. Each smoke run removes its test feed but leaves two uniquely named accounts.
- A Docker daemon connection error means Docker Desktop needs to be started.
- Port conflicts: check `make local-status` and other local servers before starting.
- Startup failures: inspect `make local-logs`; migrations must succeed before the APIs start. Compose uses database health checks and migration completion dependencies ([Docker documentation](https://docs.docker.com/compose/how-tos/startup-order/)).
- `/healthz` checks the process; `/readyz` checks PostgreSQL/Redis (or both upstream APIs at the gateway). Gateway `/metrics` exports request/error counts and latency histograms; monitoring dashboards are not included.
- Kubernetes manifests and cloud deployment targets are retained as historical reference and are not required for local tests.

---

## Architecture and reliability

```text
Browser / API client
        | localhost:2000
   API gateway ---- /metrics, structured request logs
     |       |
 auth API   feed/comment API
     |       |
 PostgreSQL (users, feeds, comments)
     |
 Redis (refresh tokens, 30-day TTL)
```

- Authentication: bcrypt passwords, 15-minute HS256 access tokens, and Redis refresh tokens. PUT/DELETE feed operations require the authenticated user to be the author (403 otherwise).
- Refresh rotation: read the token's user, load the user and generate replacement tokens, then atomically check/write/delete using a Redis Lua script. Concurrent reuse has one winner; losers receive 401. A failure before rotation leaves the old token available; a lost HTTP response after successful rotation requires login again. Logout revokes the refresh token; an existing access token remains valid until its 15-minute expiry.
- Redis deployment: the rotation script targets standalone Redis. Redis Cluster requires a hash-slot key design before use. See [Redis atomic scripting documentation](https://redis.io/docs/latest/develop/programmability/eval-intro/).
- Pagination: offset is `(page - 1) * limit`, metadata preserves the requested page, and ordering uses `created_at DESC, id DESC` to break timestamp ties. Offset pagination can still shift when new records arrive; this project does not claim snapshot pagination.
- Timeouts: gateway upstream calls have a 10-second deadline and inherit client cancellation; PostgreSQL and Redis operations have independent 5-second budgets. HTTP servers use 5-second header, 15-second read, 30-second write, and 60-second idle timeouts. Repository budgets do not currently inherit HTTP request cancellation.
- Monitoring: gateway `/metrics` is Prometheus text format with request count, HTTP 4xx/5xx count, and latency histogram. `/healthz` is liveness; `/readyz` checks dependencies with a 2-second budget. Gateway request logs use route templates, status and duration, without headers, bodies, query parameters or credentials. These metrics describe gateway traffic, not direct service requests. Restrict monitoring endpoints before a public deployment.
- Tests: `make test` runs race-enabled tests, including real controller/repository integration via SQLite, non-author rejection, page overlap regression, 20-way Redis rotation (miniredis), upstream timeout/cancellation, and concurrent metric collection. `make local-test` verifies real PostgreSQL/Redis services with an 8-way HTTP refresh race and health/metrics probes.

Useful monitoring URLs:

- http://localhost:2000/metrics
- http://localhost:2000/healthz
- http://localhost:2000/readyz

---

## Features

* JWT-based user authentication (signup/login/logout/refresh)
* Feed CRUD operations
* Comment on specific Feed by ID
* Refresh token storage in Redis
* Pagination support
* Middleware-based route protection
* Clean folder structure with MVC pattern
* Docker Compose local environment with PostgreSQL and Redis; historical GKE manifests

---

## Tech Stack

* **Backend**: Go, Gin, GORM, Testify
* **Database**: Local PostgreSQL; SQLite for tests
* **Auth**: JWT + Refresh Token
* **Token storage**: Local Redis; miniredis for tests
* **CI & Local Runtime**: GitHub Actions, Docker Compose
* **Historical Deployment**: Kubernetes / GKE (configuration retained; not an active demo)

---

## Historical cloud deployment

The previous cloud deployment used these external services. The public deployment is no longer maintained; these services are not used by the local setup:

* [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine): Hosting backend services
* [Supabase](https://supabase.com/): PostgreSQL database provider
* [Upstash](https://upstash.com/): Serverless Redis for caching / refresh token storage

---

## API Endpoints

The paths below are direct service routes. When using the gateway at `http://localhost:2000`, prefix them with `/api` (for example, `/api/user/login`).

### Auth

| Method | Endpoint      | Description                     |
| ------ | ------------- | ------------------------------- |
| POST   | /user/signup  | Register a user                 |
| POST   | /user/login   | User login + access token       |
| POST   | /user/logout  | Logout and revoke refresh token |
| POST   | /user/refresh | Issue new tokens via refresh    |

### Feed

| Method | Endpoint     | Description          |
| ------ | ----------------- | ------------------------------------------------- |
| POST   | /feed/create      | Create a new feed                                 |
| GET    | /feed/            | Get all feeds                                     |
| GET    | /feed/paginated   | Get paginated feeds (with page and limit query)   |
| GET    | /feed/\:id        | Get a feed by ID                                  |
| PUT    | /feed/\:id        | Update a feed (auth)                              |
| DELETE | /feed/\:id        | Delete a feed (auth)                              |

### Comment

| Method | Endpoint              | Description                                                                |
| ------ | --------------------- | ---------------------------------------------------------------------------|
| POST   | /comment/create       | Create a comment on a specific feed (auth)                                 |
| GET    | /comment/paginated    | Get paginated comments for a feed (with page and limit query)              |

---

## Folder Structure

```
.
├── api-gateway/        # API Gateway handling request forwarding and service routing
├── feed-service/       # Microservice for feed-related features (posts, timeline, etc.)
├── user-service/       # Microservice for user management (auth, profile, etc.)
├── k8s/                # Kubernetes deployment manifests (Deployment, Service, Ingress)
├── docker/             # Dockerfiles and related build configurations
├── .github/            # GitHub Actions for CI/CD workflows
├── .dockerignore       # Docker ignore rules
├── compose.local.yml   # Local backend, PostgreSQL, Redis, and migrations
├── scripts/            # Automated API smoke test
├── makefile            # Common build, run, and test shortcuts
├── README.md           # Project documentation

```

---

## Roadmap

* [x] Comment creation and paginated listing
* [ ] Comment deletion
* [x] Swagger/OpenAPI documentation
* [x] Unit testing with testify and mocks
* [x] API Gateway for routing
* [ ] Forget password feature
* [x] Dockerfile for containerized deployment
* [x] Kubernetes manifests for local deployment
* [ ] gRPC support with proto definitions and shared service layer
* [x] Prometheus-compatible gateway metrics endpoint
* [ ] Grafana dashboard
* [ ] Rate limiting (e.g. IP-based using middleware or Redis)
* [ ] Database performance tuning (e.g. indexes, query optimization, slow query logging)

---

## License

This project is licensed under the [MIT License](LICENSE).


