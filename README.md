# Social Feed Backend Service

[![Tests](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml/badge.svg)](https://github.com/MikeChen1109/Social_feed_backend_service/actions/workflows/ci.yml)


A simple backend service built with **Golang**, **Gin**, and **GORM**, designed to support basic social feed features including authentication, posting, and feed management. This project is suitable for learning, showcasing fullstack backend deployment, or as a starter template.

---

## Features

* JWT-based user authentication (signup/login/logout/refresh)
* Feed CRUD operations
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
DB_URL=your_postgres_url
JWT_SECRET=your_jwt_secret
REDIS_URL=your_upstash_redis_url
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Server

```bash
go run main.go
```

API will be available at: `http://localhost:3000`

---

## API Endpoints

### Auth

| Method | Endpoint | Description      |
| ------ | -------- | ---------------- |
| POST   | /signup  | Register a user  |
| POST   | /login   | User login + JWT |

### Feed

| Method | Endpoint     | Description          |
| ------ | ------------ | -------------------- |
| POST   | /feed/create | Create a new feed    |
| GET    | /feed/       | Get paginated feeds  |
| GET    | /feed/\:id   | Get a feed by ID     |
| PUT    | /feed/\:id   | Update a feed (auth) |
| DELETE | /feed/\:id   | Delete a feed (auth) |

---

## Folder Structure

```
.
├── controllers/    # Route handlers
├── middleware/     # JWT middleware, CORS etc.
├── models/         # GORM models
├── routes/         # API route definitions
├── services/       # Business logic layer
├── main.go         # App entry point
├── go.mod          # Module config
└── .env.example    # Example env file
```

---

## Roadmap

* [ ] Swagger/OpenAPI documentation
* [ ] Dockerfile for containerized deployment
* [ ] Rate limiting
* [ ] Database performance tuning (e.g. indexes, query optimization, slow query logging)

---

## License

MIT License


