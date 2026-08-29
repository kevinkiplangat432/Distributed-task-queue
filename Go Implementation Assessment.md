<!--markdownlint-disable-->
# Distributed Task Queue — Go Implementation Assessment

## Objective

Build a small **in-memory distributed task queue prototype** in Go.

The system should demonstrate your understanding of:

- Go structs and methods
- Interfaces/functions as types
- Goroutines
- Channels
- Mutexes and thread safety
- Context cancellation
- Worker pools
- Graceful shutdown
- Task lifecycle/state management
- FIFO queues
- Error handling
- Synchronization

This is **Phase 1**.

Do not use Redis, databases, HTTP servers, or external queue systems yet.

The queue is entirely in memory.

---

# Project Structure

Organize the project into the following packages:

```text
distributed-task-queue/
│
├── cmd/
│   └── worker/
│       └── main.go
│
├── internal/
│   ├── task/
│   │   └── task.go
│   │
│   ├── queue/
│   │   └── queue.go
│   │
│   └── worker/
│       └── worker.go
│
└── go.mod
```

The module name should be:

```text
distributed-task-queue
```

---

# Part 1 — Task Model

Create a `task` package.

A task represents one unit of work submitted to the queue.

## Task fields

Each task must contain:

- A unique string ID
- A string representing the task type
- A byte payload
- A lifecycle status
- The number of attempts made
- The maximum number of attempts allowed
- The time the task was created
- The time the task was last updated

The task should use Go's standard time representation for timestamps.

The ID should be generated using UUIDs.

You may use:

```text
github.com/google/uuid
```

---

# Task Statuses

A task can have the following statuses:

```text
PENDING
RUNNING
COMPLETED
FAILED
RETRYING
DLQ
```

Where:

### PENDING

The task has been created and is waiting to be processed.

### RUNNING

A worker has taken the task and is currently processing it.

### COMPLETED

The handler successfully processed the task.

### FAILED

The handler attempted to process the task but returned an error.

### RETRYING

The task is going to be attempted again.

### DLQ

The task has permanently failed and has been moved to a Dead Letter Queue state.

---

# Task Creation

Provide a constructor for creating a new task.

The constructor should:

- Generate a unique ID
- Set the provided task type
- Store the provided payload
- Set the initial status to `PENDING`
- Set attempts to zero
- Set the default maximum attempts to `3`
- Set the creation timestamp
- Set the update timestamp

The creation and update timestamps should initially represent the same moment.

---

# Task State Machine

Tasks cannot arbitrarily change status.

Implement a state-transition system.

Valid transitions:

```text
PENDING   → RUNNING

RUNNING   → COMPLETED
RUNNING   → FAILED
RUNNING   → RETRYING

RETRYING  → RUNNING
RETRYING  → DLQ

FAILED    → RETRYING
FAILED    → DLQ
```

Terminal states:

```text
COMPLETED
DLQ
```

A task in either terminal state cannot transition anywhere else.

---

# State Transition API

Provide a method that answers:

> "Can this task legally move from its current state to the requested state?"

It should return a boolean.

Then provide another method that:

1. Checks whether the transition is legal.
2. If illegal, does nothing and reports failure.
3. If legal:
   - Changes the status.
   - Updates the task's update timestamp.
   - Reports success.

Do not allow callers to directly bypass the state machine through the transition API.

---

# Part 2 — In-Memory Queue

Create a `queue` package.

The queue must behave as a:

> Thread-safe FIFO queue.

Tasks should come out in the same order they were inserted.

The queue will eventually be replaced by Redis in a later phase, so keep the queue abstraction relatively simple.

---

# Queue Requirements

The queue needs to maintain:

- A collection of tasks
- A mechanism for protecting shared state
- A mechanism for notifying workers that a task has arrived
- A closed/open state

The queue starts open.

---

# Queue Constructor

Provide a constructor that creates and returns a new queue.

The queue should initially:

- Contain no tasks.
- Be open.
- Have its task-notification mechanism initialized.

---

# Push

Provide a method for adding a task to the queue.

Behavior:

### If the queue is open

The task should:

1. Be added to the back of the queue.
2. Preserve FIFO ordering.
3. Notify waiting consumers that work is available.

### If the queue is closed

The task must not be added.

Return a specific queue-closed error.

Define this error as a package-level error so callers can compare against it.

---

# Pop

Provide a method that retrieves the next task.

The method receives a `context`.

Behavior:

### When a task exists

Immediately remove and return the oldest task.

### When the queue is empty

The method should wait until one of these things happens:

1. A task is pushed.
2. The provided context is cancelled.
3. The queue is closed.

### If the context is cancelled

Return the context's error.

