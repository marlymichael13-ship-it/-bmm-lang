package main

import (
"fmt"
"strconv"
"strings"
"time"
)

const (
p3AnsiReset  = "\x1b[0m"
p3AnsiBold   = "\x1b[1m"
p3AnsiBlue   = "\x1b[94m"
p3AnsiRed    = "\x1b[91m"
p3AnsiGreen  = "\x1b[92m"
p3AnsiYellow = "\x1b[93m"
p3AnsiCyan   = "\x1b[96m"
p3AnsiMag    = "\x1b[95m"
p3AnsiWhite  = "\x1b[97m"
)

type p3TokKind int

const (
p3TokEOF p3TokKind = iota
p3TokInt
p3TokStr
p3TokIdent
p3TokPlus
p3TokMinus
p3TokStar
p3TokSlash
p3TokEq
p3TokEqEq
p3TokLt
p3TokGt
p3TokLParen
p3TokRParen
p3TokComma
p3TokSemi
)

type p3Token struct {
Kind p3TokKind
Val  string
Line int
Col  int
}

func p3Lex(src string) []p3Token {
var toks []p3Token
line, col := 1, 1
i := 0
n := len(src)

for i < n {
ch := src[i]
if ch == '\n' {
line++
col = 1
i++
continue
}
if ch == ' ' || ch == '\t' || ch == '\r' {
col++
i++
continue
}
if ch == '#' || (ch == '/' && i+1 < n && src[i+1] == '/') {
for i < n && src[i] != '\n' {
i++
col++
}
continue
}
if ch >= '0' && ch <= '9' {
startCol := col
start := i
for i < n && src[i] >= '0' && src[i] <= '9' {
i++
col++
}
toks = append(toks, p3Token{Kind: p3TokInt, Val: src[start:i], Line: line, Col: startCol})
continue
}
if ch == '"' || ch == '\'' {
quote := ch
startCol := col
i++
col++
var sb strings.Builder
for i < n && src[i] != quote {
if src[i] == '\\' && i+1 < n {
i++
col++
switch src[i] {
case 'n':
sb.WriteByte('\n')
case 't':
sb.WriteByte('\t')
default:
sb.WriteByte(src[i])
}
i++
col++
continue
}
if src[i] == '\n' {
line++
col = 1
} else {
col++
}
sb.WriteByte(src[i])
i++
}
if i < n && src[i] == quote {
i++
col++
}
toks = append(toks, p3Token{Kind: p3TokStr, Val: sb.String(), Line: line, Col: startCol})
continue
}
if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
startCol := col
start := i
for i < n {
c := src[i]
if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' {
i++
col++
} else {
break
}
}
toks = append(toks, p3Token{Kind: p3TokIdent, Val: src[start:i], Line: line, Col: startCol})
continue
}

startCol := col
switch ch {
case '+':
toks = append(toks, p3Token{Kind: p3TokPlus, Val: "+", Line: line, Col: startCol})
i++
col++
case '-':
toks = append(toks, p3Token{Kind: p3TokMinus, Val: "-", Line: line, Col: startCol})
i++
col++
case '*':
toks = append(toks, p3Token{Kind: p3TokStar, Val: "*", Line: line, Col: startCol})
i++
col++
case '/':
toks = append(toks, p3Token{Kind: p3TokSlash, Val: "/", Line: line, Col: startCol})
i++
col++
case '=':
if i+1 < n && src[i+1] == '=' {
toks = append(toks, p3Token{Kind: p3TokEqEq, Val: "==", Line: line, Col: startCol})
i += 2
col += 2
} else {
toks = append(toks, p3Token{Kind: p3TokEq, Val: "=", Line: line, Col: startCol})
i++
col++
}
case '<':
toks = append(toks, p3Token{Kind: p3TokLt, Val: "<", Line: line, Col: startCol})
i++
col++
case '>':
toks = append(toks, p3Token{Kind: p3TokGt, Val: ">", Line: line, Col: startCol})
i++
col++
case '(':
toks = append(toks, p3Token{Kind: p3TokLParen, Val: "(", Line: line, Col: startCol})
i++
col++
case ')':
toks = append(toks, p3Token{Kind: p3TokRParen, Val: ")", Line: line, Col: startCol})
i++
col++
case ',':
toks = append(toks, p3Token{Kind: p3TokComma, Val: ",", Line: line, Col: startCol})
i++
col++
case ';':
toks = append(toks, p3Token{Kind: p3TokSemi, Val: ";", Line: line, Col: startCol})
i++
col++
default:
toks = append(toks, p3Token{Kind: p3TokIdent, Val: string(ch), Line: line, Col: startCol})
i++
col++
}
}

