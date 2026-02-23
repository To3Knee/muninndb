# MuninnDB Cluster Implementation Plan

**Date:** 2026-02-23
**Status:** Ready for Implementation
**Companion document:** `docs/plans/2026-02-23-cluster-architecture.md`
**Audience:** Senior Go engineers implementing cluster infrastructure

---

## Pre-Implementation Verification

Before writing any cluster code, the implementing engineer MUST verify these assumptions:

1. **Nil worker guards exist in Engine** — VERIFIED. `engine.go` checks `if e.hebbianWorker != nil`, `if e.decayWorker != nil`, `if e.contradictWorker != nil`, `if e.confidenceWorker != nil` before every Submit/SubmitBatch call. Passing nil workers to `NewEngine()` is safe.

2. **MBP frame type range 0x20-0x38 is available** — VERIFIED. Current frame types use 0x01-0x15 and 0xFF. The 0x20-0x38 range for cluster frames has no conflicts.

3. **Pebble key prefix 0x19 is reserved for replication** — VERIFIED. `internal/replication/log.go` uses `0x19` prefix for log entries and `0x19 | 0xFF...` for the seq counter.

4. **`internal/config/` exists** — VERIFIED. Currently contains only `plugin.go`. Cluster config will be added as a new file.

5. **REST replication handlers are stubs** — VERIFIED. `replication_handlers.go` returns hardcoded values. All three handlers need real implementations.

6. **`ReplicationLog.Append()` non-atomic seq write** — VERIFIED. Entry and seq counter are two separate Pebble writes (lines 116, 122 of `log.go`). Must fix.

7. **`ReadSince` race on `l.seq`** — VERIFIED. `l.seq` is read on line 148 after `l.mu.Unlock()` on line 140.

8. **`Prune()` is O(N) loop** — VERIFIED. Iterates from `seq=0` to `untilSeq` with individual deletes (line 195).

---

## Phase 1: Foundation (Minimum Viable Cluster)

**Goal:** A 3-node cluster with automatic failover, push-based replication, and stable operation. An operator can spin it up with minimal config. No snapshot transfer, no cognitive forwarding.

**Timeline estimate:** 4-6 weeks

---

### P1-T01: Fix ReplicationLog.Append Atomic Seq Write

**Description:** Fix the crash window in `ReplicationLog.Append()` where the entry and sequence counter are written as separate Pebble operations. Write both in a single Pebble batch with `batch.Commit(pebble.Sync)`.

**Files to modify:**
- `internal/replication/log.go` — Replace the two separate writes (lines 116 and 122) with a single `pebble.Batch`

**Acceptance criteria:**
- `Append()` uses `batch.Set(entryKey, data, nil)` + `batch.Set(seqCounterKey(), seqBuf, nil)` + `batch.Commit(pebble.Sync)`
- On rollback (marshal error), `l.seq` is decremented and no Pebble writes occur
- Existing test `TestReplicationLog_Persistence` still passes: close DB after appending 3 entries, reopen, verify `CurrentSeq() == 3` and new append returns seq 4
- New test `TestReplicationLog_Append_AtomicBatch`: verify that after a successful Append, both the entry key and seq counter key exist in Pebble by reading them directly with `db.Get()`

**Dependencies:** None
**Complexity:** S

---

### P1-T02: Fix ReadSince Race Condition

**Description:** Fix the data race in `ReadSince` where `l.seq` is read after releasing `l.mu`. Capture `currentSeq := l.seq` while holding the lock, then use `currentSeq` for the iterator upper bound.

**Files to modify:**
- `internal/replication/log.go` — Modify `ReadSince()` to capture seq under lock

**Acceptance criteria:**
- `ReadSince` captures `currentSeq := l.seq` before `l.mu.Unlock()`
- The iterator upper bound uses `replicationEntryKey(currentSeq + 1)` instead of `replicationEntryKey(l.seq + 1)`
- Existing tests `TestReplicationLog_ReadSince` and `TestReplicationLog_ReadSince_WithLimit` pass
- New test `TestReplicationLog_ReadSince_ConcurrentAppend`: launch a goroutine that appends entries in a loop while the main goroutine calls `ReadSince`. Verify no panic and returned entries have monotonically increasing seq numbers. Run with `-race` flag.

**Dependencies:** P1-T01 (seq write is now atomic, so the race fix is meaningful)
**Complexity:** S

---

### P1-T03: Fix Prune to Use Range Delete

**Description:** Replace the O(N) sequential delete loop in `ReplicationLog.Prune()` with `pebble.Batch.DeleteRange()` for a single range delete operation.

**Files to modify:**
- `internal/replication/log.go` — Replace the `for seq := uint64(0); seq <= untilSeq` loop with `batch.DeleteRange(startKey, endKey, nil)`

**Acceptance criteria:**
- `Prune(untilSeq)` computes `startKey = replicationEntryKey(0)` and `endKey = replicationEntryKey(untilSeq + 1)` and calls `batch.DeleteRange(startKey, endKey, nil)` followed by `batch.Commit(nil)`
- Existing test `TestReplicationLog_Prune` passes: append 5 entries, prune entries <= 2, ReadSince(0) returns entries 3, 4, 5
- New test `TestReplicationLog_Prune_Large`: append 10000 entries, prune entries <= 9990, verify ReadSince(0) returns 10 entries (9991-10000). Time the prune and assert it completes in under 100ms (O(1) range delete, not O(N)).

**Dependencies:** P1-T01
**Complexity:** S

---

### P1-T04: Cluster Configuration Parsing

**Description:** Add cluster configuration parsing with YAML support and environment variable overrides. This is the foundational config that all other cluster tasks depend on.

**Files to create:**
- `internal/config/cluster.go` — ClusterConfig struct, YAML parsing, env var overrides, validation

**Files to modify:**
- `cmd/muninn/server.go` — Load cluster config from YAML/flags/env vars during startup

**Config struct:**
```go
type ClusterConfig struct {
    Enabled       bool     `yaml:"enabled" json:"enabled"`
    NodeID        string   `yaml:"node_id" json:"node_id"`
    BindAddr      string   `yaml:"bind_addr" json:"bind_addr"`
    Seeds         []string `yaml:"seeds" json:"seeds"`
    ClusterSecret string   `yaml:"cluster_secret" json:"cluster_secret"`
    Role          string   `yaml:"role" json:"role"` // "auto" | "primary" | "replica"
    LeaseTTL      int      `yaml:"lease_ttl" json:"lease_ttl"`       // seconds, default 10
    HeartbeatMS   int      `yaml:"heartbeat_ms" json:"heartbeat_ms"` // default 1000
}
```

