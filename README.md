# EdgeCore Enterprise API Gateway Platform

**EdgeCore** is a production-grade, enterprise API Gateway platform built with **Go** (Data Plane & Control Plane) and **Next.js 16** (Admin Dashboard).

---

## System Architecture

```
                Internet
                    │
                    ▼
          ┌──────────────────────┐
          │   Load Balancer       │
          └──────────┬────────────┘
                     │
        ┌────────────┴────────────┐
        │                         │
 ┌──────────────┐          ┌──────────────┐
 │ Gateway #1   │          │ Gateway #2   │
 │ Data Plane   │          │ Data Plane   │
 └──────┬───────┘          └──────┬───────┘
        │                         │
        └────────────┬────────────┘
                     │
             Redis Config Sync (Pub/Sub)
                     │
      ┌──────────────┴──────────────┐
      │     Go Control Plane        │
      └──────────────┬──────────────┘
                     │
              PostgreSQL DB
                     │
             Next.js Dashboard
```

---

## Architecture Breakdown

1. **Data Plane (`apps/gateway`)**:
   - High-throughput reverse proxy based on `net/http/httputil`.
   - Dynamic Ingress Router: Exact, Prefix, Wildcard, Regex, Header, Host, Query Param, and Priority routing.
   - Load Balancer Algorithms: Round Robin, Least Connections, Random, Weighted Round Robin, IP Hash, and Consistent Hashing (Ketama ring).
   - Traffic Management & Plugins: Redis Distributed Rate Limiter, Circuit Breaker, Retries, JWT Validation, API Key Auth, Security Headers/CORS, Request/Response Transformers, Redis Response Caching, and Compression.
   - Real-Time Protocols: WebSocket TCP Hijacking pass-through & Server-Sent Events (SSE).
   - Observability: Prometheus metrics exporter (`/metrics` on port 9090) and OpenTelemetry W3C tracing context propagation.

2. **Control Plane (`apps/control-plane`)**:
   - Management REST API built on Go Fiber and Casbin RBAC framework.
   - Database layer powered by GORM and PostgreSQL.
   - Dynamic Config Event Dispatcher: Pushes route, plugin, and key updates instantly across Redis Pub/Sub (`edgecore:config:events`) to all Data Plane instances for zero-downtime reconfiguration.

3. **Admin Dashboard (`apps/dashboard`)**:
   - Built on Next.js 16 (App Router), React 19, Tailwind CSS, Shadcn UI styling, Recharts, and TanStack Query.
   - Real-time traffic analytics, interactive route matching builder, visual plugin hub toggle cards, and node status monitoring.

---

## Getting Started

### Prerequisites
- Go 1.24+
- Node.js 20+
- Docker & Docker Compose

### Quick Start with Docker Compose
```bash
docker-compose -f deployments/docker-compose.yml up --build
```

Services will start at:
- **Data Plane Gateway Node 1**: `http://localhost:8080`
- **Data Plane Gateway Node 2**: `http://localhost:8082`
- **Prometheus Metrics**: `http://localhost:9090/metrics`
- **Control Plane API**: `http://localhost:8081`
- **Next.js Admin Dashboard**: `http://localhost:3000`

---

## Running Tests
Run unit tests across all Go packages:
```bash
go test -v ./...
```
