<!--markdownlint-disable-->

# Distributed Task Queue

> **Reliable asynchronous job processing for distributed systems, built in Go.**

Modern applications rarely execute everything synchronously. Emails, image processing, data pipelines, AI inference, notifications, reports, webhooks, and background jobs often need to happen outside the user's request.

The problem is that once these jobs leave the request-response cycle, **reliability becomes a distributed-systems problem.**

What happens when a worker crashes halfway through a job?

What happens when thousands of jobs arrive simultaneously?

What happens when a job fails temporarily?

What happens when a worker disappears after claiming a task but before acknowledging it?

What happens when the system needs to retry a task without overwhelming the infrastructure?

This project is an attempt to solve those problems by building a **production-oriented distributed task queue in Go**.

---

## The Problem

A simple application might look like:

```text
Client
  │
  ▼
Application
  │
  ▼
Execute Job
```

But this becomes fragile when jobs are slow, expensive, unreliable, or numerous.

A production system needs to decouple job submission from job execution:

```text
                  ┌─────────────┐
                  │ Application │
                  └──────┬──────┘
                         │
                      submit
                         │
                         ▼
                  ┌─────────────┐
                  │ Task Queue  │
                  └──────┬──────┘
                         │
               ┌─────────┼─────────┐
               ▼         ▼         ▼
            Worker 1  Worker 2  Worker N
```

But simply putting jobs into a queue isn't enough.

A distributed task system needs to answer difficult questions around:

* **Reliability** — How do we prevent jobs from disappearing?
* **Concurrency** — How do multiple workers safely process tasks?
* **Failure recovery** — What happens when a worker dies?
* **Retries** — How do we recover from temporary failures?
* **Duplicates** — What happens if a task is executed more than once?
* **Scheduling** — How do we handle priority and delayed jobs?
* **Observability** — How do we know what the system is doing?
* **Operations** — How do we run and debug the system in production?

---

# The Solution

This project provides a distributed task execution system where applications can submit jobs without needing to execute them immediately.

```text
                         ┌─────────────────┐
                         │    Producers    │
                         │                 │
                         │ API / CLI / App │
                         └────────┬────────┘
                                  │
                                  ▼
                         ┌─────────────────┐
                         │   Dispatcher    │
                         │                 │
                         │ Scheduling      │
                         │ Task routing    │
                         │ Retry decisions │
                         └────────┬────────┘
                                  │
                                  ▼
                         ┌─────────────────┐
                         │      Redis      │
                         │                 │
                         │ Task state      │
                         │ Queues          │
                         │ Leases          │
                         │ Scheduling      │
                         └────────┬────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    ▼             ▼             ▼
              ┌──────────┐ ┌──────────┐ ┌──────────┐
              │ Worker 1 │ │ Worker 2 │ │ Worker N │
              └──────────┘ └──────────┘ └──────────┘
                    │             │             │
                    └─────────────┼─────────────┘
                                  ▼
                         ┌─────────────────┐
                         │  Observability  │
                         │                 │
                         │ Logs + Metrics  │
                         └─────────────────┘
```

The system is designed around **at-least-once task delivery**.

This means a task should not silently disappear because a worker crashed. If the system cannot confirm successful completion, the task can become available for another attempt.

That introduces the possibility of duplicate execution, so the system is designed with **idempotency and failure recovery** as first-class concerns.

---

# Core Capabilities

### Asynchronous task execution

Applications submit work and continue without waiting for the job to finish.

```text
Application
     │
     │ submit task
     ▼
   Queue
     │
     │ later
     ▼
  Worker
```

### Distributed workers

Multiple worker processes can consume tasks concurrently.

```text
                 Queue
              /    |    \
             ▼     ▼     ▼
          Worker Worker Worker
             1      2      3
```

Workers can run on separate processes, containers, or machines.

### Concurrent execution

Each worker can maintain a pool of concurrent task executions while respecting configured limits.

### Retries

Temporary failures don't necessarily mean permanent failure.

Tasks can be retried using configurable retry policies and exponential backoff.

```text
Attempt 1
   │
 failure
   ▼
 1s backoff
   │
Attempt 2
   │
 failure
   ▼
 2s backoff
   │
Attempt 3
   │
 success
   ▼
 COMPLETED
```

### Worker failure recovery

Workers periodically report their health.

Tasks claimed by workers can have leases. If a worker disappears and its lease expires, unfinished work can become available again.

### Priority queues

Important work can be processed before lower-priority tasks.

### Delayed jobs

Tasks can be scheduled for execution in the future.

### Dead-letter queues

Tasks that repeatedly fail can be moved to a dead-letter queue for inspection and manual intervention.

### Observability

The system exposes operational information such as:

* Queue depth
* Task throughput
* Task latency
* Successful tasks
* Failed tasks
* Retry counts
* Worker health
* Active workers
* Task execution duration

---

# Architecture

The architecture is intentionally divided into independent responsibilities.

```text
                    ┌──────────────────┐
                    │      Client      │
                    │                  │
                    │ CLI / API / App  │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │    Dispatcher    │
                    └────────┬─────────┘
                             │
                             ▼
                    ┌──────────────────┐
                    │      Redis       │
                    └────────┬─────────┘
                             │
                ┌────────────┼────────────┐
                ▼            ▼            ▼
           ┌────────┐   ┌────────┐   ┌────────┐
           │Worker 1│   │Worker 2│   │Worker N│
           └────────┘   └────────┘   └────────┘
                │            │            │
                └────────────┼────────────┘
                             ▼
                    ┌──────────────────┐
                    │  Metrics / Logs  │
                    └──────────────────┘
```

