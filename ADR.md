# Architecture Decision Record (ADR)

## 1. Framework Selection

### Frontend: Next.js (App Router)
**Decision:** Use Next.js 14 with the App Router.
**Rationale:**
- **Server Components:** Allows moving heavy logic to the server, reducing bundle size.
- **Routing:** File-system based routing simplifies project structure.
- **Ecosystem:** Seamless integration with Vercel, Tailwind, and TypeScript.
- **Performance:** Automatic code splitting and image optimization.

### Backend: Go (Chi Router)
**Decision:** Use Go with `chi` router.
**Rationale:**
- **Performance:** Go provides high performance and low latency, ideal for API servers.
- **Simplicity:** `chi` is a lightweight, idiomatic router that doesn't enforce a rigid structure like full frameworks (e.g., Gin or Echo), giving us more control.
- **Standard Library:** We leverage the standard `net/http` library for compatibility and maintainability.

### Database: SQLite
**Decision:** Use SQLite.
**Rationale:**
- **Portability:** The entire database is a single file, making it easy to move and deploy for small-scale projects.
- **Zero Configuration:** No need to set up a separate database server (Postgres/MySQL) for development.
- **Sufficiency:** For a single-user or small team issue tracker, SQLite's performance is more than adequate.

## 2. State Management

### Server State: TanStack Query (React Query)
**Decision:** Use React Query for all API data.
**Rationale:**
- **Caching:** Automatically caches API responses, reducing network requests.
- **Optimistic Updates:** Critical for a Kanban board. We can update the UI immediately when a card is dragged, then sync with the server in the background.
- **Deduplication:** Prevents multiple components from fetching the same data simultaneously.

### Local State: React `useState` / `useReducer`
**Decision:** Use standard React hooks for UI state.
**Rationale:**
- **Simplicity:** No need for a global store like Redux or Zustand for simple UI toggles (modals, dropdowns).

## 3. Drag and Drop (DnD) Implementation

**Decision:** Use `@dnd-kit`.
**Rationale:**
- **Modern:** Built for React hooks and functional components.
- **Accessibility:** excellent keyboard support and screen reader integration out of the box.
- **Modularity:** Separates sensors, modifiers, and collision detection, allowing for custom behaviors.
- **Weight:** Lighter than `react-beautiful-dnd` and maintained more actively.

## 4. Ordering Strategy

**Decision:** Floating Point Indexing (`order_index`).
**Rationale:**
- **Mechanism:** Each issue has a `float` order index. When moving an issue between A and B, its new index is `(A + B) / 2`.
- **Pros:** Very fast updates (only one row updated). No need to shift all subsequent items.
- **Cons:** Precision limits eventually (requires re-indexing).
- **Mitigation:** For this scale, precision issues are rare. A background job could re-normalize indices if needed.

## 5. Validation Approach

### Frontend
**Decision:** Zod + React Hook Form.
**Rationale:**
- **Type Safety:** Zod schemas infer TypeScript types, ensuring the form data matches the API expectations.
- **UX:** Immediate feedback to users before network requests.

### Backend
**Decision:** Manual struct validation.
**Rationale:**
- **Control:** We manually check fields in the handler (e.g., `if req.Title == ""`).
- **Simplicity:** Avoids reflection-heavy validation libraries for simple use cases.

## 6. Error Handling Strategy

### Backend
**Decision:** Structured JSON Errors.
**Rationale:**
- **Consistency:** All errors return `{ "error": "message", "details": {...} }`.
- **Parsability:** Frontend can programmatically handle specific error codes or field validation errors.

### Frontend
**Decision:** Global Toast Notifications.
**Rationale:**
- **Visibility:** Errors are shown as non-intrusive toasts (using `sonner` or `toast`).
- **Fallback:** React Query's `onError` global callback catches unhandled API errors.

## 7. Future Improvements
- **Real-time:** Implement WebSockets or Server-Sent Events (SSE) for real-time board updates.
- **Auth:** Add authentication (e.g., JWT or OAuth).
- **Testing:** Add comprehensive unit and integration tests.
