# Distributed Task Queue

A scalable distributed task queue system built in Go. This starter project is designed to support asynchronous job processing across multiple worker nodes, with a focus on reliability, concurrency, and operational observability.

## Overview

This project aims to provide a flexible foundation for a distributed worker ecosystem. Core capabilities include:

- Task queue management
- Worker pools
- Concurrent job execution
- Retry mechanisms
- Priority queues and delayed jobs
- Metrics and monitoring
- CLI management tools

## Architecture

The initial design splits the project into clear components:

- `cmd/queue`: application entry point and CLI interface
- `internal/task`: task model and utility types
- `internal/queue`: queue management APIs and enqueuing logic
- `internal/worker`: worker lifecycle, execution, and heartbeat patterns

## Recommended Stack

- Go
- Redis or NATS for queue persistence and coordination
- gRPC for worker and scheduler communication
- Zap for structured logging
- Prometheus for metrics
- Docker for containerization
- Cobra for CLI support (future enhancement)

## Getting Started

### Prerequisites

- Go 1.22+ installed
- Git installed
- Redis or NATS available for production integration

### Clone the repository

```bash
git clone https://github.com/yourusername/distributed-task-queue.git
cd distributed-task-queue
```

### Initialize module

```bash
go mod tidy
```

### Build

```bash
go build ./cmd/queue
```

### Run

```bash
go run ./cmd/queue --role=dispatcher
```

Switch to a worker process in another terminal:

```bash
go run ./cmd/queue --role=worker
```

## Project Structure

```text
.
├── cmd
│   └── queue
│       └── main.go
├── internal
│   ├── queue
│   │   └── queue.go
│   ├── task
│   │   └── task.go
│   └── worker
│       └── worker.go
├── go.mod
├── .gitignore
└── README.md
```

## Roadmap

Future work should include:

- Redis/NATS-backed queue persistence
- gRPC worker/shard coordination
- Dead-letter queues
- Exponential retry backoff
- Graceful shutdown and worker heartbeat monitoring
- Priority and delayed job scheduling
- Metrics instrumentation with Prometheus
- Docker Compose development setup

## Contribution

1. Fork the repository
2. Create a feature branch
3. Submit a pull request with a clear description

## License

This project is released under the MIT License.
