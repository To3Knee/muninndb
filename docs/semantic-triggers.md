# Semantic Triggers

> Push-based memory for AI agents. The database tells you when something matters.

---

## 1. The Problem With Pull

Every database primitive we have is pull-based. You query. You get results. You query again. You compare. You decide if anything changed.

This is fine for most use cases. It is fundamentally wrong for memory.

Think about how human attention works. You don't periodically poll your own brain for relevant memories. Your brain surfaces them — triggered by context, by time, by emotional salience, by contradiction. When you're deep in a debugging session and something about a Q3 incident suddenly becomes relevant, that memory pushes itself to your attention. You didn't ask.

AI agents built on pull-based databases have to simulate this push behavior. They poll at fixed intervals, make retrieval calls at every step, compare current results against previous results, and guess at what changed. They miss relevance shifts that happen between queries. They don't know when something they thought was settled has been contradicted. They don't know when a memory written last week just became critical to what they're doing right now.

The polling pattern is not just inefficient. It is architecturally wrong. The burden of knowing what's relevant falls on the agent, not the system that stores the knowledge. That's backwards.

---

## 2. What Semantic Triggers Are

A semantic trigger is a continuous subscription: *watch this context; tell me when anything becomes relevant above this threshold.*

You define the context you care about. You define the relevance level that matters. MuninnDB does the rest — continuously, automatically, from the storage layer. You don't poll. You don't compare. You don't guess.

Three things can fire a trigger:

**`new_write`** — A new engram is written to the vault and its activation score against your subscription context exceeds your threshold. You find out immediately — before your next query cycle, before your next LLM call.

**`threshold_crossed`** — An existing engram's relevance score changes and crosses your threshold. Time passed and a memory stabilized. A co-activation strengthened a Hebbian connection. The decay curve brought something back into range. The database found something for you that you didn't know was there.

**`contradiction_detected`** — An engram within your activation set has a contradiction detected. This is highest priority. No rate limiting. Immediate delivery. You need to know that your agent is working with contradictory information.

---

## 3. Why This Is a New Primitive

No other database does this. The claim is worth defending precisely.

**Relational databases** have `LISTEN/NOTIFY` (Postgres) and row-level triggers. These are data mutation events — they fire on `INSERT`, `UPDATE`, `DELETE`. They tell you a row changed. They say nothing about whether that row is relevant to anything you care about. A Postgres trigger cannot tell you "this fact just became semantically related to your task context." It can tell you a row was updated.

**Redis pub/sub** fires on explicit `PUBLISH` operations. A publisher has to decide to emit an event. Relevance score changes in a cognitive model are not explicit publish operations — they're continuous internal state changes. Redis has no model of relevance.

**Vector databases** have no push mechanism at all. They are pure query systems. You ask, they answer. The idea of a vector database proactively notifying you when a stored vector becomes relevant to a context you declared interest in — this does not exist in any vector database today.

**Graph databases** have adjacency and traversal. They have no continuous relevance scoring model that changes over time. They cannot detect when a node's relevance to a given context crosses a threshold.

The key distinction: semantic triggers fire based on *meaning and relevance*, not on data mutations. The trigger condition is evaluated inside the cognitive model — against activation scores, decay curves, Hebbian weights, contradiction signals — not at the storage layer. The trigger system understands what the data means in context. Storage-layer triggers understand only that data changed.

This is what makes it novel. It's not a faster version of something that already exists. It's a different thing.

---

## 4. How the Trigger System Works

### The TriggerWorker

One shared worker handles all subscriptions. Not one goroutine per subscription — that would be O(N) goroutines for N subscriptions, which breaks down at scale. The TriggerWorker maintains a shared event bus and processes events from four sources.

### Event Sources

**Write events** — After every successful write ACK, the trigger system receives a notification. It scores the new engram against all active subscriptions. This is O(S) where S is the subscription count — linear, bounded, predictable. High-write vaults with many subscriptions are the design load, not the edge case.

**Cognitive events** — The decay worker, Hebbian worker, and confidence worker all emit events when a score delta exceeds `NegligibleDelta`. Small score changes are filtered out. Only meaningful threshold crossings propagate into the event bus. This is the mechanism that fires `threshold_crossed` — an existing engram drifts into relevance because the cognitive model updated.

**Contradiction events** — Fired immediately on contradiction detection. Highest priority queue position. No rate limiting — contradictions are urgent and rare. An agent working with contradictory information needs to know now.

**Periodic sweep** — Every 30 seconds, all subscriptions are re-scored against the current vault state. This is the backstop. It catches anything the event stream might have missed and ensures that slow-accumulating relevance changes are eventually surfaced.

### Scoring a Subscription

Each subscription has one or more context strings and a threshold between 0.0 and 1.0. Scoring means running ACTIVATE against the subscription's context and checking whether the engram in question appears above threshold. If it does, delivery is triggered.

When the Embed plugin is active, subscription context strings are embedded to vectors. The cache stores embeddings keyed by SHA256 hash of the context string. Subscriptions with identical context strings share a single embedding — no redundant model calls for duplicate contexts.

### Rate Limiting

Token bucket per subscription. Prevents a single high-activity vault from flooding a subscriber. The bucket refills at a configured rate. When empty, non-critical events are held or dropped.

Contradiction events bypass the rate limiter entirely. You always get contradiction notifications.

### Delivery

