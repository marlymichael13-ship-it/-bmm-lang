package main

import (
"crypto/hmac"
"crypto/sha256"
"encoding/hex"
"fmt"
"math"
"strconv"
"strings"
)

// 69. bmm_sec: compute or verify HMAC-SHA256 hex digest
func lib69HmacSha256(mode, secret, message, expectedHex string) string {
mac := hmac.New(sha256.New, []byte(secret))
mac.Write([]byte(message))
sum := hex.EncodeToString(mac.Sum(nil))
if strings.ToLower(strings.TrimSpace(mode)) == "verify" {
if hmac.Equal([]byte(sum), []byte(strings.ToLower(strings.TrimSpace(expectedHex)))) {
return "HMAC_VALID"
}
return "HMAC_INVALID"
}
return sum
}

// 70. bmm_neuro: compute firing rate stats and saturation/silence check from spike CSV (e.g. "0,1,0,1,1,0")
func lib70SpikeRateCheck(spikesCsv, windowSecStr, maxHzStr string) string {
winSec, errW := strconv.ParseFloat(strings.TrimSpace(windowSecStr), 64)
if errW != nil || winSec <= 0 {
winSec = 1.0
}
maxHz, errM := strconv.ParseFloat(strings.TrimSpace(maxHzStr), 64)
if errM != nil || maxHz <= 0 {
maxHz = 100.0
}
parts := strings.Split(strings.TrimSpace(spikesCsv), ",")
var spikes int
var total int
for _, p := range parts {
s := strings.TrimSpace(p)
if s == "" {
continue
}
total++
if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 {
spikes++
}
}
if total == 0 {
return "ERR_EMPTY_SPIKES"
}
hz := float64(spikes) / winSec
status := "NOMINAL"
if hz > maxHz {
status = "OVER_RATE_ALERT"
} else if spikes == 0 {
status = "SILENT_ALERT"
}
return fmt.Sprintf("SPIKE_RATE | SPIKES=%d/%d | HZ=%.2f (max=%.2f) | STATUS=%s",
spikes, total, math.Round(hz*100)/100, maxHz, status)
}