**Environment variable mapping:**
- `MUNINN_CLUSTER_ENABLED` -> `Enabled`
- `MUNINN_CLUSTER_NODE_ID` -> `NodeID`
- `MUNINN_CLUSTER_BIND_ADDR` -> `BindAddr`
- `MUNINN_CLUSTER_SEEDS` -> `Seeds` (comma-separated)
- `MUNINN_CLUSTER_SECRET` -> `ClusterSecret`
- `MUNINN_CLUSTER_ROLE` -> `Role`

**Acceptance criteria:**
- `LoadClusterConfig(dataDir string) (ClusterConfig, error)` reads from `muninn.yaml` cluster section
- Env vars override YAML values (same priority model as plugin config)
- `Validate()` returns error if: Enabled but NodeID empty, Enabled but Seeds empty, Role not in allowed set
- If `NodeID` is empty and `Enabled` is true, auto-generate a stable ID from hostname + data dir hash
- Unit tests: parse valid YAML, env var overrides, validation errors, default values

**Dependencies:** None
**Complexity:** S

---

### P1-T05: Persistent lastApplied in Applier

**Description:** Persist `Applier.lastApplied` to Pebble so it survives restarts. Currently resets to 0, causing full replay of the replication log.

**Files to modify:**
- `internal/replication/applier.go` — Add Pebble persistence for lastApplied

**Pebble key:** `0x19 | 0x02 | "last_applied"` (bytes: `[0x19, 0x02, 0x6C, 0x61, ...]`)

**Acceptance criteria:**
- `NewApplier(db)` loads `lastApplied` from Pebble key on construction. If key does not exist, defaults to 0.
- After each successful `Apply()`, `lastApplied` is written to Pebble in the same batch as the applied entry (atomic).
- Test: create Applier, apply 5 entries, verify `LastApplied() == 5`. Close DB, reopen, create new Applier, verify `LastApplied() == 5`.
- Test: apply entries 1,2,3, close DB, reopen, apply entry 2 again (should be skipped), apply entry 4 (should succeed), verify `LastApplied() == 4`.

**Dependencies:** P1-T01
**Complexity:** S

---

### P1-T06: Persistent clusterEpoch in Pebble

**Description:** Store the cluster election epoch persistently so nodes never propose an epoch they have already seen, even after restart.

**Files to create:**
- `internal/replication/epoch.go` — `EpochStore` that reads/writes epoch from Pebble

**Pebble key:** `0x19 | 0x03 | "cluster_epoch"` — value: uint64 big-endian

**Acceptance criteria:**
- `NewEpochStore(db) *EpochStore` — loads current epoch from Pebble on construction (default 0)
- `Load() uint64` — returns the current epoch (cached in memory after load)
- `CompareAndSet(expected, new uint64) (bool, error)` — atomically updates epoch if current matches expected. Returns false if current != expected (concurrent update).
- `ForceSet(epoch uint64) error` — unconditionally sets epoch (used when accepting a CortexClaim with a higher epoch)
- Test: create store, verify Load() == 0, CompareAndSet(0, 1) succeeds, Load() == 1, CompareAndSet(0, 2) fails, CompareAndSet(1, 2) succeeds
- Test: persistence across DB close/reopen

**Dependencies:** P1-T01 (same Pebble key space)
**Complexity:** S

---

### P1-T07: Extend NodeRole and WALOp Types

**Description:** Add missing node roles and WAL operation types to `internal/replication/types.go` to support the full cluster vocabulary.

**Files to modify:**
- `internal/replication/types.go` — Add `RoleSentinel`, `RoleObserver` constants; fix `RoleUnknown` to 0; add `OpCognitive`, `OpIndex`, `OpMeta` WAL ops

**Acceptance criteria:**
- `RoleUnknown = 0`, `RolePrimary = 1`, `RoleReplica = 2`, `RoleSentinel = 3`, `RoleObserver = 4`
- `OpCognitive = 4`, `OpIndex = 5`, `OpMeta = 6`
- `NodeRole.String()` method returns "unknown", "primary", "replica", "sentinel", "observer"
- `WALOp.String()` method returns "set", "delete", "batch", "cognitive", "index", "meta"
- All existing tests pass (no constants changed, only added + RoleUnknown moved from 3 to 0)

**Dependencies:** None
**Complexity:** S

---

### P1-T08: MBP Cluster Frame Types

**Description:** Register all new MBP frame types for cluster communication in the frame type constants. This task only adds the constants and their msgpack struct definitions; the frame handling logic comes in later tasks.

**Files to modify:**
- `internal/transport/mbp/frame.go` — Add constants 0x20-0x38

**Files to create:**
- `internal/transport/mbp/cluster_frames.go` — Struct definitions for all cluster frame payloads

**Frame type constants (as defined in the architecture doc Section 8.1):**
```go
// Replication stream
TypeReplEntry    uint8 = 0x20
TypeReplBatch    uint8 = 0x21
TypeReplAck      uint8 = 0x22
TypeReplNack     uint8 = 0x23

// Snapshot transfer
TypeSnapHeader   uint8 = 0x24
TypeSnapChunk    uint8 = 0x25
TypeSnapAck      uint8 = 0x26
TypeSnapComplete uint8 = 0x27

// Cognitive forwarding
TypeCogForward   uint8 = 0x28
TypeCogAck       uint8 = 0x29

// Cluster protocol
TypeVoteRequest  uint8 = 0x30
TypeVoteResponse uint8 = 0x31
TypeCortexClaim  uint8 = 0x32
TypeSDown        uint8 = 0x33
TypeODown        uint8 = 0x34
TypeGossip       uint8 = 0x35
TypeJoinRequest  uint8 = 0x36
TypeJoinResponse uint8 = 0x37
TypeLeave        uint8 = 0x38
```

**Struct definitions (at minimum):**
```go
type VoteRequest struct {
    Epoch       uint64 `msgpack:"epoch"`
    CandidateID string `msgpack:"candidate_id"`
    LastSeq     uint64 `msgpack:"last_seq"`
    ConfigEpoch uint64 `msgpack:"config_epoch"`
}

type VoteResponse struct {
    Epoch   uint64 `msgpack:"epoch"`
    VoterID string `msgpack:"voter_id"`
    Granted bool   `msgpack:"granted"`
}

type CortexClaim struct {
    Epoch       uint64 `msgpack:"epoch"`
    FencingToken uint64 `msgpack:"fencing_token"`
    CortexID    string `msgpack:"cortex_id"`
    CortexAddr  string `msgpack:"cortex_addr"`
}

type ReplBatch struct {
    Entries []ReplEntry `msgpack:"entries"`
}

type ReplEntry struct {
    Seq         uint64 `msgpack:"seq"`
    Op          uint8  `msgpack:"op"`
    Key         []byte `msgpack:"key"`
    Value       []byte `msgpack:"value"`
    TimestampNS int64  `msgpack:"ts"`
}

type ReplAck struct {
    LastSeq uint64 `msgpack:"last_seq"`
    NodeID  string `msgpack:"node_id"`
}

type GossipMessage struct {
    SenderID string         `msgpack:"sender_id"`
    Epoch    uint64         `msgpack:"epoch"`
    Members  []GossipMember `msgpack:"members"`
}

type GossipMember struct {
    ID       string `msgpack:"id"`
    Addr     string `msgpack:"addr"`
    Role     uint8  `msgpack:"role"`
    LastSeq  uint64 `msgpack:"last_seq"`
    LastSeen int64  `msgpack:"last_seen"`
}

type JoinRequest struct {
    NodeID       string   `msgpack:"node_id"`
    Addr         string   `msgpack:"addr"`
    LastApplied  uint64   `msgpack:"last_applied"`
    Capabilities []string `msgpack:"capabilities"`
    SecretHash   []byte   `msgpack:"secret_hash"` // HMAC of cluster secret
}

type JoinResponse struct {
    Accepted    bool           `msgpack:"accepted"`
    CortexID    string         `msgpack:"cortex_id"`
    CortexAddr  string         `msgpack:"cortex_addr"`
    Epoch       uint64         `msgpack:"epoch"`
    Members     []GossipMember `msgpack:"members"`
    RejectReason string        `msgpack:"reject_reason,omitempty"`
}
```

