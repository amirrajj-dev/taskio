# Taskio - Task Management System

A production-ready, real-time task management system built with Go (Golang). Taskio provides complete organization, team, project, and task management with WebSocket real-time notifications, event-driven architecture, and comprehensive API documentation.

## ✨ Features

- **🔐 Authentication & Authorization** - JWT-based auth with refresh tokens, cookie storage, role-based access control (RBAC)
- **🏢 Organization Management** - Create organizations, invite members, manage roles (owner/admin/member)
- **👥 Team Management** - Nested teams within organizations, member roles, team-based permissions
- **📋 Project Management** - Projects under teams with complete CRUD operations
- **✅ Task Management** - Full task lifecycle, subtasks, assignments, priority & status tracking
- **💬 Comments System** - Task-based commenting with real-time notifications
- **📊 Activity Logging** - Automatic activity tracking with RabbitMQ message queue
- **🔌 Real-time WebSocket** - Live notifications for task assignments, comments, updates
- **📧 Email Integration** - Invitation emails via RabbitMQ async queue
- **🚦 Rate Limiting** - Redis-based protection for auth endpoints
- **📝 Swagger Documentation** - Complete API documentation at `/swagger/index.html`
- **🐳 Docker Ready** - Multi-stage Docker build with docker-compose for dev/prod


## 🛠️ Tech Stack

<p align="center">
  <a href="https://skillicons.dev">
    <img src="https://skillicons.dev/icons?i=go,postgres,redis,rabbitmq,docker" />
  </a>
</p>

| Technology | Purpose |
|------------|---------|
| Go 1.25.7 | Core backend language |
| Gin | Web framework |
| GORM | ORM for PostgreSQL |
| PostgreSQL | Primary database |
| Redis | Rate limiting & caching |
| RabbitMQ | Message queue (async processing) |
| Gorilla WebSocket | Real-time communication |
| JWT | Authentication |
| Swagger | API documentation |
| Docker | Containerization |


## 📁 Project Structure

```
Taskio/
├── cmd/
│ ├── api/ # Main application entry point
│ └── seeds/ # Database seeding scripts
├── internal/
│ ├── app/ # Application bootstrap
│ ├── configs/ # Configuration management
│ ├── constants/ # Application constants
│ ├── consumers/ # RabbitMQ consumers (activity, cleanup)
│ ├── dtos/ # Data Transfer Objects
│ ├── errors/ # Custom error definitions
│ ├── events/ # Activity event definitions
│ ├── handlers/ # HTTP request handlers
│ ├── helpers/ # Helper functions
│ ├── infrastructure/# Database, queue, redis connections
│ ├── middlewares/ # Auth, CORS, rate-limit, role checks
│ ├── models/ # GORM database models
│ ├── repositories/ # Data access layer
│ ├── routes/ # Route registrations
│ ├── seeds/ # Seed data
│ ├── server/ # HTTP server setup
│ ├── services/ # Business logic layer
│ ├── utils/ # Utility functions (JWT, validation)
│ └── websocket/ # WebSocket hub, client, manager
├── docs/ # Swagger generated docs
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── Dockerfile
├── Makefile
└── go.mod
```

## 🚀 API Endpoints

### 🔐 Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | User registration |
| POST | `/api/auth/login` | User login (JWT in cookie) |
| POST | `/api/auth/refresh` | Refresh token |
| POST | `/api/auth/logout` | Logout user |


### 🏢 Organizations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/orgs` | Create organization |
| GET | `/api/orgs` | Get user's organizations |
| GET | `/api/orgs/:id` | Get organization details |
| GET | `/api/orgs/:id/members` | List members |
| PATCH | `/api/orgs/:id` | Update name (owner/admin) |
| DELETE | `/api/orgs/:id` | Delete org (owner only) |
| PUT | `/api/orgs/:id/users/:orgUserId/role` | Update member role |
| POST | `/api/orgs/:id/invite` | Invite user |
| POST | `/api/orgs/invites/accept` | Accept invitation |
| POST | `/api/orgs/invites/reject` | Reject invitation |
| GET | `/api/orgs/invites/pending` | Get pending invites |


