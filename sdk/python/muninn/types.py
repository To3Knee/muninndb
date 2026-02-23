"""MuninnDB type definitions."""

from dataclasses import dataclass, field
from typing import Any


@dataclass
class WriteRequest:
    """Request to write an engram."""

    vault: str
    concept: str
    content: str
    tags: list[str] | None = None
    confidence: float = 0.9
    stability: float = 0.5
    embedding: list[float] | None = None
    associations: dict[str, Any] | None = None


@dataclass
class WriteResponse:
    """Response from writing an engram."""

    id: str
    created_at: int


@dataclass
class ActivateRequest:
    """Request to activate memory."""

    vault: str
    context: list[str]
    max_results: int = 10
    threshold: float = 0.1
    max_hops: int = 0
    include_why: bool = False
    brief_mode: str = "auto"


@dataclass
class ActivationItem:
    """A single activated memory item."""

    id: str
    concept: str
    content: str
    score: float
    confidence: float
    why: str | None = None
    hop_path: list[str] | None = None
    dormant: bool = False


@dataclass
class BriefSentence:
    """A sentence extracted by brief mode."""

    engram_id: str
    text: str
    score: float


@dataclass
class ActivateResponse:
    """Response from activating memory."""

    query_id: str
    total_found: int
    activations: list[ActivationItem]
    latency_ms: float = 0.0
    brief: list[BriefSentence] | None = None


@dataclass
class ReadResponse:
    """Response from reading an engram."""

    id: str
    concept: str
    content: str
    confidence: float
    relevance: float
    stability: float
    access_count: int
    tags: list[str]
    state: str
    created_at: int
    updated_at: int
    last_access: int | None = None
    coherence: "dict[str, CoherenceResult] | None" = None


@dataclass
class CoherenceResult:
    """Coherence metrics for a vault."""

    score: float
    orphan_ratio: float
    contradiction_density: float
    duplication_pressure: float
    decay_variance: float
    total_engrams: int


@dataclass
class StatResponse:
    """Response from stats endpoint."""

    engram_count: int
    vault_count: int
    storage_bytes: int
    coherence: dict[str, CoherenceResult] | None = None


@dataclass
class Push:
    """SSE push event from subscription."""

    subscription_id: str
    trigger: str
    push_number: int
    engram_id: str | None = None
    at: int | None = None
