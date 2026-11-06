# Distributed Network Health Checker (DNHC)

A high-performance, concurrent network health probe and telemetry collection system written in **Go**, orchestrated with **Docker Compose**, and designed for hybrid emulated environments (**Alpine Linux nodes** and **Nokia SR Linux Containerlab topologies**).

---


## 🏗️ System Architecture

```
                                    +------------------------------------------+
                                    |         Containerlab Topology (clab)     |
                                    |   +----------------------------------+   |
                                    |   | clab-srl01-srl1 (Nokia SR Linux) |   |
                                    |   +-----------------+----------------+   |
                                    |                     | e1-1               |
                                    |   +-----------------+----------------+   |
                                    |   | clab-srl01-srl2 (Nokia SR Linux) |   |
                                    |   +----------------------------------+   |
                                    +---------------------+--------------------+
                                                          |
                                           Docker Network: `clab`
                                                          |
             +--------------------------------------------+
             |
+------------v-------------------------------------------------------------------------+
| Go Health Checker Container (`health-checker-api`)                                   |
|                                                                                      |
|  +-------------------+       +-----------------------+      +---------------------+  |
|  | Discovery Engine  | ----> | Task Channel (pings)  | ---> | Worker Pool (1..N)  |  |
|  +-------------------+       +-----------------------+      +----------+----------+  |
|                                                                        |             |
|                                                                  Probe Execution     |
|                                                                (context.WithTimeout) |
|                                                                        |             |
|  +-------------------+       +-----------------------+                 v             |
|  | PostgreSQL Pool   | <---  | Log Channel (logChan) | <---------------+             |
|  | (`pgxpool.Pool`)  |       +-----------+-----------+                               |
|  +---------+---------+                   |                                           |
|            |                             v                                           |
|            |                 +-----------------------+                               |
|            |                 | Exclusive Log Writer  | ---> Host Volume: `./logs/`   |
|            |                 | (Logger Goroutine)    | ---> Host Volume: `./OUTPUT/` |
|            |                 +-----------------------+                               |
+------------|-------------------------------------------------------------------------+
             |
             | SQL (Port 5432 / health-net)
             v
+------------------------------------+
| PostgreSQL 16 (`health-checker-db`)|
| Volume: `pgdata`                   |
| Tables: `nodes`, `ping_metrics`    |
+------------------------------------+
```

---

## 🧠 Core Engineering Concepts: Why & How

### 1. Golang Concurrency & Design Patterns

#### **The Worker Pool Pattern**
* **Why:** In network probing, spawning an unbounded number of goroutines (`go ping(node)`) for hundreds of nodes causes OS thread thrashing, socket exhaustion, and network queue overflows. A bounded worker pool limits concurrent operations to an optimal, fixed number of workers.
* **How:** 
  1. A buffered task channel `pings := make(chan PingTarget, len(nodes))` is populated with targets.
  2. The producer calls `close(pings)` to signal no further work will arrive.
  3. A fixed pool of 5 workers (`for w := 1; w <= 5; w++`) consume from `range pings`. When the channel drains and closes, each worker loop terminates naturally.
  4. A `sync.WaitGroup` tracks worker lifecycles, ensuring the parent function waits (`wg.Wait()`) until all workers have exited before teardown.

#### **The Exclusive Writer Pattern (Channel-based Centralized Logging)**
* **Why:** When multiple goroutines write to the same log file or stdout concurrently, output interleaves and corrupts lines without file locks or mutexes. However, mutex locks introduce lock contention and serialize workers.
* **How:**
  1. Workers push raw string logs into an asynchronous channel `logChan := make(chan string, 100)`.
  2. Exactly **one** designated goroutine (`LoggerGoroutine`) acts as the exclusive writer for the file.
  3. When all workers finish (`wg.Wait()`), the dispatcher closes `logChan`, causing the logger loop to flush remaining items and gracefully exit (`logWg.Wait()`).

#### **Deadlock Prevention & Timeout Handling (`context.WithTimeout`)**
* **Why:** Network pings against unreachable or degraded targets can hang indefinitely if packets drop silently. A hanging worker blocks a slot in the pool forever.
* **How:**
  1. Each probe initializes a context with a strict deadline:
     ```go
     ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
     defer cancel()
     ```
  2. Commands execute via `exec.CommandContext(ctx, "ping", ...)`. If the timeout fires before completion, Go's runtime terminates the child process and frees worker resources immediately.

---

### 2. Docker & Multi-Stage Image Builds

