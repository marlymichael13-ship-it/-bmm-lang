package main

import (
"fmt"
"os"
"runtime"
"strconv"
"strings"
)

func getSystemRAM() string {
data, err := os.ReadFile("/proc/meminfo")
if err != nil {
var m runtime.MemStats
runtime.ReadMemStats(&m)
return fmt.Sprintf("AppAlloc: %d MB", m.Alloc/1024/1024)
}

var totalKB, availKB uint64
lines := strings.Split(string(data), "\n")
for _, line := range lines {
fields := strings.Fields(line)
if len(fields) < 2 {
continue
}
switch fields[0] {
case "MemTotal:":
totalKB, _ = strconv.ParseUint(fields[1], 10, 64)
case "MemAvailable:":
availKB, _ = strconv.ParseUint(fields[1], 10, 64)
}
}

if totalKB > 0 {
totalMB := totalKB / 1024
availMB := availKB / 1024
usedMB := totalMB - availMB
return fmt.Sprintf("Total: %d MB | Used: %d MB | Available: %d MB", totalMB, usedMB, availMB)
}
return "RAM_INFO_UNAVAILABLE"
}

func getSysArchInfo() string {
return fmt.Sprintf("OS: %s | ARCH: %s | CPUs: %d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}

func readFileSafe(path string) string {
content, err := os.ReadFile(strings.TrimSpace(path))
if err != nil {
return "ERR_FILE_READ: " + err.Error()
}
return string(content)
}

func writeFileSafe(path, content string) string {
err := os.WriteFile(strings.TrimSpace(path), []byte(content), 0644)
if err != nil {
return "ERR_FILE_WRITE: " + err.Error()
}
return "OK_WRITTEN"
}

func fileExistsSafe(path string) bool {
_, err := os.Stat(strings.TrimSpace(path))
return err == nil
}