### 👥 Teams

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/orgs/:id/teams` | Create team |
| GET | `/api/orgs/:id/teams` | List teams |
| PATCH | `/api/teams/:teamId` | Update team name |
| POST | `/api/teams/:teamId/members` | Add member |
| GET | `/api/teams/:teamId/members` | Get members |
| DELETE | `/api/teams/:teamId/members/:userId` | Remove member |
| PUT | `/api/teams/:teamId/members/:userId` | Change role |


### 📋 Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/teams/:teamId/projects` | Create project |
| GET | `/api/teams/:teamId/projects` | List projects |
| GET | `/api/teams/:teamId/projects/:projectId` | Get project |
| PUT | `/api/teams/:teamId/projects/:projectId` | Update project |
| DELETE | `/api/teams/:teamId/projects/:projectId` | Delete project |


### ✅ Tasks

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/projects/:projectId/tasks` | Create task |
| POST | `/api/projects/:projectId/tasks/:taskId/subtasks` | Create subtask |
| GET | `/api/projects/:projectId/tasks` | Get project tasks |
| GET | `/api/tasks/:taskId` | Get task |
| PUT | `/api/tasks/:taskId` | Update task |
| DELETE | `/api/tasks/:taskId` | Delete task |
| GET | `/api/tasks/:taskId/subtasks` | Get subtasks |


### 💬 Comments

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/tasks/:taskId/comments` | Add comment |
| GET | `/api/tasks/:taskId/comments` | Get comments |


### 📊 Activity

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/activity` | Get user activities |
| DELETE | `/api/activity/:activityId` | Delete activity |


### 🔌 WebSocket

| Endpoint | Description |
|----------|-------------|
| `GET /ws?token={jwt}` | WebSocket connection for real-time updates |


### 📝 Documentation

| Endpoint | Description |
|----------|-------------|
| `GET /swagger/*any` | Swagger UI documentation |


## 🏗️ Architecture

### Layered Architecture
```
| Layer     | Direction |
|-----------|-----------|
| Handler → | Service → | Repository → | Database |
| ↓         | ↓         | ↓           |
| DTO       | Models    | Models      |
```

### Event-Driven Cleanup

- Organization/Project/Team/Task deletion triggers RabbitMQ events
- Async consumers handle cascading deletions (comments, subtasks, members)
- Exponential backoff retry mechanism (3 attempts)


### Real-time WebSocket

- JWT authentication for WebSocket connections
- Room-based broadcasting (user, project, team, org rooms)
- Event types: `task_assigned`, `comment_added`, `task_updated`, `member_added`, etc.


### Security Features

- JWT tokens stored in HTTP-only cookies
- Refresh token rotation
- Role-based access (owner, admin, member, viewer)
- Rate limiting on auth endpoints (5/24h register, 10/15min login)
- CORS configuration for frontend


## 🚦 Prerequisites

- Go 1.25.7 or higher
- Docker & Docker Compose
- Make (optional, for Makefile commands)

## 🛠️ Installation & Setup

1. Clone the repository

```bash
git clone https://github.com/amirrajj-dev/taskio.git
cd taskio
```

2. Environment Setup

Create environment files:

```bash
cp .env.example .env.dev   # Development
cp .env.example .env.prod  # Production (update with production values)
```

Edit .env.dev with your local configuration.

3. Run with Docker (Recommended)

**Development mode** (services only, run app locally):

```bash
make docker-up-dev
# Then run app locally
make run
```

**Production mode** (full stack):

```bash
make docker-up-prod
```

4. Manual Setup (Without Docker)

Start dependencies:

```bash
# PostgreSQL
docker run -d --name postgres -e POSTGRES_DB=taskio -e POSTGRES_USER=taskio -e POSTGRES_PASSWORD=taskio -p 5432:5432 postgres:16-alpine

# Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# RabbitMQ
docker run -d --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:3-management-alpine

# MailHog (for email testing)
docker run -d --name mailhog -p 1025:1025 -p 8025:8025 mailhog/mailhog
```

Run the application:

```bash
make run
# or
go run ./cmd/api
```

5. Generate Swagger Documentation

```bash
make swagger
```
### or
```bash
	swag init --parseDependency --parseInternal --output ./docs --dir ./cmd/api,./internal/handlers
```

## 📦 Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Run application (development) |
| `make build` | Build production binary |
| `make dev` | Run with Air hot reload |
| `make prod` | Run production binary |
| `make test` | Run tests |
| `make clean` | Clean build artifacts |
| `make swagger` | Generate Swagger docs |
| `make docker-up-dev` | Start dev dependencies |
| `make docker-up-prod` | Start production stack |
| `make docker-down` | Stop all containers |

## 🔄 Database Seeding

```bash
go run ./cmd/seeds/main.go
```

## ⭐ Show Your Support

If you found this project helpful, please give it a ⭐ on GitHub!