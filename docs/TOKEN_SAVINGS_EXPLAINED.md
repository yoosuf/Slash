# Slash: How It Reduces Token Usage in Claude

**Problem:** Large codebases blow through Claude's context window, forcing expensive pagination or losing critical context.

**Solution:** Slash compresses code, logs, and JSON by 40–60% while keeping everything reversible via the `retrieve()` tool.

---

## 🎯 The Token Problem

### How Claude's Tokens Work

Claude processes text in **tokens** (~4 characters ≈ 1 token):
- **Input tokens** — everything you send to Claude (your code, the prompt, context)
- **Output tokens** — Claude's response
- **Cost** — roughly proportional to input + output tokens
- **Latency** — more tokens = slower API calls
- **Context window** — Claude has a max limit (200k, 400k, etc. depending on model)

### The Pain

A typical software engineer's workflow:

```
User: "Fix the bug in my auth service"

Claude needs:
- Your codebase (50 files, 200KB) = 50,000 tokens
- Error logs (100KB) = 25,000 tokens
- Relevant docs (50KB) = 12,500 tokens
- Your prompt (1KB) = 250 tokens
- Earlier conversation history

Total: ~90,000 tokens (just to provide context!)
```

**Result:**
- ❌ Slow API response (more tokens to process)
- ❌ Expensive (billed per token)
- ❌ Risk of hitting context limit (can't include everything)
- ❌ Quality loss (Claude loses context on large requests)

---

## ✨ How Slash Solves It

Slash compresses that same context by **40–60%** using 4 techniques:

### 1. **Diff-Only Re-Reads** (80–95% savings on re-reads)

**Problem:** You ask Claude to read a file, edit it, ask Claude to read it again.

```
First read: "Here's main.go" (5KB, 1,250 tokens)
Edit: You change 3 lines
Second read: Claude re-reads entire file (5KB, 1,250 tokens again!)
```

**Slash solution:** Track what changed. Return only the diff.

```
First read: "Here's main.go" (5KB, 1,250 tokens)
Edit: You change 3 lines
Second read: Slash returns only changed lines (50 bytes, 12 tokens!)
```

**Savings:** 1,238 tokens per re-read. On a multi-turn conversation with 10 edits, that's 12,380 tokens saved.

---

### 2. **Output Compression** (40–60% savings on all output)

**Problem:** Claude reads large files/logs/JSON and returns the full content.

#### JSON Compression

**Before:**
```json
{
  "id": 12345,
  "user": {
    "name": "Alice",
    "email": "alice@example.com",
    "profile": {
      "avatar": "https://...",
      "bio": "Software engineer...",
      "preferences": {...}
    }
  },
  "data": [100 items, each 200 bytes]
}
```

Size: 20KB = 5,000 tokens

**After (Slash skeleton):**
```json
{
  "id": "<number>",
  "user": {
    "name": "<string>",
    "email": "<string>",
    "profile": "<object with 3 keys>"
  },
  "data": "[array of 100 items]"
}
```

Size: 300 bytes = 75 tokens
**Savings: 4,925 tokens (98%!)**

Claude can still see the structure. If it needs a specific value, it calls:
```
retrieve(h_abc123, start=200, end=500)  // Get bytes 200-500 of original
```

#### Code Compression

**Before:** 1,000-line file
```go
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // 500 lines of implementation
    ...
}

func (s *Server) ValidateInput(data []byte) error {
    // 200 lines of validation
    ...
}

// 10 more functions...
```

Size: 40KB = 10,000 tokens

**After (Slash summary):**
```go
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // [implementation: 500 lines omitted]
}

func (s *Server) ValidateInput(data []byte) error {
    // [implementation: 200 lines omitted]
}

// 10 more function signatures [omitted]
```

Size: 2KB = 500 tokens
**Savings: 9,500 tokens (95%!)**

Claude sees function signatures (what it needs). Implementation details are in `retrieve(h_def456)`.

#### Log Compression

**Before:** 2,000 lines of repeated errors
```
[ERROR] Connection timeout: failed to connect to server
[ERROR] Connection timeout: failed to connect to server
[ERROR] Connection timeout: failed to connect to server
... [repeated 2,000 times]
[ERROR] Connection timeout: failed to connect to server
[WARN] Retry limit exceeded
```

Size: 80KB = 20,000 tokens

**After (Slash dedup):**
```
[ERROR] Connection timeout: failed to connect to server
  ... [repeated 1,998 more times]
[WARN] Retry limit exceeded
```

Size: 200 bytes = 50 tokens
**Savings: 19,950 tokens (99%!)**

Claude knows the error happened 2000 times without reading it 2000 times.

---

### 3. **Diff-Only On Edits** (40–80% savings on review cycles)

**Scenario:** You ask Claude to review your changes.

**Before:**
```
Claude reads entire file (5KB, 1,250 tokens)
You changed 20 lines
Claude still reads the whole file to review
```

**After (Slash):**
```
Claude gets: "20 lines changed: [the diff]" (500 bytes, 125 tokens)
Claude reviews just the changes
```

**Savings: 1,125 tokens per review**

---

### 4. **Repo Map at Session Start** (10–15% faster, fewer exploratory reads)

**Problem:** Claude asks "What's in this codebase?" and you end up copy-pasting 10 files.

**Slash solution:** Inject a symbol index at the start.

```json
{
  "modules": [
    "auth" -> ["Login", "Register", "ValidateToken"],
    "api" -> ["Server", "Router", "Handler"],
    "db" -> ["Connect", "Query", "Close"]
  ],
  "dependencies": ["express", "postgres", "jwt"]
}
```

Claude now knows the structure without reading files. Fewer "Can you show me X?" questions.

**Savings: ~10% fewer tool calls, each with large file reads**

---

## 📊 Real-World Example: Code Review

### Without Slash

```
User: "Review my authentication changes"

Claude needs:
- Original auth.go (10KB) = 2,500 tokens
- Original middleware.go (8KB) = 2,000 tokens
- Original config.go (5KB) = 1,250 tokens
- Error logs from test run (50KB) = 12,500 tokens
- Test output (20KB) = 5,000 tokens
- Conversation history = 5,000 tokens

Total: ~28,250 tokens just to get context

Claude responds: ~2,000 tokens

Total cost: ~30,250 tokens
Total time: ~2 seconds (API latency for 28k tokens)
```

### With Slash

```
User: "Review my authentication changes"

Claude needs:
- Diff of auth.go (300 bytes) = 75 tokens
- Diff of middleware.go (200 bytes) = 50 tokens
- Diff of config.go (150 bytes) = 37 tokens
- Compressed error logs (skeleton) (5KB) = 1,250 tokens
- Compressed test output (deduped) (2KB) = 500 tokens
- Conversation history = 5,000 tokens

Total: ~6,912 tokens

Claude responds: ~2,000 tokens

Total cost: ~8,912 tokens (70% reduction!)
Total time: ~0.6 seconds (4x faster)
```

**Impact:**
- **Cost:** $0.27 → $0.08 (70% cheaper)
- **Time:** 2s → 0.6s (3x faster)
- **Quality:** Same review, no context loss (can retrieve() full files if needed)

---

## 💰 Cost Savings at Scale

### Per-conversation savings

| Scenario | Tokens Before | Tokens After | Savings | Cost Savings |
|---|---|---|---|---|
| **Code review** | 30,000 | 9,000 | 70% | 70% |
| **Bug investigation** | 50,000 | 20,000 | 60% | 60% |
| **Test debugging** | 80,000 | 32,000 | 60% | 60% |
| **Feature implementation** | 100,000 | 40,000 | 60% | 60% |
| **Architecture review** | 120,000 | 48,000 | 60% | 60% |

### Annual cost (engineer using Claude daily)

```
Conversations per day: 10
Tokens per conversation (without Slash): 50,000
Days per year: 250

Total tokens/year: 10 × 50,000 × 250 = 125,000,000 tokens

Claude Sonnet pricing: ~$3 per 1M input tokens

Cost without Slash: 125M × ($3/M) = $375/year per engineer

With Slash (60% reduction):
Cost: 125M × 40% × ($3/M) = $150/year per engineer

**Savings per engineer: $225/year**
**For 100 engineers: $22,500/year**
```

---

## ⚡ Speed Improvements

**Token count directly affects latency.**

Fewer tokens = faster API response:

```
Tokens: 10,000  → API latency: ~0.5s
Tokens: 30,000  → API latency: ~1.5s
Tokens: 50,000  → API latency: ~2.5s
```

**With Slash (60% reduction):**
```
Without: 50,000 tokens → 2.5s latency
With: 20,000 tokens → 1.0s latency
```

**Improvement: 2.5x faster** (feels snappier, more iterations per hour)

---

## 🎯 Quality Benefits (Beyond Token Savings)

### 1. **No Context Loss**
Without Slash, large requests force you to prune context. You choose: include file A or file B?

With Slash, include both. Keep everything. Compression is *reversible*.

### 2. **Better Agent Performance**
Claude's agentic tools (read, bash, apply_patch) are faster and cheaper:
- More iterations per context window
- Faster per-iteration latency
- Can explore more branches before hitting token limit

### 3. **Works With Smaller Models**
Slash makes code within reach of smaller, cheaper Claude models:
- Claude Haiku is cheaper but has smaller context
- Slash lets Haiku handle tasks that would need Sonnet
- Result: cheaper + still high quality

### 4. **Longer Conversations**
Same context window, but compressed content:
```
Without Slash: 10-turn conversation (200k tokens)
With Slash: 25-turn conversation (still 200k tokens, but more useful content)
```

---

## 📈 The Math: Why 40–60% Reduction Is Possible

### JSON Example
```
Before: {key: "value", id: 12345, active: true, ...}
After: {key: "<string>", id: "<number>", active: "<boolean>", ...}
Reduction: 60–70% (structure preserved, values stripped)
```

### Code Example
```
Before: function SignUp() { ... 200 lines of implementation ... }
After: function SignUp() { ... [200 lines omitted] ... }
Reduction: 80–95% (signature visible, impl in retrieve())
```

### Log Example
```
Before: [ERROR] timeout (repeated 1000x) = 20KB
After: [ERROR] timeout ... [repeated 1000x] = 200 bytes
Reduction: 99% (intent clear, dedup works)
```

**Key insight:** Most token waste is *redundancy and detail* Claude doesn't need right now (but can retrieve if it does).

---

## 🔄 How retrieve() Keeps Quality High

Claude knows it can call `retrieve(handle, [range])` anytime:

```python
# Claude's reasoning:
# "I need to understand the full auth logic"

retrieve("h_auth123")  # → full auth.go (5KB, 1,250 tokens)
# Now Claude has detail. Tradeoff: 1,250 tokens for certainty.

# vs. if it guesses from skeleton:
# "Based on function signatures, I think auth works like..."
# Risk: wrong assumption, wastes more tokens on follow-up questions
```

**Net result:** Claude makes better decisions (fewer wasted tokens) because it can fetch detail on demand.

---

## 🚀 Real-World Impact on Claude

### Current Pain Point
Engineer works on large codebase (500 files, 200KB code):
- "Show me the auth module" → 10 files, 100KB → 25,000 tokens
- "Show me middleware" → 5 files, 50KB → 12,500 tokens
- "Show me config" → 3 files, 30KB → 7,500 tokens
- Total: 45,000 tokens just for context setup
- Claude has 155,000 tokens left for actual reasoning (200k context)
- Result: Can't include enough detail, quality suffers

### With Slash
Same setup:
- "Skeleton of auth module" → 1KB → 250 tokens
- "Skeleton of middleware" → 500 bytes → 125 tokens
- "Skeleton of config" → 300 bytes → 75 tokens
- Total: 450 tokens
- Claude has 199,550 tokens left for reasoning + retrieval
- Result: Can explore more, retrieve detail only when needed, higher quality

### Impact on Workflows

**Workflow 1: Debugging**
- Without Slash: "Here's the full error log, here's the code" → 60,000 tokens for context
- With Slash: "Here's the key errors (deduplicated), here's the skeleton code" → 15,000 tokens
- Claude can now spend 185,000 tokens on reasoning, analysis, and solutions

**Workflow 2: Code Review**
- Without Slash: "Review my PRwhole files context (80,000 tokens)
- With Slash: "Review my PR" (diffs only, 5,000 tokens)
- Claude can include detailed feedback, examples, and suggestions (195,000 tokens available)

**Workflow 3: Refactoring**
- Without Slash: "Refactor this service" (include relevant code: 70,000 tokens)
- With Slash: "Refactor this service" (include skeleton + retrieve as needed: 10,000 tokens)
- Claude explores more refactor options, considers edge cases

---

## 📝 Summary: The Value Proposition

| Factor | Without Slash | With Slash | Improvement |
|---|---|---|---|
| **Token Usage** | 50,000 per call | 20,000 per call | 60% reduction |
| **Cost** | $0.15 | $0.06 | 60% cheaper |
| **Latency** | 2.5s | 1.0s | 2.5x faster |
| **Context Left** | 150,000 | 180,000 | 20% more space |
| **Quality** | Limited (less context) | Higher (more reasoning) | Better |
| **Iterations** | 10 turns (200k context) | 25 turns (same context) | 2.5x more |

**Bottom line:**
Slash makes Claude:
- **Cheaper** (40–60% less tokens)
- **Faster** (2–3x lower latency)
- **Smarter** (more reasoning budget per token)
- **More capable** (can handle larger codebases without pagination)

---

## 🎓 For Claude Engineers

When Claude processes code with Slash:

**Without Slash:**
```
Input: Full codebase (50k tokens) + prompt (500 tokens)
Processing: Read 50k tokens, find relevant bits, reason on 150k tokens
Output: 2,000 tokens
Total: 52.5k input + 2k output = ~$0.15
```

**With Slash:**
```
Input: Skeleton codebase (10k tokens) + prompt (500 tokens)
Processing: Read 10.5k tokens, reason on 189.5k tokens, retrieve (5k) as needed
Output: 3,000 tokens (better quality, more thorough)
Total: 18.5k input + 3k output = ~$0.06
```

**Claude's perspective:**
- Spends 60% less time reading boilerplate
- Spends 25% more time thinking
- Can retrieve detail when uncertain
- Result: Better code, better advice

---

## 🔗 How to Use Slash With Claude

```bash
# 1. Start Slash daemon
slash daemon

# 2. Use Claude normally (plugin auto-compresses)
# Your code gets compressed before Claude sees it

# 3. If Claude needs detail
# It calls retrieve(handle) automatically
# You see: "[slash: fetching full auth.go...]"

# 4. Continue (compression is transparent)
```

**You don't change your workflow.** Slash runs in the background.

---

**Result: Claude gets the job done faster, cheaper, with better quality.**

**That's what Slash does for token usage.** 🚀
