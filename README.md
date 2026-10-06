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

### Auto-refreshing API dashboard

Open **[Grafana: Social Feed API Monitoring](http://localhost:3001/d/social-feed-api/social-feed-api-monitoring)** after `make local-up`. It is provisioned automatically and local viewing does not require login. Grafana refreshes every **5 seconds**, and Prometheus scrapes all three services every **5 seconds**.

Select the service matching how you tested:

| Request destination | Dashboard service |
| --- | --- |
| Auth Swagger on localhost:4000 | `user-service` (default) |
| Feed Swagger on localhost:3000 | `feed-service` |
| Gateway API on localhost:2000/api/... | `api-gateway` |

The dashboard shows per-method/per-route calls, HTTP status counts, request rate, error rate, average response time, and P95 latency. Calls passing through the gateway appear at both the gateway and the destination service; select one service to avoid double counting. Gateway latency includes forwarding overhead; direct service latency is measured inside that service.

- Count panels show totals **since the selected service restarted**, not totals for the time-picker range. Prometheus/Grafana history persists in local named volumes.
- Allow roughly 5-10 seconds for a new request to appear. Rate/latency-window charts need at least two scrapes and calls inside the selected rate window. With no traffic, these panels can be empty; the counter panels still show recorded calls.
- P95 is estimated from histogram buckets, not an exact percentile of stored requests. At very low traffic it should be read cautiously.
- Dynamic IDs use a route template such as `/feed/:id`; query strings and raw IDs are not metric labels. Health probes, metrics scrapes, and Swagger page assets are excluded.
- Prometheus targets: http://localhost:9090/targets
- `make monitoring-test` verifies direct API metrics, scrapes, all dashboard queries, and the provisioned dashboard. It leaves one local test account and removes its test feed.
- Grafana is bound to localhost with anonymous Viewer access for this local demo. Change authentication and network exposure before deploying it publicly.

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
- `/healthz` checks the process; `/readyz` checks PostgreSQL/Redis (or both upstream APIs at the gateway). Each service's `/metrics` exports request counts by status and latency histograms; Grafana provides the local API dashboard at localhost:3001.
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
- Pagination: feeds and comments use keyset cursors in `created_at DESC, id DESC` order. A continuation reads rows strictly below the previous last timestamp/ID, so insertions at the front and deletion of already-read rows cannot shift the next page. Cursors are scoped to feeds or one comment feed. PostgreSQL partial indexes support these lookups. This is not a database snapshot: unread deleted rows disappear; backdated inserts can appear later; changing sort keys externally is unsupported.
- Timeouts: gateway upstream calls have a 10-second deadline and inherit client cancellation; PostgreSQL and Redis operations derive 5-second budgets from the HTTP request context, retaining its cancellation, values, and any earlier deadline. HTTP servers use 5-second header, 15-second read, 30-second write, and 60-second idle timeouts. Redis uses pooled Redigo `DoContext` operations so cancellation closes blocked connections instead of only changing a deadline. A canceled request cannot undo an already-executed write/commit; retrying mutations must account for that uncertainty.
- Monitoring: every service exposes `/metrics` in Prometheus text format. Request counters include `service`, `method`, normalized `route`, and `status`; latency histograms include `service`, `method`, and `route`. Grafana error panels aggregate 4xx/5xx counters. `/healthz` is liveness; `/readyz` checks dependencies with a 2-second budget. All three services use the shared observability package and log route templates, status and duration, without headers, bodies, query parameters or credentials. Direct Swagger requests are recorded by the destination service. Restrict monitoring endpoints before a public deployment.
- Tests: `make test` runs race-enabled tests, including real controller/repository integration via SQLite, non-author rejection, cursor regression under insert/delete and tied timestamps, 20-way Redis rotation (miniredis), upstream timeout/cancellation, and concurrent metric collection. `make local-test` verifies real PostgreSQL/Redis services with an 8-way HTTP refresh race and health/metrics probes.

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
| GET    | /feed/paginated   | Get feeds using cursor and limit   |
| GET    | /feed/\:id        | Get a feed by ID                                  |
| PUT    | /feed/\:id        | Update a feed (auth)                              |
| DELETE | /feed/\:id        | Delete a feed (auth)                              |

### Cursor pagination (API change)

`/feed/paginated` and `/comment/paginated` now accept `cursor` and `limit` (default 10; range 1-100). Comments also require `id`, the feed ID.

```bash
# First page; use /api prefix through the gateway.
curl 'http://localhost:2000/api/feed/paginated?limit=10'
# Next page: copy meta.nextCursor from the previous response.
curl 'http://localhost:2000/api/feed/paginated?limit=10&cursor=YOUR_NEXT_CURSOR'
```

Responses retain `data` and `meta.limit`/`meta.hasMore`, replacing `meta.page` with `meta.nextCursor`. Stop when `hasMore` is false (`nextCursor` is omitted). Cursor continuation still works if the record encoded by that cursor was deleted. A cursor from another list/feed, malformed cursor, duplicate cursor/limit, or invalid limit returns 400. Any `page` parameter now returns 400 rather than silently applying offset pagination. Clients using page numbers must migrate to cursor traversal; random page-number jumps are no longer supported.

For real PostgreSQL tests (rolled back fixtures), run:

```bash
cd feed-service
TEST_POSTGRES_DSN='postgres://social:social_local@localhost:5433/social_feed?sslmode=disable' go test -race ./repositories -run TestPostgres
```

### Comment

| Method | Endpoint              | Description                                                                |
| ------ | --------------------- | ---------------------------------------------------------------------------|
| POST   | /comment/create       | Create a comment on a specific feed (auth)                                 |
| GET    | /comment/paginated    | Get comments using id, cursor, and limit              |

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
├── scripts/            # API smoke test and monitoring integration verification
├── observability/      # Shared Go instrumentation module
├── monitoring/         # Prometheus config and provisioned Grafana dashboard
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
* [x] Per-API Prometheus metrics for gateway, authentication, and feed services
* [x] Provisioned Grafana dashboard with 5-second auto-refresh
* [ ] Rate limiting (e.g. IP-based using middleware or Redis)
* [x] Request-context propagation to PostgreSQL and Redis
* [x] Cursor pagination with supporting PostgreSQL indexes
* [ ] Additional database performance tuning and slow query logging

---

## License

This project is licensed under the [MIT License](LICENSE).


