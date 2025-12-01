# Issue Board - Frontend

A modern Kanban-style issue tracking board built with Next.js 16, React Query, and dnd-kit.

## 🚀 Features

- **Drag & Drop**: Intuitive drag-and-drop interface for managing issues across columns
- **Real-time Updates**: Optimistic UI updates with React Query
- **Filtering**: Filter issues by assignee, priority, and labels
- **Modal Routing**: Parallel routes for seamless modal navigation
- **Responsive Design**: Mobile-friendly interface with touch support
- **Accessibility**: ARIA labels and keyboard navigation support
- **Error Handling**: Global error boundaries for graceful error recovery

## 📋 Prerequisites

- Node.js 18+ 
- npm, yarn, pnpm, or bun
- Backend API running (default: `http://localhost:8080/api`)

## 🛠️ Setup

### 1. Install Dependencies

```bash
npm install
# or
yarn install
# or
pnpm install
```

### 2. Environment Variables

Create a `.env.local` file in the root directory:

```bash
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080/api

# API Key (WARNING: Currently exposed in client bundle - needs refactoring)
NEXT_PUBLIC_API_KEY=your-api-key-here
```

> **⚠️ Security Note**: The API key is currently exposed in the client-side bundle due to the `NEXT_PUBLIC_` prefix. This should be refactored to use Next.js API routes to keep the key server-side.

### 3. Run Development Server

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## 📁 Project Structure

```
web/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── issues/            # Issues page with parallel routes
│   │   │   ├── @modal/        # Modal slot for intercepted routes
│   │   │   ├── [id]/          # Individual issue page
│   │   │   └── page.tsx       # Main board view
│   │   ├── layout.tsx         # Root layout with providers
│   │   └── globals.css        # Global styles
│   ├── components/            # React components
│   │   ├── ui/                # shadcn/ui components
│   │   ├── board.tsx          # Main Kanban board
│   │   ├── column.tsx         # Column component
│   │   ├── issue-card.tsx     # Draggable issue card
│   │   ├── error-boundary.tsx # Error boundary component
│   │   └── ...
│   ├── lib/                   # Utilities
│   │   ├── api.ts             # API client with axios
│   │   └── utils.ts           # Helper functions
│   └── types/                 # TypeScript types
│       └── index.ts           # Shared type definitions
├── public/                    # Static assets
└── package.json
```

## 🏗️ Architecture

### State Management
- **React Query**: Server state management with caching and optimistic updates
- **React State**: Local UI state (modals, forms, drag state)

### Routing
- **App Router**: Next.js 16 App Router with parallel routes
- **Modal Interception**: `@modal/(.)id` pattern for modal overlays
- **URL State**: Filters and highlights managed via URL parameters

### Drag & Drop
- **dnd-kit**: Modern drag-and-drop with touch support
- **Order Index**: Fractional indexing for flexible reordering
- **Optimistic Updates**: Immediate UI feedback with server sync

### API Integration
- **Axios**: HTTP client with interceptors
- **Error Handling**: Global error interceptor with toast notifications
- **Type Safety**: Full TypeScript coverage

## 🎨 Tech Stack

- **Framework**: Next.js 16 (App Router)
- **Language**: TypeScript (strict mode)
- **Styling**: Tailwind CSS v4
- **UI Components**: Radix UI + shadcn/ui
- **State Management**: TanStack React Query
- **Drag & Drop**: dnd-kit
- **Forms**: React Hook Form (planned)
- **Validation**: Zod (planned)
- **Linting**: ESLint + Prettier

## 📜 Available Scripts

```bash
# Development
npm run dev          # Start dev server on port 3000

# Production
npm run build        # Build for production
npm run start        # Start production server

# Code Quality
npm run lint         # Run ESLint
npx prettier --write . # Format code with Prettier
```

## 🔑 API Endpoints

The frontend expects the following API endpoints:

### Issues
- `GET /api/issues` - List issues (with optional filters)
- `GET /api/issues/:id` - Get single issue
- `POST /api/issues` - Create issue
- `PATCH /api/issues/:id` - Update issue
- `PATCH /api/issues/:id/move` - Move issue (status + order)
- `DELETE /api/issues/:id` - Delete issue

### Users
- `GET /api/users` - List users

### Labels
- `GET /api/labels` - List labels

## 🎯 Key Features Explained

### Drag & Drop
Issues can be dragged within the same column (reordering) or across columns (status change). The order is maintained using fractional indexing:

```typescript
// First position: previous - 1
// Last position: previous + 1  
// Middle: (prev + next) / 2
```

### Optimistic Updates
Changes are immediately reflected in the UI while the server request is in flight. If the request fails, the UI reverts to the previous state.

### Modal Routing
Clicking an issue card opens it in a modal with URL update. Refreshing the page loads the issue as a full page. This is achieved using Next.js parallel routes.

### Filtering
Filters are stored in URL parameters, making them shareable and bookmarkable:
- `/issues?assignee=user-id`
- `/issues?priority=High`
- `/issues?labels=bug`

## 🐛 Known Issues & TODOs

1. **Security**: API key is exposed in client bundle (needs API route refactoring)
2. **Pagination**: All issues loaded at once (needs virtual scrolling or pagination)
3. **Testing**: No tests currently (infrastructure needed)
4. **Activity Tracking**: Hardcoded mock data in issue details
5. **Keyboard Navigation**: Drag-and-drop is pointer-only
6. **Code Splitting**: Limited dynamic imports

## 🚧 Future Enhancements

- [ ] Implement pagination or virtual scrolling
- [ ] Add comprehensive test suite (Jest + React Testing Library + Playwright)
- [ ] Refactor API key to server-side
- [ ] Add keyboard navigation for drag-and-drop
- [ ] Implement real activity tracking
- [ ] Add dark mode toggle UI
- [ ] Internationalization (i18n)
- [ ] Offline support with service workers
- [ ] Analytics and error monitoring integration

## 🤝 Contributing

1. Follow the existing code style (enforced by ESLint + Prettier)
2. Use TypeScript strict mode
3. Add tests for new features (when testing is set up)
4. Update documentation for significant changes

## 📄 License

[Your License Here]

## 🔗 Related

- [Backend Repository](link-to-backend)
- [Design System](link-to-design-system)
- [API Documentation](link-to-api-docs)
