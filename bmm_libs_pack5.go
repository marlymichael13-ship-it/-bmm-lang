package main

import (
"crypto/md5"
"crypto/rand"
"crypto/sha256"
"encoding/hex"
"fmt"
"math"
"net"
"os"
"path/filepath"
"strconv"
"strings"
"time"
)

// 41. sec_hash_lib: compute sha256/md5 or verify expected hash
func lib41SecHash(mode, algo, target, expected string) string {
a := strings.ToLower(strings.TrimSpace(algo))
var digest string
switch a {
case "md5":
sum := md5.Sum([]byte(target))
digest = hex.EncodeToString(sum[:])
default:
sum := sha256.Sum256([]byte(target))
digest = hex.EncodeToString(sum[:])
}
if strings.EqualFold(strings.TrimSpace(mode), "verify") {
exp := strings.ToLower(strings.TrimSpace(expected))
if exp != "" && digest == exp {
return "INTEGRITY_VALID: " + digest
}
return "INTEGRITY_MISMATCH: got=" + digest
}
return digest
}

// 42. sec_token_lib: cryptographically secure random hex token
func lib42SecToken(byteLen int) string {
if byteLen < 4 {
byteLen = 4
}
if byteLen > 64 {
byteLen = 64
}
buf := make([]byte, byteLen)
if _, err := rand.Read(buf); err != nil {
return "ERR_RNG: " + err.Error()
}
return hex.EncodeToString(buf)
}

// 43. sec_mask_lib: mask sensitive string keeping visible prefix/suffix
func lib43SecMask(secret string, keepEdge int) string {
r := []rune(secret)
if len(r) == 0 {
return ""
}
if keepEdge < 0 {
keepEdge = 0
}
if len(r) <= keepEdge*2 {
return strings.Repeat("*", len(r))
}
prefix := string(r[:keepEdge])
suffix := string(r[len(r)-keepEdge:])
midLen := len(r) - keepEdge*2
return prefix + strings.Repeat("*", midLen) + suffix
}

// 44. sec_perm_lib: audit file/dir permissions for risky modes
func lib44SecPermAudit(path string) string {
p := filepath.Clean(strings.TrimSpace(path))
info, err := os.Stat(p)
if err != nil {
return "ERR_STAT: " + err.Error()
}
mode := info.Mode().Perm()
worldWrite := (mode & 0002) != 0
groupWrite := (mode & 0020) != 0
risk := "LOW"
if worldWrite {
risk = "HIGH_WORLD_WRITABLE"
} else if groupWrite {
risk = "MEDIUM_GROUP_WRITABLE"
}
return fmt.Sprintf("PATH: %s | MODE: %04o | WORLD_W: %v | GROUP_W: %v | RISK: %s",
p, mode, worldWrite, groupWrite, risk)
}

// 45. sec_payload_lib: scan input for common injection/traversal patterns
func lib45SecPayloadScan(input string) string {
low := strings.ToLower(input)
var hits []string
patterns := map[string][]string{
"PATH_TRAVERSAL": {"../", "..\\"},
"SQLI_PATTERN":   {"' or ", "\" or ", "union select", "drop table", "--"},
"XSS_PATTERN":    {"<script", "javascript:", "onerror=", "onload="},
"CMD_INJECTION":  {"; rm ", "| sh", "| bash", "$(", "`"},
}
for label, list := range patterns {
for _, pat := range list {
if strings.Contains(low, pat) {
hits = append(hits, label)
break
}
}
}
if len(hits) == 0 {
return "CLEAN | RISK: NONE"
}
return fmt.Sprintf("ALERT | RISK: HIGH | HITS: %s", strings.Join(hits, ","))
}

// 46. sys_sec_lib: audit host/process/env security posture
func lib46SysSecAudit() string {
uid := os.Getuid()
euid := os.Geteuid()
ldPreload, hasPreload := os.LookupEnv("LD_PRELOAD")
pathVal, _ := os.LookupEnv("PATH")
var findings []string
if uid == 0 || euid == 0 {
findings = append(findings, "ROOT_PRIVILEGE")
}
if hasPreload && strings.TrimSpace(ldPreload) != "" {
findings = append(findings, "LD_PRELOAD_SET")
}
for _, seg := range strings.Split(pathVal, ":") {
s := strings.TrimSpace(seg)
if s == "" || s == "." {
findings = append(findings, "UNSAFE_DOT_IN_PATH")
break
}
}
if len(findings) == 0 {
return fmt.Sprintf("UID: %d | EUID: %d | POSTURE: HARDENED_BASELINE", uid, euid)
}
return fmt.Sprintf("UID: %d | EUID: %d | POSTURE: REVIEW_NEEDED | FLAGS: %s",
uid, euid, strings.Join(findings, ","))
}

