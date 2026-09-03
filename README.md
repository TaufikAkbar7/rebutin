# Rebutin

A high-concurrency ticket rush simulation where real users compete against automated bots to secure limited tickets. This project demonstrates how modern distributed systems tackle core backend challenges: **race condition mitigation**, **distributed locking**, **asynchronous task processing**, and **real-time system observability**.

---

## Problem Statement

Rush ticketing platforms inherently face severe concurrency hurdles:

* **Race Conditions:** Thousands of concurrent users/bots competing for the exact same inventory slot simultaneously.
* **Overselling / Overbooking:** Inconsistent stock decrementing when state changes lack atomic protection.
* **Fairness Under Load:** Managing fair queue progression when incoming traffic exceeds system capacity.
* **Eventual Consistency in Async Flows:** Resolving late-arriving payment confirmations after reservation holds have already expired.

**Rebutin** models these failure modes in a controllable simulation environment (customizable bot counts, network throttling, and ticket inventory) to demonstrate robust failure-handling architectures under extreme load.

---

## Tech Stack

| Layer | Technology | Responsibilities |
| :--- | :--- | :--- |
| **Distributed Lock & Cache** | **Redis** | Atomic stock counters, session TTL, Virtual Waiting Room queue (`ZSET`), active session tracking |
| **Message Broker** | **RabbitMQ** | Asynchronous event processing (notifications, expiry events, DB reconciliation) |
| **CI/CD** | **Jenkins** | Automated linting, test suites, container builds, and deployment pipelines |
| **Observability** | **Grafana + Prometheus** | Real-time traffic metrics, latency percentiles, error rates, and queue lag |
| **Backend** | *Go / PostgreSQL* | Core REST API, concurrency logic, atomic state transitions, worker pools |
| **Frontend** | *Vue 3 / TypeScript* | Simulation config dashboard, real-time waiting room, live countdowns |

---

## System Flow
``Config ──► Waiting Room (Conditional) ──► Category Selection ──► Payment ──► Success / Refund``

**1. Configuration View**
<br>
Users configure baseline simulation parameters:
- Bot Count: 0 – 10,000 concurrent bot workers.
- Bot Throttle / Network Delay: 1 – 5 seconds.
- Total Inventory: Automatically partitioned equally across 3 categories (Total / 3).
- Max Concurrent Users: Ingress threshold before the Virtual Waiting Room activates.
- Simulate Delayed Payment Webhook: Emulates payment processing delays colliding with TTL expirations.

**2. Virtual Waiting Room (Traffic Absorber)**
- Trigger: Activates automatically whenever active concurrent sessions exceed max_concurrent. If traffic remains within thresholds, participants bypass the queue straight to Category Selection.
- Mechanism: Strict FIFO ordering powered by Redis `ZSET` (scored by arrival epoch) paired with a Leaky Bucket rate limiter to release users at a steady, manageable pace.
- Real-time Feedback: Live queue position and estimated wait times streamed to participants.
- Session Start: Releasing from the queue immediately provisions a unified Session TTL.

**3. Category Selection & Reservation**
- Displays 3 ticket categories, each allocated `total_ticket/3`.
- Just-In-Time Reservation: Quota is reserved on click, never on initial page render.
- Unified Session Binding: The reservation lock is tied directly to the parent session window rather than an independent timer.
- Atomic Category Swap: Participants may switch categories seamlessly. The system executes an atomic swap (releasing the old hold while securing the new slot) without resetting the global Session TTL, eliminating timer-reset exploits.
- Bot Behavior: Bots instantly select a random category upon entering this phase.

**4. Payment / Checkout**
- Backend-Driven Countdown: The remaining time is derived from the Redis key TTL on every request, preventing client-side clock drift or refresh tampering.
- Natural Processing: Payments proceed through regular validation and gateway handshakes without artificial delays.
- Pre-Finalization Safety Check: Prior to committing the sale, the backend verifies whether the held slot remains valid and unexpired.

**5. Final State Resolution**
- Success: Payment is approved while the reservation slot remains valid. Inventory is permanently assigned.- Refund: Payment is approved, but the confirmation arrives after the session expired and the slot was returned to the pool. This emerges naturally from network latency races and triggers an automatic compensation/refund flow.
- Timeout (Abandonment): The session expires with no checkout attempt. Slots are released back into the pool.
---

## Bot Behaviour
- Progresses through the Virtual Waiting Room under identical rate-limiting rules as real users.
- Randomly selects 1 of the 3 available categories upon entry.
- Applies configured network jitter and throttle delays across all steps.
- Completes checkout immediately once a reservation slot is successfully acquired.
---

## [Entity Relationship Diagram](https://mermaid.ai/d/becc4f69-78f6-401a-b5bd-0b3465e0d78c)
Data Storage Responsibilities
- Redis: Live stock counters, session expiration keys, waiting room `ZSET`, active concurrency counters, and distributed lock tokens.
- PostgreSQL: Permanent transaction history, audit trails, and post-simulation analytics.
---

## [Flowchart](https://mermaid.ai/d/6d836a8f-ec5c-40ce-a331-ae0f7f20df82)
---

## Key Architectural Decisions
|Decision|Rationale  |
|--|--|
| Reservation on Click (Not Page Load) |  Prevents inventory hoarding by window shoppers and eliminates spam-navigation exploits.|
| Unified Session TTL|Enables users to switch ticket tiers without resetting the checkout window or extending their time illegally. |
| Redis `ZSET` + Leaky Bucket Queue | Guarantees strict FIFO ordering while flattening traffic spikes to protect downstream services.|
| Early Inventory Decrement |Decrementing stock during reservation prevents overselling and guarantees that every active checkout holds a real seat. |
| Natural Race-Condition Refunds | Reflects realistic production failure modes where delayed payment confirmations collide with automatic inventory release.|
| Atomic Redis Operations |Guarantees zero race conditions during stock decrements and tier-switching operations. |
| Server-Sent Events (SSE) for Waiting Room | Efficient unidirectional live data delivery over single HTTP connections; avoids HTTP polling thundering-herd and WebSocket handshake overhead. |
