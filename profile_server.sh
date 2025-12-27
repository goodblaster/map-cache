#!/bin/bash

# Profile map-cache server under load
#
# This script captures CPU and memory profiles while running sustained load tests.
# The profiles can be analyzed with: go tool pprof <profile-file>

set -e

PROFILE_DIR="./profiles"
DURATION=30s

echo "=== Map-Cache Server Profiling ==="
echo ""

# Create profiles directory
mkdir -p "$PROFILE_DIR"

echo "1. Capturing CPU profile (${DURATION})..."
curl -s "http://localhost:8080/debug/pprof/profile?seconds=30" > "${PROFILE_DIR}/cpu.prof" &
CPU_PID=$!

echo "2. Starting load test..."
go test -v -run TestRESP_SustainedLoad -timeout 2m ./tests/resp_stress_test.go ./tests/resp_new_commands_test.go > "${PROFILE_DIR}/load_test.log" 2>&1 &
LOAD_PID=$!

# Wait for CPU profile to complete
wait $CPU_PID
echo "   CPU profile saved to ${PROFILE_DIR}/cpu.prof"

# Wait for load test to complete
wait $LOAD_PID
echo "   Load test completed"

echo ""
echo "3. Capturing heap profile..."
curl -s "http://localhost:8080/debug/pprof/heap" > "${PROFILE_DIR}/heap.prof"
echo "   Heap profile saved to ${PROFILE_DIR}/heap.prof"

echo ""
echo "4. Capturing goroutine profile..."
curl -s "http://localhost:8080/debug/pprof/goroutine" > "${PROFILE_DIR}/goroutine.prof"
echo "   Goroutine profile saved to ${PROFILE_DIR}/goroutine.prof"

echo ""
echo "5. Capturing allocation profile..."
curl -s "http://localhost:8080/debug/pprof/allocs" > "${PROFILE_DIR}/allocs.prof"
echo "   Allocation profile saved to ${PROFILE_DIR}/allocs.prof"

echo ""
echo "6. Capturing mutex contention profile..."
curl -s "http://localhost:8080/debug/pprof/mutex" > "${PROFILE_DIR}/mutex.prof"
echo "   Mutex profile saved to ${PROFILE_DIR}/mutex.prof"

echo ""
echo "=== Profiling Complete ==="
echo ""
echo "Profiles saved in: ${PROFILE_DIR}/"
echo ""
echo "Analyze profiles with:"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/cpu.prof"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/heap.prof"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/goroutine.prof"
echo "  go tool pprof -http=:8081 ${PROFILE_DIR}/mutex.prof"
echo ""
echo "Or use text mode:"
echo "  go tool pprof -top ${PROFILE_DIR}/cpu.prof"
echo "  go tool pprof -top ${PROFILE_DIR}/heap.prof"
echo ""