### If the queue is closed

Return the queue-closed error.

### FIFO requirement

If tasks are inserted:

```text
A
B
C
```

they must be returned:

```text
A
B
C
```

---

# Concurrency Requirement

Multiple goroutines must be able to call `Push` and `Pop` safely.

There must be no data races when:

- Multiple producers push simultaneously.
- Multiple workers pop simultaneously.
- Producers and consumers operate concurrently.

Protect the queue's shared state appropriately.

Do not rely on callers to perform synchronization.

---

# Queue Length

Provide a method that returns:

> The current number of pending tasks.

It must be safe to call concurrently with `Push` and `Pop`.

---

# Queue Shutdown

Provide a method to close the queue.

Closing the queue should:

- Prevent future pushes.
- Wake consumers that are waiting for work.
- Be safe to call more than once.
- Cause future pops to eventually return the queue-closed error once no work remains / shutdown is observed.

Do not panic if the queue is closed multiple times.

---

# Part 3 — Worker Pool

Create a `worker` package.

The worker pool is responsible for processing tasks from the queue.

A worker pool consists of:

- A reference to the queue
- A task handler
- A configurable number of workers
- Synchronization for tracking worker shutdown

---

# Task Handler

Define a handler function type.

The handler receives:

- A context
- A pointer/reference to the task being processed

It returns an error.

Conceptually:

```text
handler(context, task) → error
```

The handler is responsible for performing the actual work.

The worker pool should not know what a specific task type means.

---

# Worker Pool Constructor

Provide a constructor accepting:

- The queue
- The number of workers
- The task handler

Store these values inside the pool.

---

# Starting the Pool

Provide a method that starts the worker pool.

If the pool size is `3`, exactly three worker goroutines should be started.

The start method should return immediately rather than waiting for the workers to finish.

Each worker should continuously:

1. Wait for a task.
2. Receive a task.
3. Transition it to `RUNNING`.
4. Execute the handler.
5. Update the task according to the result.
6. Continue waiting for more work.

---

# Successful Task

If the handler returns no error:

```text
RUNNING → COMPLETED
```

The task should end in the completed state.

Log that the worker successfully processed the task.

---

# Failed Task

If the handler returns an error:

```text
RUNNING → FAILED
```

The task should enter the failed state.

Log:

- Worker identifier
- Task identifier
- Error

For this first phase, do not implement the actual retry algorithm yet.

The retry-related states exist because they will be used in a later phase.

---

# Worker Shutdown

Workers should stop when their context is cancelled or when the queue is closed.

A worker must not continue indefinitely after shutdown has been requested.

The pool must provide a method that blocks until **all workers have exited**.

Use Go synchronization primitives rather than polling.

---

# Context Behavior

The worker pool receives a parent context when started.

That context should be passed to:

- The queue's blocking pop operation.
- The task handler.

If the context is cancelled while workers are waiting for tasks, they should exit.

If the handler respects the context, cancellation should propagate into task processing.

---

# Part 4 — Demo Application

Create a `main` package under:

```text
cmd/worker/
```

The application should demonstrate the complete system.

---

# Application Startup

When the program starts:

1. Create a root context that responds to operating-system termination signals.
2. Create a new in-memory queue.
3. Create a task handler.
4. Create a worker pool.
5. Start the worker pool.

Use **three workers**.

---

# Demo Handler

The demo handler should simulate work.

Each task should take approximately:

```text
100 milliseconds
```

to process.

For now, the handler should successfully process every task.

The task type does not need special dispatch logic yet.

---

# Demo Tasks

Before waiting for shutdown, submit:

```text
5 tasks
```

to the queue.

Every task should:

```text
type = "demo"
payload = "payload"
```

---

# Shutdown

The application should remain running until it receives an interrupt or termination signal.

When shutdown occurs:

1. Log that shutdown was requested.
2. Close the queue.
3. Wait for all workers to exit.
4. Log that all workers exited cleanly.
5. Exit.

The shutdown sequence should not abruptly kill the worker goroutines.

---

# Important Design Constraint

The queue and workers must support this scenario:

```text
main
 │
 ├── creates queue
 │
 ├── starts worker pool
 │      ├── worker 1
 │      ├── worker 2
 │      └── worker 3
 │
 ├── pushes tasks
 │
 └── waits for shutdown
```

Workers should continuously consume work until shutdown.

---

# Expected Concurrency

With three workers and five tasks, multiple tasks should be capable of executing concurrently.

You should therefore expect behavior roughly equivalent to:

```text
Worker 1 → Task A
Worker 2 → Task B
Worker 3 → Task C

Worker 1 → Task D
Worker 2 → Task E
```

