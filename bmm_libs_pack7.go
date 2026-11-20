package main

import (
"fmt"
"net"
"os"
"os/exec"
"runtime"
"strconv"
"strings"
"time"
)

// 51. hw_cpu_info: inspect CPU cores, arch, and /proc/cpuinfo details
func lib51HwCpuInfo() string {
arch := runtime.GOARCH
goos := runtime.GOOS
cpus := runtime.NumCPU()
model := "unknown"
if data, err := os.ReadFile("/proc/cpuinfo"); err == nil {
for _, ln := range strings.Split(string(data), "\n") {
low := strings.ToLower(ln)
if strings.HasPrefix(low, "model name") || strings.HasPrefix(low, "processor") || strings.HasPrefix(low, "hardware") {
parts := strings.SplitN(ln, ":", 2)
if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
model = strings.TrimSpace(parts[1])
break
}
}
}
}
return fmt.Sprintf("ARCH: %s | OS: %s | CORES: %d | CPU_REF: %s", arch, goos, cpus, model)
}

// 52. hw_cpu_bind: bind Go runtime threads/affinity count and measure fast cycle
func lib52HwCpuBind(threadsStr string) string {
n, err := strconv.Atoi(strings.TrimSpace(threadsStr))
if err != nil || n <= 0 {
n = runtime.NumCPU()
}
maxProcs := runtime.NumCPU()
if n > maxProcs {
n = maxProcs
}
prev := runtime.GOMAXPROCS(n)
start := time.Now()
var acc uint64
for i := uint64(1); i <= 250000; i++ {
acc += (i * 2654435761) & 0xffffffff
}
elapsedUs := time.Since(start).Microseconds()
return fmt.Sprintf("CPU_BOUND: GOMAXPROCS=%d (prev=%d) | ACC=%d | FAST_US=%d", n, prev, acc, elapsedUs)
}

// 53. hw_to_asm: fast emitter translating simple ops into real target assembly
func lib53HwToAsm(opSpec string) string {
start := time.Now()
arch := runtime.GOARCH
parts := strings.Fields(strings.ToUpper(strings.TrimSpace(opSpec)))
if len(parts) == 0 {
return "ERR_EMPTY_ASM_SPEC"
}
cmd := parts[0]
a := "X0"
b := "X1"
if len(parts) >= 2 {
a = parts[1]
}
if len(parts) >= 3 {
b = parts[2]
}
var asmLines []string
if arch == "arm64" {
switch cmd {
case "ADD":
asmLines = []string{"ADD X0, " + a + ", " + b, "RET"}
case "SUB":
asmLines = []string{"SUB X0, " + a + ", " + b, "RET"}
case "MOV":
asmLines = []string{"MOV " + a + ", " + b, "RET"}
case "XOR":
asmLines = []string{"EOR X0, " + a + ", " + b, "RET"}
default:
asmLines = []string{"NOP", "RET"}
}
} else {
switch cmd {
case "ADD":
asmLines = []string{"mov rax, " + a, "add rax, " + b, "ret"}
case "SUB":
asmLines = []string{"mov rax, " + a, "sub rax, " + b, "ret"}
case "MOV":
asmLines = []string{"mov " + a + ", " + b, "ret"}
case "XOR":
asmLines = []string{"mov rax, " + a, "xor rax, " + b, "ret"}
default:
asmLines = []string{"nop", "ret"}
}
}
us := time.Since(start).Microseconds()
return fmt.Sprintf("ASM[%s %dus]: %s", arch, us, strings.Join(asmLines, " ; "))
}

