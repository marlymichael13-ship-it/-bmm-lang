package main

import (
"fmt"
"strconv"
"strings"
)

const (
uiReset  = "\x1b[0m"
uiRed    = "\x1b[31m"
uiBlue   = "\x1b[34m"
uiYellow = "\x1b[33m"
uiGreen  = "\x1b[32m"
uiCyan   = "\x1b[36m"
uiBold   = "\x1b[1m"
uiClear  = "\x1b[2J\x1b[H"
)

// 21. ui_num_red: format numbers in red ANSI color
func lib21UiNumRed(val string) string {
return uiRed + uiBold + strings.TrimSpace(val) + uiReset
}

// 22. ui_item_blue: format items or labels in blue ANSI color
func lib22UiItemBlue(val string) string {
return uiBlue + uiBold + val + uiReset
}

// 23. ui_table_yellow: format pipe-separated rows as a yellow table
func lib23UiTableYellow(title, rowsSpec string) string {
var b strings.Builder
b.WriteString(uiYellow + uiBold)
if strings.TrimSpace(title) != "" {
b.WriteString("=== " + strings.TrimSpace(title) + " ===\n")
}
b.WriteString("+----------------------+----------------------+\n")
for _, line := range strings.Split(rowsSpec, ";") {
row := strings.TrimSpace(line)
if row == "" {
continue
}
cols := strings.SplitN(row, "|", 2)
left := strings.TrimSpace(cols[0])
right := ""
if len(cols) > 1 {
right = strings.TrimSpace(cols[1])
}
b.WriteString(fmt.Sprintf("| %-20s | %-20s |\n", left, right))
}
b.WriteString("+----------------------+----------------------+" + uiReset)
return b.String()
}

// 24. ui_banner: render a framed BMM OS header banner
func lib24UiBanner(sysName, subtitle string) string {
line := strings.Repeat("=", 44)
return uiCyan + uiBold + line + "\n" +
"  " + strings.TrimSpace(sysName) + " :: " + strings.TrimSpace(subtitle) + "\n" +
line + uiReset
}

// 25. ui_badge: render status badge (ok, warn, err, info)
func lib25UiBadge(kind, msg string) string {
k := strings.ToLower(strings.TrimSpace(kind))
switch k {
case "ok", "success":
return uiGreen + uiBold + "[OK] " + uiReset + msg
case "warn", "warning":
return uiYellow + uiBold + "[WARN] " + uiReset + msg
case "err", "error":
return uiRed + uiBold + "[ERR] " + uiReset + msg
default:
return uiBlue + uiBold + "[INFO] " + uiReset + msg
}
}

// 26. ui_progress: render progress bar from 0 to 100
func lib26UiProgress(label string, pct int) string {
if pct < 0 {
pct = 0
}
if pct > 100 {
pct = 100
}
width := 20
filled := (pct * width) / 100
bar := strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
color := uiGreen
if pct < 30 {
color = uiRed
} else if pct < 70 {
color = uiYellow
}
return fmt.Sprintf("%s%s [%s] %s%d%%%s", uiBlue, strings.TrimSpace(label), bar, color, pct, uiReset)
}

// 27. ui_box: render a bordered info box
func lib27UiBox(title, body string) string {
top := uiBlue + "+------------------------------------------+" + uiReset
hdr := uiBlue + "| " + uiBold + strings.TrimSpace(title) + uiReset + uiBlue + " |" + uiReset
mid := uiBlue + "+------------------------------------------+" + uiReset
var out strings.Builder
out.WriteString(top + "\n" + hdr + "\n" + mid + "\n")
for _, ln := range strings.Split(body, "\n") {
out.WriteString(uiBlue + "| " + uiReset + ln + "\n")
}
out.WriteString(uiBlue + "+------------------------------------------+" + uiReset)
return out.String()
}

// 28. ui_divider: render a styled section divider
func lib28UiDivider(ch string, count int) string {
c := strings.TrimSpace(ch)
if c == "" {
c = "-"
}
if count <= 0 || count > 80 {
count = 36
}
return uiYellow + strings.Repeat(c, count) + uiReset
}

// 29. ui_kv: format key/value with blue key and red or normal value
func lib29UiKv(key, val string, redVal bool) string {
v := val
if redVal {
v = uiRed + uiBold + val + uiReset
}
return uiBlue + strings.TrimSpace(key) + ": " + uiReset + v
}

// 30. ui_clear: return clear-screen sequence or prompt line
func lib30UiClear(mode string) string {
if strings.EqualFold(strings.TrimSpace(mode), "clear") {
return uiClear
}
return uiGreen + uiBold + "BMM-OS> " + uiReset + strconv.Itoa(30)
}
