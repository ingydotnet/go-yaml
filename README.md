# Proof of Concept: yaml v3 Dependency Test

This directory contains a test consumer module to verify whether go-yaml/v4's test-only dependency on v3 propagates to downstream users.

**Context:** This relates to [PR #194](https://github.com/yaml/go-yaml/pull/194) where concerns were raised about yaml v3 being a dependency.

## Quick Start

### Option A: Test with this PR branch (data-driven-test-style)

```bash
# Clone the go-yaml repo and checkout the PR branch
git clone https://github.com/yaml/go-yaml.git
cd go-yaml
git checkout data-driven-test-style

# Navigate to this proof-of-concept
cd .proof-of-concept-consumer

# Run tests (see Test Procedure below)
```

### Option B: Standalone test (modify go.mod)

If you're running this as a standalone gist/repo:

1. Clone this directory
2. Edit `go.mod` and change the replace directive to point to your local go-yaml clone:
   ```
   replace go.yaml.in/yaml/v4 => /path/to/your/go-yaml
   ```

## Automated Test

Run all tests automatically:

```bash
./test.sh
```

This script runs all four tests below and provides a summary.

## Manual Test Procedure

### Test 1: Clean Build (No v3 Download)

```bash
# Start fresh - remove any cached dependencies
rm -f go.sum test-consumer
rm -rf ~/go/pkg/mod/go.yaml.in/yaml

# Build without go mod tidy
go build -v

# Expected: Builds successfully WITHOUT downloading v3
# Expected: No go.sum file is created
```

**Result:** ✅ Build succeeds without needing v3

### Test 2: After go mod tidy

```bash
go mod tidy
cat go.sum
```

**Expected output in go.sum:**
```
go.yaml.in/yaml/v3 v3.0.4 h1:tfq32ie2Jv2UxXFdLJdh3jXuOzWiL1fo0bu/FbuKpbc=
go.yaml.in/yaml/v3 v3.0.4/go.mod h1:DhzuOOF2ATzADvBadXxruRBLzYTpT36CKvDb3+aBEFg=
```

**Result:** ❌ v3 appears in go.sum after `go mod tidy`

### Test 3: Check Module Graph

```bash
go mod graph | grep yaml
```

**Expected output:**
```
example.com/test-consumer go.yaml.in/yaml/v4@v4.0.0-rc.3
go.yaml.in/yaml/v4@v4.0.0-rc.3 go.yaml.in/yaml/v3@v3.0.4
```

**Result:** ❌ v3 appears in the module dependency graph

### Test 4: Verify Runtime Behavior

```bash
go run .
```

**Expected output:**
```
Successfully unmarshaled: map[key:value]
```

**Result:** ✅ Runs successfully without requiring v3 download

## Findings Summary

| Operation | Requires v3? | Notes |
|-----------|--------------|-------|
| `go build` | ❌ No | Builds without downloading v3 |
| `go run` | ❌ No | Runs without downloading v3 |
| `go mod tidy` | ✅ Yes | Adds v3 to go.sum |
| `go mod graph` | ✅ Yes | Shows v3 in dependency graph |
| `go test all` | ✅ Yes | Would need v3 to test v4's tests |

## Conclusion

**The nuanced truth:**

- **For normal use** (build/run): v3 is NOT required and never downloaded
- **For module hygiene** (tidy/graph): v3 DOES appear in the dependency tree
- **Root cause**: Go has no mechanism to mark test-only dependencies in go.mod

Even though v3 is only imported in `_test.go` files in go-yaml/v4, Go's module system still tracks it as a dependency. However, module graph pruning ensures it's not downloaded during normal builds.

## Full Analysis

See [ANALYSIS.md](./ANALYSIS.md) for detailed analysis and potential solutions to this issue.

## Related Links

- [PR #194: Add data-driven tests](https://github.com/yaml/go-yaml/pull/194)
- [Go Modules Reference: Module Graph Pruning](https://go.dev/ref/mod#graph-pruning)
- [Stack Overflow: Test dependencies in Go](https://stackoverflow.com/questions/64071364/best-way-to-use-test-dependencies-in-go-but-prevent-export-them)