toks = append(toks, p3Token{Kind: p3TokEOF, Val: "EOF", Line: line, Col: col})
return toks
}

type p3NodeKind int

const (
p3NodeProgram p3NodeKind = iota
p3NodeInt
p3NodeStr
p3NodeVar
p3NodeLet
p3NodeAssign
p3NodeBinary
p3NodeCall
p3NodeIf
p3NodeLoop
)

type p3Node struct {
Kind  p3NodeKind
Name  string
Op    string
Str   string
Num   int64
Left  *p3Node
Right *p3Node
Cond  *p3Node
Then  *p3Node
Else  *p3Node
Args  []*p3Node
Stmts []*p3Node
}

type p3Parser struct {
toks []p3Token
pos  int
errs []string
}

func (p *p3Parser) cur() p3Token {
if p.pos >= len(p.toks) {
return p3Token{Kind: p3TokEOF, Val: "EOF"}
}
return p.toks[p.pos]
}

func (p *p3Parser) eat() p3Token {
t := p.cur()
if p.pos < len(p.toks) {
p.pos++
}
return t
}

func p3Parse(src string) (*p3Node, []string) {
p := &p3Parser{toks: p3Lex(src)}
prog := &p3Node{Kind: p3NodeProgram}
for p.cur().Kind != p3TokEOF {
for p.cur().Kind == p3TokSemi {
p.eat()
}
if p.cur().Kind == p3TokEOF {
break
}
stmt := p.parseStmt()
if stmt != nil {
prog.Stmts = append(prog.Stmts, stmt)
}
for p.cur().Kind == p3TokSemi {
p.eat()
}
}
return prog, p.errs
}

func (p *p3Parser) parseStmt() *p3Node {
t := p.cur()
if t.Kind == p3TokIdent {
low := strings.ToLower(t.Val)
switch low {
case "let", "var", "set":
p.eat()
nameTok := p.eat()
if p.cur().Kind == p3TokEq {
p.eat()
}
val := p.parseExpr()
return &p3Node{Kind: p3NodeLet, Name: nameTok.Val, Right: val}
case "if":
p.eat()
cond := p.parseExpr()
if p.cur().Kind == p3TokIdent && strings.ToLower(p.cur().Val) == "then" {
p.eat()
}
thenN := p.parseStmt()
var elseN *p3Node
if p.cur().Kind == p3TokIdent && strings.ToLower(p.cur().Val) == "else" {
p.eat()
elseN = p.parseStmt()
}
return &p3Node{Kind: p3NodeIf, Cond: cond, Then: thenN, Else: elseN}
case "loop", "repeat":
p.eat()
cnt := p.parseExpr()
if p.cur().Kind == p3TokIdent && strings.ToLower(p.cur().Val) == "do" {
p.eat()
}
body := p.parseStmt()
return &p3Node{Kind: p3NodeLoop, Cond: cnt, Then: body}
}
if p.pos+1 < len(p.toks) && p.toks[p.pos+1].Kind == p3TokEq {
name := p.eat().Val
p.eat()
val := p.parseExpr()
return &p3Node{Kind: p3NodeAssign, Name: name, Right: val}
}
}
return p.parseExpr()
}

func (p *p3Parser) parseExpr() *p3Node {
left := p.parseTerm()
for p.cur().Kind == p3TokPlus || p.cur().Kind == p3TokMinus || p.cur().Kind == p3TokEqEq || p.cur().Kind == p3TokLt || p.cur().Kind == p3TokGt {
opTok := p.eat()
right := p.parseTerm()
left = &p3Node{Kind: p3NodeBinary, Op: opTok.Val, Left: left, Right: right}
}
return left
}

func (p *p3Parser) parseTerm() *p3Node {
left := p.parseFactor()
for p.cur().Kind == p3TokStar || p.cur().Kind == p3TokSlash {
opTok := p.eat()
right := p.parseFactor()
left = &p3Node{Kind: p3NodeBinary, Op: opTok.Val, Left: left, Right: right}
}
return left
}

