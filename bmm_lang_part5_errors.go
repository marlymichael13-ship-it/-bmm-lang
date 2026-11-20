package main

import (
"fmt"
"strings"
"unicode/utf8"
)

// Part5ErrorRecord holds a rich diagnostic snapshot of an error inside BMM code.
type Part5ErrorRecord struct {
Code       string
Category   string
SourceCode string
Line       int
Column     int
BytePos    int
RuneChar   string
Summary    string
WhyHappens string
HowToFix   string
}

var part5LastError *Part5ErrorRecord

// Part5RecordError captures an error with exact line, column, rune, and educational guidance.
func Part5RecordError(code, category, source string, line, col int, summary, whyHappens, howToFix string) *Part5ErrorRecord {
if line < 1 {
line = 1
}
if col < 1 {
col = 1
}

lines := strings.Split(source, "\n")
targetLine := ""
if line-1 < len(lines) {
targetLine = lines[line-1]
}

bytePos, runeChar := part5LocateRuneAtColumn(targetLine, col)

rec := &Part5ErrorRecord{
Code:       code,
Category:   category,
SourceCode: source,
Line:       line,
Column:     col,
BytePos:    bytePos,
RuneChar:   runeChar,
Summary:    summary,
WhyHappens: whyHappens,
HowToFix:   howToFix,
}
part5LastError = rec
return rec
}

func part5LocateRuneAtColumn(lineStr string, col int) (int, string) {
if lineStr == "" {
return 0, "<EOL>"
}
idx := 0
currentCol := 1
for idx < len(lineStr) {
r, size := utf8.DecodeRuneInString(lineStr[idx:])
if currentCol == col {
return idx, string(r)
}
idx += size
currentCol++
}
lastR, _ := utf8.DecodeLastRuneInString(lineStr)
if lastR == utf8.RuneError {
return len(lineStr), "<END>"
}
return len(lineStr) - utf8.RuneLen(lastR), string(lastR)
}

// Part5FormatBrief prints a concise notice without forcing full teaching mode unless requested.
func (e *Part5ErrorRecord) Part5FormatBrief() string {
if e == nil {
return ProColorizeAlways("No error recorded.")
}
msg := fmt.Sprintf("[%s] line %d, col %d: %s (run: beam teaching the error)", e.Code, e.Line, e.Column, e.Summary)
return ProColorizeAlways(msg)
}

// Part5ExplainDetailed returns a deep, step-by-step explanation of where, why, and how to fix the error.
func (e *Part5ErrorRecord) Part5ExplainDetailed() string {
if e == nil {
return ProColorizeAlways("No error is currently stored. Nothing to teach.")
}

lines := strings.Split(e.SourceCode, "\n")
lineContent := ""
if e.Line-1 >= 0 && e.Line-1 < len(lines) {
lineContent = lines[e.Line-1]
}

caretPad := strings.Repeat(" ", e.Column-1)
pointerLine := caretPad + "^"

var b strings.Builder
b.WriteString(ProColorizeAlways("=== BMM PART 5: DETAILED ERROR TEACHING ===\n"))
b.WriteString(ProColorizeAlways(fmt.Sprintf("Error Code : %s | Category: %s\n", e.Code, e.Category)))
b.WriteString(ProColorizeAlways(fmt.Sprintf("Exact Spot : Line %d, Column %d, Byte %d, Character/Rune: %q\n", e.Line, e.Column, e.BytePos, e.RuneChar)))
b.WriteString("\n")
b.WriteString(ProColorizeAlways("--- 1. EXACT LOCATION IN CODE ---\n"))
b.WriteString(ProColorizeAlways(fmt.Sprintf("Line %d: %s\n", e.Line, lineContent)))
b.WriteString(AnsiCyan + fmt.Sprintf("        %s <-- exact spot (col %d, char %q)\n", pointerLine, e.Column, e.RuneChar) + AnsiReset)
b.WriteString("\n")
b.WriteString(ProColorizeAlways("--- 2. HOW & WHY THIS ERROR HAPPENED ---\n"))
b.WriteString(ProColorizeAlways(e.WhyHappens + "\n"))
b.WriteString("\n")
b.WriteString(ProColorizeAlways("--- 3. HOW TO FIX IT STEP BY STEP ---\n"))
b.WriteString(ProColorizeAlways(e.HowToFix + "\n"))
b.WriteString(ProColorizeAlways("============================================"))
return b.String()
}

// Part5HandleCommand checks if the user invoked the explicit teaching trigger:
// "beam teaching the error"
func Part5HandleCommand(input string) (string, bool) {
normalized := strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(input))), " ")
if normalized == "beam teaching the error" {
if part5LastError == nil {
return ProColorizeAlways("[PART5] No error has been recorded yet. Trigger or record an error first, then run: beam teaching the error"), true
}
return part5LastError.Part5ExplainDetailed(), true
}
return "", false
}

// Part5DemoScenario records a sample error and demonstrates "beam teaching the error".
func Part5DemoScenario() string {
sampleCode := "x = 10 + @bad\nbeam(x)"
Part5RecordError(
"E501_UNEXPECTED_SYMBOL",
"Lexical / Symbol Resolution",
sampleCode,
1,
12,
"Unexpected symbol '@' in arithmetic expression",
"At line 1, column 12, the parser encountered '@'. In this expression context, '@' is not a valid operand or arithmetic operator, so evaluation stopped right at that character.",
"1) Go to line 1, column 12 where '@' sits.\n2) Replace '@bad' with a valid number, variable, or registered symbol.\n3) Re-run your BMM script and verify with beam(...).",
)
out, _ := Part5HandleCommand("beam teaching the error")
return out
}