### Components

```text
cmd/
└── queue/
    └── main.go

internal/
├── task/
│   └── task.go
│
├── queue/
│   └── queue.go
│
├── worker/
│   └── worker.go
│
├── scheduler/
│   └── scheduler.go
│
├── retry/
│   └── retry.go
│
└── ...
```

The system will evolve incrementally. The initial implementation will begin with an in-memory queue and local worker pool before introducing distributed persistence and coordination.

---

# Delivery Semantics

The queue targets **at-least-once delivery**.

A simplified task lifecycle is:

```text
                 ┌─────────┐
                 │ PENDING │
                 └────┬────┘
                      │
                    claim
                      │
                      ▼
                 ┌─────────┐
                 │ RUNNING │
                 └────┬────┘
                      │
              ┌───────┴────────┐
              │                │
           success           failure
              │                │
              ▼                ▼
         ┌───────────┐    ┌──────────┐
         │ COMPLETED │    │ RETRYING │
         └───────────┘    └─────┬────┘
                                │
                         max attempts
                                │
                                ▼
                           ┌─────────┐
                           │   DLQ   │
                           └─────────┘
```

A worker failure should not permanently lose a task.

However, because a worker can fail after executing a task but before acknowledging it, **duplicate execution is possible**.

Therefore consumers should design task handlers to be idempotent where necessary.

---

# Technology

| Component            | Technology                 |
| -------------------- | -------------------------- |
| Language             | Go                         |
| Queue / Coordination | Redis                      |
| Worker communication | gRPC                       |
| Logging              | Zap                        |
| Metrics              | Prometheus                 |
| Containerization     | Docker                     |
| CLI                  | Cobra                      |
| Testing              | Go testing + race detector |
| Local development    | Docker Compose             |

The initial version will prioritize Redis. NATS may be evaluated later as an alternative backend rather than being implemented simultaneously.

---

# Project Structure

```text
distributed-task-queue/
│
├── cmd/
│   └── queue/
│       └── main.go
│
├── internal/
│   ├── task/
│   │   └── task.go
│   │
│   ├── queue/
│   │   └── queue.go
│   │
│   ├── worker/
│   │   └── worker.go
│   │
│   ├── scheduler/
│   │   └── scheduler.go
│   │
│   ├── retry/
│   │   └── retry.go
│   │
│   └── ...
│
├── api/
│   └── proto/
│
├── demo/
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── web/
│   └── terminal/
│
├── tests/
│   ├── integration/
│   └── load/
│
├── docs/
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── .gitignore
└── README.md
```

---

# Development Roadmap

The system will be built incrementally rather than attempting the entire distributed architecture at once.

### Phase 1 — Core Engine

* Task model
* Task state machine
* In-memory queue
* Worker pool
* Task handlers
* Graceful shutdown
* Unit tests
* Race detection

### Phase 2 — Distributed Queue

* Redis integration
* Persistent task state
* Multiple workers
* Task claiming
* Acknowledgements

### Phase 3 — Reliability

* Worker heartbeats
* Task leases
* Failure detection
* Task recovery
* Retry policies
* Exponential backoff
* Dead-letter queues
* Idempotency support

### Phase 4 — Scheduling

* Priority queues
* Delayed jobs
* Scheduled tasks

### Phase 5 — Communication

* gRPC worker protocol
* Dispatcher/worker coordination
* Worker registration

### Phase 6 — Observability

* Structured logging
* Prometheus metrics
* Health checks
* Operational statistics
* Performance benchmarks

### Phase 7 — Containerized Deployment

The complete development environment should be reproducible with:

```bash
docker compose up
```

The environment will eventually include:

```text
Redis
Dispatcher
Workers
Prometheus
Demo services
```

### Phase 8 — Interactive Demo

The project will include a browser-based interactive terminal where visitors can experiment with the queue without installing Go or Redis.

Example:

```text
$ queue workers

WORKER       STATUS     ACTIVE
worker-01    ONLINE       3
worker-02    ONLINE       2
worker-03    ONLINE       1

$ queue submit --type=email --priority=10

Task submitted: 8f31c2

$ queue task 8f31c2

STATUS: COMPLETED
WORKER: worker-02
ATTEMPTS: 1
```

The demo will run inside an isolated containerized environment and expose only approved queue commands.

---

# Running Locally

### Prerequisites

* Go 1.22+
* Docker
* Git

### Clone

```bash
git clone https://github.com/yourusername/distributed-task-queue.git
cd distributed-task-queue
```

### Run tests

```bash
go test ./...
```

Run with the race detector:

```bash
go test -race ./...
```

### Run the development environment

```bash
docker compose up
```

### Build

```bash
go build ./cmd/queue
```

---

# Why Build This?

This project is not intended to be another CRUD application.

It is an exploration of the engineering problems that appear when software has to operate across **multiple concurrent processes, machines, failures, and unreliable networks.**

The goal is to understand and implement concepts such as:

* Distributed coordination
* Concurrency
* Task scheduling
* Failure detection
* At-least-once delivery
* Idempotency
* Retry strategies
* Worker leases
* Backpressure
* Observability
* Graceful degradation

The system is being built from the ground up in Go with an emphasis on **correctness, failure handling, and operational visibility.**

---

# Contributing

Contributions are welcome.

Before submitting a pull request:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Changes should include appropriate tests and clearly describe their impact on task execution, concurrency, reliability, or operational behavior.

---

# License

MIT License