**Acceptance criteria:**
- All constants defined with no conflicts against existing 0x01-0x15 range
- All structs serialize/deserialize with msgpack (round-trip test for each struct)
- `frameTypeName(uint8) string` helper returns human-readable names for all new types
- Existing MBP tests pass unchanged

**Dependencies:** None
**Complexity:** M

---

### P1-T09: TCP Peer Connection Manager

**Description:** Build a connection manager that maintains persistent MBP TCP connections between cluster peers. The Cortex maintains outbound connections to all Lobes. Lobes maintain a connection to the Cortex. All cluster protocol frames flow over these connections.

**Files to create:**
- `internal/replication/peer_conn.go` — `PeerConnection` struct wrapping a net.Conn with MBP frame read/write
- `internal/replication/conn_manager.go` — `ConnManager` that manages peer connections with reconnection

**Design:**
```go
type PeerConnection struct {
    NodeID   string
    Addr     string
    conn     net.Conn
    encoder  *mbp.FrameWriter
    decoder  *mbp.FrameReader
    mu       sync.Mutex
    closed   bool
}

type ConnManager struct {
    nodeID      string
    clusterSecret string
    peers       map[string]*PeerConnection // nodeID -> conn
    mu          sync.RWMutex
    onFrame     func(nodeID string, frameType uint8, payload []byte) // dispatch callback
}
```

**Acceptance criteria:**
- `ConnManager.Connect(nodeID, addr string) error` — establishes MBP connection, sends HELLO with cluster capabilities, receives HELLO_OK with cluster state
- `ConnManager.Send(nodeID string, frameType uint8, payload []byte) error` — sends a frame to a specific peer
- `ConnManager.Broadcast(frameType uint8, payload []byte)` — sends to all connected peers
- `ConnManager.Disconnect(nodeID string)` — gracefully closes a peer connection
- Automatic reconnection: if a connection drops, the manager retries every 2 seconds with exponential backoff (max 30s)
- HMAC-based authentication: on HELLO, the connecting node sends `HMAC-SHA256(cluster_secret, node_id + nonce)`. The receiving node verifies.
- Test: two ConnManagers connect to each other via localhost, exchange frames, verify receipt
- Test: disconnect one side, verify the other detects disconnection within 3 heartbeat intervals
- Test: invalid cluster secret rejected during HELLO handshake

**Dependencies:** P1-T08 (frame types), P1-T04 (cluster config for secret)
**Complexity:** L

---

### P1-T10: MSP Heartbeat and Failure Detection

**Description:** Implement the MuninnDB Sentinel Protocol (MSP) heartbeat goroutine and failure detection logic. Every node runs this goroutine. It sends PING every 1s to all peers, tracks PONG responses, and declares SDOWN after 3 missed heartbeats.

**Files to create:**
- `internal/replication/msp.go` — `MSP` struct with heartbeat loop, SDOWN/ODOWN detection

**Design:**
```go
type MSP struct {
    nodeID      string
    connManager *ConnManager
    epochStore  *EpochStore
    peers       map[string]*PeerState // tracked by heartbeat
    mu          sync.RWMutex
    onSDown     func(nodeID string)
    onODown     func(nodeID string)
}

type PeerState struct {
    NodeID       string
    LastPong     time.Time
    Subjective   NodeStatus // UP, SDOWN
    Objective    NodeStatus // UP, ODOWN
    LastSeq      uint64
    Role         NodeRole
}

type NodeStatus uint8
const (
    StatusUp    NodeStatus = 0
    StatusSDown NodeStatus = 1
    StatusODown NodeStatus = 2
)
```

**Acceptance criteria:**
- `MSP.Run(ctx)` starts: (a) a heartbeat ticker sending TypePing every 1s to all peers, (b) a receiver that processes TypePong frames and updates `PeerState.LastPong`
- After 3 consecutive missed PONGs (3s), the peer is marked `StatusSDown` and `onSDown` is called
- SDOWN notifications are exchanged via `TypeSDown` frames. When a quorum of nodes agree a peer is SDOWN, it is promoted to `StatusODown` and `onODown` is called
- Quorum for ODOWN = floor(totalNodes/2) + 1 (including the node itself)
- `MSP.PeerStates() []PeerState` — returns current state of all peers (for health endpoints)
- Test: 3-node MSP cluster (localhost ports), stop heartbeats from one node, verify SDOWN detected at ~3s and ODOWN at ~3.1s by the other two
- Test: 2-node cluster, one goes SDOWN, ODOWN cannot be declared (no quorum among live nodes if total was 3)

**Dependencies:** P1-T09 (ConnManager), P1-T08 (frame types)
**Complexity:** L

---

### P1-T11: MSP Election Protocol

**Description:** Implement the election protocol that runs when ODOWN is declared for the current Cortex. Lobes propose themselves as candidates, request votes, and the winner claims the Cortex role with a fencing token.

**Files to create:**
- `internal/replication/election.go` — `ElectionRunner` struct

**Design:**
```go
type ElectionRunner struct {
    nodeID      string
    connManager *ConnManager
    epochStore  *EpochStore
    repLog      *ReplicationLog
    mu          sync.Mutex
    votedForEpoch map[uint64]string // epoch -> candidateID voted for
    onElected   func(epoch uint64)   // callback when this node wins
    onNewCortex func(claim CortexClaim) // callback when another node wins
}
```

