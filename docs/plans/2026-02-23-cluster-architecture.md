# MuninnDB Cluster Architecture

**Date:** 2026-02-23
**Status:** Strategic Design Specification
**Authors:** Architecture team
**Audience:** Senior Go engineers implementing cluster infrastructure

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Foundational Principle: Reads Are Writes](#2-foundational-principle-reads-are-writes)
3. [Read/Write Topology Decision](#3-readwrite-topology-decision)
4. [Node Roles and Vocabulary](#4-node-roles-and-vocabulary)
5. [Leader Election and Failover](#5-leader-election-and-failover)
6. [Node Discovery](#6-node-discovery)
7. [Initial State Transfer (Cognitive Snapshot)](#7-initial-state-transfer-cognitive-snapshot)
8. [Replication Transport](#8-replication-transport)
9. [Operational Simplicity](#9-operational-simplicity)
10. [Cognitive-Specific Cluster Behaviors](#10-cognitive-specific-cluster-behaviors)
11. [Enterprise "Wow Factor" Features](#11-enterprise-wow-factor-features)
12. [Phased Implementation Roadmap](#12-phased-implementation-roadmap)
13. [Appendix: Current Codebase Gaps](#appendix-current-codebase-gaps)

---

## 1. Executive Summary

MuninnDB is a cognitive database where **every read mutates state**. Activation queries trigger Hebbian weight updates, Ebbinghaus decay resets, Bayesian confidence recalculations, and semantic push triggers. This property -- reads are writes -- invalidates the assumptions behind every traditional replication topology.

This document specifies a cluster architecture built on three pillars:

1. **Option B+D Hybrid: Cognitive Primary with Write-Forwarding Replicas** -- a single cognitive primary owns all cognitive mutations, replicas serve locally with optional write-forwarding for cognitive side effects.
2. **Self-contained Sentinel-style HA** -- no etcd, no Consul, no ZooKeeper. Nodes monitor each other and reach quorum for failover decisions. Single binary, zero external dependencies.
3. **Phased delivery** -- a stable, valuable Phase 1 cluster ships first. Advanced features land in later phases only after the foundation is proven.

The architecture targets clusters of 3-7 nodes (typical) and supports up to 50 nodes.

---

## 2. Foundational Principle: Reads Are Writes

Before any topology discussion, we must internalize what happens on an activation query:

```
Client sends: ACTIVATE ["project deadlines", "team priorities"]
                           |
                    Activation Engine
                    (6-phase pipeline)
                           |
                    +------+------+------+------+
                    |      |      |      |      |
               FTS query  HNSW   Decay  Graph   Hebbian
               (read)     search  eval  traverse boost
                    |      |      |      |      |
                    +------+------+------+------+
                           |
                    Fused scoring + ranking
                           |
                    SIDE EFFECTS (writes):
                    1. Hebbian: UpdateAssocWeight() for all co-activated pairs
                    2. Decay:   ResetDecayTimer() for activated engrams
                    3. Confidence: BayesianUpdate() on accessed engrams
                    4. Triggers: Evaluate push triggers, may fire
                    5. Activity: Update access count, last_access timestamp
```

A single activation query touching 20 engrams may produce:
- Up to 190 Hebbian weight updates (20 choose 2 pairs)
- 20 decay timer resets
- 20 confidence updates
- 20 access count increments
- N trigger evaluations

**This means a "read-only replica" that serves activations is not read-only.** Any replica that answers activation queries is generating writes.

---

## 3. Read/Write Topology Decision

### Decision: Option B+D Hybrid -- Cognitive Primary with Write-Forwarding Replicas

After analyzing all four options against MuninnDB's constraints:

| Option | Verdict | Reason |
|--------|---------|--------|
| A: Single Writer | **Partial adopt** | Foundation of the model. Replicas are useful for cold reads. |
| B: Cognitive Primary, Cold Replicas | **Adopt as base** | Clean separation. Replicas serve "cold reads" (lookup by ID, stat queries, admin). |
| C: Cognitive Mesh | **Reject for Phase 1-3** | Multi-writer Hebbian conflicts are theoretically acceptable but operationally complex. Association weight convergence under concurrent updates is not well-studied. Too risky for a foundational release. |
| D: Write-Forwarding Replicas | **Adopt as enhancement** | Replicas that receive activation queries forward the cognitive side effects to the primary asynchronously. The activation result is served immediately from local state. |

### How B+D Works Together

```
                         +-----------------+
                         |    PRIMARY      |
                         | (Cognitive Brain)|
                         |                 |
                         | - All writes    |
                         | - Cognitive     |
                         |   workers run   |
                         | - Replication   |
                         |   source        |
                         +---------+-------+
                            |          |
                  replication stream    |
                    (MBP frames)       |
                    |                  |
          +---------+-------+   +-----+----------+
          |   REPLICA #1    |   |   REPLICA #2   |
          | (Cold + Forward)|   | (Cold + Forward)|
          |                 |   |                 |
          | - Serves cold   |   | - Serves cold  |
          |   reads (by ID) |   |   reads (by ID)|
          | - Activation    |   | - Activation   |
          |   queries:      |   |   queries:     |
          |   serve locally,|   |   serve locally|
          |   forward side  |   |   forward side |
          |   effects to    |   |   effects to   |
          |   primary       |   |   primary      |
          +-----------------+   +----------------+
```

### Read Classification

Not all reads are created equal. MuninnDB defines two read classes:

| Read Class | Cognitive Side Effects | Can Run on Replica | Example |
|------------|----------------------|-------------------|---------|
| **Cold Read** | None | Yes, directly | `GET /v1/engrams/{id}`, `STAT`, `PING`, admin queries |
| **Activation** | Hebbian, Decay, Confidence, Triggers | Yes, with forwarding | `ACTIVATE`, `POST /v1/activate` |

When a replica receives an Activation:
1. Execute the 6-phase activation pipeline locally (using local data, which may be slightly stale)
2. Return results to the client immediately
3. Asynchronously package the cognitive side effects (co-activation pairs, accessed engram IDs) into a `CognitiveSideEffect` message
4. Forward that message to the primary over the replication channel
5. Primary applies the side effects and they propagate back to all replicas on the next replication cycle

**Staleness window:** The replica's local cognitive state is eventually consistent, trailing the primary by the replication lag (typically <500ms). For activation queries, this means Hebbian weights and decay scores may be one replication cycle behind. This is acceptable -- human memory has this same property. The same thought retrieved twice in rapid succession does not meaningfully change in the intervening milliseconds.

### Justification for Rejecting Cognitive Mesh (Option C)

Hebbian learning is mathematically: `w_new = w_old * (1 + lr)^signal`. When two nodes concurrently strengthen the same association edge with different signals:

- Node A computes: `w = 0.5 * (1.01)^3.2 = 0.516`
- Node B computes: `w = 0.5 * (1.01)^1.8 = 0.509`
- Last-write-wins: one update is lost

While the lost precision is small, it compounds over time and across high-activation workloads. More critically, debugging association weight divergence across a mesh is operationally nightmarish. A single cognitive primary is simple to reason about and debug.

**Revisit in Phase 4+** once single-primary has been battle-tested and we have operational tooling to detect weight divergence.

---

## 4. Node Roles and Vocabulary

MuninnDB does not use "master/slave" or "leader/follower." It uses cognitive vocabulary:

| Role | Internal Constant | Description |
|------|------------------|-------------|
| **Cortex** | `RoleCortex` (replaces `RolePrimary`) | The cognitive brain. Accepts all writes, runs cognitive workers, sources the replication stream. One per cluster. |
| **Lobe** | `RoleLobe` (replaces `RoleReplica`) | A cognitive region. Receives replication stream, serves cold reads and write-forwarded activations. Participates in leader election. |
| **Sentinel** | `RoleSentinel` | A lightweight monitor node that participates in quorum voting but stores no data. Used only when you need an odd quorum number (e.g., 2 data nodes + 1 sentinel = quorum of 2). |
| **Observer** | `RoleObserver` | A read-only node that receives the replication stream but does NOT participate in elections. For analytics, monitoring dashboards, cross-region warm standby. |

**Note on naming:** While the cognitive naming (Cortex/Lobe) is used in documentation and UI, the internal Go constants and wire protocol use `primary`/`replica` for unambiguous engineering clarity. The mapping is:

```go
const (
    RolePrimary  NodeRole = 1  // Cortex
    RoleReplica  NodeRole = 2  // Lobe
    RoleSentinel NodeRole = 3  // Sentinel (vote-only, no data)
    RoleObserver NodeRole = 4  // Observer (data, no vote)
    RoleUnknown  NodeRole = 0  // Startup / uninitialized
)
```

### Minimum Viable Cluster Configurations

```
2-node (not recommended, no automatic failover):
  Node A: Cortex
  Node B: Lobe
  Failover: manual only (no quorum)

3-node (recommended minimum):
  Node A: Cortex
  Node B: Lobe
  Node C: Lobe
  Quorum: 2/3

2-node + sentinel (budget HA):
  Node A: Cortex
  Node B: Lobe
  Node C: Sentinel (no storage, minimal resources)
  Quorum: 2/3

5-node (production):
  Node A: Cortex
  Node B-E: Lobe
  Quorum: 3/5
```

---

## 5. Leader Election and Failover

### Design: MuninnDB Sentinel Protocol (MSP)

Inspired by Redis Sentinel but adapted for cognitive databases. No external DCS required. Nodes form their own consensus layer for leadership.

### 5.1 How It Works

Every node in the cluster (except Observers) runs an MSP goroutine that:

1. **Heartbeats** every 1 second to all other known nodes via MBP `PING`/`PONG` frames
2. **Monitors** the Cortex (primary) for liveness
3. **Votes** in failover elections when the Cortex is determined to be down

```
                    MSP Heartbeat Protocol

  Node A (Cortex)  <--PING/PONG-->  Node B (Lobe)
       |                                  |
       +----------PING/PONG---------+    |
       |                             |    |
  Node C (Lobe)  <--PING/PONG-->  Node B (Lobe)
```

### 5.2 Failure Detection

Three-tier detection prevents false positives:

| Tier | Duration | Action |
|------|----------|--------|
| **SDOWN** (Subjective Down) | Node misses 3 consecutive heartbeats (3s) | This node marks the Cortex as SDOWN locally |
| **ODOWN** (Objective Down) | Quorum of nodes agree on SDOWN (exchange SDOWN votes) | Cluster agrees the Cortex is down |
| **FAILOVER** | ODOWN + election winner | New Cortex is promoted |

```
Timeline:

t=0s     Cortex crashes
t=1s     Node B: missed heartbeat 1
t=2s     Node B: missed heartbeat 2
t=3s     Node B: missed heartbeat 3 -> marks SDOWN
         Node C: also marks SDOWN (independently)
t=3.1s   Node B sends SDOWN(cortex) to Node C
         Node C sends SDOWN(cortex) to Node B
         Both have quorum (2/3 including dead node, or 2/2 live) -> ODOWN
t=3.2s   Election begins (see below)
t=3.5s   New Cortex is operational

Total failover time: ~3.5 seconds
```

### 5.3 Election Protocol

When ODOWN is declared:

1. Each eligible Lobe (with data) increments a local `electionEpoch` counter
2. Each Lobe requests votes from all other nodes: `VOTE_REQUEST{epoch, nodeID, lastSeq}`
3. Nodes vote for the candidate with the highest `lastSeq` (most up-to-date data) -- tie-broken by lowest `nodeID` (deterministic)
4. A candidate needs votes from a **quorum** to win
5. Winner sends `CORTEX_CLAIM{epoch, fencingToken}` to all nodes
6. All nodes accept if the epoch and token are higher than their current known values

```go
// Election message types (added to MBP frame types)
const (
    TypeVoteRequest  uint8 = 0x30
    TypeVoteResponse uint8 = 0x31
    TypeCortexClaim  uint8 = 0x32
    TypeSDown        uint8 = 0x33
    TypeODown        uint8 = 0x34
)

type VoteRequest struct {
    Epoch      uint64 `msgpack:"epoch"`
    CandidateID string `msgpack:"candidate_id"`
    LastSeq    uint64 `msgpack:"last_seq"`
    ConfigEpoch uint64 `msgpack:"config_epoch"`
}

type VoteResponse struct {
    Epoch      uint64 `msgpack:"epoch"`
    VoterID    string `msgpack:"voter_id"`
    Granted    bool   `msgpack:"granted"`
}
```

**Critical rule:** A node can only vote once per epoch. This prevents split-vote scenarios where two candidates each get exactly half the votes.

### 5.3.1 Fencing Token = Election Epoch (Distributed Guarantee)

The fencing token in MuninnDB's distributed mode is the **election epoch number**. Each epoch is unique per election round and provides the split-brain fencing guarantee.

**Epoch lifecycle:**
- Candidates propose `lastKnownEpoch + 1` in `VoteRequest.Epoch`
- The `CortexClaim.FencingToken` field is set to the winning epoch
- All nodes store `currentEpoch` persistently in Pebble at key `0x19 | 0x03 | "cluster_epoch"`
- On startup, each node loads its stored epoch so it never proposes an epoch it has already seen

**Single-epoch guarantee proof:**

Two nodes cannot win the same epoch because: (a) votes are single-cast per epoch per voter, (b) quorum requires a majority, (c) if two candidates both propose epoch N, at most one can receive quorum.

*Proof by pigeonhole principle:*
- In a cluster of size N, quorum Q = floor(N/2) + 1
- For two candidates to both win epoch N, they would need Q + Q = 2Q votes total
- Only N votes are available (one per node per epoch)
- Since 2 * (floor(N/2) + 1) > N for all N >= 2, two simultaneous winners are **mathematically impossible**

*Concrete examples:*
- **3-node cluster:** Q = 2. Two winners would need 2 + 2 = 4 votes, but only 3 exist. Impossible.
- **5-node cluster:** Q = 3. Two winners would need 3 + 3 = 6 votes, but only 5 exist. Impossible.
- **7-node cluster:** Q = 4. Two winners would need 4 + 4 = 8 votes, but only 7 exist. Impossible.

**Persistent epoch storage format:**
```
Key:   0x19 | 0x03 | "cluster_epoch"  (variable length)
Value: uint64 big-endian (8 bytes)
```

This key is read on startup and updated atomically with each `CortexClaim` acceptance.

### 5.4 Split-Brain Protection

Multiple layers prevent split-brain:

**Layer 1: Fencing Tokens (existing)**
The existing `fence.go` implementation is sound. Every Cortex claim increments a monotonically increasing fencing token. Any write bearing a stale token is rejected by `ValidateFencingToken()`.

**Layer 2: Quorum Writes**
The Cortex must be able to reach a quorum of Lobes to confirm it is still the legitimate primary. If a network partition isolates the Cortex from the majority:

```
Partition scenario:

  [Minority side]          |  [Majority side]
  Node A (old Cortex)      |  Node B (Lobe)
                           |  Node C (Lobe)

Node A: Cannot reach quorum -> enters READ-ONLY mode after lease expiry
Node B+C: Elect new Cortex (Node B wins, has highest lastSeq)
Node B: Starts accepting writes with fencing token +1

When partition heals:
Node A: Sees higher fencing token -> demotes self to Lobe
Node A: Receives replication stream from Node B (new Cortex)
```

**Layer 3: Lease Expiry**
The existing `LeaderElector` lease TTL (10s) ensures a partitioned Cortex stops accepting writes within 10 seconds. Combined with fencing tokens, this provides a hard upper bound on the split-brain window.

**Layer 4: Node Epoch Verification**
On every write, the Cortex verifies it can still reach a **quorum** of Lobes (async, non-blocking on the write path). If it cannot reach a quorum of Lobes (floor(N/2) + 1 - 1, since the Cortex counts as 1 toward quorum) for longer than `lease_ttl / 2` (5s), it proactively demotes itself. This is the "pre-emptive demotion" strategy -- the Cortex doesn't wait for the lease to expire; it steps down early.

**Quorum examples:**
- In a **3-node cluster** (1 Cortex + 2 Lobes), quorum = 2. The Cortex counts as 1, so it needs to reach **at least 1 Lobe**.
- In a **5-node cluster** (1 Cortex + 4 Lobes), quorum = 3. The Cortex counts as 1, so it needs to reach **at least 2 Lobes**.
- In a **7-node cluster** (1 Cortex + 6 Lobes), quorum = 4. The Cortex counts as 1, so it needs to reach **at least 3 Lobes**.

### 5.5 Post-Failover: Cognitive Warmup

When a Lobe is promoted to Cortex, it must start the cognitive workers:

1. `HebbianWorker` -- starts immediately, processes new co-activation events
2. `DecayWorker` -- starts immediately, resumes from the schedule heap (which was replicated)
3. `ContradictWorker` -- starts immediately
4. `ConfidenceWorker` -- starts immediately

The cognitive workers operate on the data already present in the Lobe's Pebble store, which was kept up-to-date by the replication stream. **No warmup period is needed for data** -- only the worker goroutines need to start.

However, the `HebbianWorker`'s `lastDecayAt` timestamp (in-memory only) must be persisted. Add a Pebble key at `0x19 | 0x01` for cognitive worker metadata:

```
Key: 0x19 | 0x01 | "hebbian_last_decay"
Value: int64 (unix nanos)
```

This ensures a promoted Lobe doesn't immediately re-run the 6-hour decay cycle.

---

## 6. Node Discovery

### Design: Static Seed List + Gossip Convergence

**Why not pure gossip?** Gossip alone requires a bootstrap problem (who do you gossip with first?). Pure DNS requires infrastructure MuninnDB shouldn't depend on. Multicast is blocked in most cloud environments.

**Solution:** Start with a static seed list, then use gossip to discover new nodes and handle topology changes.

### 6.1 Configuration

```yaml
# muninn.yaml (or equivalent flags)
cluster:
  node_id: "muninn-01"           # unique, stable across restarts
  bind_addr: "10.0.1.10:8474"    # MBP listen address (replication uses same port)
  seeds:                          # initial peers to contact
    - "10.0.1.11:8474"
    - "10.0.1.12:8474"
  role: "auto"                    # "auto" | "primary" | "replica" | "sentinel" | "observer"
```

**`role: auto`** (the default) means: "join the cluster and participate in elections. If no Cortex exists, compete to become one." This is the Redis Sentinel model -- you don't pre-assign roles; the cluster decides.

### 6.2 Join Protocol

When a node starts:

```
1. Read config, determine seeds
2. Connect to each seed via MBP
3. Send HELLO with cluster capabilities:
   HelloRequest{
     Version: "1.0",
     AuthMethod: "cluster",
     Token: cluster_secret,
     Client: "muninn-node",
     Capabilities: ["replication", "election", "gossip"],
   }
4. Receive HELLO_OK with cluster state:
   HelloResponse{
     ServerVersion: "0.x.y",
     ClusterState: {
       CortexID: "muninn-02",
       Epoch: 42,
       FencingToken: 17,
       Members: [...],
     },
   }
5. If a Cortex exists:
   - Register as Lobe
   - Begin replication stream from Cortex
   - If fresh node (no data): trigger Initial State Transfer (Section 7)
6. If no Cortex exists:
   - Participate in election
```

### 6.3 Gossip Protocol

Once connected, nodes exchange cluster state via a lightweight gossip protocol piggybacked on MBP heartbeat frames:

```
Every 2 seconds, each node sends to 2 random peers:
  GOSSIP{
    sender_id: "muninn-01",
    epoch: 42,
    members: [
      {id: "muninn-01", addr: "10.0.1.10:8474", role: primary, last_seq: 12345, last_seen: now},
      {id: "muninn-02", addr: "10.0.1.11:8474", role: replica, last_seq: 12340, last_seen: now-200ms},
      {id: "muninn-03", addr: "10.0.1.12:8474", role: replica, last_seq: 12338, last_seen: now-400ms},
    ]
  }
```

Gossip convergence: In a 10-node cluster, all nodes learn about a new member within ~4 gossip rounds (8 seconds). This is acceptable because node joins/leaves are rare events.

### 6.4 Graceful Leave

```
1. Node sends LEAVE{node_id} to Cortex
2. Cortex removes node from cluster membership
3. Cortex gossips updated membership
4. If leaving node was Cortex:
   - Cortex releases lease first
   - Remaining nodes detect absence and elect new Cortex
```

### 6.5 Ungraceful Leave (Crash)

Handled by failure detection (Section 5.2). After 3 missed heartbeats, the node is marked SDOWN. After quorum agreement, it's removed from the active member list (but remembered for 24 hours in case it comes back -- "resurrection window").

---

## 7. Initial State Transfer (Cognitive Snapshot)

### The Problem

A fresh Lobe joining a running cluster has an empty Pebble store. It needs the full cognitive state:
- All engrams (concept, content, tags, metadata)
- All association edges with current Hebbian weights
- All decay scores and stability values
- All confidence scores
- FTS index data
- HNSW vector index data
- Cognitive worker metadata (last decay timestamp, etc.)

This is not just "data" -- it's a living cognitive state at a point in time.

### Design: Snapshot + Replication Log Tail

Inspired by Redis RDB + AOF approach:

```
Phase 1: Cognitive Snapshot
  Cortex creates a consistent point-in-time snapshot of its Pebble store
  Snapshot is a sequential stream of key-value pairs (Pebble SST format or raw KV stream)
  Snapshot includes a "snapshot sequence number" (the replication log seq at snapshot time)

Phase 2: Stream to new Lobe
  Cortex streams the snapshot to the new Lobe over MBP (using streaming frames)
  New Lobe writes all KV pairs to its local Pebble
  Lobe acknowledges completion

Phase 3: Catch up from replication log
  Lobe requests replication entries from snapshot_seq onward
  Cortex sends the replication log tail (entries accumulated during the snapshot transfer)
  Lobe applies entries via Applier
  Lobe is now caught up and enters normal replication mode
```

```
Timeline:

t=0     Lobe joins, sends JOIN_REQUEST
t=0.1   Cortex starts snapshot at seq=50000
t=5     Snapshot completes (5 seconds for a medium DB)
        During this time, writes continued, seq is now 50200
t=5.1   Cortex streams snapshot to Lobe
t=30    Snapshot transfer completes over network
t=30.1  Lobe requests entries since seq=50000
t=30.5  Lobe applies 200+ entries, catches up to seq=50250
t=30.6  Lobe enters normal streaming replication
```

### Snapshot Implementation

Pebble supports `db.NewSnapshot()` which provides a point-in-time consistent view. The snapshot protocol:

```go
type SnapshotHeader struct {
    SnapshotSeq uint64 `msgpack:"snapshot_seq"`
    NodeID      string `msgpack:"node_id"`
    TotalKeys   uint64 `msgpack:"total_keys"`  // estimate for progress bars
    Timestamp   int64  `msgpack:"timestamp"`
}

type SnapshotChunk struct {
    Pairs     []KVPair `msgpack:"pairs"`
    ChunkNum  uint32   `msgpack:"chunk_num"`
    LastChunk bool     `msgpack:"last_chunk"`
}

type KVPair struct {
    Key   []byte `msgpack:"key"`
    Value []byte `msgpack:"value"`
}
```

**Chunk size:** 1MB per chunk. A 10GB database transfers in ~10,000 chunks. At 100MB/s network throughput, transfer takes ~100 seconds.

**Resource protection:** The Cortex limits concurrent snapshot transfers to 1 (configurable). Snapshot reads use Pebble's snapshot isolation, so they don't block writes. However, they do consume disk I/O, so we rate-limit the scan to avoid starving the write path.

### Partial Resync (Pebble-Backed Replication Log)

For short disconnections (network blip, rolling restart), a full snapshot is wasteful. The Pebble-backed `ReplicationLog` serves as the resync source directly -- **no in-memory ring buffer is needed**.

**Design:** The `ReplicationLog` already stores all entries in Pebble with sequential keys (`0x19 | seq_be64`). The existing `ReadSince(afterSeq, batchSize)` method provides efficient range reads. This eliminates the need for a separate `ReplicationBacklog` struct, saving ~100MB of in-memory allocation and removing the crash-recovery confusion of an in-memory ring buffer (which would lose state on crash).

When a Lobe reconnects:
1. Lobe sends its `lastApplied` sequence number
2. Cortex calls `repLog.ReadSince(lastApplied, batchSize)` to fetch entries
   - **If entries exist:** Send them to the Lobe (partial resync)
   - **If `lastApplied` is below the Pebble log's oldest entry** (after pruning begins in Phase 2+): Return `FULL_RESYNC_REQUIRED` and trigger snapshot transfer (full resync)
3. Lobe applies entries and continues normal replication

**Advantages over in-memory ring buffer:**
- **Crash-safe:** Pebble data survives restarts. No "cold start" after Cortex restart where the backlog is empty.
- **Unbounded window in Phase 1:** Since Prune() is not called in Phase 1, any Lobe can resync from any point.
- **Memory-efficient:** Zero in-memory allocation for the backlog. Pebble manages its own block cache.
- **Simpler code:** One data structure (ReplicationLog) instead of two (ReplicationLog + ReplicationBacklog).

---

## 8. Replication Transport

### Decision: MBP (Muninn Binary Protocol) for Replication

**Why not gRPC?** gRPC adds a dependency (protobuf), requires HTTP/2 framing overhead, and does not integrate with MuninnDB's existing MBP connection handling. Since MBP already has streaming support (`FlagStreaming`, `FlagLastFrame`), compression (`FlagCompressed`), and correlation IDs, building replication on MBP is natural and avoids a second connection protocol.

**Why not a separate TCP channel?** Additional ports increase firewall complexity. MBP's multiplexed frame design already supports multiple concurrent streams over a single TCP connection.

### 8.1 Replication Frame Types

New MBP frame types for cluster communication:

```go
const (
    // Replication stream
    TypeReplEntry     uint8 = 0x20  // single replication entry
    TypeReplBatch     uint8 = 0x21  // batch of replication entries
    TypeReplAck       uint8 = 0x22  // replica acknowledges seq
    TypeReplNack      uint8 = 0x23  // replica requests resend from seq

    // Snapshot transfer
    TypeSnapHeader    uint8 = 0x24  // snapshot header
    TypeSnapChunk     uint8 = 0x25  // snapshot data chunk
    TypeSnapAck       uint8 = 0x26  // chunk acknowledged
    TypeSnapComplete  uint8 = 0x27  // snapshot transfer complete

    // Cognitive side-effect forwarding
    TypeCogForward    uint8 = 0x28  // forward cognitive side effects to Cortex
    TypeCogAck        uint8 = 0x29  // Cortex acknowledges cognitive forward

    // Cluster protocol
    TypeVoteRequest   uint8 = 0x30
    TypeVoteResponse  uint8 = 0x31
    TypeCortexClaim   uint8 = 0x32
    TypeSDown         uint8 = 0x33
    TypeODown         uint8 = 0x34
    TypeGossip        uint8 = 0x35
    TypeJoinRequest   uint8 = 0x36
    TypeJoinResponse  uint8 = 0x37
    TypeLeave         uint8 = 0x38
)
```

### 8.2 Replication Stream Protocol

The Cortex pushes entries to connected Lobes over persistent MBP connections:

```
Cortex                          Lobe
  |                               |
  |--- ReplBatch{entries} ------->|
  |                               | Apply entries
  |<-- ReplAck{last_seq} --------|
  |                               |
  |--- ReplBatch{entries} ------->|
  |                               | Apply entries
  |<-- ReplAck{last_seq} --------|
  |                               |
```

**Push model, not poll:** The current `Streamer` uses a 100ms poll loop. The new design replaces this with an event-driven push. When the `ReplicationLog.Append()` completes, it notifies all connected streamers via a channel. Streamers immediately push the new entry to their Lobe. This reduces replication latency from 100ms (worst case poll delay) to <1ms.

```go
// Replace polling streamer with push-based notification
type ReplicationLog struct {
    // ... existing fields ...
    subscribers []chan struct{} // notified on each Append
}

func (l *ReplicationLog) Append(...) {
    // ... existing logic ...
    // Notify subscribers
    for _, ch := range l.subscribers {
        select {
        case ch <- struct{}{}:
        default: // non-blocking
        }
    }
}
```

**Atomic entry + seq counter write (bug fix):** In the current `ReplicationLog.Append()`, the entry and sequence counter are written as separate Pebble operations, creating a crash window where the entry is persisted but the seq counter is not updated. Fix: write both in the same Pebble batch:

```go
func (l *ReplicationLog) Append(op WALOp, key, value []byte) (uint64, error) {
    l.mu.Lock()
    defer l.mu.Unlock()
    // ... existing seq increment and marshal logic ...

    batch := l.db.NewBatch()
    batch.Set(entryKey, data, nil)
    batch.Set(seqCounterKey(), seqBuf, nil)
    if err := batch.Commit(pebble.Sync); err != nil {
        l.seq-- // rollback
        batch.Close()
        return 0, err
    }
    batch.Close()
    return l.seq, nil
}
```

### 8.3 Cognitive Side-Effect Forwarding

When a Lobe serves an activation query, it generates cognitive side effects that must reach the Cortex:

```go
type CognitiveSideEffect struct {
    QueryID        string            `msgpack:"query_id"`
    OriginNodeID   string            `msgpack:"origin_node_id"`
    Timestamp      int64             `msgpack:"timestamp"`
    CoActivations  []CoActivationRef `msgpack:"co_activations,omitempty"`
    AccessedIDs    [][16]byte        `msgpack:"accessed_ids,omitempty"`
    WS             [8]byte           `msgpack:"ws"`  // vault workspace prefix
}

type CoActivationRef struct {
    ID    [16]byte `msgpack:"id"`
    Score float64  `msgpack:"score"`
}
```

The Lobe sends this as a `TypeCogForward` MBP frame. The Cortex receives it and:
1. Submits co-activations to its `HebbianWorker`
2. Submits accessed IDs to its `DecayWorker`
3. Submits confidence updates to its `ConfidenceWorker`

These are best-effort and non-blocking. If the Cortex is under heavy load, cognitive forwards from replicas may be dropped (the worker channels are bounded). This is acceptable -- missing one co-activation event slightly reduces Hebbian learning fidelity, but the next activation of the same engrams will correct it.

### 8.4 Wire Efficiency

Replication batching for throughput:

- **Batch window:** 5ms or 100 entries, whichever comes first
- **Compression:** `FlagCompressed` (zstd) for batches > 1KB
- **Backpressure:** If a Lobe is slow to ACK, the Cortex buffers up to 10,000 unACK'd entries before pausing the stream (TCP-style flow control)
- **Keepalive:** MBP `PING`/`PONG` every 1 second doubles as heartbeat and connection liveness check

---

## 9. Operational Simplicity

### 9.1 Minimal 3-Node Configuration

```yaml
# /etc/muninn/muninn.yaml on node 1
data_dir: /var/lib/muninn
cluster:
  enabled: true
  node_id: "muninn-01"
  bind_addr: "10.0.1.10:8474"
  seeds:
    - "10.0.1.11:8474"
    - "10.0.1.12:8474"
  cluster_secret: "your-shared-secret-here"
```

```yaml
# /etc/muninn/muninn.yaml on node 2
data_dir: /var/lib/muninn
cluster:
  enabled: true
  node_id: "muninn-02"
  bind_addr: "10.0.1.11:8474"
  seeds:
    - "10.0.1.10:8474"
    - "10.0.1.12:8474"
  cluster_secret: "your-shared-secret-here"
```

```yaml
# /etc/muninn/muninn.yaml on node 3
data_dir: /var/lib/muninn
cluster:
  enabled: true
  node_id: "muninn-03"
  bind_addr: "10.0.1.12:8474"
  seeds:
    - "10.0.1.10:8474"
    - "10.0.1.11:8474"
  cluster_secret: "your-shared-secret-here"
```

That's it. Three files, nearly identical. No etcd to deploy. No Consul to configure. No TLS certificates to generate for inter-node communication (TLS is optional but recommended for production).

**Environment variable override:** Every config field can be set via env var for container deployments:

```bash
MUNINN_CLUSTER_ENABLED=true
MUNINN_CLUSTER_NODE_ID=muninn-01
MUNINN_CLUSTER_BIND_ADDR=10.0.1.10:8474
MUNINN_CLUSTER_SEEDS=10.0.1.11:8474,10.0.1.12:8474
MUNINN_CLUSTER_SECRET=your-shared-secret-here
```

### 9.2 Startup Behavior

```
$ muninn start --config /etc/muninn/muninn.yaml

[INFO] MuninnDB v0.x.y starting
[INFO] cluster mode enabled, node_id=muninn-01
[INFO] connecting to seed 10.0.1.11:8474... connected
[INFO] connecting to seed 10.0.1.12:8474... connected
[INFO] cluster state: no cortex detected
[INFO] participating in election, epoch=1
[INFO] elected as cortex (received 2/3 votes)
[INFO] cognitive workers started
[INFO] replication stream ready, waiting for lobes
[INFO] MBP listening on :8474
[INFO] REST listening on :8475
[INFO] MCP listening on :8750
[INFO] gRPC listening on :8477
[INFO] UI listening on :8476
[INFO] MuninnDB ready (role=cortex, cluster_size=3)
```

On the other two nodes (which start and lose the election):

```
$ muninn start --config /etc/muninn/muninn.yaml

[INFO] MuninnDB v0.x.y starting
[INFO] cluster mode enabled, node_id=muninn-02
[INFO] connecting to seed 10.0.1.10:8474... connected
[INFO] cluster state: cortex=muninn-01 at 10.0.1.10:8474
[INFO] registering as lobe
[INFO] initial state transfer: receiving snapshot (seq=0, fresh node)
[INFO] snapshot transfer: 45,231 engrams, 12.4 MB [===========] 100%
[INFO] snapshot applied, catching up from replication log...
[INFO] replication caught up (seq=45231, lag=0)
[INFO] MuninnDB ready (role=lobe, cortex=muninn-01, cluster_size=3)
```

### 9.3 Failure and Recovery (Operator Experience)

```
Scenario: Cortex node crashes

t=0s    muninn-01 (cortex) process killed
        [No operator action needed]

t=3.5s  muninn-02 log:
        [WARN] cortex muninn-01 unreachable (3 missed heartbeats)
        [INFO] SDOWN declared for muninn-01
        [INFO] ODOWN confirmed (2/2 live nodes agree)
        [INFO] election started, epoch=2
        [INFO] elected as cortex (received 2/2 votes, highest seq=45680)
        [INFO] cognitive workers started
        [INFO] promoted to cortex successfully

        muninn-03 log:
        [WARN] cortex muninn-01 unreachable (3 missed heartbeats)
        [INFO] voted for muninn-02 (epoch=2, seq=45680)
        [INFO] new cortex: muninn-02

        [Cluster continues operating normally. No data loss for
         entries that were replicated before the crash.]

Later: Operator restarts muninn-01

        muninn-01 log:
        [INFO] MuninnDB v0.x.y starting
        [INFO] cluster mode enabled, node_id=muninn-01
        [INFO] connecting to seed 10.0.1.11:8474... connected
        [INFO] cluster state: cortex=muninn-02 at 10.0.1.11:8474
        [INFO] rejoining as lobe (was previously cortex, epoch=1 < current=2)
        [INFO] partial resync from replication log (last_applied=45650)
        [INFO] caught up in 0.3s (applied 230 entries)
        [INFO] MuninnDB ready (role=lobe, cortex=muninn-02, cluster_size=3)
```

### 9.4 Health and Metrics Endpoints

REST API (`/v1/cluster/*`):

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/cluster/info` | GET | Full cluster state: members, roles, epochs, connectivity |
| `/v1/cluster/health` | GET | Returns 200 if this node is healthy, 503 if degraded |
| `/v1/cluster/topology` | GET | ASCII art topology diagram (see below) |
| `/v1/cluster/replication/status` | GET | Per-lobe replication lag, throughput, last ACK time |
| `/v1/cluster/replication/lag` | GET | Summary lag stats across all lobes |
| `/v1/cluster/failover` | POST | Manual failover trigger (admin only) |
| `/v1/cluster/cognitive/status` | GET | Cognitive worker status per node (Hebbian, Decay, etc.) |
| `/v1/cluster/cognitive/drift` | GET | Cognitive state drift between cortex and lobes |

**Topology endpoint example response:**

```
GET /v1/cluster/topology

  MuninnDB Cluster (epoch=2, config_epoch=3)

  +------------------+     replication     +------------------+
  | muninn-02        |-------------------->| muninn-01        |
  | CORTEX           |     lag: 0          | LOBE             |
  | seq: 45880       |                     | seq: 45880       |
  | uptime: 2h14m    |-------------------->| uptime: 1h02m    |
  | cognitive: active |     lag: 0         | cognitive: idle   |
  +------------------+                     +------------------+
          |
          |     replication
          |     lag: 2
          v
  +------------------+
  | muninn-03        |
  | LOBE             |
  | seq: 45878       |
  | uptime: 2h14m    |
  | cognitive: idle   |
  +------------------+
```

**Prometheus metrics** (exposed on `/metrics`):

```
# Cluster
muninn_cluster_role{node_id="muninn-01"} 1  # 1=primary, 2=replica
muninn_cluster_epoch 2
muninn_cluster_fencing_token 17
muninn_cluster_members_total 3
muninn_cluster_healthy_members 3

# Replication
muninn_replication_lag_entries{peer="muninn-03"} 2
muninn_replication_lag_seconds{peer="muninn-03"} 0.012
muninn_replication_throughput_entries_per_sec 1450
muninn_replication_log_oldest_seq 1
muninn_replication_log_current_seq 45880
muninn_replication_full_resyncs_total 1
muninn_replication_partial_resyncs_total 3

# Cognitive (per-node)
muninn_cognitive_hebbian_processed_total 892341
muninn_cognitive_hebbian_state 0  # 0=active, 1=idle, 2=dormant
muninn_cognitive_decay_processed_total 1203456
muninn_cognitive_forwarded_effects_total 12840  # from lobes
muninn_cognitive_forwarded_effects_dropped_total 3  # dropped due to backpressure

# Failover
muninn_failover_total 1
muninn_failover_duration_seconds 3.5
muninn_failover_last_epoch 2
```

---

## 10. Cognitive-Specific Cluster Behaviors

### 10.1 Cognitive Workers on Lobes

| Worker | Runs on Cortex | Runs on Lobe | Notes |
|--------|---------------|-------------|-------|
| HebbianWorker | YES (primary) | NO | Only the Cortex produces Hebbian updates. Lobes forward co-activations. |
| DecayWorker | YES (primary) | NO | Decay is computed centrally. The replicated decay scores are applied on lobes via replication. |
| ContradictWorker | YES (primary) | NO | Contradiction detection runs centrally. |
| ConfidenceWorker | YES (primary) | NO | Bayesian updates are centralized. |
| TriggerSystem | YES (primary) | PARTIAL | Lobes evaluate triggers locally for responsiveness, but canonical trigger state lives on Cortex. |
| ActivityTracker | YES (primary) | FORWARD | Lobes forward access counts to Cortex. |

**Why not run cognitive workers on Lobes?**

If Lobes ran their own Hebbian workers, they would produce cognitive writes that conflict with the Cortex's cognitive writes. Two nodes computing `w_new = w_old * (1+lr)^signal` with different `w_old` values (due to replication lag) produce divergent weights. Centralizing cognitive computation on the Cortex ensures a single linearizable timeline of cognitive state mutations.

#### 10.1.1 Nil Worker Guard — Double-Apply Prevention

**Architectural decision:** When `muninn start` initializes the Engine on a Lobe node, cognitive workers (HebbianWorker, DecayWorker, ConfidenceWorker, ContradictWorker) are passed as `nil` in the `engine.NewEngine()` call.

**Codebase verification (confirmed):** The existing Engine already checks for nil workers before submitting work:
- `if e.hebbianWorker != nil` — guarded in `Activate()` and trigger callbacks
- `if e.decayWorker != nil` — guarded in `Activate()` and trigger callbacks
- `if e.contradictWorker != nil` — guarded in `Write()` and `Link()`
- `if e.confidenceWorker != nil` — guarded in `Write()`, `Activate()`, and `Link()`

This means passing `nil` workers is safe today — no additional nil guards are needed in the Engine.

**Phase 1 behavior:** Lobes serve activations but with nil workers. Cognitive side effects (Hebbian, Decay, Confidence, Contradiction) are silently dropped. No double-apply is possible because the workers simply do not exist on Lobe nodes.

**Phase 2 behavior:** Lobes still have nil workers locally. The activation path in the Engine is extended to collect side effects (co-activated engram IDs, accessed IDs) and forward them to the Cortex via `TypeCogForward` MBP frames instead of submitting to local workers.

**Implementation requirement:** The `server.go` construction logic must branch on `NodeRole`:
```go
if clusterCfg.Enabled && coordinator.Role() != replication.RolePrimary {
    // Lobe: nil cognitive workers — side effects dropped in Phase 1,
    // forwarded to Cortex in Phase 2
    eng = engine.NewEngine(store, authStore, ftsIdx, actEngine, trigSystem,
        nil, nil, nil, nil, embedder)
} else {
    // Cortex or standalone: full cognitive workers
    eng = engine.NewEngine(store, authStore, ftsIdx, actEngine, trigSystem,
        hebbianWorkerImpl, decayWorkerImpl,
        contradictWorkerImpl.Worker, confidenceWorkerImpl.Worker, embedder)
}
```

### 10.2 Cortex Crash During Hebbian Update

**Scenario:** The Cortex's HebbianWorker is processing a batch of 100 co-activation pairs. It has written 60 weight updates to Pebble. The process crashes.

**Guarantee:** The 60 written updates are persisted in Pebble (they were committed via `batch.Commit(nil)`). The remaining 40 are lost.

**Impact:** 40 co-activation pairs miss one learning event. This is equivalent to a human briefly forgetting a connection and re-learning it on the next exposure. The Hebbian learning rate (`0.01`) is intentionally small, so missing one event has negligible impact on association weights.

**Recovery:** On the next activation that co-activates those same engrams, the missed updates are effectively recomputed. No manual intervention is needed.

**Improvement for Phase 2:** Make `HebbianWorker.processBatch` write all pair updates in a single Pebble `Batch` so they are atomic. Either all 100 updates commit or none do. This is straightforward since `UpdateAssocWeight` already uses Pebble writes.

### 10.3 Cognitive State Queries from Any Node

**Cold cognitive queries** (read-only, no side effects) can be served from any node:

```
GET /v1/engrams/{id}/cognitive
{
  "relevance": 0.87,
  "confidence": 0.92,
  "stability": 28.4,
  "access_count": 47,
  "last_access": "2026-02-23T14:30:00Z",
  "associations": [
    {"target_id": "...", "weight": 0.73, "rel_type": 1},
    ...
  ]
}
```

This is a cold read -- it returns the current cognitive state without triggering any updates. Safe on any Lobe.

**Activation queries** (cognitive side effects) follow the write-forwarding path described in Section 3.

### 10.4 Replication Entry Classification

Not all replication entries are equal. MuninnDB classifies them for priority handling:

```go
type WALOp uint8

const (
    OpSet       WALOp = 1  // data write (engram create/update)
    OpDelete    WALOp = 2  // data delete (engram forget)
    OpBatch     WALOp = 3  // atomic multi-op
    OpCognitive WALOp = 4  // cognitive state update (Hebbian, Decay, Confidence)
    OpIndex     WALOp = 5  // index update (FTS, HNSW)
    OpMeta      WALOp = 6  // cluster metadata (cognitive worker state, etc.)
)
```

This classification enables:
- **Priority replication:** Data writes (`OpSet`, `OpDelete`) are replicated first. Cognitive updates (`OpCognitive`) can lag slightly without affecting correctness.
- **Selective replication for Observers:** An Observer that only needs raw data can filter out `OpCognitive` entries, reducing replication bandwidth.
- **Metrics granularity:** Track replication lag separately for data vs. cognitive updates.

---

## 11. Enterprise "Wow Factor" Features

These are features that make enterprise evaluators say "I can't believe they thought of that." Each is stable, implementable, and not a gimmick.

### 11.1 Cognitive Consistency Score (CCS)

A real-time metric exposed per cluster that answers: "How consistent is the cognitive state across all nodes?"

```
GET /v1/cluster/cognitive/consistency

{
  "score": 0.9987,
  "details": {
    "hebbian_weight_divergence": 0.0008,   // max weight delta between cortex and worst lobe
    "decay_score_divergence": 0.0003,       // max relevance delta
    "confidence_divergence": 0.0002,        // max confidence delta
    "replication_lag_ms": 12,               // current worst-case lag
    "cognitive_forwarding_loss_rate": 0.001 // % of forwarded effects dropped
  },
  "assessment": "excellent"  // "excellent" | "good" | "degraded" | "critical"
}
```

This is sampled periodically (every 30 seconds) by having each Lobe report a hash of its cognitive state for a random sample of engrams. The Cortex compares these hashes to its own state. No other database provides this.

### 11.2 Graceful Cognitive Handoff

During a **planned** failover (e.g., rolling upgrade), the Cortex performs a "cognitive handoff" instead of an abrupt demotion:

```
1. Cortex enters DRAINING state
   - Stops accepting new writes
   - Flushes all cognitive worker queues (HebbianWorker, DecayWorker, etc.)
   - Waits for all Lobes to catch up (replication lag = 0)
2. Cortex sends HANDOFF{target_node_id} to the designated successor
3. Successor starts cognitive workers
4. Successor sends HANDOFF_ACK
5. Cortex releases lease and demotes to Lobe
6. Successor becomes Cortex with zero data loss and zero cognitive gap
```

**Operator command:**

```bash
muninn cluster failover --target muninn-02 --graceful

[INFO] initiating graceful cognitive handoff to muninn-02
[INFO] draining cognitive workers... done (flushed 342 pending events)
[INFO] waiting for replication convergence... done (all lobes at seq=89012)
[INFO] handoff sent to muninn-02
[INFO] handoff acknowledged, muninn-02 is now cortex
[INFO] demoted to lobe successfully
[INFO] zero cognitive events lost during handoff
```

### 11.3 Cognitive Telemetry Dashboard

The Web UI (`:8476`) gets a real-time cluster view showing:

- Live replication stream visualization (entries flowing from Cortex to Lobes)
- Cognitive worker activity heatmaps (which workers are active/idle/dormant on each node)
- Association weight evolution graphs (Hebbian learning in real-time)
- Failover event timeline with detailed event log
- Cognitive consistency score gauge

This leverages the existing WebSocket broadcast infrastructure (`uiSrv.Broadcast`) and the `logging.RingBuffer`.

### 11.4 Replication Lag Alarm with Cognitive Context

When replication lag exceeds a threshold, the alarm includes cognitive context:

```json
{
  "alarm": "replication_lag_high",
  "lobe": "muninn-03",
  "lag_entries": 5000,
  "lag_seconds": 3.2,
  "cognitive_impact": {
    "hebbian_updates_pending": 847,
    "decay_resets_pending": 212,
    "estimated_weight_drift": 0.003,
    "assessment": "Lobe muninn-03 cognitive state is 3.2s behind. Activation queries on this lobe will use slightly stale Hebbian weights (est. 0.3% drift). No action required unless lag persists."
  }
}
```

No other database tells you what replication lag *means* for your query results.

### 11.5 Automatic Cognitive Reconciliation

After a partition heals or a Lobe catches up from a large lag, the cluster runs an automatic reconciliation process:

1. Cortex and Lobe exchange cognitive checksums for the top-K most-accessed engrams
2. Any divergence triggers a targeted cognitive state sync (not a full resync -- just the divergent weights/scores)
3. The reconciliation is logged with full detail for auditability

This ensures that even after network events, the cognitive state converges across all nodes within one reconciliation cycle.

---

## 12. Phased Implementation Roadmap

### Phase 1: Foundation (Minimum Viable Cluster)

**Goal:** A 3-node cluster that an operator can spin up with minimal config, with automatic failover and basic replication. This is the "it works, it's stable, ship it" phase.

**Deliverables:**

1. **Cluster configuration parsing** (`internal/config/cluster.go`)
   - YAML + env var support for cluster settings
   - Node ID, bind address, seeds, cluster secret

2. **MBP cluster frame types**
   - Add `TypeReplEntry`, `TypeReplBatch`, `TypeReplAck` to `internal/transport/mbp/frame.go`
   - Add `TypeVoteRequest`, `TypeVoteResponse`, `TypeCortexClaim` frame types
   - Add `TypeGossip`, `TypeSDown`, `TypeODown` frame types

3. **Network-aware LeaseBackend** (`internal/replication/pebble_lease.go`)
   - Replace `MemoryLeaseBackend` with a Pebble-backed lease that uses MBP heartbeats for distributed agreement
   - Implement the MSP protocol (heartbeat, SDOWN, ODOWN, election)

4. **Push-based Streamer** (`internal/replication/network_streamer.go`)
   - Replace poll-based `Streamer` with push-based notification
   - MBP frame serialization for replication entries
   - Per-Lobe connection management with reconnection

5. **Persistent lastApplied** (`internal/replication/applier.go`)
   - Persist `lastApplied` to Pebble key `0x19 | 0x02`
   - Load on startup for crash recovery

6. **Partial resync via Pebble log** (no new file needed)
   - Resync reads directly from the Pebble-backed `ReplicationLog.ReadSince()`
   - No in-memory ring buffer. The Pebble log IS the backlog.
   - If `lastApplied` is below the oldest Pebble entry (Phase 2+ after pruning): return `FULL_RESYNC_REQUIRED`

7. **Server integration** (`cmd/muninn/server.go`)
   - Wire the replication `Coordinator` into the server startup
   - Start MSP goroutine when cluster mode is enabled
   - Route writes through replication log

8. **REST handlers** (`internal/transport/rest/replication_handlers.go`)
   - Replace stub handlers with real implementations
   - Add `/v1/cluster/info` and `/v1/cluster/health`

9. **Basic Prune fix**
   - Replace O(N) sequential loop in `ReplicationLog.Prune()` with `pebble.Batch.DeleteRange()`

10. **ReadSince race fix**
    - Fix the lock/race on `l.seq` in `ReadSince` by reading `l.seq` under the lock and using the captured value

11. **Phase 1 Node Join Constraint**
    - In Phase 1, the Pebble-backed `ReplicationLog` is **NEVER pruned**. The `Prune()` function exists but is not called.
    - This means: any node joining at any time can request entries from `lastApplied+1` from the Pebble log.
    - The Pebble log grows unboundedly in Phase 1 — this is **intentional and documented**.
    - Operators should plan for ~100 bytes per replication entry x total writes.
    - A warning is logged at startup if cluster mode is enabled: `"Log pruning is disabled in Phase 1. Monitor disk usage."`
    - Phase 2 adds snapshot-based initial state transfer, after which pruning is safe.

**NOT in Phase 1:**
- Initial state transfer (new nodes must start empty and catch up from replication log)
- Cognitive side-effect forwarding (activations on Lobes don't forward to Cortex)
- Graceful handoff
- Sentinel nodes
- Observer nodes
- Cognitive consistency scoring
- TLS between nodes

**Timeline estimate:** 4-6 weeks

---

### Phase 2: Cognitive Cluster

**Goal:** Cognitive side-effect forwarding works. Activations on Lobes are fully functional. Initial state transfer allows new nodes to join a running cluster.

**Deliverables:**

1. **Cognitive side-effect forwarding**
   - `TypeCogForward` / `TypeCogAck` MBP frames
   - Lobe activation engine collects co-activations and forwards to Cortex
   - Cortex dispatches to HebbianWorker, DecayWorker, ConfidenceWorker

2. **Initial state transfer (Cognitive Snapshot)**
   - Pebble snapshot + streaming over MBP
   - Snapshot header, chunked transfer, catch-up from replication log
   - Rate limiting to protect Cortex I/O

3. **Cognitive worker metadata persistence**
   - Persist `HebbianWorker.lastDecayAt` to Pebble
   - Persist cognitive worker stats for dashboard

4. **Atomic Hebbian batch writes**
   - Refactor `HebbianWorker.processBatch` to use a single Pebble Batch for all pair updates

5. **WALOp classification**
   - Add `OpCognitive`, `OpIndex`, `OpMeta` to distinguish replication entry types
   - Priority replication (data before cognitive)

6. **Graceful cognitive handoff**
   - `DRAINING` state on Cortex
   - Flush cognitive workers, wait for convergence, hand off

7. **Replication lag alarm with cognitive context**
   - Compute cognitive impact of lag
   - Expose via REST and Prometheus

**Timeline estimate:** 4-6 weeks after Phase 1

---

### Phase 3: Enterprise Polish

**Goal:** Production-ready for enterprise deployment. Security, observability, operational tooling.

**Deliverables:**

1. **Mutual TLS between nodes**
   - Auto-generated certificates with cluster CA
   - Certificate rotation

2. **Sentinel and Observer roles**
   - Lightweight sentinel nodes for quorum
   - Observer nodes for analytics / cross-region standby

3. **Cognitive Consistency Score (CCS)**
   - Periodic checksum comparison
   - REST endpoint and Prometheus metrics

4. **Automatic cognitive reconciliation**
   - Post-partition reconciliation of divergent cognitive state
   - Targeted sync (not full resync)

5. **Web UI cluster dashboard**
   - Real-time replication stream visualization
   - Cognitive worker heatmaps
   - Failover event timeline
   - Cluster topology diagram

6. **CLI cluster management commands**
   ```bash
   muninn cluster info
   muninn cluster status
   muninn cluster failover --target <node> [--graceful]
   muninn cluster add-node <addr>
   muninn cluster remove-node <node-id>
   muninn cluster rebalance
   ```

7. **Comprehensive integration tests**
   - Multi-node test harness using `t.TempDir()` Pebble instances
   - Simulated network partitions
   - Failover correctness tests
   - Cognitive consistency tests

**Timeline estimate:** 6-8 weeks after Phase 2

---

### Phase 4+: Future (Do NOT Build Until Phase 3 is Proven)

These are ideas to track but NOT implement until the foundation is battle-tested:

- **Cognitive Mesh (Option C):** Multi-writer Hebbian updates with CRDT-style conflict resolution. Requires extensive research into association weight convergence properties.
- **Read replicas with local cognitive inference:** Lobes run a lightweight local Hebbian approximation that is periodically corrected by the Cortex. Reduces forwarding traffic.
- **Cross-region replication:** Async replication with configurable consistency SLA per region.
- **Sharding:** Partition the cognitive graph by vault, with different Cortex nodes owning different vaults. True horizontal scale.
- **Cascade replication:** Lobes replicate to sub-Lobes, reducing Cortex fan-out.
- **Kubernetes operator:** Custom resource definitions for MuninnDB clusters, auto-scaling, rolling upgrades.

---

## Appendix: Current Codebase Gaps

Concrete issues in the existing `internal/replication/` package that Phase 1 must address:

### A.1 Orphaned Package

Nothing in `cmd/muninn/server.go` references the replication package. The `Coordinator`, `LeaderElector`, `ReplicationLog`, `Applier`, and `Streamer` are entirely disconnected from the running server.

**Fix:** Wire `Coordinator` into `runServer()` behind a `cluster.enabled` config flag.

### A.2 In-Memory lastApplied

`Applier.lastApplied` resets to 0 on restart, causing full replay of the replication log.

**Fix:** Persist to Pebble key `0x19 | 0x02 | "last_applied"`. Load on `NewApplier()`.

### A.3 MemoryLeaseBackend is Single-Process

Cannot coordinate across nodes. This is documented but must be replaced.

**Fix:** Implement `PebbleLeaseBackend` that uses MBP heartbeats to distribute lease state. The lease is effectively an agreement among all MSP participants.

### A.4 Streamer is In-Process Only

Pushes to a Go channel. Cannot send entries over the network.

**Fix:** `NetworkStreamer` wraps replication entries in MBP frames and sends over persistent TCP connections to Lobes.

### A.5 Prune is O(N)

`ReplicationLog.Prune()` iterates from `seq=0` to `untilSeq`, issuing individual deletes.

**Fix:** Use `pebble.Batch.DeleteRange(startKey, endKey)` for a single range delete operation.

### A.6 ReadSince Race

After releasing `l.mu`, `ReadSince` uses `l.seq` to compute iterator bounds. Another goroutine could increment `l.seq` between the unlock and the iterator creation.

**Fix:** Capture `currentSeq := l.seq` while holding the lock, then use `currentSeq` for the iterator upper bound.

### A.7 REST Handlers are Stubs

`replication_handlers.go` returns hardcoded values. All three handlers (`HandleReplicationStatus`, `HandleReplicationLag`, `HandlePromoteReplica`) need real implementations that delegate to the `Coordinator`.

### A.8 Missing Cluster Auth

The `HelloRequest` has an `AuthMethod` field but no cluster-level authentication. Nodes need a shared secret to prevent unauthorized nodes from joining.

**Fix:** Add `cluster_secret` config. On HELLO, verify HMAC of the secret. Reject connections with invalid secrets.

---

## Design Decision Summary

| Decision | Choice | Key Reason |
|----------|--------|------------|
| Topology | B+D Hybrid (Cognitive Primary + Write-Forwarding) | Single linearizable cognitive timeline; replicas still useful |
| Leader election | Self-contained MSP (no etcd) | Zero external dependencies, Redis Sentinel model |
| Discovery | Static seeds + gossip | Simple bootstrap, dynamic membership |
| Snapshot | Pebble snapshot + MBP stream | Consistent point-in-time, uses existing protocol |
| Replication transport | MBP (not gRPC) | Already exists, streaming support, single port |
| Replication model | Push (not poll) | Sub-millisecond latency vs 100ms poll |
| Partial resync | Pebble-backed log (no in-memory backlog) | Prevents full resync on short disconnections; crash-safe |
| Cognitive workers on Lobes | Off (forward to Cortex) | Prevents weight divergence |
| Phase 1 scope | No snapshot, no forwarding, no TLS | Ship stable foundation first |

---

## Appendix B: Resolved Issues

This section documents issues found during adversarial review and their resolutions. All fixes have been applied inline to the relevant sections above.

### Resolved C1: Double-Apply of Cognitive Side Effects

**Issue:** The Engine unconditionally fires cognitive workers on every activation regardless of node role. If Lobes run the same Engine code, Hebbian/Decay/Confidence updates would be applied twice (once on Lobe, once replicated from Cortex).

**Resolution:** Section 10.1.1 specifies that Lobe nodes pass `nil` cognitive workers to `engine.NewEngine()`. Codebase verification confirms the Engine already has nil guards on all worker submissions. No additional code changes are needed for the guard itself; only `server.go` construction logic must branch on role.

### Resolved C2: Fencing Token Not Specified for Distributed Mode

**Issue:** The design referenced fencing tokens but did not specify their format, storage, or proof of uniqueness in a distributed election.

**Resolution:** Section 5.3.1 now specifies: fencing token = election epoch, persisted at Pebble key `0x19 | 0x03 | "cluster_epoch"`, with a mathematical proof via the pigeonhole principle that two candidates cannot win the same epoch.

### Resolved C3: No Path for Node Join in Phase 1 (Snapshot is Phase 2)

**Issue:** Phase 1 has no snapshot transfer, but also no guarantee that the replication log contains all entries since genesis.

**Resolution:** Section 12 Phase 1 deliverable #11 now specifies: Prune() is NEVER called in Phase 1. The replication log grows unboundedly, ensuring any joining node can resync from `lastApplied+1`. A startup warning is logged. Phase 2 adds snapshot-based IST, after which pruning is safe.

### Resolved C4: In-Memory Backlog Crash Fragility

**Issue:** The `ReplicationBacklog` ring buffer was in-memory, meaning all 1M entries would be lost on Cortex crash, forcing full resyncs after every restart.

**Resolution:** Section 7 "Partial Resync" now eliminates the `ReplicationBacklog` struct entirely. Partial resync reads directly from the Pebble-backed `ReplicationLog.ReadSince()`. This is crash-safe (Pebble survives restarts), memory-efficient (zero allocation), and simpler (one data structure instead of two).

### Resolved I2: Pre-emptive Demotion Uses Wrong Threshold

**Issue:** The Cortex demoted itself if it could not reach ANY Lobe, which is too aggressive — in a 5-node cluster, losing contact with 1 of 4 Lobes should not trigger demotion.

**Resolution:** Section 5.4 Layer 4 now uses quorum-based demotion: the Cortex demotes if it cannot reach `floor(N/2) + 1 - 1` Lobes (counting itself toward quorum). Concrete examples provided for 3, 5, and 7-node clusters.

### Resolved Minor: Non-Atomic Seq Counter Write

**Issue:** `ReplicationLog.Append()` writes the entry and seq counter as separate Pebble operations, creating a crash window.

**Resolution:** Section 8.2 now specifies that both must be written in the same Pebble batch with `batch.Commit(pebble.Sync)`.
