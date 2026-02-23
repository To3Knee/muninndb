# How Memory Works

*The neuroscience behind MuninnDB's architecture — why this model is correct, not just clever.*

---

## The question databases never asked

Every database ever built answers the same question: "What data do I have?"

MuninnDB answers a different question: "What do I *remember*?"

These sound similar. They are not.

A library and a brain are both ways of storing information. A library is static. Every book is equally present, equally accessible, equally relevant. The library doesn't know which shelf you've visited most. It doesn't know that two books are about the same idea. It won't pull a book off the shelf and set it on your desk because you might need it. You go to the library, ask for something specific, and the library returns it exactly.

A brain is dynamic. It knows that what you thought about yesterday is more present than what you thought about a year ago. It discovers connections between ideas you never consciously linked. It surfaces memories you didn't know you needed. It forgets what stopped mattering.

Every database before MuninnDB was a library. Perfectly reliable. Perfectly static.

MuninnDB is a brain. It does something no database has ever done.

---

## Hermann Ebbinghaus and the forgetting curve (1885)

In 1885, a German psychologist named Hermann Ebbinghaus did something that seems slightly absurd: he spent years memorizing thousands of meaningless syllable combinations — "DAX," "BUP," "ZOL" — and then meticulously recorded how fast he forgot them. He did this to himself, repeatedly, for years.

The result was the most important quantitative contribution to memory science in history.

He discovered that forgetting isn't random. It follows a precise mathematical curve. You lose most of a memory quickly — in the first hours and days. Then the rate slows. The curve flattens. A well-consolidated memory doesn't disappear; it becomes dormant. The forgetting curve never reaches zero.

The formula is:

```
R(t) = e^(-t/S)
```

Where:
- **R** is retention — how much of the memory remains
- **t** is time elapsed since the memory was last active
- **S** is stability — the strength of the memory's consolidation

Stability is the key variable. A memory you've returned to repeatedly has high stability — it decays slowly. A memory you've encountered once has low stability — it fades fast. This is why spaced repetition works as a learning technique: returning to material at increasing intervals raises stability, which flattens the decay curve.

MuninnDB implements this directly. Every engram carries a relevance score that decays according to this function, running continuously in the background. The database moves even when you don't. An engram you haven't activated in six months has a lower relevance score than it did last week. This isn't a scheduled batch job — it's continuous recalculation by the cognitive worker layer.

One design choice deserves attention: MuninnDB applies a floor to the decay function. Relevance never reaches zero. Nothing is ever truly deleted — engrams become dormant. They can be reactivated. This matches biological reality: you don't permanently lose consolidated memories, you lose access pathways to them.

A 19th century German psychologist studying nonsense syllables gave us the mathematical model that MuninnDB uses to decide what's relevant right now. That's not a coincidence. Ebbinghaus discovered something true about the structure of memory. We used it.

---

## Donald Hebb and associative learning (1949)

In 1949, Canadian neuropsychologist Donald Hebb published *The Organization of Behavior*, where he proposed what became the most cited principle in neuroscience:

**Neurons that fire together, wire together.**

The idea is precise: when two neurons activate at the same time, the synaptic connection between them strengthens. Do it enough times, and the connection becomes so strong that activating one neuron automatically activates the other. This is the mechanism behind habits, expertise, and associative memory. It's why the smell of something can pull a memory out of nowhere — sensory neurons and memory neurons fired together enough times that the smell became a retrieval cue.

This is also how expertise consolidates. A chess grandmaster doesn't analyze every move from first principles; patterns activate related patterns automatically. Years of co-activation built a network where activating one node floods the relevant neighborhood instantly.

MuninnDB implements Hebb's Rule in the association graph. Every engram carries a set of association weights — edges to other engrams. When two engrams are retrieved in the same activation event, their association weight increases. The update is multiplicative:

```
w_new = min(1.0, w_old × (1 + η)^n)
```

Where:
- **η** is the learning rate (0.01)
- **n** is co-activation count in this event

Associations strengthen with every co-activation. The `min(1.0, ...)` cap ensures weights stay bounded. The learning rate is intentionally small — associations build gradually through repeated co-activation, not from a single event.

You didn't define this relationship. You never created an edge in a schema. You never wrote a migration saying "payment service connects to idempotency keys." The connection emerged from how you used the system — exactly the way expertise emerges from practice.

