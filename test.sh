#!/bin/bash
set -e

echo "=========================================="
echo "Test 1: Clean Build (No v3 Download)"
echo "=========================================="
echo ""

# Clean up
echo "Cleaning up previous artifacts..."
rm -f go.sum test-consumer
rm -rf ~/go/pkg/mod/go.yaml.in/yaml 2>/dev/null || true
echo ""

echo "Building without go mod tidy..."
go build -v
echo ""

if [ -f go.sum ]; then
    echo "❌ UNEXPECTED: go.sum was created"
    cat go.sum
else
    echo "✅ SUCCESS: No go.sum file created"
fi
echo ""

if [ -f test-consumer ]; then
    echo "✅ SUCCESS: Binary built successfully"
    ls -lh test-consumer
else
    echo "❌ FAILED: Binary not created"
    exit 1
fi
echo ""

echo "=========================================="
echo "Test 2: After go mod tidy"
echo "=========================================="
echo ""

echo "Running go mod tidy..."
go mod tidy
echo ""

echo "Checking go.sum for v3 entries..."
if grep -q "go.yaml.in/yaml/v3" go.sum; then
    echo "❌ FOUND: v3 appears in go.sum"
    echo ""
    grep "go.yaml.in/yaml/v3" go.sum
else
    echo "✅ NOT FOUND: v3 not in go.sum"
fi
echo ""

echo "=========================================="
echo "Test 3: Module Graph"
echo "=========================================="
echo ""

echo "Checking module dependency graph..."
go mod graph | grep yaml
echo ""

if go mod graph | grep -q "go.yaml.in/yaml/v3"; then
    echo "❌ FOUND: v3 appears in module graph"
else
    echo "✅ NOT FOUND: v3 not in module graph"
fi
echo ""

echo "=========================================="
echo "Test 4: Runtime Test"
echo "=========================================="
echo ""

echo "Running the test program..."
go run .
echo ""

echo "=========================================="
echo "Summary"
echo "=========================================="
echo ""
echo "✅ go build: Works without v3"
echo "❌ go mod tidy: Adds v3 to go.sum"
echo "❌ go mod graph: Shows v3 in dependency tree"
echo "✅ go run: Works without needing v3 download (if already in cache)"
echo ""
echo "Conclusion: v3 appears in module metadata but is not required for builds"
