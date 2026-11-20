package main

import (
"fmt"
"os"
"path/filepath"
"strings"
"unicode"
)

const (
AnsiReset  = "\033[0m"
AnsiBlue   = "\033[34m"
AnsiRed    = "\033[31m"
AnsiCyan   = "\033[96m"
AnsiYellow = "\033[33m"
)

func ProColorizeAlways(input string) string {
if strings.TrimSpace(input) == "" {
return AnsiBlue + input + AnsiReset
}

var out strings.Builder
lines := strings.Split(input, "\n")
for li, line := range lines {
if li > 0 {
out.WriteByte('\n')
}
trimmed := strings.TrimSpace(line)
isTableLine := strings.HasPrefix(trimmed, "|") ||
strings.HasPrefix(trimmed, "+-") ||
strings.HasPrefix(trimmed, "TABLE") ||
strings.Contains(line, "\t")

if isTableLine {
out.WriteString(AnsiYellow)
out.WriteString(line)
out.WriteString(AnsiReset)
continue
}

i := 0
n := len(line)
for i < n {
ch := line[i]

if unicode.IsDigit(rune(ch)) {
start := i
for i < n && (unicode.IsDigit(rune(line[i])) || line[i] == '.') {
i++
}
out.WriteString(AnsiRed)
out.WriteString(line[start:i])
out.WriteString(AnsiReset)
continue
}

if strings.HasPrefix(line[i:], "p3_") || strings.HasPrefix(line[i:], "p2_") {
start := i
for i < n && (unicode.IsLetter(rune(line[i])) || unicode.IsDigit(rune(line[i])) || line[i] == '_') {
i++
}
out.WriteString(AnsiCyan)
out.WriteString(line[start:i])
out.WriteString(AnsiReset)
continue
}

if strings.ContainsRune("@$#%^&*+-=<>?/\\|~", rune(ch)) {
start := i
for i < n && strings.ContainsRune("@$#%^&*+-=<>?/\\|~", rune(line[i])) {
i++
}
out.WriteString(AnsiCyan)
out.WriteString(line[start:i])
out.WriteString(AnsiReset)
continue
}

start := i
for i < n {
c := line[i]
if unicode.IsDigit(rune(c)) ||
strings.HasPrefix(line[i:], "p3_") ||
strings.HasPrefix(line[i:], "p2_") ||
strings.ContainsRune("@$#%^&*+-=<>?/\\|~", rune(c)) {
break
}
i++
}
out.WriteString(AnsiBlue)
out.WriteString(line[start:i])
out.WriteString(AnsiReset)
}
}
return out.String()
}

func ProBeamPrint(raw string) string {
rendered := ProColorizeAlways("[BEAM] " + raw)
fmt.Println(rendered)
return rendered
}

type ProDiagnostic struct {
Code string
Line int
Msg  string
Hint string
}

func (d ProDiagnostic) Format() string {
if d.Line > 0 {
return ProColorizeAlways(fmt.Sprintf("[%s] line %d: %s (hint: %s)", d.Code, d.Line, d.Msg, d.Hint))
}
return ProColorizeAlways(fmt.Sprintf("[%s] %s (hint: %s)", d.Code, d.Msg, d.Hint))
}

type ProTable struct {
Headers []string
Rows    [][]string
}

func NewProTable(headers ...string) *ProTable {
return &ProTable{Headers: headers}
}

func (t *ProTable) AddRow(cols ...string) {
t.Rows = append(t.Rows, cols)
}

func (t *ProTable) RenderYellow() string {
var b strings.Builder
b.WriteString(AnsiYellow)
if len(t.Headers) > 0 {
b.WriteString("| " + strings.Join(t.Headers, " | ") + " |\n")
b.WriteString("|" + strings.Repeat("---|", len(t.Headers)) + "\n")
}
for _, r := range t.Rows {
b.WriteString("| " + strings.Join(r, " | ") + " |\n")
}
b.WriteString(AnsiReset)
return b.String()
}

func RunProTestSuite(dir string) string {
bridge := NewPart4Bridge()
_, okP3 := bridge.Resolve("p3_color")
_, okP2 := bridge.Resolve("p2_futura")

matches, _ := filepath.Glob(filepath.Join(dir, "test_*.bmm"))
tbl := NewProTable("Check", "Status", "Detail")
tbl.AddRow("Part4Bridge p3_color", boolStatus(okP3), "symbol resolution")
tbl.AddRow("Part4Bridge p2_futura", boolStatus(okP2), "symbol resolution")
tbl.AddRow("Test Scripts", "FOUND", fmt.Sprintf("%d files", len(matches)))

_, statErr := os.Stat(filepath.Join(dir, "bmm_lang_part4_bridge.go"))
tbl.AddRow("Part4 Bridge File", boolStatus(statErr == nil), "bmm_lang_part4_bridge.go")
return tbl.RenderYellow()
}

func boolStatus(ok bool) string {
if ok {
return "PASS"
}
return "FAIL"
}
