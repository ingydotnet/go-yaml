# Response to ccoVeille's Concern About yaml v3 Dependency

## Summary

You're correct that v3 appears in the module graph. I've tested this thoroughly and here are the findings:

## Test Results

Created a test consumer module that imports go-yaml/v4:

```bash
# Test module setup
cd /tmp/test-v4-consumer
go mod init example.com/test-consumer
# Added: replace go.yaml.in/yaml/v4 => /path/to/local/v4
```

### What Works (Good News)
- ✅ `go build` succeeds **without downloading v3**
- ✅ `go run` works **without downloading v3**
- ✅ Normal usage doesn't require v3 in module cache

### What Doesn't Work (Your Concern is Valid)
- ❌ `go mod tidy` **does add v3 to go.sum**
- ❌ `go mod graph` **does show v3 as a dependency**
- ❌ v3 appears in the dependency tree even though it's test-only

## Root Cause

**Go doesn't have a mechanism to mark dependencies as "test-only" in go.mod.**

Even though v3 is only imported in `_test.go` files (`internal/libyaml/api_test.go` and `internal/libyaml/yaml_data_test.go`), when you run `go mod tidy`, it adds v3 to the require block because:

1. Go's module system treats all dependencies equally
2. There's no `[test]` section like other languages (Python's `[dev-dependencies]`, etc.)
3. Module graph pruning (Go 1.17+) helps but doesn't eliminate test deps from go.mod

## The Dilemma

Your original concern stands:
> "something added as a dependency, even for tests, becomes a dependency of the code that would use go-yaml"

This is **true in the module graph sense** but **false in the practical build sense**.

## Possible Solutions

### Option 1: Accept the Status Quo (Current Approach)
- Pros: Simplest, standard Go practice
- Cons: v3 appears in module graphs

### Option 2: Separate Test Module
Move test code to a separate module with its own go.mod:
```
internal/libyaml/tests/
├── go.mod        # Can depend on both v3 and v4
├── api_test.go
└── yaml_data_test.go
```
- Pros: v3 truly isolated
- Cons: More complex, non-standard structure

### Option 3: Remove v3, Use Alternative
Rewrite test data parsing to not use any YAML library (manual parsing or JSON).
- Pros: No v3 dependency at all
- Cons: Significant work, loses benefit of YAML-based test data

### Option 4: Use Go Workspaces
Keep tests in a workspace-only module.
- Pros: Clean separation during development
- Cons: Requires Go 1.18+ workspaces

## Recommendation

I lean toward **Option 2** (separate test module) if this is a blocker. It's the cleanest way to truly isolate v3 from v4's module graph.

What do you prefer?

---

**Test files available at:**
- `/tmp/test-v4-consumer/` - proof-of-concept consumer module
- Logs showing `go build` succeeds without v3
- Logs showing `go mod tidy` adds v3 to go.sum
