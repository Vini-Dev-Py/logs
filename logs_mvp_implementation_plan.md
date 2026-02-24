# Logs — MVP Implementation Plan (Visual Log Tracing Platform)

> **Goal:** Build an observability platform that turns raw logs into an **interactive flow graph** (nodes + edges) per request/trace, with a clean UX (login-centered, sidebar navigation, animations), scalable ingestion, and storage optimized for high write throughput.

---

## 0) Scope & MVP Success Criteria

### MVP must deliver

- **Multi-tenant** (Company / Users / API Key)
- **Auth + Permissions (basic)**: user belongs to a company; can only see their company’s traces.
- **Log ingestion** endpoint using **Company API Key** (write path is async via queue).
- **Trace viewer** in the UI:
  - A list/table of traces (filtered by date/time, status, service)
  - Opening a trace renders a **React Flow** graph with nodes/edges
  - Nodes show: type, label, duration, status, timestamps, metadata
  - Basic highlights: error nodes, slow nodes (threshold)
  - **Annotations nodes**: user can add annotation cards to the flow (stored in Postgres)
- **Docker Compose** for local dev: Postgres + RabbitMQ + Cassandra + services + Traefik/Nginx LB.
- **Seed** script to create first admin user + company + apiKey.
- **Replicas**: at least ingestion service replicated behind LB in compose (for testing stateless scale).

### Not required for MVP (Phase 2+)

- Full-text search (OpenSearch/Elasticsearch)
- Advanced RBAC (per-service visibility rules)
- Metrics correlation (CPU spikes) beyond timestamp-based heuristics
- Distributed tracing auto-instrumentation (OpenTelemetry SDK) — we’ll keep it as integration later

---

## 1) Architecture Overview (MVP)

### Services

1. **logs-bff** (Go) — “Backend for Frontend”
   - Auth (JWT), sessions
   - Companies, users, apiKeys
   - Trace listing (query API), fetch trace graph
   - Annotation CRUD (stored in Postgres)
   - Orchestrates calls to query service (internal network)

2. **logs-ingest** (Go) — ingestion API (high availability)
   - Receives log events: `POST /v1/log-events`
   - Validates company via apiKey (fast lookup in Postgres cache or Cassandra api_keys table)
   - Publishes to **RabbitMQ** (async)
   - Responds **202 Accepted** quickly

3. **logs-worker** (Go) — consumers
   - Consumes events from RabbitMQ
   - Persists into **Cassandra** (logs/traces/nodes/edges)
   - Ensures idempotency (dedupe by eventId)

4. **logs-query** (Go) — read model API (optimized for reads)
   - Fetch trace graph by `traceId`
   - Lists traces by time range / status / service
   - Returns graph structure ready for React Flow

### Data Stores

- **PostgreSQL**: tenants, users, api keys, annotations, UI preferences
- **Cassandra**: high-volume event storage and trace materialization
- **RabbitMQ**: buffering + backpressure + async ingestion

### Load Balancer / Edge

- **Traefik** (preferred in Docker Compose) or Nginx:
  - Routes `/api/*` to **logs-bff**
  - Routes `/ingest/*` to **logs-ingest** (replicated)
  - Optionally routes `/query/*` to **logs-query**

---

## 2) Tech Stack (Recommended)

### Backend (Go)

- Language: **Go 1.22+**
- HTTP: **chi** (or fiber if you prefer, but chi keeps it simple)
- DB (Postgres): **pgx** + migrations (goose)
- Cassandra: **gocql**
- RabbitMQ: **amqp091-go**
- Auth: JWT (access token + refresh optional)
- Validation: **go-playground/validator**
- Observability: structured logging (zap or slog)

### Frontend (React)

- **React + Vite + TypeScript**
- **Tailwind CSS**
- **Framer Motion** for sidebar animations, hover micro-interactions
- **React Flow** for the trace graph
- Icons: **lucide-react**
- State: TanStack Query + Zustand (or Redux Toolkit if you prefer)

---

## 3) Domain Concepts (DDD vocabulary)

- **Company (Tenant)**: owns traces, users, apiKey
- **User**: belongs to company; has role (`ADMIN`, `MEMBER`)
- **Trace**: request execution timeline (traceId) for a single “journey”
- **Node (Span-like)**: operation step (db query, http call, if, loop, service transition)
- **Edge**: parent-child relation between nodes
- **Annotation**: user-created node (note/comment) attached to trace graph

