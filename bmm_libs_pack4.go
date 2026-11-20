package main

import (
"fmt"
"net"
"os"
"os/exec"
"sort"
"strconv"
"strings"
"time"
)

// 31. net_check_lib: resolve host or test TCP port reachability
func lib31NetCheck(target string, timeoutMs int) string {
t := strings.TrimSpace(target)
if t == "" {
return "ERR_EMPTY_TARGET"
}
if timeoutMs <= 0 || timeoutMs > 15000 {
timeoutMs = 2500
}
d := net.Dialer{Timeout: time.Duration(timeoutMs) * time.Millisecond}
if strings.Contains(t, ":") {
conn, err := d.Dial("tcp", t)
if err != nil {
return "UNREACHABLE: " + err.Error()
}
conn.Close()
return "REACHABLE_TCP: " + t
}
ips, err := net.LookupHost(t)
if err != nil {
return "DNS_FAIL: " + err.Error()
}
return "RESOLVED: " + strings.Join(ips, ",")
}

// 32. net_ip_lib: list local non-loopback IP addresses
func lib32NetLocalIPs() string {
ifaces, err := net.Interfaces()
if err != nil {
return "ERR_IFACES: " + err.Error()
}
var addrs []string
for _, iface := range ifaces {
if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
continue
}
ifAddrs, err := iface.Addrs()
if err != nil {
continue
}
for _, a := range ifAddrs {
if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP != nil {
addrs = append(addrs, iface.Name+":"+ipNet.IP.String())
}
}
}
if len(addrs) == 0 {
return "NO_ACTIVE_IFACE"
}
return strings.Join(addrs, " | ")
}

// 33. shell_exec_lib: run a shell command and return trimmed output
func lib33ShellExec(cmdStr string) string {
c := strings.TrimSpace(cmdStr)
if c == "" {
return "ERR_EMPTY_CMD"
}
out, err := exec.Command("sh", "-c", c).CombinedOutput()
res := strings.TrimSpace(string(out))
if err != nil {
if res == "" {
return "CMD_ERR (" + err.Error() + ")"
}
return res + " [EXIT_ERR: " + err.Error() + "]"
}
if res == "" {
return "OK_EMPTY_OUTPUT"
}
return res
}

// 34. diag_log_lib: append timestamped log entry to file
func lib34DiagLog(logPath, level, msg string) string {
p := strings.TrimSpace(logPath)
lvl := strings.ToUpper(strings.TrimSpace(level))
if lvl == "" {
lvl = "INFO"
}
entry := fmt.Sprintf("[%s] [%s] %s\n", time.Now().UTC().Format(time.RFC3339), lvl, msg)
f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
if err != nil {
return "ERR_LOG_OPEN: " + err.Error()
}
defer f.Close()
if _, err := f.WriteString(entry); err != nil {
return "ERR_LOG_WRITE: " + err.Error()
}
return "LOGGED_" + lvl
}

// 35. config_lib: get or set KEY=VALUE in a config file
func lib35Config(mode, cfgPath, key, val string) string {
p := strings.TrimSpace(cfgPath)
k := strings.TrimSpace(key)
m := strings.ToLower(strings.TrimSpace(mode))
switch m {
case "set":
lines := []string{}
if data, err := os.ReadFile(p); err == nil {
for _, ln := range strings.Split(string(data), "\n") {
t := strings.TrimSpace(ln)
if t == "" || strings.HasPrefix(t, k+"=") {
continue
}
lines = append(lines, ln)
}
}
lines = append(lines, k+"="+val)
if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
return "ERR_CFG_SET: " + err.Error()
}
return "CFG_SET_" + k
case "get":
data, err := os.ReadFile(p)
if err != nil {
return "ERR_CFG_READ: " + err.Error()
}
for _, ln := range strings.Split(string(data), "\n") {
t := strings.TrimSpace(ln)
if strings.HasPrefix(t, k+"=") {
return strings.TrimPrefix(t, k+"=")
}
}
return "KEY_NOT_FOUND"
default:
return "ERR_CFG_MODE"
}
}

func splitListItems(raw, sep string) []string {
if sep == "" {
sep = ","
}
parts := strings.Split(raw, sep)
var out []string
for _, p := range parts {
t := strings.TrimSpace(p)
if t != "" {
out = append(out, t)
}
}
return out
}

// 36. list_len_lib: return number of items in delimited list
func lib36ListLen(raw, sep string) string {
return strconv.Itoa(len(splitListItems(raw, sep)))
}

// 37. list_get_lib: return item at 0-based index
func lib37ListGet(raw, sep string, idx int) string {
items := splitListItems(raw, sep)
if idx < 0 || idx >= len(items) {
return "ERR_IDX_RANGE"
}
return items[idx]
}

// 38. list_has_lib: check if item exists in list
func lib38ListHas(raw, sep, target string) string {
t := strings.TrimSpace(target)
for _, it := range splitListItems(raw, sep) {
if it == t {
return "true"
}
}
return "false"
}

// 39. list_sort_lib: sort list ascending or descending
func lib39ListSort(raw, sep, order string) string {
items := splitListItems(raw, sep)
if strings.EqualFold(strings.TrimSpace(order), "desc") {
sort.Slice(items, func(i, j int) bool { return items[i] > items[j] })
} else {
sort.Strings(items)
}
return strings.Join(items, sep)
}

// 40. list_unique_lib: deduplicate list preserving order
func lib40ListUnique(raw, sep string) string {
seen := make(map[string]bool)
var out []string
for _, it := range splitListItems(raw, sep) {
if !seen[it] {
seen[it] = true
out = append(out, it)
}
}
if sep == "" {
sep = ","
}
return strings.Join(out, sep)
}