// 47. cps_sensor_lib: detect out-of-bounds or abrupt sensor spikes (CSV numbers)
func lib47CpsSensorGuard(csvVals, minStr, maxStr, maxStepStr string) string {
minV, err1 := strconv.ParseFloat(strings.TrimSpace(minStr), 64)
maxV, err2 := strconv.ParseFloat(strings.TrimSpace(maxStr), 64)
stepLimit, err3 := strconv.ParseFloat(strings.TrimSpace(maxStepStr), 64)
if err1 != nil || err2 != nil || err3 != nil {
return "ERR_CPS_PARAMS"
}
var vals []float64
for _, part := range strings.Split(csvVals, ",") {
t := strings.TrimSpace(part)
if t == "" {
continue
}
v, err := strconv.ParseFloat(t, 64)
if err != nil {
return "ERR_SENSOR_VAL: " + t
}
vals = append(vals, v)
}
if len(vals) == 0 {
return "ERR_NO_SAMPLES"
}
var anomalies []string
for i, v := range vals {
if math.IsNaN(v) || math.IsInf(v, 0) || v < minV || v > maxV {
anomalies = append(anomalies, fmt.Sprintf("idx%d_out_of_range(%.2f)", i, v))
}
if i > 0 && math.Abs(v-vals[i-1]) > stepLimit {
anomalies = append(anomalies, fmt.Sprintf("idx%d_spike(%.2f)", i, math.Abs(v-vals[i-1])))
}
}
if len(anomalies) == 0 {
return fmt.Sprintf("CPS_SENSOR_OK | SAMPLES: %d", len(vals))
}
return fmt.Sprintf("CPS_ANOMALY_DETECTED | FLAGS: %s", strings.Join(anomalies, ";"))
}

// 48. robo_act_lib: validate robotic actuator command against safety envelope
func lib48RoboActuatorGuard(cmdValStr, safeMinStr, safeMaxStr, maxTempStr, curTempStr string) string {
cmdV, err1 := strconv.ParseFloat(strings.TrimSpace(cmdValStr), 64)
minV, err2 := strconv.ParseFloat(strings.TrimSpace(safeMinStr), 64)
maxV, err3 := strconv.ParseFloat(strings.TrimSpace(safeMaxStr), 64)
maxT, err4 := strconv.ParseFloat(strings.TrimSpace(maxTempStr), 64)
curT, err5 := strconv.ParseFloat(strings.TrimSpace(curTempStr), 64)
if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
return "ERR_ACTUATOR_PARAMS"
}
var faults []string
if curT >= maxT {
faults = append(faults, "THERMAL_LIMIT_EXCEEDED")
}
if cmdV < minV || cmdV > maxV {
faults = append(faults, "COMMAND_OUT_OF_ENVELOPE")
}
if len(faults) > 0 {
return "ACTUATOR_INTERLOCK_TRIPPED | ACTION: SAFE_STOP | FAULTS: " + strings.Join(faults, ",")
}
return fmt.Sprintf("ACTUATOR_PERMITTED | CMD: %.2f | TEMP: %.1f/%.1fC", cmdV, curT, maxT)
}

// 49. net_sec_lib: check target port risk and TCP reachability
func lib49NetSecCheck(hostPort string, timeoutMs int) string {
hp := strings.TrimSpace(hostPort)
host, portStr, err := net.SplitHostPort(hp)
if err != nil {
return "ERR_HOST_PORT: " + err.Error()
}
portNum, _ := strconv.Atoi(portStr)
riskyPorts := map[int]string{
21: "FTP_CLEARTEXT", 23: "TELNET_CLEARTEXT", 80: "HTTP_UNENCRYPTED",
445: "SMB_EXPOSURE", 3389: "RDP_EXPOSURE",
}
portTag := "PORT_STANDARD"
if r, ok := riskyPorts[portNum]; ok {
portTag = "PORT_RISK_" + r
}
if timeoutMs <= 0 || timeoutMs > 10000 {
timeoutMs = 2000
}
conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, portStr), time.Duration(timeoutMs)*time.Millisecond)
if err != nil {
return fmt.Sprintf("TARGET: %s | %s | TCP: CLOSED_OR_FILTERED", hp, portTag)
}
conn.Close()
return fmt.Sprintf("TARGET: %s | %s | TCP: OPEN", hp, portTag)
}

// 50. neuro_sec_lib: audit neuromorphic/SNN spike rate and weight stability (CSV)
func lib50NeuroSecAudit(spikeRatesCsv, weightsCsv, maxRateStr, maxWeightStr string) string {
maxRate, errR := strconv.ParseFloat(strings.TrimSpace(maxRateStr), 64)
maxW, errW := strconv.ParseFloat(strings.TrimSpace(maxWeightStr), 64)
if errR != nil || errW != nil || maxRate <= 0 || maxW <= 0 {
return "ERR_NEURO_THRESHOLDS"
}
var alerts []string
var sumRate float64
var rateCount int
for i, part := range strings.Split(spikeRatesCsv, ",") {
t := strings.TrimSpace(part)
if t == "" {
continue
}
r, err := strconv.ParseFloat(t, 64)
if err != nil {
return "ERR_SPIKE_VAL: " + t
}
rateCount++
sumRate += r
if r > maxRate {
alerts = append(alerts, fmt.Sprintf("spike_runaway_n%d(%.2f)", i, r))
}
if r < 0 {
alerts = append(alerts, fmt.Sprintf("negative_rate_n%d", i))
}
}
for i, part := range strings.Split(weightsCsv, ",") {
t := strings.TrimSpace(part)
if t == "" {
continue
}
w, err := strconv.ParseFloat(t, 64)
if err != nil {
return "ERR_WEIGHT_VAL: " + t
}
if math.Abs(w) > maxW {
alerts = append(alerts, fmt.Sprintf("weight_drift_w%d(%.2f)", i, w))
}
}
avgRate := 0.0
if rateCount > 0 {
avgRate = sumRate / float64(rateCount)
}
if len(alerts) == 0 {
return fmt.Sprintf("NEURO_SEC_STABLE | NEURONS: %d | AVG_HZ: %.2f", rateCount, avgRate)
}
return fmt.Sprintf("NEURO_SEC_ALERT | AVG_HZ: %.2f | FLAGS: %s", avgRate, strings.Join(alerts, ";"))
}