func (p *p3Parser) parseFactor() *p3Node {
t := p.cur()
switch t.Kind {
case p3TokInt:
p.eat()
v, _ := strconv.ParseInt(t.Val, 10, 64)
return &p3Node{Kind: p3NodeInt, Num: v, Str: t.Val}
case p3TokStr:
p.eat()
return &p3Node{Kind: p3NodeStr, Str: t.Val}
case p3TokLParen:
p.eat()
inner := p.parseExpr()
if p.cur().Kind == p3TokRParen {
p.eat()
}
return inner
case p3TokIdent:
name := p.eat().Val
if p.cur().Kind == p3TokLParen {
p.eat()
var args []*p3Node
for p.cur().Kind != p3TokRParen && p.cur().Kind != p3TokEOF {
args = append(args, p.parseExpr())
if p.cur().Kind == p3TokComma {
p.eat()
} else {
break
}
}
if p.cur().Kind == p3TokRParen {
p.eat()
}
return &p3Node{Kind: p3NodeCall, Name: name, Args: args}
}
return &p3Node{Kind: p3NodeVar, Name: name, Str: name}
default:
p.eat()
return &p3Node{Kind: p3NodeStr, Str: t.Val}
}
}

type p3Val struct {
Num int64
Str string
}

func p3ValString(v p3Val) string {
if v.Str != "" {
return v.Str
}
return strconv.FormatInt(v.Num, 10)
}

func p3Truthy(v p3Val) bool {
if v.Num != 0 {
return true
}
return strings.TrimSpace(v.Str) != "" && strings.ToLower(strings.TrimSpace(v.Str)) != "0" && strings.ToLower(strings.TrimSpace(v.Str)) != "false"
}

type p3Interp struct {
env map[string]p3Val
out []string
}

func p3Eval(src string) string {
ast, errs := p3Parse(src)
if len(errs) > 0 {
return "P3_PARSE_ERR: " + strings.Join(errs, "; ")
}
ip := &p3Interp{env: make(map[string]p3Val)}
var last p3Val
for _, st := range ast.Stmts {
last = ip.evalNode(st)
}
if len(ip.out) > 0 {
return strings.Join(ip.out, " || ")
}
return p3ValString(last)
}

