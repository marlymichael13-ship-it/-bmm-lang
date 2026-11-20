package main

import (
"fmt"
"os"
"path/filepath"
"runtime"
"strings"
"time"
)

// nexusMeshMap links project files, packs, and bmm_* namespaces into one unified view
func nexusMeshMap(baseDir string) string {
dir := strings.TrimSpace(baseDir)
if dir == "" {
dir = "/data/data/com.termux/files/home/bmm_engine"
}
entries, err := os.ReadDir(dir)
if err != nil {
return "NEXUS_MESH_ERR: " + err.Error()
}
var goFiles, bmmFiles []string
for _, e := range entries {
if e.IsDir() {
continue
}
name := e.Name()
if strings.HasSuffix(name, ".go") {
goFiles = append(goFiles, name)
} else if strings.HasSuffix(name, ".bmm") {
bmmFiles = append(bmmFiles, name)
}
}
pkgs := libBmmPkgList()
return fmt.Sprintf("NEXUS_MESH | DIR=%s | GO_MODULES=%d | BMM_SCRIPTS=%d | PACKAGES=[%s]",
dir, len(goFiles), len(bmmFiles), pkgs)
}

// nexusLibBridge unifies cross-library calls into a single coordinated pipeline
func nexusLibBridge(pipelineSpec string) string {
start := time.Now()
steps := strings.Split(pipelineSpec, ";")
var results []string
for _, raw := range steps {
item := strings.TrimSpace(raw)
if item == "" {
continue
}
parts := strings.SplitN(item, ":", 2)
kind := strings.ToLower(strings.TrimSpace(parts[0]))
arg := ""
if len(parts) == 2 {
arg = strings.TrimSpace(parts[1])
}
switch kind {
case "cpu":
results = append(results, lib51HwCpuInfo())
case "asm":
if arg == "" {
arg = "ADD X2 X3"
}
results = append(results, lib53HwToAsm(arg))
case "json":
kv := strings.SplitN(arg, "|", 2)
if len(kv) == 2 {
results = append(results, lib61JsonGet(kv[0], kv[1]))
} else {
results = append(results, "ERR_JSON_PIPE")
}
case "b64":
kv := strings.SplitN(arg, "|", 2)
if len(kv) == 2 {
results = append(results, lib62Base64(kv[0], kv[1]))
} else {
results = append(results, "ERR_B64_PIPE")
}
case "snap":
results = append(results, lib65FsWatchSnap(arg, ""))
case "ping":
if arg == "" {
arg = "127.0.0.1:8080"
}
results = append(results, lib67NetPingMs(arg, "400"))
case "hmac":
kv := strings.SplitN(arg, "|", 3)
if len(kv) == 3 {
results = append(results, lib69HmacSha256("sign", kv[0], kv[1], ""))
} else {
results = append(results, "ERR_HMAC_PIPE")
}
case "spike":
kv := strings.SplitN(arg, "|", 3)
if len(kv) == 3 {
results = append(results, lib70SpikeRateCheck(kv[0], kv[1], kv[2]))
} else {
results = append(results, "ERR_SPIKE_PIPE")
}
default:
results = append(results, "UNLINKED_STEP("+kind+")")
}
}
us := time.Since(start).Microseconds()
return fmt.Sprintf("NEXUS_BRIDGE[%dus] => %s", us, strings.Join(results, " || "))
}

// nexusSysLink binds the language binary, runtime, and OS environment into one unified status
func nexusSysLink(binPath string) string {
p := strings.TrimSpace(binPath)
if p == "" {
p = "/data/data/com.termux/files/usr/bin/bmm_lang"
}
binStatus := "MISSING"
if info, err := os.Stat(p); err == nil {
binStatus = fmt.Sprintf("READY(%dB)", info.Size())
}
host, _ := os.Hostname()
cwd, _ := os.Getwd()
return fmt.Sprintf("NEXUS_SYS_LINK | BIN=%s[%s] | ARCH=%s/%s | CORES=%d | HOST=%s | CWD=%s",
filepath.Base(p), binStatus, runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), host, cwd)
}