**Election flow:**
1. On ODOWN(cortex), each eligible Lobe calls `StartElection()`
2. `StartElection` increments epoch via `epochStore.CompareAndSet(current, current+1)`
3. Sends `TypeVoteRequest{Epoch, CandidateID, LastSeq}` to all peers
4. Receives `TypeVoteResponse`. Votes are granted to the candidate with highest `LastSeq`, tie-broken by lowest `NodeID`.
5. A node can only vote once per epoch (tracked in `votedForEpoch` map)
6. If this node receives quorum votes: send `TypeCortexClaim{Epoch, FencingToken=Epoch, CortexID, CortexAddr}` to all peers
7. On receiving `CortexClaim` with epoch > current: accept, update local epoch, call `onNewCortex`

**Acceptance criteria:**
- `StartElection()` proposes epoch = currentEpoch + 1
- Vote granting logic: if not yet voted for this epoch, vote for the candidate with highest LastSeq; if tied, lowest NodeID
- Election succeeds when quorum votes received. Winner broadcasts CortexClaim.
- Election epoch is persisted via EpochStore before any votes are sent (crash safety)
- Test: 3-node cluster, node A is Cortex, kill node A, nodes B and C run election. Node with highest LastSeq wins. Verify epoch incremented, CortexClaim received by all live nodes.
- Test: split vote impossible — two candidates propose same epoch, only one gets quorum (pigeonhole principle, verified by running 100 iterations with random timing)
- Test: stale vote rejection — a node that already voted for epoch N rejects a second VoteRequest for epoch N

**Dependencies:** P1-T06 (EpochStore), P1-T09 (ConnManager), P1-T10 (ODOWN triggers election), P1-T08 (frame types)
**Complexity:** L

---

### P1-T12: Push-Based Network Streamer

**Description:** Replace the poll-based `Streamer` with a push-based `NetworkStreamer` that sends replication entries over MBP frames to connected Lobes. Uses subscriber notification from `ReplicationLog.Append()`.

**Files to create:**
- `internal/replication/network_streamer.go` — `NetworkStreamer` struct

**Files to modify:**
- `internal/replication/log.go` — Add `subscribers []chan struct{}` to `ReplicationLog`, notify on each `Append()`

**Design:**
```go
type NetworkStreamer struct {
    log         *ReplicationLog
    connManager *ConnManager
    batchWindow time.Duration // 5ms
    maxBatch    int           // 100 entries
    notify      chan struct{} // from ReplicationLog subscriber
    mu          sync.RWMutex
    lobes       map[string]*LobeStream
}

type LobeStream struct {
    nodeID     string
    lastAcked  uint64
    unacked    int
    maxUnacked int // 10000, backpressure threshold
}
```

**Acceptance criteria:**
- `ReplicationLog.Subscribe() <-chan struct{}` — returns a notification channel, `Unsubscribe(ch)` removes it
- On each `Append()`, all subscriber channels are notified (non-blocking send)
- `NetworkStreamer.Run(ctx)` listens on the notify channel, batches entries (5ms window or 100 entries), serializes as `TypeReplBatch`, sends to all connected Lobes
- Backpressure: if a Lobe has > 10000 unacked entries, pause sending to that Lobe (log warning, do not block other Lobes)
- `TypeReplAck` from Lobe updates `LobeStream.lastAcked` and `LobeStream.unacked`
- Test: append 100 entries to the log, verify NetworkStreamer delivers them to a connected mock Lobe within 50ms
- Test: backpressure — mock Lobe that never ACKs, verify other Lobes still receive entries
- Test: Lobe reconnects with lastApplied=50, receives entries 51+ immediately from log

**Dependencies:** P1-T01 (atomic append), P1-T02 (ReadSince race fix), P1-T09 (ConnManager), P1-T08 (frame types)
**Complexity:** L

---

### P1-T13: Lobe Join Protocol

**Description:** Implement the node join protocol where a new Lobe connects to the cluster, authenticates, receives cluster state, and begins receiving the replication stream from the Cortex.

**Files to create:**
- `internal/replication/join.go` — join protocol handler for both sides (joining node and Cortex)

**Join flow:**
1. Joining node connects to seed nodes via ConnManager
2. Sends `TypeJoinRequest{NodeID, Addr, LastApplied, Capabilities, SecretHash}`
3. Cortex validates secret, adds node to cluster membership
4. Cortex responds with `TypeJoinResponse{Accepted, CortexID, Epoch, Members}`
5. If `LastApplied > 0` and entries exist in the log: Cortex sends entries from `lastApplied+1` via the replication stream (partial resync)
6. If `LastApplied == 0`: this is a fresh node; in Phase 1, it gets the full replication log from seq 1 (since Prune is never called)

**Acceptance criteria:**
- Joining node receives full cluster membership in JoinResponse
- Joining node with `LastApplied=0` receives all entries from the Cortex's replication log
- Joining node with `LastApplied=500` (rejoining after restart) receives only entries 501+
- Invalid cluster secret is rejected with `RejectReason: "invalid cluster secret"`
- Cortex gossips new membership to all existing Lobes after a successful join
- Test: 3-node cluster, node C joins with LastApplied=0, receives all entries
- Test: node C restarts with LastApplied=500, Cortex has 1000 entries, node C receives 501-1000
- Test: node with wrong secret is rejected

**Dependencies:** P1-T09 (ConnManager), P1-T12 (NetworkStreamer for replication), P1-T05 (persistent lastApplied), P1-T04 (cluster config)
**Complexity:** M

---

### P1-T14: Cluster Coordinator Rewrite

**Description:** Rewrite the `Coordinator` to orchestrate all cluster subsystems: MSP, election, replication streaming, join protocol, and role management. This is the "brain" of the cluster layer.

**Files to modify:**
- `internal/replication/coordinator.go` — Major rewrite: add subsystem lifecycle management

**Design:**
```go
type Coordinator struct {
    cfg          *config.ClusterConfig
    nodeID       string
    role         atomic.Uint32
    log          *ReplicationLog
    applier      *Applier
    epochStore   *EpochStore
    connManager  *ConnManager
    msp          *MSP
    election     *ElectionRunner
    streamer     *NetworkStreamer

    // Callbacks for role transitions
    onPromoteToCortex func()
    onDemoteToLobe    func()

    mu            sync.Mutex
    knownMembers  map[string]*MemberInfo
    cortexID      string
    cortexAddr    string
}
```

**Key responsibilities:**
- `Start(ctx)` — connects to seeds, runs MSP, participates in initial election
- Role transitions: when elected, start cognitive workers (via callback), start NetworkStreamer. When demoted, stop workers, start receiving replication.
- Pre-emptive demotion: if Cortex cannot reach quorum of Lobes for 5s, demote self
- Expose `Role()`, `CortexID()`, `Epoch()`, `FencingToken()`, `ClusterMembers()` for REST handlers

