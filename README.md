# Issue Board

A modern, full-stack Kanban issue tracker built with **Next.js 14** (Frontend) and **Go 1.21+** (Backend).

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-14-black?logo=next.js&logoColor=white)
![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?logo=typescript&logoColor=white)
![Tailwind CSS](https://img.shields.io/badge/Tailwind-3.0+-38B2AC?logo=tailwind-css&logoColor=white)

## 🏗 Architecture Overview

The project follows a monorepo-style structure separating the frontend and backend:

- **`api/` (Backend)**: A RESTful API written in Go using `chi` router and `SQLite` for persistence. It handles business logic, data validation, and database operations.
- **`web/` (Frontend)**: A Next.js application using the App Router. It leverages `React Query` for state management and `@dnd-kit` for the Kanban board interactions.

### Data Model

The core entities are **Issues**, **Users**, and **Labels**.

**Issue**
- `id` (UUID): Unique identifier
- `title` (String): Issue summary
- `description` (Text): Detailed description
- `status` (Enum): `backlog`, `todo`, `in_progress`, `done`, `canceled`
- `priority` (Enum): `low`, `medium`, `high`
- `assignee_id` (UUID, FK): Linked User
- `order_index` (Float): For sorting within columns
- `created_at` / `updated_at` (Timestamp)

**User**
- `id` (UUID)
- `name` (String)
- `avatar_url` (String)

**Label**
- `id` (UUID)
- `name` (String)
- `color` (String)

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- Node.js 18+
- npm

### 1. Backend Setup

Navigate to the `api` directory:
```bash
cd api
go mod download
```

**Configuration:**
The API uses environment variables. A default configuration is provided, but you can override it.
Create a `.env` file (optional):
```env
PORT=8080
DB_PATH=./issues.db
```

**Run Migrations & Seed Data:**
Initialize the database with schema and sample data:
```bash
go run cmd/seed/main.go
```

**Start the Server:**
```bash
go run cmd/api/main.go
```
The API will run on `http://localhost:8080`.

### 2. Frontend Setup

Navigate to the `web` directory:
```bash
cd web
npm install
```

**Configuration:**
Create a `.env.local` file to point to your backend:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

**Start the Development Server:**
```bash
npm run dev
```
The app will be available at `http://localhost:3000`.

## 📚 API Documentation

The API supports `X-API-Key` authentication (default key in seed: `test-api-key` if configured, otherwise open).

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/issues` | List issues. Params: `status`, `assignee`, `priority`, `page`, `page_size` |
| `POST` | `/api/issues` | Create a new issue |
| `GET` | `/api/issues/{id}` | Get issue details |
| `PATCH` | `/api/issues/{id}` | Update issue details |
| `PATCH` | `/api/issues/{id}/move` | Move issue (status/order) |
| `DELETE` | `/api/issues/{id}` | Delete an issue |
| `GET` | `/api/users` | List all users |
| `GET` | `/api/labels` | List all labels |

### Sample Requests

**Get Issues (Filtered & Paginated):**
```bash
curl "http://localhost:8080/api/issues?status=todo&page=1&page_size=10"
```

**Create Issue:**
```bash
curl -X POST http://localhost:8080/api/issues \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Fix login bug",
    "description": "Users cannot log in via email",
    "status": "todo",
    "priority": "high",
    "assignee_id": "user-uuid-here"
  }'
```

## 🛠 Tech Stack Details

- **Backend**: Go, Chi, SQLite, Go-Migrate (custom)
- **Frontend**: Next.js 14, React, TypeScript, Tailwind CSS, shadcn/ui
- **State**: TanStack Query (React Query) v5
- **DnD**: @dnd-kit (Core, Sortable, Modifiers)
- **Forms**: React Hook Form + Zod

## 🔮 Future Improvements

- **Real-time Updates**: WebSockets for live board collaboration.
- **Auth**: JWT-based authentication.
- **Testing**: E2E tests with Playwright.