---

## 4) Event Model (Log Event Contract)

### `LogEvent` (JSON)

```json
{
  "eventId": "uuid",
  "companyId": "uuid-or-int",
  "apiKey": "string",
  "traceId": "string",
  "nodeId": "string",
  "parentNodeId": "string|null",
  "serviceName": "string",
  "operation": {
    "type": "DB|HTTP|IF|LOOP|FILE_UPLOAD|EXTERNAL_API|SERVICE_TRANSITION|CUSTOM",
    "name": "string",
    "status": "OK|ERROR",
    "startAt": "2026-02-24T12:00:00.000Z",
    "endAt": "2026-02-24T12:00:00.120Z",
    "durationMs": 120
  },
  "http": {
    "method": "GET",
    "path": "/clients",
    "statusCode": 200
  },
  "db": {
    "system": "postgres",
    "query": "select ...",
    "rows": 10
  },
  "metadata": {
    "any": "json"
  }
}
```

### Notes

- `nodeId` must be unique within `traceId`
- `parentNodeId` builds edges
- `eventId` enables idempotency in worker
- `durationMs` can be computed server-side if only start/end provided

---

## 5) Cassandra Schema (MVP CQL)

> **Design principle:** Cassandra queries must be known upfront. MVP needs:
>
> 1. List traces by company + day/time
> 2. Get nodes/edges for a trace
> 3. (Optional) list nodes by trace quickly

### Keyspace

```sql
CREATE KEYSPACE IF NOT EXISTS logs
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

### Tables

#### 5.1 traces_by_company_day

List traces for a company in a time window.

```sql
CREATE TABLE IF NOT EXISTS logs.traces_by_company_day (
  company_id text,
  day text,                       -- YYYY-MM-DD
  started_at timestamp,
  trace_id text,
  status text,                    -- OK|ERROR
  root_operation text,
  service_name text,
  http_method text,
  http_path text,
  http_status int,
  duration_ms int,
  PRIMARY KEY ((company_id, day), started_at, trace_id)
) WITH CLUSTERING ORDER BY (started_at DESC);
```

#### 5.2 nodes_by_trace

Fetch all nodes for a trace.

```sql
CREATE TABLE IF NOT EXISTS logs.nodes_by_trace (
  trace_id text,
  node_id text,
  parent_node_id text,
  company_id text,
  service_name text,
  type text,
  name text,
  status text,
  start_at timestamp,
  end_at timestamp,
  duration_ms int,
  http_method text,
  http_path text,
  http_status int,
  db_system text,
  db_query text,
  db_rows int,
  metadata text,                  -- JSON string (keep flexible)
  PRIMARY KEY ((trace_id), node_id)
);
```

#### 5.3 edges_by_trace

Fetch edges for a trace (optional; can be derived from parent_node_id).

```sql
CREATE TABLE IF NOT EXISTS logs.edges_by_trace (
  trace_id text,
  from_node_id text,
  to_node_id text,
  kind text,                      -- PARENT_CHILD|SERVICE_BRANCH
  PRIMARY KEY ((trace_id), from_node_id, to_node_id)
);
```

#### 5.4 event_dedup

Idempotency.

```sql
CREATE TABLE IF NOT EXISTS logs.event_dedup (
  company_id text,
  event_id text,
  received_at timestamp,
  PRIMARY KEY ((company_id), event_id)
) WITH default_time_to_live = 604800; -- 7 days TTL
```

#### 5.5 api_keys (optional in Cassandra)

Fast apiKey validation without hitting Postgres (Phase 1 can keep in Postgres only).

```sql
CREATE TABLE IF NOT EXISTS logs.api_keys (
  api_key text,
  company_id text,
  active boolean,
  created_at timestamp,
  PRIMARY KEY ((api_key))
);
```

---

## 6) PostgreSQL Schema (MVP)

### Core tables

- `companies (id, name, plan_user_limit, api_key, created_at)`
- `users (id, company_id, name, email, password_hash, role, created_at)`
- `annotations (id, company_id, trace_id, node_id, x, y, text, created_by, created_at)`

### Indexes

- `users(email)` unique
- `annotations(company_id, trace_id)`
- `companies(api_key)` unique

---

## 7) API Design (Endpoints)

### 7.1 logs-bff (public API used by UI)

- `POST /api/auth/login`
- `POST /api/auth/logout` (optional)
- `GET  /api/me`
- `GET  /api/traces?from=&to=&status=&service=&text=` (text is Phase 2)
- `GET  /api/traces/{traceId}` → returns `{ nodes, edges, annotations }` ready for React Flow
- `POST /api/traces/{traceId}/annotations`
- `PUT  /api/annotations/{id}`
- `DELETE /api/annotations/{id}`
- `POST /api/admin/seed` (dev-only; removed in prod) OR CLI seed command

### 7.2 logs-ingest (write path)

- `POST /ingest/v1/log-events` (apiKey required)
  - returns `202 Accepted` with `{accepted: true}`
- `POST /ingest/v1/trace-start` (optional sugar)
- `POST /ingest/v1/trace-end` (optional sugar)

### 7.3 logs-query (internal)

- `GET /query/v1/traces?companyId=&from=&to=&status=&service=`
- `GET /query/v1/traces/{traceId}`

---

## 8) UI/UX Requirements (MVP)

### Global UI Principles

- Clean, minimal, text + icons
- Strong hover states
- Micro animations with **Framer Motion**
- Fast navigation, no clutter

### Screens

#### 8.1 Login (Centered)

- Centered card
- Fields: email + password
- CTA: “Entrar”
- Small brand/title “Logs”
- Error feedback inline

#### 8.2 App Layout (Private Routes)

- **Sidebar** with:
  - Logo/top section
  - Menu groups with submenus (collapse/expand)
  - Icons + simple labels
  - Hover highlight + active indicator
  - Motion animation for open/close
- Main area: topbar with search/filter (Phase 2 search)
- Content area: responsive

#### 8.3 Traces List

- Table with:
  - startedAt
  - method + path
  - status (badge)
  - duration
  - service
- Filters:
  - date range
  - status
  - service
- Click row → opens Trace Viewer

#### 8.4 Trace Viewer (React Flow)

- Left: graph canvas
- Right: inspector panel for selected node
  - operation type, name
  - duration
  - timestamps
  - metadata (pretty JSON viewer)
- Highlights:
  - status ERROR
  - duration > threshold (slow)
- Add Annotation:
  - button “Add note” creates draggable annotation node
  - saved to Postgres

---

## 9) Folder Structure (DDD + CQRS + Clean Architecture)

> Recommendation: **monorepo** with multiple services + shared packages.

### Repo layout

```
logs/
  apps/
    logs-bff/
    logs-ingest/
    logs-worker/
    logs-query/
    web/
  packages/
    shared-contracts/         # JSON schemas, DTOs, event contract
    shared-logger/            # logging helpers
    shared-auth/              # jwt utils (optional)
  infra/
    docker/
    migrations/
    cassandra/
    rabbitmq/
  docs/
  Makefile
  docker-compose.yml
