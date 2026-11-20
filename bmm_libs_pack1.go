package main

import (
"fmt"
"os"
"path/filepath"
"runtime"
"strconv"
"strings"
"time"
)

// 1. ram_lib: read total, used, and available RAM from /proc/meminfo
func lib1RamInfo() string {
data, err := os.ReadFile("/proc/meminfo")
if err != nil {
var m runtime.MemStats
runtime.ReadMemStats(&m)
return fmt.Sprintf("AppAlloc: %d MB", m.Alloc/1024/1024)
}
var totalKB, availKB uint64
for _, line := range strings.Split(string(data), "\n") {
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

// 2. cpu_lib: read OS, architecture, and CPU count
func lib2CpuInfo() string {
return fmt.Sprintf("OS: %s | ARCH: %s | CPUs: %d", runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}

// 3. battery_lib: read battery capacity and status if exposed in /sys
func lib3BatteryInfo() string {
capBytes, errCap := os.ReadFile("/sys/class/power_supply/battery/capacity")
statBytes, errStat := os.ReadFile("/sys/class/power_supply/battery/status")
capStr := "N/A"
statStr := "N/A"
if errCap == nil {
capStr = strings.TrimSpace(string(capBytes)) + "%"
}
if errStat == nil {
statStr = strings.TrimSpace(string(statBytes))
}
if errCap != nil && errStat != nil {
return "BATTERY_UNAVAILABLE_IN_ENV"
}
return fmt.Sprintf("Capacity: %s | Status: %s", capStr, statStr)
}

// 4. disk_lib: read home directory path and basic storage info
func lib4DiskInfo() string {
home, err := os.UserHomeDir()
if err != nil || home == "" {
home = "/data/data/com.termux/files/home"
}
info, err := os.Stat(home)
if err != nil {
return "DISK_ERR: " + err.Error()
}
return fmt.Sprintf("Home: %s | Mode: %s", home, info.Mode().String())
}

// 5. env_lib: read environment variable or hostname
func lib5EnvInfo(key string) string {
k := strings.TrimSpace(key)
if k == "" || strings.EqualFold(k, "hostname") {
h, err := os.Hostname()
if err != nil {
return "ERR_HOSTNAME: " + err.Error()
}
return h
}
val, ok := os.LookupEnv(k)
if !ok {
return "ENV_NOT_SET"
}
return val
}

// 6. fs_read_lib: read file content safely
func lib6FsRead(path string) string {
data, err := os.ReadFile(strings.TrimSpace(path))
if err != nil {
return "ERR_FS_READ: " + err.Error()
}
return string(data)
}

// 7. fs_write_lib: write or append text to a file
func lib7FsWrite(path, content string, appendMode bool) string {
cleanPath := strings.TrimSpace(path)
if !appendMode {
if err := os.WriteFile(cleanPath, []byte(content), 0644); err != nil {
return "ERR_FS_WRITE: " + err.Error()
}
return "OK_WRITTEN"
}
f, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
if err != nil {
return "ERR_FS_APPEND: " + err.Error()
}
defer f.Close()
if _, err := f.WriteString(content); err != nil {
return "ERR_FS_APPEND_WRITE: " + err.Error()
}
return "OK_APPENDED"
}

// 8. fs_manage_lib: copy, move, or remove a file safely
func lib8FsManage(action, src, dst string) string {
act := strings.ToLower(strings.TrimSpace(action))
s := strings.TrimSpace(src)
d := strings.TrimSpace(dst)
switch act {
case "copy":
data, err := os.ReadFile(s)
if err != nil {
return "ERR_COPY_READ: " + err.Error()
}
if err := os.WriteFile(d, data, 0644); err != nil {
return "ERR_COPY_WRITE: " + err.Error()
}
return "OK_COPIED"
case "move", "rename":
if err := os.Rename(s, d); err != nil {
return "ERR_MOVE: " + err.Error()
}
return "OK_MOVED"
case "delete", "remove":
if err := os.Remove(s); err != nil {
return "ERR_DELETE: " + err.Error()
}
return "OK_DELETED"
default:
return "ERR_UNKNOWN_ACTION"
}
}

// 9. fs_dir_lib: create directory or list entries
func lib9FsDir(action, dirPath string) string {
act := strings.ToLower(strings.TrimSpace(action))
p := strings.TrimSpace(dirPath)
switch act {
case "mkdir":
if err := os.MkdirAll(p, 0755); err != nil {
return "ERR_MKDIR: " + err.Error()
}
return "OK_MKDIR"
case "list", "ls":
entries, err := os.ReadDir(p)
if err != nil {
return "ERR_LS: " + err.Error()
}
var names []string
for _, e := range entries {
n := e.Name()
if e.IsDir() {
n += "/"
}
names = append(names, n)
}
return strings.Join(names, ", ")
default:
return "ERR_UNKNOWN_DIR_ACTION"
}
}

// 10. fs_meta_lib: return file existence, size, and modified timestamp
func lib10FsMeta(path string) string {
p := filepath.Clean(strings.TrimSpace(path))
info, err := os.Stat(p)
if err != nil {
return "EXISTS: false | ERR: " + err.Error()
}
return fmt.Sprintf("EXISTS: true | IS_DIR: %v | SIZE: %d | MODIFIED: %s",
info.IsDir(), info.Size(), info.ModTime().UTC().Format(time.RFC3339))
}
