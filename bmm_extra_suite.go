package main

import (
"crypto/sha256"
"encoding/hex"
"fmt"
"os"
"path/filepath"
"runtime"
"strings"
"sync/atomic"
"time"
)

var bmmSuiteCounter uint64

// BmmSmallSlug converts text into a clean URL/identifier-friendly slug.
func BmmSmallSlug(input string) string {
s := strings.ToLower(strings.TrimSpace(input))
var b strings.Builder
prevDash := false
for _, r := range s {
if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
b.WriteRune(r)
prevDash = false
} else if !prevDash {
b.WriteByte('-')
prevDash = true
}
}
return strings.Trim(b.String(), "-")
}

// BmmSmallClamp restricts an integer value between minVal and maxVal.
func BmmSmallClamp(val, minVal, maxVal int) int {
if minVal > maxVal {
minVal, maxVal = maxVal, minVal
}
if val < minVal {
return minVal
}
if val > maxVal {
return maxVal
}
return val
}

// BmmSmallEnvOr returns an environment variable or a fallback value.
func BmmSmallEnvOr(key, fallback string) string {
if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
return v
}
return fallback
}

// BmmMedJoinPath safely joins path segments using OS-clean rules.
func BmmMedJoinPath(parts ...string) string {
return filepath.Join(parts...)
}

// BmmMedCompactSpaces normalizes whitespace for clean log/payload formatting.
func BmmMedCompactSpaces(raw string) string {
return strings.Join(strings.Fields(raw), " ")
}

// BmmMedRetryPlan formats a deterministic backoff plan for medium services.
func BmmMedRetryPlan(attempts, baseDelayMs int) string {
if attempts < 1 {
attempts = 1
}
if baseDelayMs < 1 {
baseDelayMs = 10
}
var parts []string
for i := 1; i <= attempts; i++ {
parts = append(parts, fmt.Sprintf("try#%d=%dms", i, baseDelayMs*i))
}
return strings.Join(parts, " -> ")
}

// BmmLargeMemMB reports current heap allocation in megabytes.
func BmmLargeMemMB() float64 {
var m runtime.MemStats
runtime.ReadMemStats(&m)
return float64(m.HeapAlloc) / (1024.0 * 1024.0)
}

// BmmLargeGoroutines reports active Go runtime goroutines.
func BmmLargeGoroutines() int {
return runtime.NumGoroutine()
}

// BmmLargeTraceTag creates a timestamped execution trace tag.
func BmmLargeTraceTag(scope string) string {
n := atomic.AddUint64(&bmmSuiteCounter, 1)
return fmt.Sprintf("[%s#%d@%d]", scope, n, time.Now().UTC().UnixNano())
}

// BmmMegaSHA256 computes a standard SHA-256 hex digest.
func BmmMegaSHA256(data string) string {
sum := sha256.Sum256([]byte(data))
return hex.EncodeToString(sum[:])
}

// BmmMegaUID generates a unique monotonic identifier with prefix.
func BmmMegaUID(prefix string) string {
if strings.TrimSpace(prefix) == "" {
prefix = "BMM"
}
n := atomic.AddUint64(&bmmSuiteCounter, 1)
return fmt.Sprintf("%s-%d-%06d", strings.ToUpper(prefix), time.Now().UTC().Unix(), n)
}

// BmmMegaAuditCheck verifies required non-empty configuration keys.
func BmmMegaAuditCheck(pairs map[string]string) (bool, []string) {
var missing []string
for k, v := range pairs {
if strings.TrimSpace(v) == "" {
missing = append(missing, k)
}
}
return len(missing) == 0, missing
}

// RunBmmExtraSuiteDemo executes a live demonstration across all 4 project tiers.
func RunBmmExtraSuiteDemo() string {
slug := BmmSmallSlug("BMM Engine Next Gen Build")
clamped := BmmSmallClamp(145, 0, 100)
envVal := BmmSmallEnvOr("BMM_MODE", "production")

joined := BmmMedJoinPath("/var/lib/bmm", "modules", "auth.bmm")
compact := BmmMedCompactSpaces("status   ok   latency   12ms")
retry := BmmMedRetryPlan(3, 50)

mem := BmmLargeMemMB()
gr := BmmLargeGoroutines()
trace := BmmLargeTraceTag("CORE-PIPELINE")

hash := BmmMegaSHA256("bmm-enterprise-payload")
uid := BmmMegaUID("ENT")
ok, missing := BmmMegaAuditCheck(map[string]string{
"SERVICE_NAME": "bmm-core",
"REGION":       "me-central-1",
})

return fmt.Sprintf(
"SMALL: slug=%s | clamp=%d | env=%s\n"+
"MEDIUM: path=%s | compact=%s | retry=%s\n"+
"LARGE: mem=%.2fMB | goroutines=%d | trace=%s\n"+
"MEGA: uid=%s | sha256=%s... | audit_ok=%v (missing=%v)",
slug, clamped, envVal,
joined, compact, retry,
mem, gr, trace,
uid, hash[:16], ok, missing,
)
}
