# Social Feed Backend Service

[![Tests](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml/badge.svg)](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml)


A simple backend service built with **Golang**, **Gin**, and **GORM**, designed to support basic social feed features including authentication, posting, and feed management. This project is suitable for learning, showcasing fullstack backend deployment, or as a starter template.

---

## Features

* JWT-based user authentication (signup/login/logout/refresh)
* Feed CRUD operations
* Comment on specific Feed by ID
* Refresh token storage in Redis (Upstash)
* Pagination support
* Middleware-based route protection
* Clean folder structure with MVC pattern

---

## Tech Stack

* **Backend**: Go, Gin, GORM, Testify
* **Database**: PostgreSQL (via Supabase), sqlite (for tests), miniredis(for tests)
* **Auth**: JWT + Refresh Token  
* **Cache**: Redis (via Upstash)
* **Deployment**: Render
* **CI**: GitHub Actions

---

## Infrastructure / Deployment

This project uses the following external services:

* [Render](https://render.com/): Hosting backend API
* [Supabase](https://supabase.com/): PostgreSQL database provider
* [Upstash](https://upstash.com/): Serverless Redis for caching / refresh token storage

---

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/MikeChen1109/Social_feed_backend_service.git
cd Social_feed_backend_service
```

### 2. Environment Setup

Create a `.env` file and fill in:

```env
PORT=your_custom_port
DB_URL=your_postgres_url
JWT_SECRET=your_jwt_secret
REDIS_URL=your_upstash_redis_url
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Test the Server

```bash
make test 
```

### 5. Run the Server

```bash
make run 
```

API will be available at: `http://localhost:YOUR_PORT_FROM_ENVIRMENT`

---

## API Endpoints

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
├── common/             # Shared utilities and error definitions
│   ├── appErrors/      # Custom error types
│   └── helpers/        # Helper functions
│
├── controllers/        # Route handlers
├── docs/               # Swagger/OpenAPI documentation
├── initializers/       # App initialization (DB, Redis, env loading)
├── middleware/         # Middleware (JWT, CORS, etc.)
├── migrate/            # Database migration files or tools
├── models/             # GORM models and API request/response structs
├── repositories/       # Data access layer (DB operations)
├── routes/             # Route definitions and registration
├── services/           # Business logic layer
│
├── .env                # Environment config
├── .env.example        # Example env file
├── .gitignore          # Git ignore rules
├── cover.out           # Code coverage report
├── go.mod              # Go module config
├── go.sum              # Module dependency checksums
├── main.go             # App entry point
└── makefile            # run/test shortcuts

```

---

## Roadmap

* [x] Comment feature (create, list, delete)
* [x] Swagger/OpenAPI documentation
* [x] Unit testing with testify and mocks
* [x] API Gateway for routing
* [ ] Forget password feature
* [ ] Dockerfile for containerized deployment
* [ ] Kubernetes manifests for local deployment
* [ ] gRPC support with proto definitions and shared service layer
* [ ] Prometheus metrics endpoint and Grafana dashboard
* [ ] Rate limiting (e.g. IP-based using middleware or Redis)
* [ ] Database performance tuning (e.g. indexes, query optimization, slow query logging)

---

## License

MIT License