// 54. hw_net_mod: inspect or toggle network interfaces/modules
func lib54HwNetMod(action, iface string) string {
act := strings.ToLower(strings.TrimSpace(action))
ifName := strings.TrimSpace(iface)
if act == "list" || act == "" {
ifaces, err := net.Interfaces()
if err != nil {
return "ERR_NET_IFACE: " + err.Error()
}
var out []string
for _, it := range ifaces {
state := "DOWN"
if it.Flags&net.FlagUp != 0 {
state = "UP"
}
out = append(out, it.Name+"("+state+")")
}
return "NET_MODULES: " + strings.Join(out, ", ")
}
if ifName == "" {
return "ERR_IFACE_REQUIRED"
}
var cmd *exec.Cmd
if act == "up" {
cmd = exec.Command("sh", "-c", "ip link set dev "+ifName+" up 2>&1 || ifconfig "+ifName+" up 2>&1")
} else if act == "down" {
cmd = exec.Command("sh", "-c", "ip link set dev "+ifName+" down 2>&1 || ifconfig "+ifName+" down 2>&1")
} else {
return "ERR_NET_MOD_ACT"
}
out, err := cmd.CombinedOutput()
txt := strings.TrimSpace(string(out))
if err != nil {
if txt == "" {
txt = err.Error()
}
return "NET_MOD_RESTRICTED(" + ifName + "): " + txt
}
if txt == "" {
txt = "APPLIED"
}
return "NET_MOD_OK(" + ifName + " -> " + act + "): " + txt
}

// 55. hw_bt_mod: inspect or control Bluetooth module state
func lib55HwBtMod(action string) string {
act := strings.ToLower(strings.TrimSpace(action))
if act == "status" || act == "" {
if _, err := exec.LookPath("termux-bluetooth-enable"); err == nil {
return "BT_MODULE: TERMUX_API_AVAILABLE"
}
if entries, err := os.ReadDir("/sys/class/bluetooth"); err == nil && len(entries) > 0 {
var names []string
for _, e := range entries {
names = append(names, e.Name())
}
return "BT_MODULE: SYS_BT(" + strings.Join(names, ",") + ")"
}
return "BT_MODULE: KERNEL_OR_HAL_MANAGED"
}
if act == "on" || act == "enable" {
if _, err := exec.LookPath("termux-bluetooth-enable"); err == nil {
out, err := exec.Command("termux-bluetooth-enable").CombinedOutput()
if err == nil {
return "BT_ENABLE_SENT: " + strings.TrimSpace(string(out))
}
}
return "BT_CTRL_NOTE: USE_ANDROID_OR_ROOT_RFKILL"
}
if act == "off" || act == "disable" {
return "BT_CTRL_NOTE: USE_ANDROID_OR_ROOT_RFKILL"
}
return "ERR_BT_ACTION"
}

// 56. hw_wifi_mod: inspect or control Wi-Fi module/interface
func lib56HwWifiMod(action, iface string) string {
act := strings.ToLower(strings.TrimSpace(action))
dev := strings.TrimSpace(iface)
if dev == "" {
dev = "wlan0"
}
if act == "status" || act == "" {
if it, err := net.InterfaceByName(dev); err == nil {
up := (it.Flags & net.FlagUp) != 0
addrs, _ := it.Addrs()
var ips []string
for _, a := range addrs {
ips = append(ips, a.String())
}
return fmt.Sprintf("WIFI_DEV: %s | UP: %v | ADDRS: %s", dev, up, strings.Join(ips, ","))
}
return "WIFI_DEV: " + dev + " | NOT_PRESENT"
}
if act == "scan" {
if _, err := exec.LookPath("termux-wifi-scaninfo"); err == nil {
out, err := exec.Command("termux-wifi-scaninfo").CombinedOutput()
if err == nil {
s := strings.TrimSpace(string(out))
if len(s) > 180 {
s = s[:180] + "..."
}
return "WIFI_SCAN: " + s
}
}
return "WIFI_SCAN_NOTE: TERMUX_API_UNAVAILABLE_OR_RESTRICTED"
}
return lib54HwNetMod(act, dev)
}

