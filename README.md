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
- `status` (Enum): `Backlog`, `Todo`, `In Progress`, `Done`, `Canceled`
- `priority` (Enum): `Low`, `Medium`, `High`, `Critical`
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
- Docker (optional, for containerized API)

### 1. Backend Setup

#### Option A: Local (SQLite)

Navigate to the `api` directory:
```bash
cd api
go mod download
```

**Run Migrations & Seed Data:**
```bash
go run cmd/seed/main.go
```
This will:
- Apply database migrations from `/db/migrations/`
- Create 3 users (Alice, Bob, Charlie)
- Create 4 labels (Bug, Feature, Enhancement, Documentation)
- Create 20 issues spread across all statuses

**Start the Server:**
```bash
go run cmd/api/main.go
```
The API will run on `http://localhost:8080`.

#### Option B: Docker

Build and run the API container:
```bash
docker build -t issue-board-api ./api
docker run -p 8080:8080 issue-board-api
```
*Note: The Docker container uses an internal SQLite database. Data will not persist across restarts unless you mount a volume.*

### 2. Frontend Setup

Navigate to the `web` directory:
```bash
cd web
npm install
```

**Configuration:**
Create a `.env.local` file:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_API_KEY=J3yPAMuS0j5w4AWj6P0bh2l7prZKBSq6
```

**Start the Development Server:**
```bash
npm run dev
```
The app will be available at `http://localhost:3000`.

## ✨ Features

- **Kanban Board**: Drag and drop issues between columns (Backlog, Todo, In Progress, Done, Canceled).
- **Issue Details**: Modal and full-page view for creating and editing issues.
- **Filtering**: Filter by Assignee, Priority, and Labels (URL-synced).
- **Command Palette**: Press `Cmd+K` (or `Ctrl+K`) to quickly navigate or create issues.
- **Optimistic Updates**: UI updates immediately for a snappy experience.
- **Skeleton Loaders**: Polished loading states.
- **Authentication**: Simple `X-API-Key` header support for API security.

## 📚 API Documentation

The API supports `X-API-Key` authentication.
- **Default Key**: `J3yPAMuS0j5w4AWj6P0bh2l7prZKBSq6`
- **Header**: `X-API-Key: J3yPAMuS0j5w4AWj6P0bh2l7prZKBSq6`
- **Swagger Docs**: Available at `http://localhost:8080/docs`

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/issues` | List issues. Params: `status`, `assignee`, `priority`, `labels`, `page`, `page_size` |
| `POST` | `/api/issues` | Create a new issue |
| `GET` | `/api/issues/{id}` | Get issue details |
| `PATCH` | `/api/issues/{id}` | Update issue details |
| `PATCH` | `/api/issues/{id}/move` | Move issue (status/order) |
| `DELETE` | `/api/issues/{id}` | Delete an issue |
| `GET` | `/api/users` | List all users |
| `GET` | `/api/labels` | List all labels |

## 🛠 Tech Stack Details

- **Backend**: Go, Chi, SQLite, Go-Migrate
- **Frontend**: Next.js 14 (App Router), React, TypeScript, Tailwind CSS, shadcn/ui
- **State**: TanStack Query (React Query) v5
- **DnD**: @dnd-kit (Core, Sortable, Modifiers)
- **Forms**: React Hook Form + Zod

## ✅ Deliverables Checklist

- [x] Repository with source
- [x] README with architecture, setup, and API docs
- [x] ADR capturing key choices
- [x] Dockerfile for API
- [x] Seed script for data population

## 🔮 Future Improvements

- **Audit Logs**: Implement a real history tracking system to record all changes (status updates, assignments, edits) instead of the current static activity list.
- **Real-time Updates**: WebSockets or Server-Sent Events (SSE) for live board collaboration.
- **Notifications**: Email or in-app notifications when a user is assigned to an issue or mentioned.
- **Advanced Search**: Full-text search and advanced filtering (e.g., `is:open assignee:@me`).
- **Auth**: JWT-based authentication with OAuth providers (GitHub/Google).
- **Testing**: E2E tests with Playwright.
