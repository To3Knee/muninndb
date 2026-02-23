# MuninnDB

**The world's first cognitive database. Memory, not storage.**

![Go Version](https://img.shields.io/badge/go-1.22%2B-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)
![Status](https://img.shields.io/badge/status-alpha-orange)

---

## What this is

Every database you have ever used stores data and waits to be asked.

MuninnDB stores memory. It knows what matters right now, what connects to what, what you have forgotten, and it tells you before you ask. It continuously recalculates relevance as time passes. It discovers associations you never defined. When something becomes important, it pushes that memory to you.

This is what the human brain does. No database has ever done this.

---

## Every other database gets this wrong

| Database Type | What it does | What it can't do |
|---|---|---|
| **Relational** (Postgres, MySQL) | Stores structured rows. Queries by exact match or index. | Doesn't know two rows are semantically related. Doesn't know one is more important today than it was last week. Relevance is not a concept it has. |
| **Document** (MongoDB, Firestore) | Stores JSON blobs. Queries by fields. Flexible schema. | Same fundamental problem. No sense of time. No sense of relevance. No associations between documents unless you build them yourself and maintain them forever. |
| **Key-Value** (Redis, DynamoDB) | Blazing fast lookups by key. Simple, predictable, fast. | Zero context. Zero relationships. Zero memory of what matters. The database has no idea what it's holding. |
| **Graph** (Neo4j, Amazon Neptune) | Models relationships between nodes. Traverses connections. | Relationships are static. Hand-defined. They don't strengthen or weaken. There's no decay. No relevance scoring. The graph doesn't learn — you update it manually or it stays wrong. |
| **Vector** (Pinecone, Weaviate, pgvector) | Finds semantically similar content via embeddings. | Great at "find similar." Cannot tell you what's relevant *right now* given your current context. No decay. No push delivery. No cognitive model. Similarity is not memory. |
| **MuninnDB** | Stores memory traces (engrams). Knows what's relevant now. Pushes to you when something matters. Learns which concepts connect. Forgets what you no longer need. | This is the new category. |

The problem is not that other databases are slow. Postgres is fast. Redis is extremely fast. The problem is that they are solving the wrong problem.

They are libraries. A library holds everything perfectly and returns exactly what you ask for. That is valuable. But a library does not know which book you need before you know you need it. A library does not notice that two books are about the same idea. A library does not forget the books that no longer matter to your work.

MuninnDB is a brain. It holds memory, tracks relevance, discovers connections, and initiates contact when something matters. That is a different thing entirely.

---

## The code moment

Three things are about to happen that no other database in the world can do.

```go
// Tell MuninnDB something
err := mem.Remember(ctx, "payment service",
    "We switched to idempotency keys after the double-charge incident in Q3")

// Ask it what's relevant RIGHT NOW
results, err := mem.Activate(ctx, "debugging the payment retry logic")
// Returns the Q3 incident — even though you never mentioned it.
// MuninnDB connected the concepts. No query. No join. No index hint.

// Watch for memories that become relevant as your work evolves
mem.Watch(ctx, "payment service", func(e muninn.Engram) {
    fmt.Printf("Relevant: %s (%.0f%% confidence)\n", e.Concept, e.Confidence*100)
})
// The DB pushes to you. You don't poll. You don't query. It remembers.
```

1. `Activate` retrieved a memory by *meaning and context* — not by field, not by key, not by vector distance alone. It understood that "payment retry logic" and "idempotency keys after a double-charge incident" are the same conversation.
2. The connection between those two concepts was *inferred*. You never defined it. You never created a relationship in a schema. The association emerged from co-activation and semantic proximity.
3. The `Watch` call means the database will push future relevant memories to you — without you asking — because relevance changed. Time passed. Related things were activated. A contradiction was found. MuninnDB initiates. You receive.

---

## This is not a wrapper

MuninnDB is built from scratch. Every core component is purpose-built for cognitive memory.

**Built from scratch:**

- **ERF (Engram Record Format)** — a custom binary storage format designed specifically for cognitive memory traces. Not JSON. Not protobuf on disk. A purpose-built format with decay scores, association weights, and confidence values baked into every record at the byte level. The format itself encodes time.

- **MBP (Muninn Binary Protocol)** — a custom TCP wire protocol. Not REST with extra steps. A pipelined binary protocol with correlation IDs, designed for sub-10ms memory operations. Every frame knows what request it belongs to. Responses arrive out of order and are reassembled correctly at the client. This is not a design you can bolt onto an existing protocol.

- **6-phase Activation Engine** — a query pipeline that runs full-text search, vector search, and decay scoring in parallel, then fuses the results with Reciprocal Rank Fusion, applies Hebbian co-activation boosts, traverses the association graph with BFS, and scores by confidence — in under 20ms. This is not a feature of any existing database engine. It is the engine.

- **Cognitive workers** — async goroutines that continuously run Ebbinghaus decay, Hebbian learning, Bayesian confidence updating, and contradiction detection in the background. While you are working, the database is working. Relevance scores are recalculating. Associations are strengthening or weakening. Confidence is being updated. The database gets smarter while you are not looking.

- **Semantic trigger system** — a pub/sub primitive native to the database itself. When any memory crosses a relevance threshold — because time passed, because you activated related things, because a contradiction was detected — the database pushes to your subscriber. No polling. No cron job. No external message queue. The database initiates.

**Pebble is our filesystem.** We use CockroachDB's Pebble as the raw key-value layer — the same way Postgres uses the operating system's filesystem. Pebble stores bytes reliably. Everything above that — the cognitive architecture, the indexes, the wire protocols, the storage format, the activation engine — is MuninnDB.

---

## Five things that make it different

**1. Memory Decay**

Your brain knows that a memory you haven't touched in months matters less than something you thought about yesterday. MuninnDB knows this too. It continuously recalculates relevance using the Ebbinghaus forgetting curve — a formula derived from 19th-century memory research that describes, with mathematical precision, how biological memory fades over time. Every engram has a decay score. That score changes whether or not you do anything. The database moves.

*Technically: Ebbinghaus decay is applied by background cognitive workers on a continuous schedule, adjusting the relevance weight of every engram based on time since last activation and original encoding strength.* [How memory works](docs/how-memory-works.md)

---

**2. Hebbian Association**

"Neurons that fire together, wire together." Hebb's Rule, 1949. When two memories are retrieved together repeatedly, their connection strengthens automatically. You don't define relationships. You don't maintain a schema of associations. You don't write a migration when two concepts start traveling together. The associations emerge from use — and they fade when use stops.

*Technically: Co-activation events are recorded in the association graph. Edge weights are updated using a Hebbian learning function applied by the cognitive worker layer. Weights increase on co-activation and decay symmetrically with the Ebbinghaus function.* [Cognitive primitives](docs/cognitive-primitives.md)

---

**3. Semantic Push Triggers**

This is the one that has no analogue anywhere else. You subscribe to a context. The database tells you when something becomes relevant to that context. Not because you queried. Because relevance *changed* — time passed, related memories were activated by someone else, a contradiction was discovered, a new memory was added that connects to your context. The database is the initiator. You are the receiver.

*Technically: The trigger system evaluates subscriptions against the cognitive state after every background worker cycle and every activation event, pushing engrams to subscribers when relevance scores cross configured thresholds.* [Semantic triggers](docs/semantic-triggers.md)

---

**4. Bayesian Confidence**

Every memory has a confidence score. That score is not a label you assign — it updates automatically. A new memory that contradicts an existing one lowers confidence. A new memory that reinforces an existing one raises it. The math is Bayesian: principled, calibrated, updatable, not a heuristic someone tuned by hand. You can ask MuninnDB not just what it remembers, but how sure it is, and the answer is grounded in evidence.

*Technically: Confidence is tracked per-engram using a Beta distribution prior. Evidence for and against is accumulated by the contradiction detection and reinforcement systems. Posterior confidence is computed and stored at update time.* [Cognitive primitives](docs/cognitive-primitives.md)

---

**5. Retroactive Enrichment**

Install the embed plugin. Every existing memory in your vault gets vector embeddings automatically, in the background, without blocking reads or writes. The activation engine immediately becomes smarter. Install the enrich plugin and every memory gets LLM-generated summaries, entity extraction, and typed relationship detection — retroactively applied to everything already in the database.

You don't rewrite your application. You don't migrate your data. You add a plugin and the database upgrades its own understanding of what it holds.

*Technically: Both plugins implement the MuninnDB plugin interface, registering as background enrichment workers. They process the engram backlog asynchronously using configurable concurrency and rate limits.* [Plugins](docs/plugins.md)

---

## Quickstart

```bash
# Single binary. Zero config. Zero dependencies.
go run ./cmd/muninndb -data ./my-memory

# Three ports come up:
# 8474 — MBP (native binary protocol, fastest)
# 8475 — REST (JSON)
# 8750 — MCP (for AI agents)
# 8476 — Web UI
```

**Via REST (available today):**

```bash
# Write a memory
curl -X POST http://localhost:8475/api/engrams \
  -H "Content-Type: application/json" \
  -d '{"concept":"architecture decision","content":"We chose event sourcing over CRUD for the order service"}'

# Activate by context
curl -X POST http://localhost:8475/api/activate \
  -H "Content-Type: application/json" \
  -d '{"context":"designing the order refund flow"}'
```

**Via MCP (for AI agents — Claude, Cursor, and any MCP-compatible client):**

MuninnDB exposes 17 MCP tools at `http://localhost:8750`. Point your MCP client at that address and MuninnDB becomes your agent's persistent memory with no additional integration work.

> **Go SDK** is in active development. Watch this repo for updates.

---

## Access & Authentication

MuninnDB uses a two-layer model that reflects how a cognitive database is actually used — not how a relational database was designed in 1974.

**Layer 1 — Admin (operators):** Username and password. Used for the web UI and shell. The first run prints a generated root password. Admins create vaults, manage API keys, and configure access.

**Layer 2 — Vault API keys:** Each vault is either **open** (no auth, default) or **locked** (requires a bearer key). A vault can have multiple keys — one per integration point. Keys come in two modes:

| Mode | What it does |
|---|---|
| `full` | Participates in cognitive state. Activations update decay timers, access counts, and Hebbian weights. The vault *learns* from this connection. |
| `observe` | Ephemeral reads only. Relevance scores are computed but nothing is written back. The cognitive state of the vault is never altered. |

This distinction matters because MuninnDB is the first database where **reads have side effects on the data**. When you activate a memory, you strengthen its connections, reset its decay timer, and influence what surfaces next time. An `observe` key gives you a window into the vault's learned state without contaminating it. A dashboard, an analytics job, or a read-only partner integration should never reshape the brain. A `full` key is for agents that are part of the brain.

```bash
# Create a key (admin only)
curl -X POST http://localhost:8475/api/admin/keys \
  -H "Content-Type: application/json" \
  -d '{"vault":"default","label":"production-agent","mode":"full"}'
# → {"token":"mk_xK9m...","key":{...}}   ← token shown once

# Use the key on any vault request
curl -X POST http://localhost:8475/api/activate \
  -H "Authorization: Bearer mk_xK9m..." \
  -H "Content-Type: application/json" \
  -d '{"context":"debugging the payment retry logic"}'
```

[Full auth documentation →](docs/auth.md)

---

## Documentation

| Document | What's in it |
|---|---|
| [How Memory Works](docs/how-memory-works.md) | The neuroscience behind why MuninnDB works the way it does |
| [MuninnDB vs. Every Other Database](docs/vs-other-databases.md) | Full comparison with relational, document, key-value, graph, and vector databases |
| [Architecture Deep Dive](docs/architecture.md) | ERF binary format, 6-phase activation engine, wire protocols, cognitive workers |
| [Cognitive Primitives](docs/cognitive-primitives.md) | Decay, Hebbian learning, Bayesian confidence, contradiction detection |
| [Semantic Triggers](docs/semantic-triggers.md) | How push-memory works and why nothing else has it |
| [What Is An Engram?](docs/engram.md) | The core data model — why it's not a row, not a document, not a node |
| [Plugins: Embed & Enrich](docs/plugins.md) | Vector embeddings and LLM enrichment, added without changing a line of your code |
| [Access & Authentication](docs/auth.md) | Two-layer auth model, vault API keys, full vs. observe mode |

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Pull requests are welcome. If you are thinking about a large change, open an issue first.

---

## License

Apache 2.0. See [LICENSE](LICENSE).

---

*Named after Muninn — one of Odin's two ravens, whose name means "memory" in Old Norse. In myth, Muninn flies across the nine worlds and returns what has been forgotten. We thought that was a good name for a database.*
