# Cluster Management UI/UX Design

**Date:** 2026-02-23
**Status:** Approved

## Overview

Enable full cluster lifecycle management from the MuninnDB Web UI and CLI: enable cluster mode, join nodes, monitor health, trigger failovers, and manage security — all without touching config files.

---

## 1. Tab Structure

The existing Cluster tab gains three sub-tabs:

- **Overview** — health banner, members table, CCS gauge, failover timeline, security posture card
- **Management** — add node, remove node, trigger failover, live activity feed
- **Settings** — join token, TLS certificate, cluster behavior tuning

When cluster mode is **off**, the tab shows only an "Enable Cluster Mode" card (no sub-tabs).

---

## 2. Node Join Flow

**Token model:** kubeadm-style, time-limited (15 min default), derived from cluster secret via HMAC. Operators distribute the token — not the raw secret. Token is displayed with a copy button and countdown timer.

**Add Node modal (Management tab):**
1. Enter node address (IP:port)
2. Enter join token
3. Click **Test Reachability** — server-side MBP handshake ping, returns latency or error inline
4. Click **Add Node** — confirmation modal with quorum impact warning if applicable
5. Live progress stream in modal: "Sending join request → Validating token → Streaming snapshot (if needed) → Catching up log entries → Node joined ✓"

**CLI:** `muninn cluster add-node --addr=10.0.0.2:7777 --token=<token>`

---

## 3. Remove Node & Manual Failover

**Remove Node:**
- Per-row Remove button in members table
- Confirmation modal shows: node ID, role, last-seen, quorum impact warning
- Option: "Drain before removing" (waits for LastApplied == Cortex.CurrentSeq(), 30s default timeout)
- Progress stream: "Sending leave signal → Draining (if selected) → Removed ✓"

**Manual Failover:**
- "Trigger Failover" button at top of Management tab (disabled unless current node is Cortex)
- Target node selector (dropdown of eligible Lobes, shows lag in real-time)
- Confirmation modal with safety check badge (entries behind)
- Progress stream: "Sending handoff → New Cortex elected → Handoff acknowledged → Complete ✓"

**CLI:** `muninn cluster remove-node --id=<nodeID>`, `muninn cluster failover --target=<nodeID>`

---

## 4. Cluster Settings

Three grouped sections in the Settings sub-tab:

**Join Token**
- Current token with copy button and validity countdown
- "Regenerate Token" button (confirmation modal)
- Token validity window (configurable, default 15m)

**TLS**
- Cert expiry display (color-coded: green > 30d, yellow 7–30d, red < 7d)
- "Rotate Certificate" button (confirmation modal)
- mTLS toggle (locked to ON once cluster has >1 member)

**Cluster Behavior**
- Heartbeat interval (ms)
- SDOWN timeout (missed beats)
- CCS probe interval (seconds)
- Reconcile on partition-heal (toggle)
- Save button per section, inline "Saved ✓" confirmation

---

## 5. Enable Cluster Flow

Single centered card when cluster mode is off:

**Fields:**
- Role: Cortex | Lobe | Sentinel
- Bind address (pre-filled with detected IP:port)
- Cluster secret (optional; inline warning if empty)
- Cortex address (Lobe/Sentinel only)

**On Enable:**
- Confirmation modal: "This will restart the replication subsystem. Existing data is not affected."
- Full-screen overlay progress stream: "Initializing TLS → Generating join token → Starting heartbeat → Cluster active ✓"
- Errors shown inline with "View logs" link

**CLI:** `muninn cluster enable --role=cortex --addr=10.0.0.1:7777 --secret=<secret>`

---

## 6. Security Posture Panel

Read-only card in the Overview sub-tab:

| Item | Status |
|------|--------|
| mTLS | Enabled / Disabled |
| Cert expires | Date (color-coded) |
| Join token | Valid for Xm Ys / Expired |
| Cluster secret | Configured / Not set |
| In-flight encryption | TLS 1.3 |

---

## 7. Live Activity Feed

Scrolling stream in the Management sub-tab, SSE-powered (`/api/admin/cluster/events`), showing:
- Replication entries applied (seq, key prefix, op type)
- Node join/leave events
- Election events (candidate, promoted, demoted)
- CCS score changes
- Errors and warnings with log-level coloring

Auto-pauses on hover, resumes on mouse-out. Max 200 entries in DOM.

---

## 8. API Endpoints (new)

All under `/api/admin/cluster/` (session auth required):

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/enable` | Enable cluster mode |
| POST | `/disable` | Disable cluster mode |
| POST | `/nodes` | Add a node (join) |
| DELETE | `/nodes/:id` | Remove a node |
| POST | `/failover` | Trigger manual failover |
| POST | `/token/regenerate` | Regenerate join token |
| POST | `/tls/rotate` | Rotate TLS certificate |
| PUT | `/settings` | Update cluster behavior settings |
| GET | `/token/test` | Test reachability of a node address |
| GET | `/events` | SSE stream of cluster events |

---

## 9. Data Flow

```
Browser → Admin REST API → ClusterCoordinator
                        ↘ JoinHandler / JoinClient
                        ↘ Election
                        ↘ TLS (RotateCert)
                        ↘ ReplicationLog (Subscribe → SSE)
```

Config changes persist to the cluster config file (same path as `--config`). Sensitive fields (cluster secret) are write-only from UI — never returned in GET responses.

---

## 10. Error Handling

- All modals show inline errors (no page reload)
- "View logs" deep-links to the Logs tab filtered to `cluster` source
- Quorum warnings are non-blocking (can proceed with explicit confirmation)
- Test Reachability failures show specific error: timeout, TLS mismatch, wrong protocol, refused

---

## Implementation Notes

- Match existing UI look and feel (Tailwind classes, same modal component pattern as existing admin modals)
- SSE event stream reuses `http.Flusher` pattern already used in log tail
- Join token stored in-memory on Cortex (not persisted — regenerated on restart); Lobes obtain it out-of-band
- No new dependencies required