```

### Service internal structure (example: apps/logs-query)

```
apps/logs-query/
  cmd/api/main.go
  internal/
    app/
      usecase/                # orchestration
      query/                  # CQRS query handlers
    domain/
      model/                  # entities/value objects
      service/                # domain services
    infra/
      http/                   # controllers/handlers
      cassandra/              # repositories
      config/
      observability/
    ports/
      in/                     # interfaces for controllers
      out/                    # repository interfaces
  go.mod
```

### CQRS rule of thumb

- **Commands** mutate state (mostly ingestion/worker/annotations)
- **Queries** read optimized models (logs-query)

---

## 10) Docker & Local Infrastructure

### docker-compose services

- traefik (LB)
- postgres
- cassandra
- rabbitmq
- logs-bff
- logs-ingest (replicas = 2)
- logs-worker (replicas = 2)
- logs-query (replicas = 1)
- web

### Routing (Traefik)

- `http://localhost:5173` → web
- `http://localhost/api/*` → logs-bff
- `http://localhost/ingest/*` → logs-ingest
- `http://localhost/query/*` → logs-query (optional internal only)

---

## 11) Implementation Stages (Ordered Plan)

## Stage 1 — Repo bootstrap + infra (Day 1)

1. Create monorepo structure (`apps/`, `packages/`, `infra/`).
2. Setup `docker-compose.yml` with:
   - Postgres
   - Cassandra
   - RabbitMQ
   - Traefik
3. Add migrations folder for Postgres + CQL scripts for Cassandra.
4. Add Makefile tasks:
   - `make up`, `make down`, `make logs`, `make migrate`, `make seed`

**Deliverable:** infra up locally and reachable.

## Stage 2 — Postgres Auth + Seed (Day 2)

