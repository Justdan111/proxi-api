# Proxi API

> Location-based reminder backend — built with Go, MongoDB, and JWT authentication.

Proxi is a REST API that powers the Proxi mobile app. It handles user accounts, reminder storage, and activity history. The mobile app handles all real-time GPS and geofencing logic; this backend focuses on data persistence and authentication.

---

## Table of Contents

- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [Running the API](#running-the-api)
- [Authentication](#authentication)
- [API Reference](#api-reference)
  - [Health](#health)
  - [Auth Endpoints](#auth-endpoints)
  - [Reminder Endpoints](#reminder-endpoints)
  - [Activity Endpoints](#activity-endpoints)
- [Data Models](#data-models)
- [Error Handling](#error-handling)
- [Deployment](#deployment)

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| Router | Chi |
| Database | MongoDB Atlas |
| Auth | JWT (HS256) |
| Password hashing | bcrypt |
| Logging | Uber Zap |
| Containerization | Docker |
| Deployment | Railway |

---

## Project Structure

```
proxi-backend/
├── cmd/server/main.go          # Entry point — wires everything together
├── internal/
│   ├── auth/
│   │   ├── handler.go          # HTTP handlers for auth routes
│   │   ├── middleware.go       # JWT middleware for protected routes
│   │   └── service.go          # Business logic — signup, login, token generation
│   ├── reminder/
│   │   ├── handler.go          # HTTP handlers for reminder routes
│   │   ├── model.go            # Reminder struct and input DTOs
│   │   ├── repository.go       # MongoDB queries
│   │   └── service.go          # Business logic — CRUD operations
│   ├── activity/
│   │   ├── handler.go          # HTTP handlers for activity routes
│   │   ├── model.go            # Activity struct and input DTOs
│   │   ├── repository.go       # MongoDB queries
│   │   └── service.go          # Business logic — logging events
│   ├── user/
│   │   ├── model.go            # User struct
│   │   └── repository.go       # User DB queries
│   └── config/
│       └── config.go           # Loads env variables
├── pkg/
│   ├── database/mongodb.go     # MongoDB connection
│   ├── logger/logger.go        # Structured logging setup
│   └── response/response.go    # Standard JSON response helpers
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── .env.example
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- A MongoDB Atlas account (free tier works fine)

### Clone and install

```bash
git clone https://github.com/yourusername/proxi-backend.git
cd proxi-backend
go mod download
```

### Set up environment variables

```bash
cp .env.example .env
# Edit .env with your values (see Environment Variables section below)
```

---

## Environment Variables

Create a `.env` file at the project root. Never commit this file.

```env
# Application
APP_ENV=development         # "development" or "production"
PORT=8080                   # Port the server listens on

# MongoDB
MONGODB_URI=mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority
MONGODB_NAME=proxi          # Database name

# JWT
JWT_SECRET=your-secret-key  # Generate with: openssl rand -hex 32
JWT_EXPIRY_HOURS=72         # How long tokens stay valid (72 = 3 days)

# CORS
ALLOWED_ORIGINS=http://localhost:3000,exp://localhost:8081
```

> **JWT_SECRET** must be a long, random string in production. Generate one with `openssl rand -hex 32`.

---

## Running the API

### With Docker (recommended)

Starts the API, MongoDB, and Mongo Express (database GUI at `localhost:8081`):

```bash
docker-compose up --build
```

### Locally without Docker

Requires MongoDB running separately:

```bash
make dev        # hot reload — restarts on file save
# or
make run        # single build and run
```

### Verify it's running

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "proxi-api",
  "mongo": "ok",
  "env": "development"
}
```

### Useful Makefile commands

```bash
make dev          # Run with hot reload (requires air)
make build        # Compile binary to bin/server
make test         # Run all tests
make docker-up    # Start full stack with Docker
make docker-down  # Stop all Docker services
make docker-logs  # Follow API logs
make mongo-shell  # Open MongoDB shell
make lint         # Run linter
```

---

## Authentication

The API uses **JWT (JSON Web Token)** authentication.

### How it works

1. The client calls `/api/auth/signup` or `/api/auth/login`
2. The server returns a `token` in the response
3. The client stores this token securely (use `expo-secure-store` in React Native — never AsyncStorage)
4. The client sends the token in the `Authorization` header on every subsequent request:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

5. If the token is missing or expired, the server returns `401 Unauthorized`
6. Tokens expire after `JWT_EXPIRY_HOURS` hours (default: 72 hours / 3 days)

### Security notes

- Passwords are hashed with **bcrypt** before storage — the raw password is never saved
- The `password` field is never returned in any API response
- All reminder and activity endpoints are **scoped to the authenticated user** — a user can never read or modify another user's data, even if they know the ID

---

## API Reference

All responses follow this consistent structure:

**Success**
```json
{
  "success": true,
  "message": "human readable message",
  "data": { }
}
```

**Error**
```json
{
  "success": false,
  "error": "description of what went wrong"
}
```

---

### Health

#### `GET /health`

Check if the API and database are running. No authentication required.

**Response**
```json
{
  "status": "ok",
  "service": "proxi-api",
  "mongo": "ok",
  "env": "production"
}
```

---

### Auth Endpoints

#### `POST /api/auth/signup`

Create a new user account.

**Request body**
```json
{
  "name": "Adebayo Okafor",
  "email": "adebayo@example.com",
  "password": "securepassword123"
}
```

| Field | Type | Rules |
|---|---|---|
| name | string | required, min 2 characters |
| email | string | required, valid email format |
| password | string | required, min 6 characters |

**Response** `201 Created`
```json
{
  "success": true,
  "message": "account created",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "65f3a2b1c4e5f6a8b9c0d1e2",
      "name": "Adebayo Okafor",
      "email": "adebayo@example.com",
      "createdAt": "2025-03-03T10:00:00Z"
    }
  }
}
```

**Errors**

| Status | Cause |
|---|---|
| `400` | Missing or invalid fields |
| `409` | Email already registered |

---

#### `POST /api/auth/login`

Log in with an existing account.

**Request body**
```json
{
  "email": "adebayo@example.com",
  "password": "securepassword123"
}
```

**Response** `200 OK`
```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": {
      "id": "65f3a2b1c4e5f6a8b9c0d1e2",
      "name": "Adebayo Okafor",
      "email": "adebayo@example.com",
      "createdAt": "2025-03-03T10:00:00Z"
    }
  }
}
```

**Errors**

| Status | Cause |
|---|---|
| `400` | Missing fields |
| `401` | Wrong email or password |

---

#### `POST /api/auth/logout`

Log out. The server acknowledges the request — the client is responsible for discarding the token.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "logged out",
  "data": null
}
```

---

#### `GET /api/auth/me` 🔒

Get the authenticated user's profile.

**Headers**
```
Authorization: Bearer <token>
```

**Response** `200 OK`
```json
{
  "success": true,
  "message": "user profile",
  "data": {
    "id": "65f3a2b1c4e5f6a8b9c0d1e2",
    "name": "Adebayo Okafor",
    "email": "adebayo@example.com",
    "createdAt": "2025-03-03T10:00:00Z"
  }
}
```

---

#### `DELETE /api/auth/me` 🔒

Permanently delete the authenticated user's account. Required by Apple App Store
guideline 5.1.1(v).

This is a **hard delete, not a deactivation**. It cascades to everything the user
owns — their reminders and their activity log — and cannot be undone. The account
is identified from the JWT, so a user can only ever delete themselves.

No request body.

**Headers**
```
Authorization: Bearer <token>
```

**Response** `200 OK`
```json
{
  "success": true,
  "message": "account deleted"
}
```

The call is **idempotent** — deleting an account that is already gone also returns
`200`, so a retry or a double tap still lets the client finish its local teardown.

Owned data is purged before the user record itself, so a failure part-way through
leaves the account intact and safe to retry rather than stranding orphaned
documents.

Note that JWTs are stateless: a token issued before deletion stays
cryptographically valid until it expires. It grants no access to data, though —
`GET /api/auth/me` returns `404` and every user-scoped collection returns empty.

| Status | Meaning |
|---|---|
| `200` | Account deleted, or already did not exist |
| `401` | Missing, invalid, or expired token |
| `500` | Deletion failed part-way; the account still exists and the call can be retried |

---

### Reminder Endpoints

All reminder endpoints require authentication. Every query is automatically scoped to the logged-in user — users cannot access each other's reminders.

---

#### `GET /api/reminders` 🔒

Get all reminders for the authenticated user, sorted by newest first.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "reminders fetched",
  "data": [
    {
      "id": "65f3a2b1c4e5f6a8b9c0d1e2",
      "userId": "65f3a2b1c4e5f6a8b9c0d1e1",
      "title": "Buy fuel",
      "location": "NNPC Station Wuse",
      "address": "Wuse Zone 5, Abuja, Nigeria",
      "radius": 300,
      "enabled": true,
      "icon": "⛽",
      "frequency": "always",
      "timeframe": {
        "startTime": "08:00",
        "endTime": "20:00"
      },
      "coordinates": {
        "latitude": 9.0820,
        "longitude": 7.4800
      },
      "triggered": false,
      "createdAt": "2025-03-03T10:00:00Z",
      "updatedAt": "2025-03-03T10:00:00Z"
    }
  ]
}
```

Returns an empty array `[]` if the user has no reminders.

---

#### `POST /api/reminders` 🔒

Create a new reminder.

**Request body**
```json
{
  "title": "Buy fuel",
  "location": "NNPC Station Wuse",
  "address": "Wuse Zone 5, Abuja, Nigeria",
  "radius": 300,
  "icon": "⛽",
  "frequency": "always",
  "timeframe": {
    "startTime": "08:00",
    "endTime": "20:00"
  },
  "coordinates": {
    "latitude": 9.0820,
    "longitude": 7.4800
  }
}
```

| Field | Type | Rules |
|---|---|---|
| title | string | required, 1–100 characters |
| location | string | required — short place name |
| address | string | required — full address |
| radius | integer | required, 50–5000 (metres) |
| icon | string | required — emoji |
| frequency | string | required — `"once"` or `"always"` |
| timeframe | object | optional — active hours window |
| timeframe.startTime | string | `"HH:MM"` format |
| timeframe.endTime | string | `"HH:MM"` format |
| coordinates.latitude | float | required |
| coordinates.longitude | float | required |

> **frequency `"once"`** — fires one time then never again.  
> **frequency `"always"`** — fires every time you enter the radius.

New reminders are created with `enabled: true` and `triggered: false` by default.

**Response** `201 Created`
```json
{
  "success": true,
  "message": "reminder created",
  "data": { ...full reminder object... }
}
```

---

#### `GET /api/reminders/:id` 🔒

Get a single reminder by ID.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "reminder fetched",
  "data": { ...reminder object... }
}
```

**Errors**

| Status | Cause |
|---|---|
| `404` | Reminder not found or belongs to a different user |

---

#### `PUT /api/reminders/:id` 🔒

Update a reminder. Only send the fields you want to change — unset fields stay as they are.

**Request body** (all fields optional)
```json
{
  "title": "Buy fuel and engine oil",
  "radius": 500
}
```

**Response** `200 OK`
```json
{
  "success": true,
  "message": "reminder updated",
  "data": { ...updated reminder object... }
}
```

---

#### `PATCH /api/reminders/:id/toggle` 🔒

Toggle a reminder's `enabled` state. If it's on, it turns off. If it's off, it turns on.

No request body needed.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "reminder toggled",
  "data": {
    "id": "65f3a2b1c4e5f6a8b9c0d1e2",
    "enabled": false,
    ...
  }
}
```

---

#### `DELETE /api/reminders/:id` 🔒

Permanently delete a reminder.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "reminder deleted",
  "data": null
}
```

---

### Activity Endpoints

The activity log is a history of events — when reminders were created, deleted, toggled, or triggered by proximity.

---

#### `GET /api/activities` 🔒

Get the last 50 activity events for the authenticated user, sorted by newest first.

**Response** `200 OK`
```json
{
  "success": true,
  "message": "activities fetched",
  "data": [
    {
      "id": "65f3a2b1c4e5f6a8b9c0d1e3",
      "userId": "65f3a2b1c4e5f6a8b9c0d1e1",
      "reminderId": "65f3a2b1c4e5f6a8b9c0d1e2",
      "reminderTitle": "Buy fuel",
      "location": "NNPC Station Wuse",
      "icon": "⛽",
      "eventType": "triggered",
      "triggeredAt": "2025-03-03T14:22:00Z"
    }
  ]
}
```

---

#### `POST /api/activities` 🔒

Log a new activity event. Called by the mobile app when a geofence triggers or when the user creates/deletes/toggles a reminder.

**Request body**
```json
{
  "reminderId": "65f3a2b1c4e5f6a8b9c0d1e2",
  "reminderTitle": "Buy fuel",
  "location": "NNPC Station Wuse",
  "icon": "⛽",
  "eventType": "triggered"
}
```

| Field | Type | Values |
|---|---|---|
| reminderId | string | valid reminder ID |
| reminderTitle | string | required |
| location | string | required |
| icon | string | emoji |
| eventType | string | `triggered`, `created`, `deleted`, `toggled` |

**Response** `201 Created`
```json
{
  "success": true,
  "message": "activity logged",
  "data": { ...activity object... }
}
```

---

## Data Models

### User

```
id          ObjectID    Unique user identifier
name        string      Display name
email       string      Unique, used for login
password    string      bcrypt hash — never returned in responses
createdAt   timestamp
updatedAt   timestamp
```

### Reminder

```
id          ObjectID    Unique reminder identifier
userId      ObjectID    Owner — all queries are scoped to this
title       string      What to remember
location    string      Short place name (e.g. "NNPC Station Wuse")
address     string      Full address string
radius      integer     Geofence radius in metres (50–5000)
enabled     boolean     Whether the reminder is active
icon        string      Emoji icon
frequency   string      "once" or "always"
timeframe   object      Optional — { startTime: "HH:MM", endTime: "HH:MM" }
coordinates object      { latitude: float, longitude: float }
triggered   boolean     Whether a "once" reminder has already fired
createdAt   timestamp
updatedAt   timestamp
```

### Activity

```
id              ObjectID    Unique activity identifier
userId          ObjectID    Owner
reminderId      ObjectID    Which reminder this relates to
reminderTitle   string      Snapshot of reminder title at time of event
location        string      Snapshot of location name
icon            string      Emoji
eventType       string      "triggered" | "created" | "deleted" | "toggled"
triggeredAt     timestamp   When the event happened
```

---

## Error Handling

All errors return a consistent JSON body:

```json
{
  "success": false,
  "error": "description of the error"
}
```

| Status Code | Meaning |
|---|---|
| `400 Bad Request` | Missing or invalid fields in request body |
| `401 Unauthorized` | Missing token, invalid token, or expired token |
| `403 Forbidden` | Valid token but insufficient permissions |
| `404 Not Found` | Resource doesn't exist or belongs to another user |
| `409 Conflict` | Duplicate — e.g. email already registered |
| `500 Internal Server Error` | Unexpected server error |

---

## Deployment

The API is deployed on **Railway** and connects to **MongoDB Atlas**.

### Environment setup

1. Push this repository to GitHub
2. Create a new project on railway.app → Deploy from GitHub repo
3. Add a MongoDB Atlas cluster (free M0 tier is sufficient)
4. In Atlas → Network Access → add `0.0.0.0/0` to allow Railway's IPs
5. Copy your Atlas connection string from Atlas → Connect → Drivers → Go
6. Set all environment variables in Railway's Variables tab (see [Environment Variables](#environment-variables))
7. Railway auto-deploys on every push to `main`

### MongoDB Atlas connection string format

```
mongodb+srv://USERNAME:PASSWORD@cluster.mongodb.net/?retryWrites=true&w=majority
```

### Production checklist

- [ ] `APP_ENV` set to `production`
- [ ] `JWT_SECRET` is a long random string (`openssl rand -hex 32`)
- [ ] MongoDB Atlas IP whitelist includes `0.0.0.0/0`
- [ ] `MONGODB_URI` uses `mongodb+srv://` (not `mongodb://`)
- [ ] Password in connection string is URL-encoded (e.g. `@` becomes `%40`)

### Local development with Docker

```bash
docker-compose up --build
# API:          http://localhost:8080
# Mongo Express: http://localhost:8081  (database GUI)
```

---

## License

MIT