**Acceptance criteria:**
- `Start(ctx)` connects to seed peers, discovers cluster state, and reaches steady state within 5 seconds
- If no Cortex exists (fresh cluster), an election occurs and one node becomes Cortex
- If a Cortex exists, the joining node registers as a Lobe and begins receiving replication
- `Stop()` gracefully disconnects from all peers and releases any held lease
- Role transitions fire callbacks: `onPromoteToCortex` and `onDemoteToLobe`
- Test: start 3 Coordinators with localhost ports, verify one becomes Cortex and two become Lobes
- Test: stop the Cortex coordinator, verify one of the Lobes is promoted within 5 seconds

**Dependencies:** P1-T09, P1-T10, P1-T11, P1-T12, P1-T13, P1-T06
**Complexity:** L

---

### P1-T15: Wire Coordinator into server.go

**Description:** Integrate the cluster `Coordinator` into the server startup path behind a `cluster.enabled` config flag. This task makes the cluster layer actually run when the server starts.

**Files to modify:**
- `cmd/muninn/server.go` — Add cluster startup logic after Pebble open, before engine construction

**Integration points:**
1. After Pebble opens, check `clusterCfg.Enabled`
2. Create `ReplicationLog`, `Applier`, `EpochStore`, `ConnManager`, `MSP`, `ElectionRunner`, `NetworkStreamer`, `Coordinator`
3. Start `Coordinator` in a goroutine
4. Wait for initial role determination (Cortex or Lobe) before constructing the Engine
5. If Lobe: pass nil cognitive workers to `NewEngine()`
6. If Cortex: pass full cognitive workers to `NewEngine()`
7. Set up role transition callbacks:
   - `onPromoteToCortex`: construct and start cognitive workers, update engine
   - `onDemoteToLobe`: stop cognitive workers, set to nil in engine
8. Log Phase 1 pruning warning: `"Log pruning is disabled in Phase 1. Monitor disk usage."`
9. On shutdown: stop Coordinator before closing Pebble

**Acceptance criteria:**
- `muninn start` with `cluster.enabled: false` (default) behaves exactly as before (no regressions)
- `muninn start` with `cluster.enabled: true` starts the Coordinator, connects to seeds, and logs cluster state
- Role-dependent engine construction: Lobe nodes do not start cognitive workers
- Graceful shutdown: Coordinator.Stop() is called before Pebble close
- Test: start a single-node cluster (seeds pointing to self), verify it becomes Cortex
- Manual test: start 3 nodes with correct seeds, verify cluster formation in logs

**Dependencies:** P1-T14 (Coordinator), P1-T04 (cluster config)
**Complexity:** M

---

### P1-T16: Replace Stub REST Handlers

**Description:** Replace the hardcoded stub handlers in `replication_handlers.go` with real implementations that delegate to the `Coordinator`.

**Files to modify:**
- `internal/transport/rest/replication_handlers.go` — Real implementations
- `internal/transport/rest/server.go` — Wire coordinator into REST server, register new routes

**Handler implementations:**
- `HandleReplicationStatus`: query `Coordinator.Role()`, `Coordinator.FencingToken()`, `Coordinator.Epoch()`, `ReplicationLog.CurrentSeq()`
- `HandleReplicationLag`: query `Coordinator.GetKnownReplicas()`, compute lag per Lobe
- `HandlePromoteReplica`: call `Coordinator.TriggerElection()` if force=true, or `Coordinator.GracefulFailover(targetNodeID)` if target specified

**Acceptance criteria:**
- `GET /v1/replication/status` returns actual node role, seq, fencing token, epoch
- `GET /v1/replication/lag` returns per-Lobe lag with actual seq numbers
- `POST /v1/replication/promote` triggers an election and returns the result
- When cluster mode is disabled, all three endpoints return a `501 Not Implemented` with message "cluster mode not enabled"
- Test: mock Coordinator, verify handlers return correct data
- Test: cluster disabled, verify 501 response

**Dependencies:** P1-T14 (Coordinator), P1-T15 (wiring into server.go)
**Complexity:** M

---

### P1-T17: Add Cluster Info and Health Endpoints

**Description:** Add new REST endpoints for cluster-wide information and health checks.

**Files to create:**
- `internal/transport/rest/cluster_handlers.go` — New cluster-specific handlers

**Files to modify:**
- `internal/transport/rest/server.go` — Register new routes

**Endpoints:**
- `GET /v1/cluster/info` — Returns: members, roles, epochs, connectivity, fencing token, cluster size
- `GET /v1/cluster/health` — Returns 200 if healthy (connected to Cortex or is Cortex with quorum), 503 if degraded (lost quorum, SDOWN detected, high replication lag)

**Acceptance criteria:**
- `/v1/cluster/info` returns JSON with: `node_id`, `role`, `cortex_id`, `epoch`, `fencing_token`, `members[]` (each with id, addr, role, last_seq, status)
- `/v1/cluster/health` returns 200 with `{"status": "healthy", "role": "cortex", "cluster_size": 3}` when healthy
- `/v1/cluster/health` returns 503 with `{"status": "degraded", "reason": "cannot reach quorum"}` when degraded
- Both endpoints return 501 when cluster mode is disabled
- Test: mock Coordinator with healthy state, verify 200 response
- Test: mock Coordinator with SDOWN detected, verify 503 response

**Dependencies:** P1-T14, P1-T15, P1-T16
**Complexity:** S

---

### P1-T18: Phase 1 Integration Test

**Description:** Comprehensive integration test that starts a 3-node cluster using in-process components (separate Pebble dirs, localhost ports), performs writes, verifies replication, kills the Cortex, and verifies failover.

**Files to create:**
- `internal/replication/integration_test.go` — Full integration test (build tag: `integration`)

**Test scenario:**
```
1. Start 3 nodes (node1, node2, node3) with t.TempDir() Pebble instances
2. Wait for election: verify exactly one Cortex and two Lobes
3. Write 100 entries to the Cortex
4. Wait for replication: verify both Lobes have LastApplied >= 100
5. ReadSince(0, 100) on each Lobe: verify all 100 entries match
6. Stop the Cortex node (simulate crash)
7. Wait for failover: verify one Lobe becomes Cortex within 5 seconds
8. Write 50 more entries to the new Cortex
9. Wait for replication: verify the surviving Lobe has LastApplied >= 150
10. Restart the original Cortex node
11. Verify it rejoins as a Lobe (epoch is lower than current)
12. Verify it catches up via partial resync (LastApplied >= 150)
13. Verify fencing tokens are monotonically increasing across all transitions
```

**Acceptance criteria:**
- Test passes with `-race` flag
- Test completes in under 30 seconds
- All replication entries verified byte-for-byte (key + value match)
- Failover time < 5 seconds (3s SDOWN + election)
- Fencing token monotonically increases across role transitions
- No goroutine leaks (verify with `goleak` or manual `runtime.NumGoroutine()` check)

**Dependencies:** All P1 tasks (P1-T01 through P1-T17)
**Complexity:** L

---