func (ip *p3Interp) evalNode(n *p3Node) p3Val {
if n == nil {
return p3Val{}
}
switch n.Kind {
case p3NodeInt:
return p3Val{Num: n.Num, Str: strconv.FormatInt(n.Num, 10)}
case p3NodeStr:
return p3Val{Str: n.Str}
case p3NodeVar:
if v, ok := ip.env[n.Name]; ok {
return v
}
return p3Val{Str: n.Name}
case p3NodeLet, p3NodeAssign:
v := ip.evalNode(n.Right)
ip.env[n.Name] = v
return v
case p3NodeIf:
if p3Truthy(ip.evalNode(n.Cond)) {
return ip.evalNode(n.Then)
}
if n.Else != nil {
return ip.evalNode(n.Else)
}
return p3Val{}
case p3NodeLoop:
cntVal := ip.evalNode(n.Cond)
steps := cntVal.Num
if steps < 0 {
steps = 0
}
if steps > 64 {
steps = 64
}
var last p3Val
for i := int64(0); i < steps; i++ {
ip.env["idx"] = p3Val{Num: i, Str: strconv.FormatInt(i, 10)}
last = ip.evalNode(n.Then)
}
return last
case p3NodeBinary:
l := ip.evalNode(n.Left)
r := ip.evalNode(n.Right)
switch n.Op {
case "+":
if l.Str != "" && r.Str != "" && strconv.FormatInt(l.Num, 10) != l.Str {
s := l.Str + r.Str
return p3Val{Str: s}
}
sum := l.Num + r.Num
return p3Val{Num: sum, Str: strconv.FormatInt(sum, 10)}
case "-":
diff := l.Num - r.Num
return p3Val{Num: diff, Str: strconv.FormatInt(diff, 10)}
case "*":
prod := l.Num * r.Num
return p3Val{Num: prod, Str: strconv.FormatInt(prod, 10)}
case "/":
if r.Num == 0 {
return p3Val{Str: "ERR_DIV_ZERO"}
}
q := l.Num / r.Num
return p3Val{Num: q, Str: strconv.FormatInt(q, 10)}
case "==":
eq := l.Num == r.Num && p3ValString(l) == p3ValString(r)
if eq {
return p3Val{Num: 1, Str: "1"}
}
return p3Val{Num: 0, Str: "0"}
case "<":
if l.Num < r.Num {
return p3Val{Num: 1, Str: "1"}
}
return p3Val{Num: 0, Str: "0"}
case ">":
if l.Num > r.Num {
return p3Val{Num: 1, Str: "1"}
}
return p3Val{Num: 0, Str: "0"}
}
case p3NodeCall:
var evalArgs []string
for _, a := range n.Args {
evalArgs = append(evalArgs, p3ValString(ip.evalNode(a)))
}
joined := strings.Join(evalArgs, " ")
switch strings.ToLower(n.Name) {
case "color", "dyn_color", "paint":
res := p3ColorizeDynamic(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "lexicon", "lex":
res := p3LexiconScan(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "futura", "p2_futura":
res := bmmPart2Translate(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "alias", "p2_alias":
res := bmmPart2Alias(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "mesh", "nexus_mesh":
res := nexusMeshMap(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "sys", "nexus_sys":
res := nexusSysLink(joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
case "emit", "print", "say":
ip.out = append(ip.out, joined)
return p3Val{Str: joined}
default:
res := fmt.Sprintf("P3_CALL[%s(%s)]", n.Name, joined)
ip.out = append(ip.out, res)
return p3Val{Str: res}
}
}
return p3Val{}
}

type p3Instr struct {
Op  string
Arg string
}

type p3Compiler struct {
code []p3Instr
}

func p3Compile(src string) string {
ast, errs := p3Parse(src)
if len(errs) > 0 {
return "P3_COMPILE_ERR: " + strings.Join(errs, "; ")
}
c := &p3Compiler{}
for _, st := range ast.Stmts {
c.emitNode(st)
}
c.code = append(c.code, p3Instr{Op: "HALT", Arg: "0"})
var lines []string
for idx, ins := range c.code {
lines = append(lines, fmt.Sprintf("%03d: %-14s %s", idx, ins.Op, ins.Arg))
}
return fmt.Sprintf("P3_BYTECODE[%d_OPS] => %s", len(c.code), strings.Join(lines, " | "))
}

func (c *p3Compiler) emitNode(n *p3Node) {
if n == nil {
return
}
switch n.Kind {
case p3NodeInt:
c.code = append(c.code, p3Instr{Op: "PUSH_INT", Arg: strconv.FormatInt(n.Num, 10)})
case p3NodeStr:
c.code = append(c.code, p3Instr{Op: "PUSH_STR", Arg: fmt.Sprintf("%q", n.Str)})
case p3NodeVar:
c.code = append(c.code, p3Instr{Op: "LOAD_VAR", Arg: n.Name})
case p3NodeLet, p3NodeAssign:
c.emitNode(n.Right)
c.code = append(c.code, p3Instr{Op: "STORE_VAR", Arg: n.Name})
case p3NodeBinary:
c.emitNode(n.Left)
c.emitNode(n.Right)
c.code = append(c.code, p3Instr{Op: "BIN_OP", Arg: n.Op})
case p3NodeCall:
for _, a := range n.Args {
c.emitNode(a)
}
c.code = append(c.code, p3Instr{Op: "CALL_NATIVE", Arg: fmt.Sprintf("%s/%d", n.Name, len(n.Args))})
case p3NodeIf:
c.emitNode(n.Cond)
c.code = append(c.code, p3Instr{Op: "JMP_UNLESS", Arg: "ELSE_OR_END"})
c.emitNode(n.Then)
if n.Else != nil {
c.code = append(c.code, p3Instr{Op: "JMP", Arg: "END"})
c.emitNode(n.Else)
}
c.code = append(c.code, p3Instr{Op: "LABEL", Arg: "END"})
case p3NodeLoop:
c.emitNode(n.Cond)
c.code = append(c.code, p3Instr{Op: "LOOP_BEGIN", Arg: "BODY"})
c.emitNode(n.Then)
c.code = append(c.code, p3Instr{Op: "LOOP_END", Arg: "NEXT"})
}
}

var p3LexiconTable = map[string]string{
"cyber":    "CYBER_DOMAIN",
"bmm":      "BMM_CORE",
"pp":       "PULSE_PAIR_PP",
"mm":       "MEM_PAIR_MM",
"p":        "PULSE_UNIT_P",
"m":        "MEM_UNIT_M",
"quantum":  "QUANTUM_STATE",
"neuro":    "NEURO_SPIKE",
"matrix":   "MATRIX_GRID",
"shield":   "SHIELD_GUARD",
"flux":     "FLUX_STREAM",
"warp":     "WARP_SHIFT",
"nexus":    "NEXUS_HUB",
"core":     "CORE_ENGINE",
"pulse":    "PULSE_SIGNAL",
"echo":     "ECHO_RESONANCE",
"cipher":   "CIPHER_LOCK",
"beacon":   "BEACON_PULSE",
"forge":    "FORGE_BUILD",
"drift":    "DRIFT_VECTOR",
"syn":      "SYNC_LINK",
"vox":      "VOX_CHANNEL",
"grid":     "GRID_TOPOLOGY",
"aether":   "AETHER_RUNTIME",
"futura":   "FUTURA_MODE",
"hyper":    "HYPER_SCALE",
"omni":     "OMNI_BRIDGE",
"nano":     "NANO_KERNEL",
"tensor":   "TENSOR_FLOW",
"vector":   "VECTOR_SPACE",
"photon":   "PHOTON_BUS",
"volt":     "VOLT_CHARGE",
"helix":    "HELIX_CHAIN",
"zenith":   "ZENITH_PEAK",
"vortex":   "VORTEX_LOOP",
"prism":    "PRISM_SPLIT",
"titan":    "TITAN_NODE",
"nova":     "NOVA_BURST",
"pulse2":   "PULSE_DUAL",
"mem2":     "MEM_DUAL",
"sec":      "SECURITY_LAYER",
"opt":      "OPTIMIZER_PASS",
"trace":    "TRACE_LOG",
"link":     "LINK_BIND",
"mesh":     "MESH_MAP",
"stream":   "STREAM_FLOW",
"node":     "NODE_UNIT",
"spark":    "SPARK_TRIGGER",
"horizon":  "HORIZON_LINE",
"chronos":  "CHRONOS_TIME",
"ether":    "ETHER_BUS",
"ion":      "ION_CARRIER",
"plasma":   "PLASMA_CORE",
"orbit":    "ORBIT_PATH",
"radar":    "RADAR_SCAN",
"sonar":    "SONAR_PING",
"hydra":    "HYDRA_MULTI",
"atlas":    "ATLAS_INDEX",
"omega":    "OMEGA_FINAL",
"alpha":    "ALPHA_INIT",
"sigma":    "SIGMA_SUM",
"delta":    "DELTA_CHANGE",
"lambda":   "LAMBDA_EXPR",
"kinetix":  "KINETIX_MOTION",
"cybernet": "CYBERNET_LINK",
"bmmlang":  "BMMLANG_SPEC",
"bmmcore":  "BMMCORE_BASE",
"ppx":      "PP_EXT",
"mmx":      "MM_EXT",
"ppxcore":  "PP_CORE",
"mmxcore":  "MM_CORE",
}

func p3LexiconScan(input string) string {
words := strings.Fields(strings.ToLower(input))
var hits []string
for _, w := range words {
clean := strings.Trim(w, ",.;:!?()[]{}\"'")
if tag, ok := p3LexiconTable[clean]; ok {
hits = append(hits, clean+"->"+tag)
continue
}
if strings.Contains(clean, "cyber") {
hits = append(hits, clean+"->CYBER_SUB")
} else if strings.Contains(clean, "bmm") {
hits = append(hits, clean+"->BMM_SUB")
} else if strings.Contains(clean, "pp") {
hits = append(hits, clean+"->PP_SUB")
} else if strings.Contains(clean, "mm") {
hits = append(hits, clean+"->MM_SUB")
}
}
if len(hits) == 0 {
hits = append(hits, "GENERAL_TOKEN")
}
return fmt.Sprintf("P3_LEXICON[%d/%d_MATCHES] => %s", len(hits), len(words), strings.Join(hits, " | "))
}

// p3ColorizeDynamic applies real ANSI terminal colors dynamically to cyber, bmm, mm, pp and keywords
func p3ColorizeDynamic(text string) string {
words := strings.Split(text, " ")
var out []string
for _, w := range words {
low := strings.ToLower(w)
colored := w
switch {
case strings.Contains(low, "cyber"):
colored = p3AnsiBold + p3AnsiBlue + w + p3AnsiReset
case strings.Contains(low, "bmm"):
colored = p3AnsiBold + p3AnsiRed + w + p3AnsiReset
case strings.Contains(low, "mm"):
colored = p3AnsiBold + p3AnsiGreen + w + p3AnsiReset
case strings.Contains(low, "pp"):
colored = p3AnsiBold + p3AnsiYellow + w + p3AnsiReset
case strings.Contains(low, "quantum") || strings.Contains(low, "neuro") || strings.Contains(low, "nexus"):
colored = p3AnsiCyan + w + p3AnsiReset
case strings.Contains(low, "flux") || strings.Contains(low, "warp") || strings.Contains(low, "matrix"):
colored = p3AnsiMag + w + p3AnsiReset
}
out = append(out, colored)
}
return fmt.Sprintf("%s[DYN_COLOR_%s]%s %s",
p3AnsiWhite, time.Now().UTC().Format("150405"), p3AnsiReset, strings.Join(out, " "))
}