This is fundamentally different from a graph database, where edges are static and hand-defined. In Neo4j, you create a relationship and it exists until you delete it, unchanged, unweighted, indifferent to whether you ever traversed it again. Hebb's Rule produces dynamic relationships that strengthen when they're useful and fade when they're not.

---

## Bayesian confidence: not all memory is equally trustworthy

Here's a problem that no database has ever seriously addressed: you can be wrong.

You remember something confidently. Then you encounter contradicting information. How confident should you be now? If someone tells you the opposite thing three times, but you've heard the original ten times, what's the right confidence level?

The answer is Bayesian updating. MuninnDB maintains a confidence score between 0 and 1 for every engram. That score is not a label you set manually — it updates automatically when new information arrives.

The update formula:

```
posterior = (p × s) / (p × s + (1 - p) × (1 - s))
```

Where:
- **p** is the current confidence (prior)
- **s** is the signal strength of the new evidence

When contradicting information is detected, s is low — confidence updates downward. When reinforcing information arrives, s is high — confidence updates upward. With Laplace smoothing applied, the system is stable at the extremes: confidence doesn't collapse to zero or lock at one.

The practical implication: if you tell MuninnDB something and later tell it the opposite, MuninnDB doesn't blindly trust the latest message. It tracks the tension. It flags the contradiction explicitly. It updates confidence based on accumulated evidence, not recency alone.

This is how expert knowledge management actually works. A good analyst doesn't replace their model every time they see new data — they update it. MuninnDB is doing the same thing with every engram.

Confidence is also visible as a query parameter. You can ask MuninnDB to return only engrams above a confidence threshold. You can filter out low-confidence memories when you need reliable information. You can surface low-confidence memories specifically when you're doing a review or audit. The confidence score is a first-class property of every memory, not an afterthought.

---

## The spacing effect: access patterns change stability

Ebbinghaus discovered something beyond the forgetting curve: the spacing effect.

Two people memorize the same list. One reviews it 50 times in a single afternoon. The other reviews it 50 times over six months, with increasing gaps between sessions. Six months later, the second person remembers far more. Same number of repetitions. Completely different outcome.

Spaced repetition builds stability in a way that massed repetition does not. Each time you return to a memory after a gap, the retrieval itself — pulling the memory back into consciousness — strengthens it. The effort of retrieval is the consolidation mechanism. Cramming doesn't produce stable memories because you're not retrieving; you're just re-reading.

MuninnDB tracks this. Stability increases with access count, but the pattern of access matters. An engram activated 50 times over six months has higher stability than one activated 50 times in a single day. The database weights the history, not just the count.

The result: memories that are used consistently, sustainably, over time become increasingly resistant to decay. They become long-term fixtures. Memories that were activated in a burst — during an intensive project, say — decay faster once the project ends. This is correct behavior. The system reflects the real epistemic weight of the memory.

---

## Why this matters for AI agents

AI agents today have a memory problem.

The most common approaches are:

**Context windows.** The agent holds recent messages in its prompt. This works for a single session. It resets when the session ends. There is no persistence, no decay, no learning, no connection between sessions. The agent wakes up with amnesia every time.

**Vector databases.** Embed everything, retrieve by similarity. This is the current state of the art for AI memory, and it's still fundamentally wrong. Similarity is not relevance. "Find things similar to my current query" is not the same as "tell me what I should be thinking about right now, given what I've learned, how recently I learned it, and what connects to what." A vector database returns the 10 most similar chunks. It has no concept of whether those chunks were important last month and forgotten for good reason. It has no concept of confidence in their accuracy. It never initiates contact.

**Key-value stores with retrieval logic.** Hand-written systems where someone decided which keys to write, how to look them up, and when to expire them. These work until the domain gets complex, at which point they require constant maintenance and still miss connections that weren't explicitly programmed.

None of these are memory. They're storage with retrieval. The distinction matters.

MuninnDB gives AI agents the same memory model that evolution spent hundreds of millions of years perfecting in biological brains. Relevant things surface. Irrelevant things fade. Connections form automatically. Contradictions get flagged. And when something becomes urgent — because related concepts were activated, because time shifted the relevance landscape, because a contradiction appeared — the database tells the agent before the agent asks.

That last part is new. Every database before MuninnDB was passive. You query it and it responds. MuninnDB has a native push mechanism: subscribe to a context, and the database will deliver relevant engrams to you when relevance changes. Not when you ask. When it matters.

This is the architecture that Ebbinghaus, Hebb, and Bayes were pointing toward. They described how biological memory actually works. MuninnDB is the implementation.