#### **Multi-Stage Build Pattern**
* **Why:** A standard Go build environment requires the Go toolchain, git, and compilers (~800MB–1.2GB). Shipping that image into production increases container attack surfaces, vulnerability scan noise, and deployment transfer latency.
* **How:**
  * **Stage 1 (`builder`)**: Uses `golang:1.26-alpine3.24` (or `golang:1.23-alpine`) to compile a statically linked Linux binary (`CGO_ENABLED=0 GOOS=linux go build -o health-checker`).
  * **Stage 2 (`runtime`)**: Uses a pristine `alpine:latest` image (~7MB), copies **only** the compiled binary (`COPY --from=builder /app/health-checker .`), installs runtime network utilities (`iputils-ping`), and creates `/app/logs` and `/app/output`.
  * **Result:** Production container size drops to under **20MB**.

#### **Debug Mode Execution via Compose**
* **Why:** In early development and network probing integration, you often need the container to stay alive so you can inspect mounted volumes, execute ad-hoc pings, or run tools manually without the container exiting when the binary finishes.
* **How:**
  * In `docker-compose.yml`, `command: ["sleep", "infinity"]` overrides `CMD ["./health-checker"]`.
  * This keeps the container active in the background, allowing you to trigger runs using `docker exec` or `make run-health-check`.

---

### 3. Docker Compose Orchestration & Networking

#### **Modular Multi-File Composition (`include:`)**
* **Why:** Monolithic `docker-compose.yml` files become unmaintainable when orchestrating application services alongside independent test topologies.
* **How:** Docker Compose’s top-level `include:` directive cleanly imports external Compose files:
  ```yaml
  include:
    - docker-network-compose.yml
  ```
  This imports `alpine0` through `alpine4` directly into the main Compose stack while keeping their service definitions segregated.

#### **Isolated Bridge Networks (`health-net`)**
* **Why:** Containers on Docker's default bridge network cannot resolve each other by container name (DNS resolution is disabled). Custom user-defined bridge networks enable automatic embedded DNS.
* **How:** Both `db`, `api`, and the test Alpine nodes declare membership in `health-net`. The Go app can resolve PostgreSQL by the hostname `db` and test nodes by `alpine0`, `alpine1`, etc., without hardcoded IP addresses.

#### **Granular Service Lifecycle (`make restart-app`)**
* **Why:** Running `docker compose down && docker compose up` tears down database state, disconnects test nodes, and drops named volumes.
* **How:** The command `docker compose -f ./compose/docker-compose.yml up -d --build api` scopes execution **only** to the Go service. Docker recompiles the Go binary, swaps the container, and leaves the database and test nodes completely uninterrupted.

---

### 4. PostgreSQL Architecture & Telemetry Persistence

#### **Entrypoint Initialization Hook (`/docker-entrypoint-initdb.d/`)**
* **Why:** Manual table creation and schema provisioning creates friction for team members and CI/CD runners.
* **How:**
  1. SQL migrations are stored in `./compose/initdb/01-init.sql`.
  2. Docker mounts this directory into `/docker-entrypoint-initdb.d:ro`.
  3. The official PostgreSQL image automatically executes all `.sql` files in alphabetical order on the very first container initialization, provisioning the `nodes` and `ping_metrics` tables without manual intervention.

#### **Database Connection Pooling (`pgxpool.Pool`)**
* **Why:** Standard individual connections (`pgx.Conn`) are **not thread-safe**. If 5 workers execute queries concurrently on a single connection, the driver panics or fails with socket sync errors.
* **How:** 
  1. `internal/db/db.go` initializes a `pgxpool.Pool` via `pgxpool.New(ctx, dbURL)`.
  2. The pool manages an internal queue of open connections. When a worker writes a metric, it checks out an available connection, runs the query, and returns it to the pool automatically.

---

## 🚀 Section 1: Alpine Nodes Pinging & Discovery

This section details the local simulation environment where lightweight Alpine Linux containers represent network endpoints.

### Workflow & Execution Steps

#### Step 1: Initialize Database & Compose Environment
Start PostgreSQL and the Alpine test nodes in detached mode:
```bash
cd /home/clab/Distributed-Network-Health-Checker

# Build the Go application and start all services
make build
make up

# Verify all 7 containers are running (db, api, alpine0..4)
docker compose -f ./compose/docker-compose.yml ps
```

#### Step 2: Verify PostgreSQL Initialization & Tables
Check that PostgreSQL executed `01-init.sql` and seeded the schema:
```bash
docker exec -it health-checker-db psql -U admin -d mydatabase -c "\dt"
docker exec -it health-checker-db psql -U admin -d mydatabase -c "SELECT * FROM nodes;"
```