Non-blocking async channel per subscriber. If the subscriber's channel is full because the consumer is processing too slowly, the notification is dropped. A metric is incremented. This is intentional — a subscriber that cannot keep up must not slow down the rest of the system. The periodic sweep provides a backstop: even dropped events will be caught at the next 30-second pass.

---

## 5. Using Triggers

### Basic Watch

```go
client := muninn.New("localhost:8747")
mem := client.Vault("my-project")

err := mem.Watch(ctx, "payment service architecture",
    muninn.WithThreshold(0.7),
    muninn.OnEngram(func(e muninn.Engram) {
        fmt.Printf("[MEMORY] %s: %s\n", e.Concept, e.Content)
    }),
    muninn.OnContradiction(func(a, b muninn.Engram) {
        fmt.Printf("[CONFLICT] %s contradicts %s\n", a.Concept, b.Concept)
    }),
)
```

The `Watch` call returns immediately. Delivery is via gRPC streaming. The callback fires on the trigger system's schedule, not yours.

### Trigger Class Filtering

```go
// Only care about new writes, not score drift
err := mem.Watch(ctx, "database architecture",
    muninn.WithThreshold(0.8),
    muninn.WithTriggerClasses(muninn.TriggerNewWrite, muninn.TriggerContradiction),
    muninn.OnEngram(func(e muninn.Engram) {
        // fired only on new_write and contradiction events
    }),
)
```

### Multiple Contexts

```go
// One subscription, multiple context strings
// Fires if the engram is relevant to any of them
err := mem.Watch(ctx, []string{
    "payment retry logic",
    "idempotency key implementation",
    "exponential backoff strategy",
},
    muninn.WithThreshold(0.65),
    muninn.OnEngram(func(e muninn.Engram) {
        // relevant to at least one context string above 0.65
    }),
)
```

### Vault Scope

```go
// Scoped to a single vault
mem := client.Vault("payment-service")
err := mem.Watch(ctx, "charge deduplication", muninn.WithThreshold(0.7), ...)

// Cross-vault (global scope — use with a higher threshold)
err := client.WatchGlobal(ctx, "payment system architecture",
    muninn.WithThreshold(0.85),
    muninn.OnEngram(func(e muninn.Engram) {
        fmt.Printf("[VAULT: %s] %s\n", e.VaultID, e.Concept)
    }),
)
```

---

## 6. The Agent Memory Pattern

The canonical integration for AI agents looks like this:

```go
func RunAgent(ctx context.Context, task string) {
    mem := client.Vault("agent-session")

    // Subscribe at session start
    // The DB will push memories as they become relevant — no polling
    mem.Watch(ctx, task,
        muninn.WithThreshold(0.7),
        muninn.OnEngram(func(e muninn.Engram) {
            // Inject into the agent's context window before next LLM call
            agent.InjectContext(e.Content)
        }),
        muninn.OnContradiction(func(a, b muninn.Engram) {
            // Surface the conflict — let the agent reason about it
            agent.InjectConflict(a.Content, b.Content)
        }),
    )

    // Agent works; the DB handles what's relevant
    agent.Run(ctx, task)
}
```

The session flow:

1. **Session start** — Create a Watch subscription for the agent's task context. One call. Done.

2. **As the agent works** — Memories relevant to the task are pushed automatically. The agent doesn't poll. It receives.

3. **Before each LLM call** — The agent's context window already contains the most relevant memories. No explicit retrieval query needed at each step.

4. **The DB handles** — what's relevant now (decay scoring), what connects (Hebbian associations), what's uncertain (confidence weighting), what's contradictory (contradiction detection and notification).

5. **The agent handles** — using pushed memories to make better decisions.

### This Is Not RAG

Retrieval-Augmented Generation is pull: the agent queries before it generates. It decides what to retrieve, when to retrieve it, and how much to retrieve. The retrieval step is explicit, synchronous, and agent-managed.

Semantic triggers are push: the database decides when something is relevant and sends it. The agent's context window accumulates relevant memories automatically across the session. Retrieval is continuous, async, and database-managed.

These are not competing approaches for the same problem. RAG is the right pattern for "I need specific information right now." Semantic triggers are the right pattern for "I want the database to surface what matters as the session evolves."

In practice, most agent architectures benefit from both: semantic triggers for continuous context accumulation, explicit ACTIVATE calls when the agent needs to deliberately search.

---

## 7. Operational Notes

**Subscription limits** — No hard limit per vault, but each active subscription adds O(1) work per write event. Monitor `trigger_subscriptions_active` in the metrics endpoint.

**Threshold calibration** — Start at 0.7. Lower thresholds mean more events (higher recall, lower precision). Higher thresholds mean fewer events (lower recall, higher precision). The periodic sweep catches anything the event stream misses, so a threshold that's slightly too high doesn't create permanent blind spots — just 30-second lag at most.

**Subscription lifetime** — Tied to the context lifetime. When the context is cancelled, the subscription is removed. Leaked contexts leak subscriptions. Use `defer cancel()`.

**Contradiction rate limiting** — Contradiction events bypass the token bucket. If a vault has many contradictions firing simultaneously (e.g., after a large batch write), each subscribed context receives all of them. Design contradiction handlers to be fast or to queue internally.

**Metrics** — The trigger system exposes:
- `trigger_events_total` — by event type and vault
- `trigger_deliveries_total` — successful deliveries
- `trigger_drops_total` — dropped events (consumer too slow)
- `trigger_subscriptions_active` — current active count
- `trigger_sweep_duration_seconds` — periodic sweep timing