// 57. hw_chip_ctrl: inspect device chips, buses, and sysfs nodes
func lib57HwChipCtrl(busFilter string) string {
f := strings.ToLower(strings.TrimSpace(busFilter))
paths := []string{"/sys/bus/i2c/devices", "/sys/bus/spi/devices", "/sys/bus/platform/devices"}
var found []string
for _, p := range paths {
if f != "" && f != "all" && !strings.Contains(p, "/"+f+"/") {
continue
}
entries, err := os.ReadDir(p)
if err != nil {
continue
}
busName := strings.Split(p, "/")[3]
cnt := len(entries)
sample := ""
if cnt > 0 {
sample = entries[0].Name()
}
found = append(found, fmt.Sprintf("%s(count=%d,sample=%s)", busName, cnt, sample))
}
if len(found) == 0 {
return "CHIPS_BUSES: RESTRICTED_SYSFS_IN_USERLAND"
}
return "CHIPS_BUSES: " + strings.Join(found, " | ")
}

// 58. hw_sock_ctrl: inspect active sockets or bind/test a local TCP socket
func lib58HwSockCtrl(mode, addr string) string {
m := strings.ToLower(strings.TrimSpace(mode))
target := strings.TrimSpace(addr)
if m == "tcp_count" || m == "" {
cnt := 0
for _, p := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
if data, err := os.ReadFile(p); err == nil {
lines := strings.Split(strings.TrimSpace(string(data)), "\n")
if len(lines) > 1 {
cnt += len(lines) - 1
}
}
}
return fmt.Sprintf("SOCKETS_TCP_ENTRIES: %d", cnt)
}
if m == "probe" {
if target == "" {
target = "127.0.0.1:8080"
}
conn, err := net.DialTimeout("tcp", target, 1200*time.Millisecond)
if err != nil {
return "SOCK_PROBE(" + target + "): UNREACHABLE (" + err.Error() + ")"
}
local := conn.LocalAddr().String()
remote := conn.RemoteAddr().String()
conn.Close()
return fmt.Sprintf("SOCK_PROBE(%s): CONNECTED | LOCAL=%s -> REMOTE=%s", target, local, remote)
}
if m == "listen_test" {
if target == "" {
target = "127.0.0.1:0"
}
ln, err := net.Listen("tcp", target)
if err != nil {
return "SOCK_LISTEN_ERR: " + err.Error()
}
bound := ln.Addr().String()
ln.Close()
return "SOCK_BOUND_OK: " + bound
}
return "ERR_SOCK_MODE"
}

// 59. hw_board_info: inspect motherboard / SoC / system identity
func lib59HwBoardInfo() string {
var fields []string
sysFiles := map[string]string{
"BOARD":   "/sys/devices/virtual/dmi/id/board_name",
"VENDOR":  "/sys/devices/virtual/dmi/id/board_vendor",
"PRODUCT": "/sys/devices/virtual/dmi/id/product_name",
"SOC":     "/sys/devices/soc0/family",
}
for k, p := range sysFiles {
if b, err := os.ReadFile(p); err == nil {
v := strings.TrimSpace(string(b))
if v != "" {
fields = append(fields, k+"="+v)
}
}
}
if len(fields) == 0 {
if out, err := exec.Command("uname", "-a").Output(); err == nil {
fields = append(fields, "KERNEL="+strings.TrimSpace(string(out)))
}
}
if len(fields) == 0 {
return "BOARD_INFO: UNAVAILABLE"
}
return "BOARD_SYS: " + strings.Join(fields, " | ")
}

// 60. hw_mod_map: unified hardware module & bus map summary
func lib60HwModMap() string {
return fmt.Sprintf("HW_MAP[arch=%s,os=%s,cpus=%d,gomaxprocs=%d] => CPU+ASM+NET+BT+WIFI+CHIP+SOCK+BOARD",
runtime.GOARCH, runtime.GOOS, runtime.NumCPU(), runtime.GOMAXPROCS(0))
}