The exact worker/task assignment does not matter.

FIFO applies to **task retrieval from the queue**, not which worker receives which task.

---

# Error Handling Requirements

Do not silently ignore meaningful errors inside the queue or worker implementation.

The following conditions must be represented explicitly:

```text
queue closed
context cancelled
invalid task state transition
handler failure
```

The caller should be able to distinguish queue closure from context cancellation.

---

# Logging

Workers should log enough information to understand what is happening.

At minimum, successful processing should identify:

```text
worker ID
task ID
```

Failures should identify:

```text
worker ID
task ID
error
```

Shutdown should also be visible in application logs.

---

# Acceptance Criteria

Your implementation is considered complete when all of the following are true.

## Task

- [ ] Tasks receive unique IDs.
- [ ] New tasks begin in `PENDING`.
- [ ] New tasks have zero attempts.
- [ ] New tasks have a maximum of three attempts.
- [ ] Creation and update timestamps are initialized.
- [ ] Valid state transitions are enforced.
- [ ] Invalid transitions are rejected.
- [ ] Successful transitions update the timestamp.
- [ ] `COMPLETED` is terminal.
- [ ] `DLQ` is terminal.

## Queue

- [ ] Queue is FIFO.
- [ ] Queue is safe for concurrent access.
- [ ] Tasks can be pushed.
- [ ] Tasks can be popped.
- [ ] Empty queue causes consumers to wait.
- [ ] Context cancellation interrupts waiting.
- [ ] Closing the queue wakes waiting consumers.
- [ ] Push after close fails.
- [ ] Closing twice does not panic.
- [ ] Queue length is available.

## Workers

- [ ] Configurable number of workers.
- [ ] Workers run concurrently.
- [ ] Workers consume tasks from the queue.
- [ ] Tasks transition to `RUNNING`.
- [ ] Successful handlers produce `COMPLETED`.
- [ ] Failed handlers produce `FAILED`.
- [ ] Worker failures are logged.
- [ ] Workers stop on cancellation.
- [ ] Workers stop when the queue closes.
- [ ] The application can wait for every worker to exit.

## Application

- [ ] OS shutdown signals are handled.
- [ ] Three workers are started.
- [ ] Five demo tasks are submitted.
- [ ] Tasks take approximately 100ms to process.
- [ ] Shutdown is graceful.
- [ ] The application waits for workers before exiting.

---

# Testing Challenges

After implementing the system, verify the following manually or with tests.

### Challenge 1 — FIFO

Push:

```text
task-1
task-2
task-3
```

Confirm they are popped in that order.

### Challenge 2 — Concurrent producers

Start multiple goroutines that push tasks simultaneously.

Confirm:

- No race conditions.
- No lost tasks.
- No corrupted queue state.

### Challenge 3 — Concurrent consumers

Start multiple workers consuming simultaneously.

Confirm:

- Each task is processed once by the in-memory queue.
- No task disappears.
- No task is returned to two consumers.

### Challenge 4 — Empty queue

Start a worker when there are no tasks.

Confirm it blocks rather than spinning in a CPU-heavy loop.

### Challenge 5 — Cancellation

Start a worker with an active context and no tasks.

Cancel the context.

Confirm the worker exits.

### Challenge 6 — Queue closure

Start a worker waiting on an empty queue.

Close the queue.

Confirm the worker wakes up and exits.

### Challenge 7 — Invalid state

Attempt illegal transitions such as:

```text
COMPLETED → RUNNING
DLQ → RUNNING
PENDING → COMPLETED
```

They must be rejected.

### Challenge 8 — Handler failure

Create a handler that returns an error.

Confirm the task ends in:

```text
FAILED
```

and the worker continues processing other tasks.

---

# Engineering Rules

Do not add Redis yet.

Do not add HTTP yet.

Do not add a database yet.

Do not implement retries yet.

Do not implement the DLQ storage yet.

Do not over-engineer the system.

The goal of this phase is to prove that you understand the fundamentals required to build a concurrent task queue in Go.

---

# Final Goal

When this phase is complete, you should have a program capable of doing this:

```text
             ┌──────────────┐
             │    Producer  │
             └──────┬───────┘
                    │
                    ▼
             ┌──────────────┐
             │ In-Memory    │
             │ FIFO Queue   │
             └──────┬───────┘
                    │
          ┌─────────┼─────────┐
          ▼         ▼         ▼
      Worker 1  Worker 2  Worker 3
          │         │         │
          └─────────┼─────────┘
                    ▼
                Handler
                    │
              ┌─────┴─────┐
              ▼           ▼
          COMPLETED     FAILED
```

The next phase will replace the in-memory queue with a persistent/distributed implementation.