1. Implement **logs-bff**:
   - Connect to Postgres
   - Migrations: companies/users/annotations
2. Implement auth:
   - password hashing (bcrypt)
   - JWT issue/validate middleware
3. Create **seed** CLI:
   - creates default company + admin user + apiKey
4. Endpoints: `/api/auth/login`, `/api/me`

**Deliverable:** login works with seeded user.

## Stage 3 — Ingestion API + Queue (Day 3)

1. Implement **logs-ingest**:
   - `POST /ingest/v1/log-events`
   - validate apiKey (Postgres lookup + optional in-memory cache)
   - publish message to RabbitMQ queue `log_events`
2. Add rate-limit / basic protection (optional)
3. Add ingest replicas in compose

**Deliverable:** client can send events and get 202 quickly.

## Stage 4 — Worker + Cassandra persistence (Day 4)

1. Implement **logs-worker**:
   - consume from RabbitMQ
   - idempotency using `event_dedup`
   - write nodes into `nodes_by_trace`
   - maintain trace summary in `traces_by_company_day`
   - (edges optional) insert into `edges_by_trace` or derive from parent_node_id
2. Define “root node” logic for trace summary (first node)

**Deliverable:** events are stored in Cassandra and visible via CQL.

## Stage 5 — Query API (Day 5)

1. Implement **logs-query**:
   - `GET /query/v1/traces` lists trace summaries
   - `GET /query/v1/traces/{traceId}` returns nodes+edges
2. Implement “graph shaping”:
   - Convert Cassandra nodes to React Flow nodes `{id, position?, data}`
   - Compute edges from `parent_node_id` if `edges_by_trace` not used
   - Server can return nodes without positions (frontend auto-layout in MVP)

**Deliverable:** query returns a graph JSON for a trace.

## Stage 6 — BFF Orchestration (Day 6)

1. logs-bff calls logs-query internally:
   - `/api/traces` proxies query list (enforces company scoping)
   - `/api/traces/{traceId}` returns `{nodes, edges, annotations}`
2. Annotation CRUD in Postgres (trace-bound)

**Deliverable:** UI can get list + graph with auth enforced.

## Stage 7 — Frontend UI (Day 7–8)

1. **Login centered** page (nice error UX)
2. Private layout with **sidebar + motion**:
   - menu: “Traces”, “Settings” (Settings can be placeholder)
3. Traces list page:
   - filters (date range, status, service)
   - table, click to open trace
4. Trace viewer:
   - React Flow graph render
   - node inspector panel
   - highlight errors & slow nodes
   - add annotation node + persist

**Deliverable:** end-to-end: ingest → queue → cassandra → query → UI graph.

## Stage 8 — Hardening & DX (Day 9)

1. Contract validation: JSON schema for LogEvent
2. Load test ingestion (k6) locally
3. Improve caching for apiKey validation
4. Add pagination to traces list
5. Add basic audit logs (optional)

---

## 12) MVP Open Questions (Decide later, don’t block MVP)

- Node positioning strategy:
  - MVP: auto-layout on frontend (dagre) or simple vertical stacking
  - Later: persist positions per user/company
- How to group nodes visually:
  - per service (swimlanes)
  - per operation type
- Retention policy:
  - TTL in Cassandra for nodes? (Phase 2)

---

## 13) Quick “Hello Trace” Example (Manual Test)

1. Send a small trace with 5 nodes:
   - ROOT HTTP
   - DB SELECT
   - IF decision
   - EXTERNAL_API call
   - END

2. Verify:
   - Traces list shows it
   - Trace viewer renders correct edges
   - Error highlight works

---

## 14) What should do first (Actionable checklist)

1. Create the repo structure + docker-compose + Traefik routing.
2. Implement logs-bff auth + seed.
3. Implement logs-ingest publish to RabbitMQ.
4. Implement logs-worker persist to Cassandra.
5. Implement logs-query endpoints.
6. Implement web UI with login + sidebar + traces + trace viewer.

---

## 15) Naming & UX micro details (as requirements)

- Sidebar items:
  - **Traces**
  - **Settings**
- Submenus (later):
  - Settings → Users, API Keys, Retention
- Use Framer Motion for:
  - Sidebar expand/collapse
  - Menu item hover slide/fade
  - Submenu accordion animation
- Simple typography, strong spacing, clean cards

---

## Done ✅

This file is the **implementation roadmap** for the MVP.
