package main

import (
"fmt"
"io/fs"
"net"
"net/http"
"os"
"path/filepath"
"strconv"
"strings"
"time"
)

// 65. bmm_fs: snapshot file size & mod unix nano, or compare against prev "size:modUnixNano"
func lib65FsWatchSnap(filePath, prevSnap string) string {
info, err := os.Stat(strings.TrimSpace(filePath))
if err != nil {
return "ERR_STAT: " + err.Error()
}
cur := fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano())
prev := strings.TrimSpace(prevSnap)
if prev == "" {
return "SNAP: " + cur
}
if cur == prev {
return "UNCHANGED | SNAP: " + cur
}
return "CHANGED | PREV: " + prev + " -> CUR: " + cur
}

// 66. bmm_fs: find files under root dir matching suffix/extension (e.g. ".bmm" or ".go")
func lib66FsFind(rootDir, suffix string) string {
root := strings.TrimSpace(rootDir)
if root == "" {
root = "."
}
ext := strings.TrimSpace(suffix)
var matches []string
_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
if err != nil || d == nil || d.IsDir() {
return nil
}
if ext == "" || strings.HasSuffix(d.Name(), ext) {
matches = append(matches, p)
}
return nil
})
if len(matches) == 0 {
return "NO_MATCHES"
}
if len(matches) > 15 {
return fmt.Sprintf("COUNT=%d | SAMPLE=%s", len(matches), strings.Join(matches[:15], ","))
}
return fmt.Sprintf("COUNT=%d | FILES=%s", len(matches), strings.Join(matches, ","))
}

// 67. bmm_net: measure TCP connect latency in milliseconds to host:port
func lib67NetPingMs(hostPort, timeoutMsStr string) string {
hp := strings.TrimSpace(hostPort)
tMs, err := strconv.Atoi(strings.TrimSpace(timeoutMsStr))
if err != nil || tMs <= 0 || tMs > 15000 {
tMs = 2000
}
start := time.Now()
conn, err := net.DialTimeout("tcp", hp, time.Duration(tMs)*time.Millisecond)
ms := float64(time.Since(start).Microseconds()) / 1000.0
if err != nil {
return fmt.Sprintf("TARGET: %s | REACHABLE: false | ELAPSED_MS: %.2f | ERR: %s", hp, ms, err.Error())
}
conn.Close()
return fmt.Sprintf("TARGET: %s | REACHABLE: true | LATENCY_MS: %.2f", hp, ms)
}

// 68. bmm_net: quick HTTP GET status check with timeout
func lib68HttpProbe(rawURL, timeoutSecStr string) string {
u := strings.TrimSpace(rawURL)
if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
u = "http://" + u
}
sec, err := strconv.Atoi(strings.TrimSpace(timeoutSecStr))
if err != nil || sec <= 0 || sec > 20 {
sec = 4
}
client := &http.Client{Timeout: time.Duration(sec) * time.Second}
resp, err := client.Get(u)
if err != nil {
return "HTTP_ERR(" + u + "): " + err.Error()
}
defer resp.Body.Close()
return fmt.Sprintf("HTTP_OK | URL: %s | STATUS: %d %s", u, resp.StatusCode, http.StatusText(resp.StatusCode))
}