#### Step 3: Test Cross-Container Reachability from Go App
Confirm DNS resolution and ICMP ping from `health-checker-api` across `health-net`:
```bash
# Ping test nodes by container name
docker exec -it health-checker-api ping -c 3 alpine0
docker exec -it health-checker-api ping -c 3 alpine1

# Check TCP reachability to database
docker exec -it health-checker-api nc -zv db 5432
```

#### Step 4: Run the Go Health Checker Binary
Execute the health check engine inside the running container:
```bash
make run-health-check
# Or directly:
docker compose -f ./compose/docker-compose.yml exec api ./health-checker
```

#### Step 5: Inspect Generated Artifacts & Logs
Verify that the host-mounted directories received the probe telemetry:
```bash
# Check raw ping output written by the exclusive logger goroutine
cat ./OUTPUT/ping_output.txt

# Check application event and discovery logs
cat ./logs/App.log
cat ./logs/Discovery.log
cat ./logs/Ping.log
```

#### Step 6: Code Iteration Loop
When modifying Go code in `main.go` or `internal/`:
```bash
# Rebuild and restart ONLY the Go container in seconds (DB remains active)
make restart-app

# Re-run health check
make run-health-check
```

---

## 🌐 Section 2: Containerlab Nodes (Nokia SR Linux)

*This section documents the integration with Containerlab for production-grade Network Operating System (NOS) emulation. (Detailed commands and topologies will be expanded as work progresses).*

### Overview & Architecture
* **Containerlab (`clab`)**: An open-source orchestrator that deploys network topologies using containerized Network Operating Systems (Nokia SR Linux, Arista cEOS, Cisco XRd).
* **Topology Definition (`/home/clab/topology/example.clab.yml`)**:
  * Defines two Nokia SR Linux nodes: `srl1` (`clab-srl01-srl1`) and `srl2` (`clab-srl01-srl2`).
  * Connects them via a point-to-point emulated Ethernet link (`e1-1` <-> `e1-1`).
  * Management interfaces attach to the Docker bridge network named `clab` (`172.20.20.0/24`).

### Future Network Bridge Integration
1. **Deploying the Topology**:
   ```bash
   sudo clab deploy -t /home/clab/topology/example.clab.yml
   sudo clab inspect -t /home/clab/topology/example.clab.yml
   ```
2. **Attaching Health Checker to Containerlab**:
   * Connect `health-checker-api` to the `clab` management network:
     ```bash
     docker network connect clab health-checker-api
     ```
   * Or uncomment the `clab` external network in `compose/docker-compose.yml`.
3. **Evolution Beyond ICMP**:
   * While V1 utilizes ICMP ping, the Containerlab integration will expand into:
     * **SSH Telemetry**: Executing structured command parsing.
     * **gNMI / gRPC Probes**: Streaming telemetry using protobuf over HTTP/2 for real-time interface and BGP peer health metrics.

---

## 🛠️ Project Structure & File Index

```text
Distributed-Network-Health-Checker/
├── cmd/
│   ├── composegen/main.go       # CLI tool to dynamically generate docker-network-compose files
│   └── health-checker/main.go   # Main orchestrator (Logger -> Discovery -> Workers -> Output)
├── compose/
│   ├── Dockerfile               # Multi-stage Docker build recipe (Go builder -> Alpine runtime)
│   ├── Dockerfile.debug         # Debug Dockerfile with mounted volumes
│   ├── docker-compose.yml       # Primary orchestration (PostgreSQL, Go API, includes network nodes)
│   ├── docker-network-compose.yml # 5 Alpine Linux test nodes on health-net
│   └── initdb/
│       └── 01-init.sql          # PostgreSQL schema definition and seed data
├── internal/
│   ├── compose/composer.go      # Dynamic YAML generation logic for variable node counts
│   ├── db/
│   │   ├── db.go                # Thread-safe pgxpool connection manager & health checks
│   │   └── queries.go           # CRUD query operations
│   ├── discovery/discovery.go   # Node discovery engine
│   └── ping/
│       ├── ping.go              # Worker pool dispatcher (WaitGroups, Channels)
│       ├── ping_external.go     # External target probe orchestrator
│       └── worker.go            # Worker pool consumer implementation
├── utils/
│   ├── logger/
│   │   ├── logger.go            # Custom MultiWriter logger setup
│   │   └── logger_goroutine.go  # Exclusive writer goroutine preventing file contention
│   └── ping_helper/
│       └── ping_helper.go       # System ping wrapper with context.WithTimeout
├── logs/                        # Host bind mount for application logs
├── OUTPUT/                      # Host bind mount for probe result outputs
├── Makefile                     # Developer workflow automation
├── PLAN.md                      # Milestone execution plan
└── README.md                    # System architecture and interview report
```

---