## Phase 2: Cognitive Cluster

**Goal:** Cognitive side-effect forwarding works. Activations on Lobes are fully functional. Initial state transfer allows new nodes to join a running cluster.

**Timeline estimate:** 4-6 weeks after Phase 1

---

### P2-T01: TypeCogForward / TypeCogAck Frame Implementation

**Description:** Implement the frame handling for cognitive side-effect forwarding. Lobes serialize cognitive side effects into `TypeCogForward` frames and the Cortex sends `TypeCogAck` back.

**Files to modify:**
- `internal/transport/mbp/cluster_frames.go` — Add `CognitiveSideEffect` and `CogAck` structs (definitions already in arch doc)
- `internal/replication/conn_manager.go` — Handle TypeCogForward/TypeCogAck dispatch

**Acceptance criteria:**
- `CognitiveSideEffect` struct round-trips through msgpack correctly
- ConnManager dispatches `TypeCogForward` frames to registered handler
- `CogAck{QueryID}` sent back to Lobe after Cortex processes the forward
- Test: serialize a CognitiveSideEffect with 20 CoActivationRefs and 20 AccessedIDs, verify round-trip

**Dependencies:** P1-T08, P1-T09
**Complexity:** S

---

### P2-T02: Lobe Activation Side-Effect Collection

**Description:** Modify the Engine's activation path to collect cognitive side effects instead of submitting to (nil) local workers, and forward them to the Cortex. This is the key change that makes Lobe activations fully functional.

**Files to modify:**
- `internal/engine/engine.go` — Add side-effect collection mode to `Activate()`
- `internal/replication/coordinator.go` — Add `ForwardCognitiveEffects(effects CognitiveSideEffect) error`

**Design:** When cognitive workers are nil (Lobe mode), the Engine collects the co-activated engram IDs and accessed IDs that would have been submitted to workers, packages them into a `CognitiveSideEffect`, and calls `coordinator.ForwardCognitiveEffects()`.

**Acceptance criteria:**
- On a Lobe, `Activate()` returns correct results (same as Cortex minus the side effects)
- The side effects (co-activation pairs, accessed IDs) are packaged into a `CognitiveSideEffect` and forwarded
- Forwarding is async and non-blocking — if the Cortex is unreachable, the effects are dropped (best-effort)
- Test: Lobe activation produces correct results and generates a CognitiveSideEffect with correct engram IDs

**Dependencies:** P2-T01, P1-T14
**Complexity:** M

---

### P2-T03: Cortex Cognitive Forward Dispatch

**Description:** On the Cortex, receive `TypeCogForward` frames from Lobes and dispatch the contained side effects to the appropriate cognitive workers (HebbianWorker, DecayWorker, ConfidenceWorker).

**Files to modify:**
- `internal/replication/coordinator.go` — Add `handleCogForward()` method

**Acceptance criteria:**
- Co-activation events from the forward are submitted to `HebbianWorker.Submit()`
- Accessed IDs are submitted to `DecayWorker.SubmitBatch()`
- Confidence updates are submitted to `ConfidenceWorker.SubmitBatch()`
- If any worker channel is full, the event is dropped (not blocking the Cortex)
- A counter `cognitive_forwarded_effects_total` is incremented for observability
- Test: send a CognitiveSideEffect with 5 co-activation pairs, verify HebbianWorker receives them

**Dependencies:** P2-T01, P2-T02
**Complexity:** M

---

### P2-T04: WALOp Classification Extension

**Description:** Classify replication log entries by type (data, cognitive, index, meta) to enable priority replication and selective replication for Observers.

**Files to modify:**
- `internal/replication/types.go` — Already done in P1-T07 (constants added)
- `internal/replication/log.go` — `Append()` now accepts `WALOp` correctly
- `internal/engine/engine.go` — Use `OpCognitive` for Hebbian/Decay/Confidence writes, `OpIndex` for FTS updates, `OpMeta` for cluster metadata

**Acceptance criteria:**
- Cognitive worker writes use `OpCognitive` in the replication log
- FTS index updates use `OpIndex`
- Cluster metadata (epoch changes, membership) use `OpMeta`
- `Applier.Apply()` handles all new op types correctly (OpCognitive, OpIndex, OpMeta all apply the key-value write)
- Test: append entries of each type, ReadSince returns them with correct Op values

**Dependencies:** P1-T07
**Complexity:** S

---

### P2-T05: Pebble Snapshot Streaming (Initial State Transfer)

**Description:** Implement the snapshot-based initial state transfer that allows a fresh node to join a running cluster. The Cortex creates a Pebble snapshot, streams it to the new Lobe, and then the Lobe catches up from the replication log tail.

**Files to create:**
- `internal/replication/snapshot.go` — `SnapshotSender` (Cortex side) and `SnapshotReceiver` (Lobe side)

**Design:** Uses `pebble.DB.NewSnapshot()` for consistent point-in-time reads. Streams key-value pairs in 1MB chunks via `TypeSnapHeader`, `TypeSnapChunk`, `TypeSnapComplete` frames.

**Acceptance criteria:**
- `SnapshotSender.Stream(ctx, conn, snapshot)` streams all KV pairs from the snapshot, 1MB per chunk
- `SnapshotReceiver.Receive(ctx, conn)` writes all received KV pairs to local Pebble
- Snapshot includes a `SnapshotSeq` (the replication log seq at snapshot time)
- After snapshot, the Lobe requests entries from `SnapshotSeq+1` and applies them
- Rate limiting: max 100MB/s to avoid starving the Cortex write path
- Concurrent transfer limit: 1 (configurable)
- Test: Cortex with 10000 KV pairs, new Lobe joins, receives snapshot, catches up from log tail, ends with identical data

**Dependencies:** P1-T13 (join protocol triggers snapshot for fresh nodes), P1-T08
**Complexity:** L

---

### P2-T06: Snapshot Catch-Up from Replication Log Tail

**Description:** After a Lobe receives a snapshot, it applies the replication log entries that were written during the snapshot transfer. This closes the gap between the snapshot point-in-time and the current state.

**Files to modify:**
- `internal/replication/join.go` — Add catch-up phase after snapshot completion

**Acceptance criteria:**
- After snapshot completion, Lobe requests entries from `snapshotSeq+1`
- Entries are applied via the Applier (idempotent, so duplicate application is safe)
- Once caught up (lastApplied == Cortex currentSeq), Lobe enters normal streaming replication
- Test: during snapshot transfer, Cortex processes 200 writes. After snapshot, Lobe catches up and has all 200 entries.

**Dependencies:** P2-T05, P1-T12
**Complexity:** M

---

### P2-T07: Atomic Hebbian Batch Writes

**Description:** Refactor `HebbianWorker.processBatch` to write all pair updates in a single Pebble Batch, making them atomic. Currently, a crash mid-batch can leave partial updates.

**Files to modify:**
- `internal/cognitive/hebbian.go` (or wherever processBatch is implemented)

**Acceptance criteria:**
- All Hebbian weight updates in a single batch are written atomically via `pebble.Batch.Commit()`
- If the batch fails, no partial updates are persisted
- Test: submit a batch of 50 co-activation pairs, verify all 50 weights are updated atomically

**Dependencies:** None (can be done in parallel with other Phase 2 work)
**Complexity:** M

---

### P2-T08: Cognitive Worker Metadata Persistence

**Description:** Persist cognitive worker metadata (e.g., `HebbianWorker.lastDecayAt`) to Pebble so a newly promoted Lobe does not re-run expensive operations.

**Files to modify:**
- `internal/cognitive/hebbian.go` — Persist `lastDecayAt` to Pebble key `0x19 | 0x01 | "hebbian_last_decay"`
- Other workers as needed

**Pebble keys:**
- `0x19 | 0x01 | "hebbian_last_decay"` — int64 unix nanos
- `0x19 | 0x01 | "decay_last_run"` — int64 unix nanos

**Acceptance criteria:**
- Worker metadata is written to Pebble every 30 seconds and on graceful shutdown
- On startup (or promotion to Cortex), workers load their metadata from Pebble
- A promoted Lobe does not re-run the 6-hour decay cycle if it ran recently
- Test: write metadata, restart, verify loaded correctly

**Dependencies:** None
**Complexity:** S

---

### P2-T09: Graceful Cognitive Handoff (DRAINING State)

**Description:** Implement the graceful cognitive handoff for planned failovers (e.g., rolling upgrades). The Cortex enters DRAINING state, flushes cognitive workers, waits for replication convergence, and hands off to a successor.

**Files to modify:**
- `internal/replication/coordinator.go` — Add `GracefulFailover(targetNodeID string) error`

**Handoff flow:**
1. Cortex enters DRAINING state (rejects new writes)
2. Flushes all cognitive worker queues
3. Waits for all Lobes to catch up (replication lag = 0)
4. Sends `HANDOFF{targetNodeID}` to the designated successor
5. Successor starts cognitive workers, sends `HANDOFF_ACK`
6. Cortex releases lease and demotes to Lobe

**Acceptance criteria:**
- Graceful handoff results in zero data loss and zero cognitive gap
- DRAINING state rejects new writes with a specific error code
- Cognitive worker queues are fully flushed before handoff
- Test: start 3-node cluster, trigger graceful failover to node B, verify zero lost cognitive events

**Dependencies:** P2-T03, P2-T08, P1-T14
**Complexity:** L

---

### P2-T10: Phase 2 Integration Tests

**Description:** Integration tests covering new node joins with snapshot transfer and cognitive forwarding correctness.

**Files to create:**
- `internal/replication/integration_phase2_test.go`

**Test scenarios:**
1. **New node joins running cluster:** Start 2-node cluster, write 1000 entries, start 3rd node. Verify snapshot transfer completes and 3rd node has all data.
2. **Cognitive forwarding correctness:** Activate on a Lobe, verify co-activation events reach the Cortex's HebbianWorker and produce weight updates visible on all nodes after replication.
3. **Graceful handoff:** Trigger graceful failover, verify zero cognitive events lost during handoff.

**Acceptance criteria:**
- All tests pass with `-race` flag
- Snapshot-based join verified: new node data matches Cortex byte-for-byte
- Cognitive forwarding verified: Hebbian weight change originated from Lobe activation is visible on all nodes

**Dependencies:** All P2 tasks
**Complexity:** L

---

## Phase 3: Enterprise Polish

**Goal:** Production-ready for enterprise deployment. Security, observability, operational tooling.

**Timeline estimate:** 6-8 weeks after Phase 2

---

### P3-T01: Mutual TLS Between Nodes

**Description:** Auto-generate a cluster CA and node certificates. All inter-node MBP connections use mutual TLS.

**Files to create:**
- `internal/replication/tls.go` — Certificate generation and management
- `internal/config/tls.go` — TLS configuration

**Acceptance criteria:**
- On first cluster startup, a CA key pair is generated and stored in the data directory
- Each node generates a CSR signed by the CA
- All peer connections use mutual TLS (both sides present certificates)
- Certificate rotation without downtime
- Test: cluster with TLS enabled, verify connection fails with invalid certificate

**Dependencies:** P1-T09
**Complexity:** L

---

### P3-T02: Sentinel Node Role

**Description:** Implement the Sentinel node role: a lightweight node that participates in MSP voting but stores no data in Pebble.

**Files to modify:**
- `internal/replication/coordinator.go` — Handle `RoleSentinel` role
- `internal/replication/msp.go` — Sentinel participates in heartbeat and voting only

**Acceptance criteria:**
- Sentinel node starts without Pebble (or with an empty Pebble for epoch storage only)
- Sentinel participates in SDOWN/ODOWN detection and election voting
- Sentinel never becomes Cortex (not eligible)
- Test: 2 data nodes + 1 sentinel, kill Cortex, verify failover succeeds with sentinel providing quorum vote

**Dependencies:** P1-T10, P1-T11, P1-T07
**Complexity:** M

---

### P3-T03: Observer Node Role

**Description:** Implement the Observer node role: receives the replication stream but does not participate in elections.

**Files to modify:**
- `internal/replication/coordinator.go` — Handle `RoleObserver` role

**Acceptance criteria:**
- Observer receives replication entries and applies them locally
- Observer does not participate in SDOWN/ODOWN voting or elections
- Observer can serve cold reads but not activations (or activations without forwarding)
- Test: add Observer to a 3-node cluster, verify it receives all data but does not affect election quorum

**Dependencies:** P1-T12, P1-T14, P1-T07
**Complexity:** M

---

### P3-T04: Cognitive Consistency Score (CCS)

**Description:** Implement the real-time Cognitive Consistency Score that measures cognitive state divergence across cluster nodes.

**Files to create:**
- `internal/replication/ccs.go` — CCS computation logic

**Acceptance criteria:**
- Every 30 seconds, each Lobe computes a hash of cognitive state for a random sample of engrams
- Cortex compares Lobe hashes to its own and computes divergence
- Exposed via `GET /v1/cluster/cognitive/consistency` with score 0.0-1.0
- Assessment categories: "excellent" (>0.99), "good" (>0.95), "degraded" (>0.90), "critical" (<0.90)
- Test: cluster with zero replication lag has CCS > 0.99; cluster with intentional 5s lag has lower CCS

**Dependencies:** P2-T02, P1-T17
**Complexity:** M

---

### P3-T05: Automatic Cognitive Reconciliation Post-Partition

**Description:** After a partition heals, automatically reconcile divergent cognitive state by comparing checksums and syncing only the divergent weights/scores.

**Files to create:**
- `internal/replication/reconcile.go` — Reconciliation protocol

**Acceptance criteria:**
- After partition heal, top-K most-accessed engrams are checksum-compared
- Divergent weights/scores are synced from Cortex to Lobe (Cortex is authoritative)
- Reconciliation is logged with full detail
- Test: simulate partition, write different cognitive state on each side, heal partition, verify reconciliation

**Dependencies:** P3-T04
**Complexity:** L

---

### P3-T06: Web UI Cluster Dashboard

**Description:** Add a cluster dashboard to the existing Web UI showing real-time cluster state.

**Files to modify:**
- `web/templates/index.html` — Add cluster tab/section
- `web/static/js/app.js` — Add cluster dashboard logic

**Features:**
- Live replication stream visualization
- Cognitive worker heatmaps per node
- Failover event timeline
- Cluster topology diagram
- CCS gauge

**Acceptance criteria:**
- Dashboard shows all cluster members with roles and status
- Replication lag displayed per Lobe
- Failover events shown in timeline
- Auto-refreshes via WebSocket (leverages existing `uiSrv.Broadcast`)

**Dependencies:** P1-T17, P3-T04
**Complexity:** L

---

### P3-T07: CLI Cluster Management Commands

**Description:** Add CLI commands for cluster management operations.

**Files to create:**
- `cmd/muninn/cluster.go` — CLI subcommands

**Commands:**
```
muninn cluster info           — show cluster state
muninn cluster status         — show health and replication lag
muninn cluster failover       — trigger manual failover
muninn cluster add-node       — add a node to the cluster
muninn cluster remove-node    — remove a node from the cluster
```

**Acceptance criteria:**
- Each command calls the appropriate REST endpoint and formats output
- `cluster info` shows members, roles, epochs in a readable table
- `cluster failover --target <node> --graceful` triggers graceful handoff
- Commands return non-zero exit code on error

**Dependencies:** P1-T16, P1-T17
**Complexity:** M

---

### P3-T08: Comprehensive Integration Test Suite

**Description:** Full integration test suite covering all cluster scenarios including network partitions, cognitive consistency, and edge cases.

**Files to create:**
- `internal/replication/integration_phase3_test.go`

**Test scenarios:**
1. Simulated network partition (drop frames between specific nodes)
2. Failover correctness under write load
3. Cognitive consistency after partition heal
4. Sentinel-based quorum in 2+1 configuration
5. Observer receives data but does not affect elections
6. Rolling upgrade (graceful failover of each node in sequence)
7. Large cluster (5-7 nodes) election convergence
8. Clock skew resilience (SDOWN timing with simulated clock drift)

**Acceptance criteria:**
- All tests pass with `-race` flag
- No goroutine leaks in any test
- Tests complete within 2 minutes total

**Dependencies:** All P3 tasks
**Complexity:** L

---

## Phase Summary

| Phase | Tasks | Total Complexity | Timeline |
|-------|-------|-----------------|----------|
| Phase 1 | 18 tasks (P1-T01 through P1-T18) | 5S + 5M + 8L | 4-6 weeks |
| Phase 2 | 10 tasks (P2-T01 through P2-T10) | 3S + 5M + 2L | 4-6 weeks |
| Phase 3 | 8 tasks (P3-T01 through P3-T08) | 0S + 4M + 4L | 6-8 weeks |
| **Total** | **36 tasks** | | **14-20 weeks** |

---

## Dependency Graph (Phase 1)

```
P1-T04 (config) ─────────────────────────────────────────┐
P1-T07 (types) ──────────────────────────────────────────┐│
P1-T08 (frame types) ───────────────────────────────────┐││
                                                         │││
P1-T01 (atomic seq) ──┬── P1-T02 (ReadSince race) ──┐  │││
                      │                               │  │││
                      ├── P1-T03 (Prune range del) │  │││
                      │                               │  │││
                      └── P1-T05 (persist lastApplied)│  │││
                                                      │  │││
P1-T06 (epoch store) ──────────────────────────────┐  │  │││
                                                    │  │  │││
P1-T09 (conn manager) ◄── P1-T08, P1-T04 ────────┐│  │  │││
                                                   ││  │  │││
P1-T10 (heartbeat/SDOWN) ◄── P1-T09, P1-T08 ────┐││  │  │││
                                                  │││  │  │││
P1-T11 (election) ◄── P1-T06, P1-T09, P1-T10 ──┐│││  │  │││
                                                 ││││  │  │││
P1-T12 (streamer) ◄── P1-T01, P1-T02, P1-T09 ──┤│││  │  │││
                                                 ││││  │  │││
P1-T13 (join) ◄── P1-T09, P1-T12, P1-T05, P1-T04│││  │  │││
                                                 │││││ │  │││
P1-T14 (coordinator) ◄── T09,T10,T11,T12,T13,T06│││││ │  │││
                                                 │││││││  │││
P1-T15 (wire server.go) ◄── P1-T14, P1-T04 ─────┤│││││   │││
                                                  ││││││   │││
P1-T16 (REST handlers) ◄── P1-T14, P1-T15 ──────┤│││││   │││
                                                  ││││││   │││
P1-T17 (cluster endpoints) ◄── P1-T14, P1-T15 ──┤│││││   │││
                                                  ││││││   │││
P1-T18 (integration test) ◄── ALL ───────────────┘┘┘┘┘┘   │││
```

**Parallelizable groups:**
- **Group A (no deps):** P1-T04, P1-T07, P1-T08 — can all start immediately in parallel
- **Group B (after T01):** P1-T02, P1-T03, P1-T05, P1-T06 — can all start once T01 is done
- **Group C (after A+B):** P1-T09 — depends on T08 and T04
- **Group D (after C):** P1-T10, P1-T11, P1-T12, P1-T13 — can partially parallelize
- **Group E (after D):** P1-T14, P1-T15, P1-T16, P1-T17 — sequential
- **Group F (after E):** P1-T18 — integration test

---

## Readiness Assessment

This plan is ready to implement. A senior Go engineer can pick up any Phase 1 task whose dependencies are met and implement it without asking architectural questions. The acceptance criteria are specific enough to write tests from, the file paths are concrete, and the dependency order is explicit.

**Remaining unknowns (zero-risk):**
- The exact MBP frame encoding format for cluster frames (msgpack vs. raw bytes with headers) needs a convention decision during P1-T08. Recommendation: use msgpack for all cluster frames to match existing MBP convention.
- The `ConnManager` reconnection backoff parameters (initial 2s, max 30s) may need tuning under real network conditions, but the defaults are safe for Phase 1.

**Confidence level:** High. Phase 1 produces a shippable 3-node cluster with automatic failover, push-based replication, and stable operation. The no-prune Phase 1 constraint eliminates snapshot complexity while still allowing node joins. The nil-worker guard for Lobes is verified against the existing codebase.
