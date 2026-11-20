package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const (
	AetherName      = "AETHER-BMM"
	AetherRelease   = "13.0.0-AETHER-COMPILED-GENESIS"
	AetherSubSystem = "HYBRID-BYTECODE-VM-AND-KERNEL"
	MaxCallFrames   = 1024
	MaxStepQuota    = int64(2000000)
	MaxSeqSpan      = 250000
	MaxConstPool    = 65536
)

type DatumKind uint8

const (
	DatumVoid DatumKind = iota
	DatumCycle
	DatumGlyph
	DatumFlag
	DatumVec
	DatumDict
	DatumPulse
	DatumUnit
	DatumCodeRef
)

type Datum struct {
	Kind     DatumKind
	Cycle    int64
	Glyph    string
	Flag     bool
	Vec      []Datum
	Dict     map[string]Datum
	Band     string
	FreqHz   int64
	UnitTag  string
	UnitMeta map[string]Datum
	FnAddr   int
}

func VoidDatum() Datum {
	return Datum{Kind: DatumVoid}
}

func CycleDatum(n int64) Datum {
	return Datum{Kind: DatumCycle, Cycle: n}
}

func GlyphDatum(s string) Datum {
	return Datum{Kind: DatumGlyph, Glyph: s}
}

func FlagDatum(b bool) Datum {
	return Datum{Kind: DatumFlag, Flag: b}
}

func VecDatum(items []Datum) Datum {
	out := make([]Datum, len(items))
	for i, it := range items {
		out[i] = it.DeepCopy()
	}
	return Datum{Kind: DatumVec, Vec: out}
}

func DictDatum(entries map[string]Datum) Datum {
	out := make(map[string]Datum, len(entries))
	for k, v := range entries {
		out[k] = v.DeepCopy()
	}
	return Datum{Kind: DatumDict, Dict: out}
}

func PulseDatum(band string, hz int64) Datum {
	return Datum{Kind: DatumPulse, Band: band, FreqHz: hz}
}

func UnitDatum(tag string, meta map[string]Datum) Datum {
	cp := make(map[string]Datum, len(meta))
	for k, v := range meta {
		cp[k] = v.DeepCopy()
	}
	return Datum{Kind: DatumUnit, UnitTag: tag, UnitMeta: cp}
}

func CodeRefDatum(addr int) Datum {
	return Datum{Kind: DatumCodeRef, FnAddr: addr}
}

func (d Datum) DeepCopy() Datum {
	switch d.Kind {
	case DatumVec:
		cp := make([]Datum, len(d.Vec))
		for i, item := range d.Vec {
			cp[i] = item.DeepCopy()
		}
		return Datum{Kind: DatumVec, Vec: cp}
	case DatumDict:
		cp := make(map[string]Datum, len(d.Dict))
		for k, v := range d.Dict {
			cp[k] = v.DeepCopy()
		}
		return Datum{Kind: DatumDict, Dict: cp}
	case DatumUnit:
		cp := make(map[string]Datum, len(d.UnitMeta))
		for k, v := range d.UnitMeta {
			cp[k] = v.DeepCopy()
		}
		return Datum{Kind: DatumUnit, UnitTag: d.UnitTag, UnitMeta: cp}
	default:
		return d
	}
}

func (d Datum) TypeName() string {
	switch d.Kind {
	case DatumCycle:
		return "cycle"
	case DatumGlyph:
		return "glyph"
	case DatumFlag:
		return "flag"
	case DatumVec:
		return "vec"
	case DatumDict:
		return "dict"
	case DatumPulse:
		return "pulse"
	case DatumUnit:
		return "unit"
	case DatumCodeRef:
		return "coderef"
	default:
		return "void"
	}
}

func (d Datum) Truthy() bool {
	switch d.Kind {
	case DatumFlag:
		return d.Flag
	case DatumCycle:
		return d.Cycle != 0
	case DatumGlyph:
		return d.Glyph != ""
	case DatumVec:
		return len(d.Vec) > 0
	case DatumDict:
		return len(d.Dict) > 0
	case DatumPulse:
		return d.FreqHz > 0
	case DatumUnit:
		return d.UnitTag != ""
	case DatumCodeRef:
		return d.FnAddr >= 0
	default:
		return false
	}
}

func (d Datum) Stringify() string {
	switch d.Kind {
	case DatumCycle:
		return strconv.FormatInt(d.Cycle, 10)
	case DatumGlyph:
		return d.Glyph
	case DatumFlag:
		if d.Flag {
			return "TRUE"
		}
		return "FALSE"
	case DatumVec:
		parts := make([]string, len(d.Vec))
		for i, item := range d.Vec {
			if item.Kind == DatumGlyph {
				parts[i] = `"` + escapeGlyphText(item.Glyph) + `"`
			} else {
				parts[i] = item.Stringify()
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case DatumDict:
		parts := make([]string, 0, len(d.Dict))
		for k, v := range d.Dict {
			if v.Kind == DatumGlyph {
				parts = append(parts, fmt.Sprintf("%s: \"%s\"", k, escapeGlyphText(v.Glyph)))
			} else {
				parts = append(parts, fmt.Sprintf("%s: %s", k, v.Stringify()))
			}
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case DatumPulse:
		return fmt.Sprintf("<pulse:%s@%dHz>", d.Band, d.FreqHz)
	case DatumUnit:
		return fmt.Sprintf("<unit:%s>", d.UnitTag)
	case DatumCodeRef:
		return fmt.Sprintf("<coderef@%d>", d.FnAddr)
	default:
		return "VOID"
	}
}

func (d Datum) Equals(other Datum) bool {
	if d.Kind != other.Kind {
		return false
	}
	switch d.Kind {
	case DatumVoid:
		return true
	case DatumCycle:
		return d.Cycle == other.Cycle
	case DatumGlyph:
		return d.Glyph == other.Glyph
	case DatumFlag:
		return d.Flag == other.Flag
	case DatumPulse:
		return d.Band == other.Band && d.FreqHz == other.FreqHz
	case DatumCodeRef:
		return d.FnAddr == other.FnAddr
	case DatumUnit:
		if d.UnitTag != other.UnitTag || len(d.UnitMeta) != len(other.UnitMeta) {
			return false
		}
		for k, v := range d.UnitMeta {
			ov, ok := other.UnitMeta[k]
			if !ok || !v.Equals(ov) {
				return false
			}
		}
		return true
	case DatumVec:
		if len(d.Vec) != len(other.Vec) {
			return false
		}
		for i := range d.Vec {
			if !d.Vec[i].Equals(other.Vec[i]) {
				return false
			}
		}
		return true
	case DatumDict:
		if len(d.Dict) != len(other.Dict) {
			return false
		}
		for k, v := range d.Dict {
			ov, ok := other.Dict[k]
			if !ok || !v.Equals(ov) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func escapeGlyphText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

type OpCode uint8

const (
	OpNop OpCode = iota
	OpLoadConst
	OpLoadVoid
	OpLoadVar
	OpStoreVar
	OpStoreLocal
	OpMakeVec
	OpMakeDict
	OpIndexGet
	OpIndexSet
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpNot
	OpEq
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte
	OpAnd
	OpOr
	OpJump
	OpJumpIfFalse
	OpJumpIfTrue
	OpCall
	OpReturn
	OpPop
	OpAssert
	OpRaise
	OpDefineUnit
	OpNewUnit
	OpGetField
	OpSetField
)

func (op OpCode) Mnemonic() string {
	switch op {
	case OpNop:
		return "NOP"
	case OpLoadConst:
		return "LOAD_CONST"
	case OpLoadVoid:
		return "LOAD_VOID"
	case OpLoadVar:
		return "LOAD_VAR"
	case OpStoreVar:
		return "STORE_VAR"
	case OpStoreLocal:
		return "STORE_LOCAL"
	case OpMakeVec:
		return "MAKE_VEC"
	case OpMakeDict:
		return "MAKE_DICT"
	case OpIndexGet:
		return "INDEX_GET"
	case OpIndexSet:
		return "INDEX_SET"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpNeg:
		return "NEG"
	case OpNot:
		return "NOT"
	case OpEq:
		return "EQ"
	case OpNeq:
		return "NEQ"
	case OpLt:
		return "LT"
	case OpLte:
		return "LTE"
	case OpGt:
		return "GT"
	case OpGte:
		return "GTE"
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpJump:
		return "JMP"
	case OpJumpIfFalse:
		return "JMP_FALSE"
	case OpJumpIfTrue:
		return "JMP_TRUE"
	case OpCall:
		return "CALL"
	case OpReturn:
		return "RET"
	case OpPop:
		return "POP"
	case OpAssert:
		return "ASSERT"
	case OpRaise:
		return "RAISE"
	case OpDefineUnit:
		return "DEF_UNIT"
	case OpNewUnit:
		return "NEW_UNIT"
	case OpGetField:
		return "GET_FIELD"
	case OpSetField:
		return "SET_FIELD"
	default:
		return "UNK"
	}
}

type Instruction struct {
	Op   OpCode
	ArgA int
	ArgB int
	Line int
}

type CompiledChunk struct {
	Code      []Instruction
	Constants []Datum
	Symbols   []string
	EntryPC   int
}

func NewCompiledChunk() *CompiledChunk {
	return &CompiledChunk{
		Code:      make([]Instruction, 0, 128),
		Constants: make([]Datum, 0, 64),
		Symbols:   make([]string, 0, 64),
		EntryPC:   0,
	}
}

func (ch *CompiledChunk) AddConst(val Datum, line int) (int, error) {
	if len(ch.Constants) >= MaxConstPool {
		return -1, fmt.Errorf("AETHER-COMPILE-FAULT [ERR_CONST_OVERFLOW @ line %d]: constant pool capacity exceeded", line)
	}
	for idx, existing := range ch.Constants {
		if existing.Equals(val) {
			return idx, nil
		}
	}
	ch.Constants = append(ch.Constants, val.DeepCopy())
	return len(ch.Constants) - 1, nil
}

func (ch *CompiledChunk) AddSymbol(name string) int {
	for idx, sym := range ch.Symbols {
		if sym == name {
			return idx
		}
	}
	ch.Symbols = append(ch.Symbols, name)
	return len(ch.Symbols) - 1
}

func (ch *CompiledChunk) Emit(op OpCode, a, b, line int) int {
	pc := len(ch.Code)
	ch.Code = append(ch.Code, Instruction{
		Op:   op,
		ArgA: a,
		ArgB: b,
		Line: line,
	})
	return pc
}

func (ch *CompiledChunk) PatchJump(pc int, target int) {
	if pc >= 0 && pc < len(ch.Code) {
		ch.Code[pc].ArgA = target
	}
}

func (ch *CompiledChunk) Disassemble() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== AETHER BYTECODE DISASSEMBLY (Consts: %d | Syms: %d | Ops: %d) ===\n",
		len(ch.Constants), len(ch.Symbols), len(ch.Code)))
	for i, ins := range ch.Code {
		b.WriteString(fmt.Sprintf("%04d [L%-3d] %-12s A=%-4d B=%-4d\n",
			i, ins.Line, ins.Op.Mnemonic(), ins.ArgA, ins.ArgB))
	}
	b.WriteString("========================================================================")
	return b.String()
}

type UnitSchema struct {
	Name   string
	Fields []string
}

type RoutineProto struct {
	Name   string
	Params []string
	Entry  int
	Lines  []CodeSpan
}

type CodeSpan struct {
	LineNo int
	Indent int
	Raw    string
	Stmt   string
}

type ScopeFrame struct {
	Slots map[string]Datum
	Label string
}

func NewScopeFrame(label string) *ScopeFrame {
	return &ScopeFrame{
		Slots: make(map[string]Datum),
		Label: label,
	}
}

type AetherDiag struct {
	Line   int
	Stage  string
	Code   string
	Detail string
	Hint   string
	Chain  []string
}

func (ad AetherDiag) Error() string {
	chainTxt := ""
	if len(ad.Chain) > 0 {
		chainTxt = " (" + strings.Join(ad.Chain, " -> ") + ")"
	}
	if ad.Hint != "" {
		return fmt.Sprintf("AETHER-%s [%s @ line %d]: %s | Hint: %s%s", ad.Stage, ad.Code, ad.Line, ad.Detail, ad.Hint, chainTxt)
	}
	return fmt.Sprintf("AETHER-%s [%s @ line %d]: %s%s", ad.Stage, ad.Code, ad.Line, ad.Detail, chainTxt)
}

func unescapeGlyphSeq(raw string) (string, error) {
	var b strings.Builder
	b.Grow(len(raw))
	esc := false
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if !esc {
			if ch == '\\' {
				esc = true
				continue
			}
			b.WriteByte(ch)
			continue
		}
		switch ch {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case '\\':
			b.WriteByte('\\')
		case '"':
			b.WriteByte('"')
		case '\'':
			b.WriteByte('\'')
		case '0':
			b.WriteByte(0)
		default:
			b.WriteByte('\\')
			b.WriteByte(ch)
		}
		esc = false
	}
	if esc {
		return "", fmt.Errorf("incomplete escape sequence in glyph literal")
	}
	return b.String(), nil
}

func isEscapedChar(s string, idx int) bool {
	backslashes := 0
	for k := idx - 1; k >= 0 && s[k] == '\\'; k-- {
		backslashes++
	}
	return backslashes%2 == 1
}

func pairCloseByte(open byte) byte {
	switch open {
	case '(':
		return ')'
	case '[':
		return ']'
	case '{':
		return '}'
	default:
		return 0
	}
}

func findPairClose(s string, openIdx int, openCh, closeCh byte) int {
	depth := 0
	quote := byte(0)
	for i := openIdx; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case openCh:
			depth++
		case closeCh:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func findPairOpen(s string, closeIdx int, openCh, closeCh byte) int {
	depth := 0
	quote := byte(0)
	for i := closeIdx; i >= 0; i-- {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case closeCh:
			depth++
		case openCh:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitTopLevelArgs(input string, sep byte) []string {
	var parts []string
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)
	start := 0

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(input, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			paren++
		case '[':
			bracket++
		case '{':
			brace++
		case ')':
			if paren > 0 {
				paren--
			}
		case ']':
			if bracket > 0 {
				bracket--
			}
		case '}':
			if brace > 0 {
				brace--
			}
		default:
			if ch == sep && paren == 0 && bracket == 0 && brace == 0 {
				parts = append(parts, input[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, input[start:])
	return parts
}

func prevNonBlankByte(s string, idx int) byte {
	for i := idx; i >= 0; i-- {
		c := s[i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return c
		}
	}
	return 0
}

func precedesUnary(b byte) bool {
	if b == 0 {
		return true
	}
	switch b {
	case '(', '[', '{', ',', ':', '=', '!', '<', '>', '+', '-', '*', '/', '%', '&', '|', '^':
		return true
	default:
		return false
	}
}

func findBinarySplit(expr string, ops []string) (int, string) {
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)

	for i := len(expr) - 1; i >= 0; i-- {
		ch := expr[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(expr, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case ')':
			paren++
		case ']':
			bracket++
		case '}':
			brace++
		case '(':
			paren--
		case '[':
			bracket--
		case '{':
			brace--
		default:
			if paren == 0 && bracket == 0 && brace == 0 {
				for _, op := range ops {
					l := len(op)
					if i+1 >= l && expr[i+1-l:i+1] == op {
						start := i + 1 - l
						if op == "-" || op == "+" {
							prev := prevNonBlankByte(expr, start-1)
							if precedesUnary(prev) {
								continue
							}
						}
						if op == "=" || op == "!" || op == "<" || op == ">" || op == "&" || op == "|" {
							if start+l < len(expr) && expr[start+l] == '=' {
								continue
							}
						}
						return start, op
					}
				}
			}
		}
	}
	return -1, ""
}

func IsIdentToken(tok string) bool {
	if tok == "" {
		return false
	}
	for idx, r := range tok {
		if idx == 0 && !(unicode.IsLetter(r) || r == '_') {
			return false
		}
		if idx > 0 && !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return false
		}
	}
	return true
}

func (k *AetherKernel) CompileExprToChunk(expr string, chunk *CompiledChunk, line int) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return k.fault(line, "ERR_COMPILE_EMPTY", "empty expression in compiler pass", "provide a valid expression")
	}

	if len(expr) >= 2 {
		openCh := expr[0]
		closeCh := expr[len(expr)-1]
		if (openCh == '(' && closeCh == ')') || (openCh == '[' && closeCh == ']') || (openCh == '{' && closeCh == '}') {
			closeIdx := findPairClose(expr, 0, openCh, pairCloseByte(openCh))
			if closeIdx == len(expr)-1 {
				inner := strings.TrimSpace(expr[1 : len(expr)-1])
				switch openCh {
				case '(':
					if inner == "" {
						chunk.Emit(OpLoadVoid, 0, 0, line)
						return nil
					}
					return k.CompileExprToChunk(inner, chunk, line)
				case '[':
					if inner == "" {
						chunk.Emit(OpMakeVec, 0, 0, line)
						return nil
					}
					elems := splitTopLevelArgs(inner, ',')
					for _, el := range elems {
						if err := k.CompileExprToChunk(el, chunk, line); err != nil {
							return err
						}
					}
					chunk.Emit(OpMakeVec, len(elems), 0, line)
					return nil
				case '{':
					if inner == "" {
						chunk.Emit(OpMakeDict, 0, 0, line)
						return nil
					}
					entries := splitTopLevelArgs(inner, ',')
					for _, entry := range entries {
						kv := splitTopLevelArgs(strings.TrimSpace(entry), ':')
						if len(kv) != 2 {
							return k.fault(line, "ERR_COMPILE_DICT", "invalid dict entry in expression", "use key: value syntax")
						}
						kRaw := strings.TrimSpace(kv[0])
						kGlyph := kRaw
						if len(kRaw) >= 2 && ((kRaw[0] == '"' && kRaw[len(kRaw)-1] == '"') || (kRaw[0] == '\'' && kRaw[len(kRaw)-1] == '\'')) {
							unq, err := unescapeGlyphSeq(kRaw[1 : len(kRaw)-1])
							if err != nil {
								return k.fault(line, "ERR_COMPILE_KEY", err.Error(), "check escape sequence in dict key")
							}
							kGlyph = unq
						}
						kIdx, err := chunk.AddConst(GlyphDatum(kGlyph), line)
						if err != nil {
							return err
						}
						chunk.Emit(OpLoadConst, kIdx, 0, line)
						if err := k.CompileExprToChunk(kv[1], chunk, line); err != nil {
							return err
						}
					}
					chunk.Emit(OpMakeDict, len(entries), 0, line)
					return nil
				}
			}
		}
	}

	if idx, op := findBinarySplit(expr, []string{"||", "or"}); idx != -1 {
		_ = op
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpOr, 0, 0, line)
		return nil
	}

	if idx, op := findBinarySplit(expr, []string{"&&", "and"}); idx != -1 {
		_ = op
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpAnd, 0, 0, line)
		return nil
	}

	if idx, op := findBinarySplit(expr, []string{"==", "!="}); idx != -1 {
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		if op == "==" {
			chunk.Emit(OpEq, 0, 0, line)
		} else {
			chunk.Emit(OpNeq, 0, 0, line)
		}
		return nil
	}

	if idx, op := findBinarySplit(expr, []string{"<=", ">=", "<", ">"}); idx != -1 {
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		switch op {
		case "<":
			chunk.Emit(OpLt, 0, 0, line)
		case "<=":
			chunk.Emit(OpLte, 0, 0, line)
		case ">":
			chunk.Emit(OpGt, 0, 0, line)
		case ">=":
			chunk.Emit(OpGte, 0, 0, line)
		}
		return nil
	}

	if idx, op := findBinarySplit(expr, []string{"+", "-"}); idx != -1 {
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		if op == "+" {
			chunk.Emit(OpAdd, 0, 0, line)
		} else {
			chunk.Emit(OpSub, 0, 0, line)
		}
		return nil
	}

	if idx, op := findBinarySplit(expr, []string{"*", "/", "%"}); idx != -1 {
		if err := k.CompileExprToChunk(expr[:idx], chunk, line); err != nil {
			return err
		}
		if err := k.CompileExprToChunk(expr[idx+len(op):], chunk, line); err != nil {
			return err
		}
		switch op {
		case "*":
			chunk.Emit(OpMul, 0, 0, line)
		case "/":
			chunk.Emit(OpDiv, 0, 0, line)
		case "%":
			chunk.Emit(OpMod, 0, 0, line)
		}
		return nil
	}

	if strings.HasPrefix(expr, "!") || strings.HasPrefix(expr, "not ") {
		var inner string
		if strings.HasPrefix(expr, "!") {
			inner = strings.TrimSpace(expr[1:])
		} else {
			inner = strings.TrimSpace(strings.TrimPrefix(expr, "not "))
		}
		if err := k.CompileExprToChunk(inner, chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpNot, 0, 0, line)
		return nil
	}

	if strings.HasPrefix(expr, "-") && len(expr) > 1 {
		inner := strings.TrimSpace(expr[1:])
		if err := k.CompileExprToChunk(inner, chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpNeg, 0, 0, line)
		return nil
	}

	if len(expr) >= 2 && expr[len(expr)-1] == ']' {
		openIdx := findPairOpen(expr, len(expr)-1, '[', ']')
		if openIdx > 0 {
			baseExpr := strings.TrimSpace(expr[:openIdx])
			idxExpr := strings.TrimSpace(expr[openIdx+1 : len(expr)-1])
			if baseExpr != "" && idxExpr != "" {
				if err := k.CompileExprToChunk(baseExpr, chunk, line); err != nil {
					return err
				}
				if err := k.CompileExprToChunk(idxExpr, chunk, line); err != nil {
					return err
				}
				chunk.Emit(OpIndexGet, 0, 0, line)
				return nil
			}
		}
	}

	switch strings.ToUpper(expr) {
	case "TRUE", "ACTIVE":
		cIdx, err := chunk.AddConst(FlagDatum(true), line)
		if err != nil {
			return err
		}
		chunk.Emit(OpLoadConst, cIdx, 0, line)
		return nil
	case "FALSE", "DORMANT":
		cIdx, err := chunk.AddConst(FlagDatum(false), line)
		if err != nil {
			return err
		}
		chunk.Emit(OpLoadConst, cIdx, 0, line)
		return nil
	case "VOID", "NULL", "NONE":
		chunk.Emit(OpLoadVoid, 0, 0, line)
		return nil
	}

	if len(expr) >= 2 && ((expr[0] == '"' && expr[len(expr)-1] == '"') || (expr[0] == '\'' && expr[len(expr)-1] == '\'')) {
		unq, err := unescapeGlyphSeq(expr[1 : len(expr)-1])
		if err != nil {
			return k.fault(line, "ERR_COMPILE_GLYPH", err.Error(), "check glyph literal quotes")
		}
		cIdx, err := chunk.AddConst(GlyphDatum(unq), line)
		if err != nil {
			return err
		}
		chunk.Emit(OpLoadConst, cIdx, 0, line)
		return nil
	}

	if n, err := strconv.ParseInt(expr, 10, 64); err == nil {
		cIdx, err := chunk.AddConst(CycleDatum(n), line)
		if err != nil {
			return err
		}
		chunk.Emit(OpLoadConst, cIdx, 0, line)
		return nil
	}

	if openParen := strings.IndexByte(expr, '('); openParen > 0 && expr[len(expr)-1] == ')' {
		closeParen := findPairClose(expr, openParen, '(', ')')
		if closeParen == len(expr)-1 {
			fnName := strings.TrimSpace(expr[:openParen])
			if IsIdentToken(fnName) {
				argRaw := strings.TrimSpace(expr[openParen+1 : len(expr)-1])
				argCount := 0
				if argRaw != "" {
					args := splitTopLevelArgs(argRaw, ',')
					for _, a := range args {
						if err := k.CompileExprToChunk(a, chunk, line); err != nil {
							return err
						}
					}
					argCount = len(args)
				}
				symIdx := chunk.AddSymbol(fnName)
				chunk.Emit(OpCall, symIdx, argCount, line)
				return nil
			}
		}
	}

	if IsIdentToken(expr) {
		symIdx := chunk.AddSymbol(expr)
		chunk.Emit(OpLoadVar, symIdx, 0, line)
		return nil
	}

	return k.fault(line, "ERR_COMPILE_TOKEN", fmt.Sprintf("cannot compile expression '%s'", expr), "verify syntax and identifiers")
}

type CallFrame struct {
	ReturnPC  int
	BaseDepth int
	Name      string
}

type AetherVM struct {
	Kernel   *AetherKernel
	Stack    []Datum
	CallInfo []CallFrame
	Units    map[string]UnitSchema
	Protos   map[string]RoutineProto
}

func NewAetherVM(kernel *AetherKernel) *AetherVM {
	return &AetherVM{
		Kernel:   kernel,
		Stack:    make([]Datum, 0, 256),
		CallInfo: make([]CallFrame, 0, 64),
		Units:    make(map[string]UnitSchema),
		Protos:   make(map[string]RoutineProto),
	}
}

func (vm *AetherVM) Push(d Datum, line int) error {
	if len(vm.Stack) >= 16384 {
		return vm.Kernel.fault(line, "ERR_VM_STACK_FULL", "operand stack capacity exceeded", "simplify nested expressions or check recursion")
	}
	vm.Stack = append(vm.Stack, d.DeepCopy())
	return nil
}

func (vm *AetherVM) Pop(line int) (Datum, error) {
	if len(vm.Stack) == 0 {
		return VoidDatum(), vm.Kernel.fault(line, "ERR_VM_STACK_UNDERFLOW", "operand stack is empty", "report compiler bytecode imbalance")
	}
	top := vm.Stack[len(vm.Stack)-1]
	vm.Stack = vm.Stack[:len(vm.Stack)-1]
	return top, nil
}

func (vm *AetherVM) Peek(line int) (Datum, error) {
	if len(vm.Stack) == 0 {
		return VoidDatum(), vm.Kernel.fault(line, "ERR_VM_STACK_EMPTY", "cannot inspect empty operand stack", "report compiler bytecode imbalance")
	}
	return vm.Stack[len(vm.Stack)-1], nil
}

func (vm *AetherVM) ExecChunk(chunk *CompiledChunk) (Datum, error) {
	if chunk == nil {
		return VoidDatum(), vm.Kernel.fault(1, "ERR_NIL_CHUNK", "compiled bytecode chunk is nil", "compile source before executing VM")
	}

	pc := chunk.EntryPC
	lastResult := VoidDatum()

	for pc >= 0 && pc < len(chunk.Code) {
		ins := chunk.Code[pc]
		if err := vm.Kernel.AdvanceCycle(ins.Line); err != nil {
			return VoidDatum(), err
		}

		switch ins.Op {
		case OpNop:
			pc++

		case OpLoadConst:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Constants) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_CONST_IDX", "constant pool index out of range", "recompile chunk")
			}
			if err := vm.Push(chunk.Constants[ins.ArgA], ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpLoadVoid:
			if err := vm.Push(VoidDatum(), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpLoadVar:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "symbol index out of range", "recompile chunk")
			}
			sym := chunk.Symbols[ins.ArgA]
			if schema, isUnit := vm.Units[sym]; isUnit {
				_ = schema
				if err := vm.Push(GlyphDatum("unit:"+sym), ins.Line); err != nil {
					return VoidDatum(), err
				}
				pc++
				break
			}
			if proto, isFn := vm.Protos[sym]; isFn {
				if err := vm.Push(CodeRefDatum(proto.Entry), ins.Line); err != nil {
					return VoidDatum(), err
				}
				pc++
				break
			}
			val, ok := vm.Kernel.LoadSymbol(sym)
			if !ok {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_UNBOUND_VAR", fmt.Sprintf("symbol '%s' is not bound", sym), "bind symbol with let/bind or assign before reading")
			}
			if err := vm.Push(val, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpStoreVar:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "symbol index out of range", "recompile chunk")
			}
			val, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			sym := chunk.Symbols[ins.ArgA]
			vm.Kernel.StoreSymbol(sym, val)
			lastResult = val.DeepCopy()
			pc++

		case OpStoreLocal:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "symbol index out of range", "recompile chunk")
			}
			val, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			sym := chunk.Symbols[ins.ArgA]
			vm.Kernel.StoreLocalSymbol(sym, val)
			lastResult = val.DeepCopy()
			pc++

		case OpMakeVec:
			count := ins.ArgA
			if count < 0 || len(vm.Stack) < count {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_VEC_STACK", "insufficient items on stack for vector construction", "verify vector literal syntax")
			}
			items := make([]Datum, count)
			start := len(vm.Stack) - count
			for i := 0; i < count; i++ {
				items[i] = vm.Stack[start+i].DeepCopy()
			}
			vm.Stack = vm.Stack[:start]
			vec := VecDatum(items)
			if err := vm.Push(vec, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpMakeDict:
			pairCount := ins.ArgA
			needed := pairCount * 2
			if pairCount < 0 || len(vm.Stack) < needed {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_DICT_STACK", "insufficient key-value pairs on stack for dict construction", "verify dict literal syntax")
			}
			entries := make(map[string]Datum, pairCount)
			base := len(vm.Stack) - needed
			for i := 0; i < pairCount; i++ {
				kDatum := vm.Stack[base+i*2]
				vDatum := vm.Stack[base+i*2+1]
				entries[kDatum.Stringify()] = vDatum.DeepCopy()
			}
			vm.Stack = vm.Stack[:base]
			d := DictDatum(entries)
			if err := vm.Push(d, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpIndexGet:
			idxVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			targetVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			res, err := vm.evalIndexGet(targetVal, idxVal, ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(res, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpIndexSet:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "symbol index out of range", "recompile chunk")
			}
			newVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			idxVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			sym := chunk.Symbols[ins.ArgA]
			baseVal, ok := vm.Kernel.LoadSymbol(sym)
			if !ok {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_UNBOUND_TARGET", fmt.Sprintf("target '%s' is not bound", sym), "bind container before indexing")
			}
			updated, err := vm.evalIndexSet(baseVal, idxVal, newVal, ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			vm.Kernel.StoreSymbol(sym, updated)
			lastResult = newVal.DeepCopy()
			pc++

		case OpAdd:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			res, err := vm.evalAdd(lhs, rhs, ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(res, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpSub:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if lhs.Kind != DatumCycle || rhs.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SUB_TYPE", "subtraction requires cycle values", "use cycle numbers with -")
			}
			if err := vm.Push(CycleDatum(lhs.Cycle-rhs.Cycle), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpMul:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if lhs.Kind != DatumCycle || rhs.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_MUL_TYPE", "multiplication requires cycle values", "use cycle numbers with *")
			}
			if err := vm.Push(CycleDatum(lhs.Cycle*rhs.Cycle), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpDiv:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if lhs.Kind != DatumCycle || rhs.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_DIV_TYPE", "division requires cycle values", "use cycle numbers with /")
			}
			if rhs.Cycle == 0 {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_DIV_ZERO", "division by zero cycle", "ensure divisor is non-zero")
			}
			if err := vm.Push(CycleDatum(lhs.Cycle/rhs.Cycle), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpMod:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if lhs.Kind != DatumCycle || rhs.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_MOD_TYPE", "modulo requires cycle values", "use cycle numbers with %")
			}
			if rhs.Cycle == 0 {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_MOD_ZERO", "modulo by zero cycle", "ensure divisor is non-zero")
			}
			if err := vm.Push(CycleDatum(lhs.Cycle%rhs.Cycle), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpNeg:
			val, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if val.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_NEG_TYPE", "unary negation requires cycle value", "negate a cycle number")
			}
			if err := vm.Push(CycleDatum(-val.Cycle), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpNot:
			val, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(FlagDatum(!val.Truthy()), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpEq:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(FlagDatum(lhs.Equals(rhs)), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpNeq:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(FlagDatum(!lhs.Equals(rhs)), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpLt, OpLte, OpGt, OpGte:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if lhs.Kind != DatumCycle || rhs.Kind != DatumCycle {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_CMP_TYPE", "relational operators require cycle operands", "compare cycle values")
			}
			var cmp bool
			switch ins.Op {
			case OpLt:
				cmp = lhs.Cycle < rhs.Cycle
			case OpLte:
				cmp = lhs.Cycle <= rhs.Cycle
			case OpGt:
				cmp = lhs.Cycle > rhs.Cycle
			case OpGte:
				cmp = lhs.Cycle >= rhs.Cycle
			}
			if err := vm.Push(FlagDatum(cmp), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpAnd:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(FlagDatum(lhs.Truthy() && rhs.Truthy()), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpOr:
			rhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lhs, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(FlagDatum(lhs.Truthy() || rhs.Truthy()), ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpJump:
			pc = ins.ArgA

		case OpJumpIfFalse:
			cond, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if !cond.Truthy() {
				pc = ins.ArgA
			} else {
				pc++
			}

		case OpJumpIfTrue:
			cond, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if cond.Truthy() {
				pc = ins.ArgA
			} else {
				pc++
			}

		case OpCall:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "routine symbol index out of range", "recompile chunk")
			}
			fnName := chunk.Symbols[ins.ArgA]
			argCount := ins.ArgB
			if argCount < 0 || len(vm.Stack) < argCount {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_CALL_STACK", "insufficient arguments on stack for routine call", "check routine call arguments")
			}
			args := make([]Datum, argCount)
			start := len(vm.Stack) - argCount
			for i := 0; i < argCount; i++ {
				args[i] = vm.Stack[start+i].DeepCopy()
			}
			vm.Stack = vm.Stack[:start]

			if schema, ok := vm.Units[fnName]; ok {
				fields := make(map[string]Datum, len(schema.Fields))
				for i, fname := range schema.Fields {
					if i < len(args) {
						fields[fname] = args[i].DeepCopy()
					} else {
						fields[fname] = VoidDatum()
					}
				}
				uDatum := UnitDatum(fnName, fields)
				if err := vm.Push(uDatum, ins.Line); err != nil {
					return VoidDatum(), err
				}
				pc++
				break
			}

			if handled, bRes, bErr := vm.callBuiltin(fnName, args, ins.Line); handled {
				if bErr != nil {
					return VoidDatum(), bErr
				}
				if err := vm.Push(bRes, ins.Line); err != nil {
					return VoidDatum(), err
				}
				pc++
				break
			}

			if handled, libRes, libErr := EvalCustomLibs(fnName, args, ins.Line); handled {
				if libErr != nil { return VoidDatum(), libErr }
				if err := vm.Push(libRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, prRes, prErr := EvalPkgReleaseLibs(fnName, args, ins.Line); handled {
				if prErr != nil { return VoidDatum(), prErr }
				if err := vm.Push(prRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, reRes, reErr := EvalRuntimeEngineLibs(fnName, args, ins.Line); handled {
				if reErr != nil { return VoidDatum(), reErr }
				if err := vm.Push(reRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, cpRes, cpErr := EvalCoreProLibs(fnName, args, ins.Line); handled {
				if cpErr != nil { return VoidDatum(), cpErr }
				if err := vm.Push(cpRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, fsRes, fsErr := EvalFullStackLibs(fnName, args, ins.Line); handled {
				if fsErr != nil { return VoidDatum(), fsErr }
				if err := vm.Push(fsRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, aioRes, aioErr := EvalAllInOneLibs(fnName, args, ins.Line); handled {
				if aioErr != nil { return VoidDatum(), aioErr }
				if err := vm.Push(aioRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, tfsRes, tfsErr := EvalTimeFSLibs(fnName, args, ins.Line); handled {
				if tfsErr != nil { return VoidDatum(), tfsErr }
				if err := vm.Push(tfsRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, netRes, netErr := EvalNetPlusLibs(fnName, args, ins.Line); handled {
				if netErr != nil { return VoidDatum(), netErr }
				if err := vm.Push(netRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, datRes, datErr := EvalDataLibs(fnName, args, ins.Line); handled {
				if datErr != nil { return VoidDatum(), datErr }
				if err := vm.Push(datRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, regRes, regErr := EvalRegexLibs(fnName, args, ins.Line); handled {
				if regErr != nil { return VoidDatum(), regErr }
				if err := vm.Push(regRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, ex4Res, ex4Err := EvalExtra4Libs(fnName, args, ins.Line); handled {
				if ex4Err != nil { return VoidDatum(), ex4Err }
				if err := vm.Push(ex4Res, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, proRes, proErr := EvalProLibs(fnName, args, ins.Line); handled {
				if proErr != nil { return VoidDatum(), proErr }
				if err := vm.Push(proRes, ins.Line); err != nil { return VoidDatum(), err }
				pc++
				break
			}
			if handled, extRes, extErr := vm.EvalExtendedBuiltin(fnName, args, ins.Line); handled {
				if extErr != nil {
					return VoidDatum(), extErr
				}
				if err := vm.Push(extRes, ins.Line); err != nil {
					return VoidDatum(), err
				}
				pc++
				break
			}

			proto, ok := vm.Protos[fnName]
			if !ok {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_UNDEF_ROUTINE", fmt.Sprintf("routine '%s' is not defined", fnName), "define routine before calling")
			}
			if len(args) != len(proto.Params) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_ROUTINE_ARITY", fmt.Sprintf("routine '%s' expects %d args, got %d", fnName, len(proto.Params), len(args)), "match argument count")
			}
			if err := vm.Kernel.OpenFrame(fnName); err != nil {
				return VoidDatum(), err
			}
			for i, pName := range proto.Params {
				vm.Kernel.StoreLocalSymbol(pName, args[i])
			}
			vm.CallInfo = append(vm.CallInfo, CallFrame{
				ReturnPC:  pc + 1,
				BaseDepth: len(vm.Stack),
				Name:      fnName,
			})
			pc = proto.Entry

		case OpReturn:
			retVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lastResult = retVal.DeepCopy()
			if len(vm.CallInfo) == 0 {
				return lastResult, nil
			}
			frame := vm.CallInfo[len(vm.CallInfo)-1]
			vm.CallInfo = vm.CallInfo[:len(vm.CallInfo)-1]
			if len(vm.Stack) > frame.BaseDepth {
				vm.Stack = vm.Stack[:frame.BaseDepth]
			}
			vm.Kernel.CloseFrame()
			if err := vm.Push(lastResult, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc = frame.ReturnPC

		case OpPop:
			val, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			lastResult = val.DeepCopy()
			pc++

		case OpAssert:
			cond, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if !cond.Truthy() {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_ASSERT_FAIL", "assertion check evaluated to false", "verify invariant expression")
			}
			lastResult = FlagDatum(true)
			pc++

		case OpRaise:
			msgVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_RAISED", msgVal.Stringify(), "inspect error condition or guard block")

		case OpGetField:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "field symbol index out of range", "recompile chunk")
			}
			fieldName := chunk.Symbols[ins.ArgA]
			obj, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			fVal, err := vm.evalFieldGet(obj, fieldName, ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			if err := vm.Push(fVal, ins.Line); err != nil {
				return VoidDatum(), err
			}
			pc++

		case OpSetField:
			if ins.ArgA < 0 || ins.ArgA >= len(chunk.Symbols) {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_SYM_IDX", "field symbol index out of range", "recompile chunk")
			}
			sym := chunk.Symbols[ins.ArgA]
			fieldName := ""
			if ins.ArgB >= 0 && ins.ArgB < len(chunk.Symbols) {
				fieldName = chunk.Symbols[ins.ArgB]
			}
			newVal, err := vm.Pop(ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			baseObj, ok := vm.Kernel.LoadSymbol(sym)
			if !ok {
				return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_UNBOUND_TARGET", fmt.Sprintf("target '%s' is not bound", sym), "bind unit or dict first")
			}
			updatedObj, err := vm.evalFieldSet(baseObj, fieldName, newVal, ins.Line)
			if err != nil {
				return VoidDatum(), err
			}
			vm.Kernel.StoreSymbol(sym, updatedObj)
			lastResult = newVal.DeepCopy()
			pc++

		default:
			return VoidDatum(), vm.Kernel.fault(ins.Line, "ERR_BAD_OPCODE", fmt.Sprintf("unknown opcode %d", int(ins.Op)), "recompile source code")
		}
	}

	if len(vm.Stack) > 0 {
		lastResult = vm.Stack[len(vm.Stack)-1].DeepCopy()
	}
	return lastResult, nil
}

func (vm *AetherVM) evalAdd(lhs, rhs Datum, line int) (Datum, error) {
	if lhs.Kind == DatumCycle && rhs.Kind == DatumCycle {
		return CycleDatum(lhs.Cycle + rhs.Cycle), nil
	}
	if lhs.Kind == DatumGlyph || rhs.Kind == DatumGlyph {
		return GlyphDatum(lhs.Stringify() + rhs.Stringify()), nil
	}
	if lhs.Kind == DatumVec && rhs.Kind == DatumVec {
		out := make([]Datum, 0, len(lhs.Vec)+len(rhs.Vec))
		for _, a := range lhs.Vec {
			out = append(out, a.DeepCopy())
		}
		for _, b := range rhs.Vec {
			out = append(out, b.DeepCopy())
		}
		return VecDatum(out), nil
	}
	return VoidDatum(), vm.Kernel.fault(line, "ERR_ADD_MISMATCH", fmt.Sprintf("cannot add %s and %s", lhs.TypeName(), rhs.TypeName()), "use compatible cycle, glyph, or vec operands")
}

func (vm *AetherVM) evalIndexGet(target, idx Datum, line int) (Datum, error) {
	switch target.Kind {
	case DatumVec:
		if idx.Kind != DatumCycle {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_VEC_IDX_TYPE", "vec index must be cycle", "use integer index")
		}
		i := int(idx.Cycle)
		if i < 0 || i >= len(target.Vec) {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_VEC_BOUNDS", fmt.Sprintf("vec index %d out of range [0..%d)", i, len(target.Vec)), "check vec length")
		}
		return target.Vec[i].DeepCopy(), nil
	case DatumDict:
		if idx.Kind != DatumGlyph {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_DICT_KEY_TYPE", "dict key must be glyph", "use glyph key")
		}
		v, ok := target.Dict[idx.Glyph]
		if !ok {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_DICT_MISSING", fmt.Sprintf("key '%s' missing in dict", idx.Glyph), "verify key existence")
		}
		return v.DeepCopy(), nil
	case DatumGlyph:
		if idx.Kind != DatumCycle {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_IDX_TYPE", "glyph index must be cycle", "use integer index")
		}
		runes := []rune(target.Glyph)
		i := int(idx.Cycle)
		if i < 0 || i >= len(runes) {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_BOUNDS", fmt.Sprintf("glyph index %d out of range [0..%d)", i, len(runes)), "check glyph length")
		}
		return GlyphDatum(string(runes[i])), nil
	default:
		return VoidDatum(), vm.Kernel.fault(line, "ERR_NOT_INDEXABLE", fmt.Sprintf("type '%s' cannot be indexed", target.TypeName()), "index vec, dict, or glyph")
	}
}

func (vm *AetherVM) evalIndexSet(base, idx, val Datum, line int) (Datum, error) {
	switch base.Kind {
	case DatumVec:
		if idx.Kind != DatumCycle {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_VEC_IDX_TYPE", "vec index must be cycle", "use integer index")
		}
		i := int(idx.Cycle)
		if i < 0 || i >= len(base.Vec) {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_VEC_BOUNDS", fmt.Sprintf("vec index %d out of range [0..%d)", i, len(base.Vec)), "check vec length")
		}
		cp := make([]Datum, len(base.Vec))
		for k := range base.Vec {
			cp[k] = base.Vec[k].DeepCopy()
		}
		cp[i] = val.DeepCopy()
		return VecDatum(cp), nil
	case DatumDict:
		if idx.Kind != DatumGlyph {
			return VoidDatum(), vm.Kernel.fault(line, "ERR_DICT_KEY_TYPE", "dict key must be glyph", "use glyph key")
		}
		cp := make(map[string]Datum, len(base.Dict)+1)
		for k, v := range base.Dict {
			cp[k] = v.DeepCopy()
		}
		cp[idx.Glyph] = val.DeepCopy()
		return DictDatum(cp), nil
	default:
		return VoidDatum(), vm.Kernel.fault(line, "ERR_IMMUTABLE_INDEX", fmt.Sprintf("type '%s' does not support indexed assignment", base.TypeName()), "assign to vec or dict")
	}
}

func (vm *AetherVM) evalFieldGet(obj Datum, field string, line int) (Datum, error) {
	switch obj.Kind {
	case DatumUnit:
		if v, ok := obj.UnitMeta[field]; ok {
			return v.DeepCopy(), nil
		}
		return VoidDatum(), vm.Kernel.fault(line, "ERR_UNIT_FIELD", fmt.Sprintf("unit '%s' has no field '%s'", obj.UnitTag, field), "check unit definition")
	case DatumDict:
		if v, ok := obj.Dict[field]; ok {
			return v.DeepCopy(), nil
		}
		return VoidDatum(), vm.Kernel.fault(line, "ERR_DICT_FIELD", fmt.Sprintf("dict has no key '%s'", field), "verify key name")
	default:
		return VoidDatum(), vm.Kernel.fault(line, "ERR_FIELD_ACCESS", fmt.Sprintf("type '%s' does not support dot field access", obj.TypeName()), "access fields on unit or dict")
	}
}

func (vm *AetherVM) evalFieldSet(obj Datum, field string, val Datum, line int) (Datum, error) {
	switch obj.Kind {
	case DatumUnit:
		cp := make(map[string]Datum, len(obj.UnitMeta)+1)
		for k, v := range obj.UnitMeta {
			cp[k] = v.DeepCopy()
		}
		cp[field] = val.DeepCopy()
		return UnitDatum(obj.UnitTag, cp), nil
	case DatumDict:
		cp := make(map[string]Datum, len(obj.Dict)+1)
		for k, v := range obj.Dict {
			cp[k] = v.DeepCopy()
		}
		cp[field] = val.DeepCopy()
		return DictDatum(cp), nil
	default:
		return VoidDatum(), vm.Kernel.fault(line, "ERR_FIELD_SET", fmt.Sprintf("type '%s' does not support dot field assignment", obj.TypeName()), "assign fields on unit or dict")
	}
}

func (vm *AetherVM) callBuiltin(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "emit_bmm", "emit", "print", "say":
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = a.Stringify()
		}
		rendered := strings.Join(parts, " ")
		fmt.Println(rendered)
		return true, GlyphDatum(rendered), nil

	case "kind_bmm", "kind", "type_of":
		if len(args) != 1 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", fmt.Sprintf("%s expects 1 argument, got %d", name, len(args)), "pass a single datum")
		}
		return true, GlyphDatum(args[0].TypeName()), nil

	case "len_bmm", "len", "len_of":
		if len(args) != 1 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", fmt.Sprintf("%s expects 1 argument, got %d", name, len(args)), "pass a vec, dict, or glyph")
		}
		switch args[0].Kind {
		case DatumVec:
			return true, CycleDatum(int64(len(args[0].Vec))), nil
		case DatumDict:
			return true, CycleDatum(int64(len(args[0].Dict))), nil
		case DatumGlyph:
			return true, CycleDatum(int64(len([]rune(args[0].Glyph)))), nil
		default:
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_LEN_TYPE_BMM", fmt.Sprintf("cannot measure length of %s", args[0].TypeName()), "use vec, dict, or glyph")
		}

	case "cycle_bmm", "to_cycle", "cycle":
		if len(args) != 1 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", fmt.Sprintf("%s expects 1 argument, got %d", name, len(args)), "pass one datum")
		}
		switch args[0].Kind {
		case DatumCycle:
			return true, args[0].DeepCopy(), nil
		case DatumFlag:
			if args[0].Flag {
				return true, CycleDatum(1), nil
			}
			return true, CycleDatum(0), nil
		case DatumPulse:
			return true, CycleDatum(args[0].FreqHz), nil
		case DatumGlyph:
			n, err := strconv.ParseInt(strings.TrimSpace(args[0].Glyph), 10, 64)
			if err != nil {
				return true, VoidDatum(), vm.Kernel.fault(line, "ERR_PARSE_CYCLE_BMM", fmt.Sprintf("cannot convert glyph '%s' to cycle", args[0].Glyph), "provide an integer glyph")
			}
			return true, CycleDatum(n), nil
		default:
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_CAST_CYCLE_BMM", fmt.Sprintf("cannot cast %s to cycle", args[0].TypeName()), "use cycle, flag, pulse, or numeric glyph")
		}

	case "glyph_bmm", "to_glyph", "glyph":
		if len(args) != 1 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", fmt.Sprintf("%s expects 1 argument, got %d", name, len(args)), "pass one datum")
		}
		return true, GlyphDatum(args[0].Stringify()), nil

	case "state_bmm", "to_state", "state":
		if len(args) != 1 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", fmt.Sprintf("%s expects 1 argument, got %d", name, len(args)), "pass one datum")
		}
		return true, FlagDatum(args[0].Truthy()), nil

	case "seq_bmm", "seq", "range_vec":
		if len(args) < 1 || len(args) > 3 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_ARITY_BMM", "seq_bmm expects 1 to 3 cycle bounds", "use seq_bmm(end), seq_bmm(start, end), or seq_bmm(start, end, step)")
		}
		var start, end, step int64 = 0, 0, 1
		if len(args) == 1 {
			if args[0].Kind != DatumCycle {
				return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_TYPE_BMM", "seq_bmm bound must be cycle", "pass cycle integer")
			}
			end = args[0].Cycle
		} else if len(args) == 2 {
			if args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
				return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_TYPE_BMM", "seq_bmm bounds must be cycle", "pass cycle integers")
			}
			start, end = args[0].Cycle, args[1].Cycle
		} else {
			for _, a := range args {
				if a.Kind != DatumCycle {
					return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_TYPE_BMM", "seq_bmm arguments must be cycle", "pass cycle integers")
				}
			}
			start, end, step = args[0].Cycle, args[1].Cycle, args[2].Cycle
			if step == 0 {
				return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_STEP_BMM", "seq_bmm step cannot be zero", "use a non-zero cycle step")
			}
		}
		var list []Datum
		if step > 0 {
			for cur := start; cur < end; cur += step {
				if len(list) >= MaxSeqSpan {
					return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_CAP_BMM", "sequence span exceeded maximum capacity", "narrow bounds or increase step")
				}
				list = append(list, CycleDatum(cur))
			}
		} else {
			for cur := start; cur > end; cur += step {
				if len(list) >= MaxSeqSpan {
					return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SEQ_CAP_BMM", "sequence span exceeded maximum capacity", "narrow bounds or increase step")
				}
				list = append(list, CycleDatum(cur))
			}
		}
		return true, VecDatum(list), nil

	case "push_bmm", "append", "push_vec":
		if len(args) != 2 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_PUSH_ARGS_BMM", "push_bmm expects (vec, datum)", "pass a vec and an element")
		}
		out := make([]Datum, len(args[0].Vec)+1)
		for i, item := range args[0].Vec {
			out[i] = item.DeepCopy()
		}
		out[len(args[0].Vec)] = args[1].DeepCopy()
		return true, VecDatum(out), nil

	case "pop_bmm", "pop", "pop_vec":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_POP_ARGS_BMM", "pop_bmm expects a single vec", "pass a non-empty vec")
		}
		if len(args[0].Vec) == 0 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_EMPTY_VEC_BMM", "cannot pop from an empty vec", "verify len_bmm(v) > 0")
		}
		return true, args[0].Vec[len(args[0].Vec)-1].DeepCopy(), nil

	case "slice_bmm", "slice":
		if len(args) != 3 || args[0].Kind != DatumVec || args[1].Kind != DatumCycle || args[2].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_SLICE_ARGS_BMM", "slice_bmm expects (vec, start_cycle, end_cycle)", "pass vec and two cycle bounds")
		}
		n := len(args[0].Vec)
		sIdx := int(args[1].Cycle)
		eIdx := int(args[2].Cycle)
		if sIdx < 0 {
			sIdx = 0
		}
		if eIdx > n {
			eIdx = n
		}
		if sIdx >= eIdx {
			return true, VecDatum(nil), nil
		}
		res := make([]Datum, eIdx-sIdx)
		for i := sIdx; i < eIdx; i++ {
			res[i-sIdx] = args[0].Vec[i].DeepCopy()
		}
		return true, VecDatum(res), nil

	case "keys_bmm", "keys", "keys_of":
		if len(args) != 1 || args[0].Kind != DatumDict {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_KEYS_ARGS_BMM", "keys_bmm expects a single dict", "pass a dict datum")
		}
		ks := make([]Datum, 0, len(args[0].Dict))
		for k := range args[0].Dict {
			ks = append(ks, GlyphDatum(k))
		}
		return true, VecDatum(ks), nil

	case "has_bmm", "has_key":
		if len(args) != 2 || args[0].Kind != DatumDict || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_HAS_ARGS_BMM", "has_bmm expects (dict, glyph_key)", "pass a dict and a glyph key")
		}
		_, ok := args[0].Dict[args[1].Glyph]
		return true, FlagDatum(ok), nil

	case "pulse_bmm", "pulse", "link_pulse":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_PULSE_ARGS_BMM", "pulse_bmm expects (band_glyph, hz_cycle)", "pass band glyph and cycle frequency")
		}
		sig := PulseDatum(args[0].Glyph, args[1].Cycle)
		vm.Kernel.carriers[args[0].Glyph] = sig
		return true, sig, nil

	case "read_pulse_bmm", "read_pulse", "pulse_read":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_PULSE_ARGS_BMM", "read_pulse_bmm expects (band_glyph)", "pass band glyph")
		}
		if sig, ok := vm.Kernel.carriers[args[0].Glyph]; ok {
			return true, sig.DeepCopy(), nil
		}
		return true, PulseDatum(args[0].Glyph, 0), nil

	case "vault_put_bmm", "vault_store":
		if len(args) != 2 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_VAULT_ARGS_BMM", "vault_put_bmm expects (slot_glyph, value)", "pass glyph slot key and value")
		}
		vm.Kernel.vaults[args[0].Glyph] = args[1].DeepCopy()
		return true, args[1].DeepCopy(), nil

	case "vault_get_bmm", "vault_load":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_VAULT_ARGS_BMM", "vault_get_bmm expects (slot_glyph)", "pass glyph slot key")
		}
		if v, ok := vm.Kernel.vaults[args[0].Glyph]; ok {
			return true, v.DeepCopy(), nil
		}
		return true, VoidDatum(), nil

	case "vault_del_bmm", "vault_drop":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_VAULT_ARGS_BMM", "vault_del_bmm expects (slot_glyph)", "pass glyph slot key")
		}
		_, existed := vm.Kernel.vaults[args[0].Glyph]
		delete(vm.Kernel.vaults, args[0].Glyph)
		return true, FlagDatum(existed), nil

	case "abs_bmm", "abs":
		if len(args) != 1 || args[0].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MATH_ARGS_BMM", "abs_bmm expects a cycle operand", "pass a cycle value")
		}
		if args[0].Cycle < 0 {
			return true, CycleDatum(-args[0].Cycle), nil
		}
		return true, args[0], nil

	case "clamp_bmm", "clamp":
		if len(args) != 3 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle || args[2].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MATH_ARGS_BMM", "clamp_bmm expects (val, lo, hi) as cycle", "pass three cycle values")
		}
		v, lo, hi := args[0].Cycle, args[1].Cycle, args[2].Cycle
		if lo > hi {
			lo, hi = hi, lo
		}
		if v < lo {
			return true, CycleDatum(lo), nil
		}
		if v > hi {
			return true, CycleDatum(hi), nil
		}
		return true, CycleDatum(v), nil

	case "min_bmm", "min_val":
		if len(args) != 2 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MATH_ARGS_BMM", "min_bmm expects two cycle operands", "pass two cycle values")
		}
		if args[0].Cycle < args[1].Cycle {
			return true, args[0], nil
		}
		return true, args[1], nil

	case "max_bmm", "max_val":
		if len(args) != 2 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MATH_ARGS_BMM", "max_bmm expects two cycle operands", "pass two cycle values")
		}
		if args[0].Cycle > args[1].Cycle {
			return true, args[0], nil
		}
		return true, args[1], nil

	case "upper_bmm", "upper":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_ARGS_BMM", "upper_bmm expects a glyph operand", "pass a glyph value")
		}
		return true, GlyphDatum(strings.ToUpper(args[0].Glyph)), nil

	case "lower_bmm", "lower":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_ARGS_BMM", "lower_bmm expects a glyph operand", "pass a glyph value")
		}
		return true, GlyphDatum(strings.ToLower(args[0].Glyph)), nil

	case "trim_bmm", "trim_str":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_ARGS_BMM", "trim_bmm expects a glyph operand", "pass a glyph value")
		}
		return true, GlyphDatum(strings.TrimSpace(args[0].Glyph)), nil

	case "split_bmm", "split_str":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_GLYPH_ARGS_BMM", "split_bmm expects (glyph, sep_glyph)", "pass two glyph values")
		}
		raw := strings.Split(args[0].Glyph, args[1].Glyph)
		out := make([]Datum, len(raw))
		for i, p := range raw {
			out[i] = GlyphDatum(p)
		}
		return true, VecDatum(out), nil

	case "join_bmm", "join_vec":
		if len(args) != 2 || args[0].Kind != DatumVec || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_JOIN_ARGS_BMM", "join_bmm expects (vec, sep_glyph)", "pass a vec and a glyph separator")
		}
		parts := make([]string, len(args[0].Vec))
		for i, item := range args[0].Vec {
			parts[i] = item.Stringify()
		}
		return true, GlyphDatum(strings.Join(parts, args[1].Glyph)), nil
	}

	return false, VoidDatum(), nil
}

type BytecodeCompiler struct {
	Kernel *AetherKernel
	VM     *AetherVM
}

func NewBytecodeCompiler(kernel *AetherKernel, vm *AetherVM) *BytecodeCompiler {
	return &BytecodeCompiler{
		Kernel: kernel,
		VM:     vm,
	}
}

func (bc *BytecodeCompiler) CompileProgram(spans []CodeSpan) (*CompiledChunk, error) {
	chunk := NewCompiledChunk()

	mainJumpPC := chunk.Emit(OpJump, 0, 0, 1)

	if err := bc.compileBlockSpans(spans, chunk, 1); err != nil {
		return nil, err
	}

	voidConst, err := chunk.AddConst(VoidDatum(), 1)
	if err != nil {
		return nil, err
	}
	chunk.Emit(OpLoadConst, voidConst, 0, 1)
	chunk.Emit(OpReturn, 0, 0, 1)

	entryPC := len(chunk.Code)
	chunk.PatchJump(mainJumpPC, entryPC)
	chunk.EntryPC = entryPC

	if err := bc.compileBlockSpans(spans, chunk, 1); err != nil {
		return nil, err
	}

	if len(chunk.Code) == entryPC || chunk.Code[len(chunk.Code)-1].Op != OpReturn {
		chunk.Emit(OpLoadVoid, 0, 0, 1)
		chunk.Emit(OpReturn, 0, 0, 1)
	}

	return chunk, nil
}

func (bc *BytecodeCompiler) compileBlockSpans(spans []CodeSpan, chunk *CompiledChunk, parentLine int) error {
	_ = parentLine
	i := 0
	for i < len(spans) {
		cur := spans[i]
		baseIndent := cur.Indent

		if bc.isBlockStart(cur.Stmt) {
			j := i + 1
			for j < len(spans) && spans[j].Indent > baseIndent {
				j++
			}
			body := spans[i+1 : j]

			if bc.hasKw(cur.Stmt, "when") || bc.hasKw(cur.Stmt, "if") {
				nextIdx, err := bc.compileConditionalChain(spans, i, baseIndent, chunk)
				if err != nil {
					return err
				}
				i = nextIdx
				continue
			}

			if bc.hasKw(cur.Stmt, "loop") || bc.hasKw(cur.Stmt, "while") {
				if err := bc.compileWhileLoop(cur, body, chunk); err != nil {
					return err
				}
				i = j
				continue
			}

			if bc.hasKw(cur.Stmt, "for") {
				if err := bc.compileForRangeLoop(cur, body, chunk); err != nil {
					return err
				}
				i = j
				continue
			}

			if bc.hasKw(cur.Stmt, "fn") || bc.hasKw(cur.Stmt, "def") || bc.hasKw(cur.Stmt, "routine") {
				if err := bc.compileRoutineDecl(cur, body, chunk); err != nil {
					return err
				}
				i = j
				continue
			}

			if bc.hasKw(cur.Stmt, "unit") || bc.hasKw(cur.Stmt, "struct") {
				if err := bc.compileUnitDecl(cur, body, chunk); err != nil {
					return err
				}
				i = j
				continue
			}

			if bc.isGuardStart(cur.Stmt) {
				rescueBody, afterIdx := bc.collectRescueBlock(spans, j, baseIndent)
				if err := bc.compileGuardBlock(cur, body, rescueBody, chunk); err != nil {
					return err
				}
				i = afterIdx
				continue
			}

			return bc.Kernel.fault(cur.LineNo, "ERR_COMPILE_BLOCK", fmt.Sprintf("unknown block construct '%s'", cur.Stmt), "use when/if, loop/while, for, fn/def/routine, unit, or guard")
		}

		if err := bc.compileStmtLine(cur, chunk); err != nil {
			return err
		}
		i++
	}
	return nil
}

func (bc *BytecodeCompiler) compileConditionalChain(spans []CodeSpan, startIdx int, baseIndent int, chunk *CompiledChunk) (int, error) {
	first := spans[startIdx]
	endFirst := startIdx + 1
	for endFirst < len(spans) && spans[endFirst].Indent > baseIndent {
		endFirst++
	}
	firstBody := spans[startIdx+1 : endFirst]

	condExpr := bc.stripKwHeader(first.Stmt, []string{"when", "if"})
	if err := bc.Kernel.CompileExprToChunk(condExpr, chunk, first.LineNo); err != nil {
		return startIdx, err
	}
	jumpFalsePC := chunk.Emit(OpJumpIfFalse, 0, 0, first.LineNo)

	if err := bc.compileBlockSpans(firstBody, chunk, first.LineNo); err != nil {
		return startIdx, err
	}
	endJumpPC := chunk.Emit(OpJump, 0, 0, first.LineNo)

	falseTargetPC := len(chunk.Code)
	chunk.PatchJump(jumpFalsePC, falseTargetPC)

	var endJumps []int
	endJumps = append(endJumps, endJumpPC)

	cursor := endFirst
	for cursor < len(spans) && spans[cursor].Indent == baseIndent {
		br := spans[cursor]
		brEnd := cursor + 1
		for brEnd < len(spans) && spans[brEnd].Indent > baseIndent {
			brEnd++
		}
		brBody := spans[cursor+1 : brEnd]

		if bc.hasKw(br.Stmt, "elif") || bc.hasKw(br.Stmt, "else if") {
			elifCond := bc.stripKwHeader(br.Stmt, []string{"elif", "else if"})
			if err := bc.Kernel.CompileExprToChunk(elifCond, chunk, br.LineNo); err != nil {
				return cursor, err
			}
			nextFalsePC := chunk.Emit(OpJumpIfFalse, 0, 0, br.LineNo)
			if err := bc.compileBlockSpans(brBody, chunk, br.LineNo); err != nil {
				return cursor, err
			}
			exitPC := chunk.Emit(OpJump, 0, 0, br.LineNo)
			endJumps = append(endJumps, exitPC)
			chunk.PatchJump(nextFalsePC, len(chunk.Code))
			cursor = brEnd
			continue
		}

		if bc.isElseStart(br.Stmt) {
			if err := bc.compileBlockSpans(brBody, chunk, br.LineNo); err != nil {
				return cursor, err
			}
			cursor = brEnd
			break
		}

		break
	}

	finalPC := len(chunk.Code)
	for _, ej := range endJumps {
		chunk.PatchJump(ej, finalPC)
	}
	return cursor, nil
}

func (bc *BytecodeCompiler) compileWhileLoop(header CodeSpan, body []CodeSpan, chunk *CompiledChunk) error {
	loopStartPC := len(chunk.Code)
	condExpr := bc.stripKwHeader(header.Stmt, []string{"loop", "while"})
	if err := bc.Kernel.CompileExprToChunk(condExpr, chunk, header.LineNo); err != nil {
		return err
	}
	exitJumpPC := chunk.Emit(OpJumpIfFalse, 0, 0, header.LineNo)

	if err := bc.compileBlockSpans(body, chunk, header.LineNo); err != nil {
		return err
	}

	chunk.Emit(OpJump, loopStartPC, 0, header.LineNo)
	afterLoopPC := len(chunk.Code)
	chunk.PatchJump(exitJumpPC, afterLoopPC)
	return nil
}

func (bc *BytecodeCompiler) compileForRangeLoop(header CodeSpan, body []CodeSpan, chunk *CompiledChunk) error {
	raw := strings.TrimSpace(bc.stripKwHeader(header.Stmt, []string{"for"}))
	inIdx := strings.Index(raw, " in ")
	if inIdx == -1 {
		return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_FOR", "for header requires 'in' clause", "use syntax: for item in collection:")
	}
	iterVar := strings.TrimSpace(raw[:inIdx])
	colExpr := strings.TrimSpace(raw[inIdx+4:])
	if !IsIdentToken(iterVar) {
		return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_ITER", fmt.Sprintf("invalid iterator identifier '%s'", iterVar), "use a valid symbol name")
	}

	colSym := fmt.Sprintf("__aether_col_%d_%d", header.LineNo, len(chunk.Code))
	idxSym := fmt.Sprintf("__aether_idx_%d_%d", header.LineNo, len(chunk.Code))
	lenSym := fmt.Sprintf("__aether_len_%d_%d", header.LineNo, len(chunk.Code))

	colSymIdx := chunk.AddSymbol(colSym)
	idxSymIdx := chunk.AddSymbol(idxSym)
	lenSymIdx := chunk.AddSymbol(lenSym)
	itemSymIdx := chunk.AddSymbol(iterVar)
	lenFnIdx := chunk.AddSymbol("len_bmm")

	if err := bc.Kernel.CompileExprToChunk(colExpr, chunk, header.LineNo); err != nil {
		return err
	}
	chunk.Emit(OpStoreLocal, colSymIdx, 0, header.LineNo)

	zeroConst, err := chunk.AddConst(CycleDatum(0), header.LineNo)
	if err != nil {
		return err
	}
	chunk.Emit(OpLoadConst, zeroConst, 0, header.LineNo)
	chunk.Emit(OpStoreLocal, idxSymIdx, 0, header.LineNo)

	chunk.Emit(OpLoadVar, colSymIdx, 0, header.LineNo)
	chunk.Emit(OpCall, lenFnIdx, 1, header.LineNo)
	chunk.Emit(OpStoreLocal, lenSymIdx, 0, header.LineNo)

	condPC := len(chunk.Code)
	chunk.Emit(OpLoadVar, idxSymIdx, 0, header.LineNo)
	chunk.Emit(OpLoadVar, lenSymIdx, 0, header.LineNo)
	chunk.Emit(OpLt, 0, 0, header.LineNo)
	exitPC := chunk.Emit(OpJumpIfFalse, 0, 0, header.LineNo)

	chunk.Emit(OpLoadVar, colSymIdx, 0, header.LineNo)
	chunk.Emit(OpLoadVar, idxSymIdx, 0, header.LineNo)
	chunk.Emit(OpIndexGet, 0, 0, header.LineNo)
	chunk.Emit(OpStoreLocal, itemSymIdx, 0, header.LineNo)

	if err := bc.compileBlockSpans(body, chunk, header.LineNo); err != nil {
		return err
	}

	oneConst, err := chunk.AddConst(CycleDatum(1), header.LineNo)
	if err != nil {
		return err
	}
	chunk.Emit(OpLoadVar, idxSymIdx, 0, header.LineNo)
	chunk.Emit(OpLoadConst, oneConst, 0, header.LineNo)
	chunk.Emit(OpAdd, 0, 0, header.LineNo)
	chunk.Emit(OpStoreLocal, idxSymIdx, 0, header.LineNo)

	chunk.Emit(OpJump, condPC, 0, header.LineNo)
	donePC := len(chunk.Code)
	chunk.PatchJump(exitPC, donePC)
	return nil
}

func (bc *BytecodeCompiler) compileRoutineDecl(header CodeSpan, body []CodeSpan, chunk *CompiledChunk) error {
	raw := strings.TrimSpace(bc.stripKwHeader(header.Stmt, []string{"routine", "fn", "def"}))
	openP := strings.IndexByte(raw, '(')
	closeP := strings.LastIndexByte(raw, ')')
	if openP <= 0 || closeP <= openP {
		return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_FN", "invalid routine declaration syntax", "use syntax: fn compute(a, b):")
	}
	fnName := strings.TrimSpace(raw[:openP])
	if !IsIdentToken(fnName) {
		return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_FN_NAME", fmt.Sprintf("invalid routine name '%s'", fnName), "use a valid identifier")
	}
	paramRaw := strings.TrimSpace(raw[openP+1 : closeP])
	var params []string
	if paramRaw != "" {
		for _, p := range strings.Split(paramRaw, ",") {
			pClean := strings.TrimSpace(p)
			if !IsIdentToken(pClean) {
				return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_PARAM", fmt.Sprintf("invalid parameter '%s'", pClean), "use valid parameter identifiers")
			}
			params = append(params, pClean)
		}
	}

	skipBodyJump := chunk.Emit(OpJump, 0, 0, header.LineNo)
	entryPC := len(chunk.Code)

	if err := bc.compileBlockSpans(body, chunk, header.LineNo); err != nil {
		return err
	}

	if len(chunk.Code) == entryPC || chunk.Code[len(chunk.Code)-1].Op != OpReturn {
		chunk.Emit(OpLoadVoid, 0, 0, header.LineNo)
		chunk.Emit(OpReturn, 0, 0, header.LineNo)
	}

	afterRoutinePC := len(chunk.Code)
	chunk.PatchJump(skipBodyJump, afterRoutinePC)

	proto := RoutineProto{
		Name:   fnName,
		Params: params,
		Entry:  entryPC,
		Lines:  body,
	}
	bc.VM.Protos[fnName] = proto
	bc.Kernel.routines[fnName] = SubRoutine{
		Ident:  fnName,
		Params: params,
		Block:  nil,
	}
	return nil
}

func (bc *BytecodeCompiler) compileUnitDecl(header CodeSpan, body []CodeSpan, chunk *CompiledChunk) error {
	_ = chunk
	raw := strings.TrimSpace(bc.stripKwHeader(header.Stmt, []string{"unit", "struct"}))
	if !IsIdentToken(raw) {
		return bc.Kernel.fault(header.LineNo, "ERR_COMPILE_UNIT", fmt.Sprintf("invalid unit schema name '%s'", raw), "use syntax: unit NodeSpec:")
	}
	var fields []string
	for _, line := range body {
		fName := strings.TrimSpace(line.Stmt)
		if strings.HasPrefix(fName, "field ") {
			fName = strings.TrimSpace(strings.TrimPrefix(fName, "field "))
		}
		if IsIdentToken(fName) {
			fields = append(fields, fName)
		}
	}
	bc.VM.Units[raw] = UnitSchema{
		Name:   raw,
		Fields: fields,
	}
	return nil
}

func (bc *BytecodeCompiler) compileGuardBlock(header CodeSpan, tryBody []CodeSpan, rescueBody []CodeSpan, chunk *CompiledChunk) error {
	_ = header
	if err := bc.compileBlockSpans(tryBody, chunk, header.LineNo); err != nil {
		if len(rescueBody) == 0 {
			return err
		}
		errSym := chunk.AddSymbol("last_fault")
		errConst, cErr := chunk.AddConst(GlyphDatum(err.Error()), header.LineNo)
		if cErr != nil {
			return cErr
		}
		chunk.Emit(OpLoadConst, errConst, 0, header.LineNo)
		chunk.Emit(OpStoreLocal, errSym, 0, header.LineNo)
		return bc.compileBlockSpans(rescueBody, chunk, header.LineNo)
	}
	return nil
}

func (bc *BytecodeCompiler) compileStmtLine(span CodeSpan, chunk *CompiledChunk) error {
	stmt := strings.TrimSpace(span.Stmt)
	line := span.LineNo
	if stmt == "" {
		return nil
	}

	for _, retPfx := range []string{"yield ", "ret ", "return "} {
		if strings.HasPrefix(stmt, retPfx) {
			expr := strings.TrimSpace(stmt[len(retPfx):])
			if expr == "" {
				chunk.Emit(OpLoadVoid, 0, 0, line)
			} else if err := bc.Kernel.CompileExprToChunk(expr, chunk, line); err != nil {
				return err
			}
			chunk.Emit(OpReturn, 0, 0, line)
			return nil
		}
	}
	if stmt == "yield" || stmt == "ret" || stmt == "return" {
		chunk.Emit(OpLoadVoid, 0, 0, line)
		chunk.Emit(OpReturn, 0, 0, line)
		return nil
	}

	if strings.HasPrefix(stmt, "assert ") {
		expr := strings.TrimSpace(strings.TrimPrefix(stmt, "assert "))
		if commaIdx := findTopLevelCommaPos(expr); commaIdx != -1 {
			expr = strings.TrimSpace(expr[:commaIdx])
		}
		if err := bc.Kernel.CompileExprToChunk(expr, chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpAssert, 0, 0, line)
		return nil
	}

	if strings.HasPrefix(stmt, "raise_fault ") {
		expr := strings.TrimSpace(strings.TrimPrefix(stmt, "raise_fault "))
		if err := bc.Kernel.CompileExprToChunk(expr, chunk, line); err != nil {
			return err
		}
		chunk.Emit(OpRaise, 0, 0, line)
		return nil
	}

	for _, declPfx := range []string{"let ", "bind ", "var "} {
		if strings.HasPrefix(stmt, declPfx) {
			rest := strings.TrimSpace(stmt[len(declPfx):])
			eqPos := findAssignCharPos(rest)
			if eqPos == -1 {
				return bc.Kernel.fault(line, "ERR_COMPILE_BIND", "declaration requires '='", "use syntax: let symbol = expr")
			}
			sym := strings.TrimSpace(rest[:eqPos])
			rhs := strings.TrimSpace(rest[eqPos+1:])
			if !IsIdentToken(sym) {
				return bc.Kernel.fault(line, "ERR_COMPILE_SYM", fmt.Sprintf("invalid symbol identifier '%s'", sym), "use valid identifier")
			}
			if err := bc.Kernel.CompileExprToChunk(rhs, chunk, line); err != nil {
				return err
			}
			symIdx := chunk.AddSymbol(sym)
			chunk.Emit(OpStoreLocal, symIdx, 0, line)
			return nil
		}
	}

	if op, lhs, rhs, ok := splitCompoundStmt(stmt); ok {
		lhs = strings.TrimSpace(lhs)
		rhs = strings.TrimSpace(rhs)
		if !IsIdentToken(lhs) {
			return bc.Kernel.fault(line, "ERR_COMPILE_COMPOUND", "compound assignment target must be a symbol", "use syntax: x += 1")
		}
		symIdx := chunk.AddSymbol(lhs)
		chunk.Emit(OpLoadVar, symIdx, 0, line)
		if err := bc.Kernel.CompileExprToChunk(rhs, chunk, line); err != nil {
			return err
		}
		switch op {
		case "+=":
			chunk.Emit(OpAdd, 0, 0, line)
		case "-=":
			chunk.Emit(OpSub, 0, 0, line)
		case "*=":
			chunk.Emit(OpMul, 0, 0, line)
		case "/=":
			chunk.Emit(OpDiv, 0, 0, line)
		case "%=":
			chunk.Emit(OpMod, 0, 0, line)
		}
		chunk.Emit(OpStoreVar, symIdx, 0, line)
		return nil
	}

	if eqPos := findAssignCharPos(stmt); eqPos != -1 {
		lhs := strings.TrimSpace(stmt[:eqPos])
		rhs := strings.TrimSpace(stmt[eqPos+1:])

		if IsIdentToken(lhs) {
			if err := bc.Kernel.CompileExprToChunk(rhs, chunk, line); err != nil {
				return err
			}
			symIdx := chunk.AddSymbol(lhs)
			chunk.Emit(OpStoreVar, symIdx, 0, line)
			return nil
		}

		if dotPos := strings.LastIndexByte(lhs, '.'); dotPos > 0 {
			objSym := strings.TrimSpace(lhs[:dotPos])
			fieldSym := strings.TrimSpace(lhs[dotPos+1:])
			if IsIdentToken(objSym) && IsIdentToken(fieldSym) {
				if err := bc.Kernel.CompileExprToChunk(rhs, chunk, line); err != nil {
					return err
				}
				objIdx := chunk.AddSymbol(objSym)
				fldIdx := chunk.AddSymbol(fieldSym)
				chunk.Emit(OpSetField, objIdx, fldIdx, line)
				return nil
			}
		}

		if len(lhs) >= 2 && lhs[len(lhs)-1] == ']' {
			openIdx := findPairOpen(lhs, len(lhs)-1, '[', ']')
			if openIdx > 0 {
				targetSym := strings.TrimSpace(lhs[:openIdx])
				idxExpr := strings.TrimSpace(lhs[openIdx+1 : len(lhs)-1])
				if IsIdentToken(targetSym) && idxExpr != "" {
					symIdx := chunk.AddSymbol(targetSym)
					if err := bc.Kernel.CompileExprToChunk(idxExpr, chunk, line); err != nil {
						return err
					}
					if err := bc.Kernel.CompileExprToChunk(rhs, chunk, line); err != nil {
						return err
					}
					chunk.Emit(OpIndexSet, symIdx, 0, line)
					return nil
				}
			}
		}
	}

	if err := bc.Kernel.CompileExprToChunk(stmt, chunk, line); err != nil {
		return err
	}
	chunk.Emit(OpPop, 0, 0, line)
	return nil
}

func (bc *BytecodeCompiler) hasKw(stmt string, kw string) bool {
	if !strings.HasPrefix(stmt, kw) {
		return false
	}
	if len(stmt) == len(kw) {
		return true
	}
	nxt := stmt[len(kw)]
	return nxt == ' ' || nxt == '\t' || nxt == '(' || nxt == ':'
}

func (bc *BytecodeCompiler) stripKwHeader(stmt string, kws []string) string {
	t := strings.TrimSpace(stmt)
	for _, kw := range kws {
		if bc.hasKw(t, kw) {
			rest := strings.TrimSpace(t[len(kw):])
			if strings.HasSuffix(rest, ":") {
				rest = strings.TrimSpace(rest[:len(rest)-1])
			}
			return rest
		}
	}
	if strings.HasSuffix(t, ":") {
		return strings.TrimSpace(t[:len(t)-1])
	}
	return t
}

func (bc *BytecodeCompiler) isElseStart(stmt string) bool {
	t := strings.TrimSpace(stmt)
	return t == "else" || t == "else:"
}

func (bc *BytecodeCompiler) isGuardStart(stmt string) bool {
	t := strings.TrimSpace(stmt)
	return t == "guard" || t == "guard:" || t == "try" || t == "try:"
}

func (bc *BytecodeCompiler) collectRescueBlock(spans []CodeSpan, startIdx int, baseIndent int) ([]CodeSpan, int) {
	idx := startIdx
	if idx < len(spans) && spans[idx].Indent == baseIndent {
		t := strings.TrimSpace(spans[idx].Stmt)
		if t == "rescue" || t == "rescue:" || t == "catch" || t == "catch:" {
			end := idx + 1
			for end < len(spans) && spans[end].Indent > baseIndent {
				end++
			}
			return spans[idx+1 : end], end
		}
	}
	return nil, idx
}

func (bc *BytecodeCompiler) isBlockStart(stmt string) bool {
	t := strings.TrimSpace(stmt)
	if strings.HasSuffix(t, ":") {
		return true
	}
	for _, kw := range []string{"when ", "if ", "loop ", "while ", "for ", "fn ", "def ", "routine ", "unit ", "struct "} {
		if strings.HasPrefix(t, kw) {
			return true
		}
	}
	return bc.isGuardStart(t)
}

func findTopLevelCommaPos(s string) int {
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			paren++
		case '[':
			bracket++
		case '{':
			brace++
		case ')':
			if paren > 0 {
				paren--
			}
		case ']':
			if bracket > 0 {
				bracket--
			}
		case '}':
			if brace > 0 {
				brace--
			}
		case ',':
			if paren == 0 && bracket == 0 && brace == 0 {
				return i
			}
		}
	}
	return -1
}

func splitCompoundStmt(s string) (string, string, string, bool) {
	ops := []string{"+=", "-=", "*=", "/=", "%="}
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			paren++
		case '[':
			bracket++
		case '{':
			brace++
		case ')':
			if paren > 0 {
				paren--
			}
		case ']':
			if bracket > 0 {
				bracket--
			}
		case '}':
			if brace > 0 {
				brace--
			}
		default:
			if paren == 0 && bracket == 0 && brace == 0 {
				for _, op := range ops {
					if i+len(op) <= len(s) && s[i:i+len(op)] == op {
						return op, s[:i], s[i+len(op):], true
					}
				}
			}
		}
	}
	return "", "", "", false
}

func findAssignCharPos(s string) int {
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			paren++
		case '[':
			bracket++
		case '{':
			brace++
		case ')':
			if paren > 0 {
				paren--
			}
		case ']':
			if bracket > 0 {
				bracket--
			}
		case '}':
			if brace > 0 {
				brace--
			}
		case '=':
			if paren == 0 && bracket == 0 && brace == 0 {
				prev := byte(0)
				if i > 0 {
					prev = s[i-1]
				}
				next := byte(0)
				if i+1 < len(s) {
					next = s[i+1]
				}
				if prev != '=' && prev != '!' && prev != '<' && prev != '>' && prev != '+' && prev != '-' && prev != '*' && prev != '/' && prev != '%' && next != '=' {
					return i
				}
			}
		}
	}
	return -1
}

type AetherProgram struct {
	SourcePath string
	RawSource  string
	Spans      []CodeSpan
	Chunk      *CompiledChunk
	Checksum   uint64
}

func ComputeSourceHash(data string) uint64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(data); i++ {
		h ^= uint64(data[i])
		h *= 1099511628211
	}
	return h
}

func StripInlineComment(line string) string {
	quote := byte(0)
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(line, i) {
				quote = 0
			}
			continue
		}
		if ch == '"' || ch == '\'' {
			quote = ch
			continue
		}
		if ch == '#' || (ch == '/' && i+1 < len(line) && line[i+1] == '/') {
			return line[:i]
		}
	}
	return line
}

func CalcLeadingIndent(s string) int {
	width := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			width++
		} else if s[i] == '\t' {
			width += 4
		} else {
			break
		}
	}
	return width
}

func ParseAetherSpans(raw string) ([]CodeSpan, error) {
	norm := strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	spans := make([]CodeSpan, 0, len(lines))
	indentStack := []int{0}

	for idx, rawLine := range lines {
		lineNo := idx + 1
		noComment := StripInlineComment(rawLine)
		stmt := strings.TrimSpace(noComment)
		if stmt == "" {
			continue
		}

		indent := CalcLeadingIndent(noComment)
		topIndent := indentStack[len(indentStack)-1]

		if indent > topIndent {
			indentStack = append(indentStack, indent)
		} else if indent < topIndent {
			found := false
			for len(indentStack) > 1 {
				indentStack = indentStack[:len(indentStack)-1]
				if indentStack[len(indentStack)-1] == indent {
					found = true
					break
				}
			}
			if !found {
				return nil, AetherDiag{
					Line:   lineNo,
					Stage:  "LEXER",
					Code:   "ERR_INDENT_ALIGN",
					Detail: fmt.Sprintf("indent column %d does not match any outer block", indent),
					Hint:   "align line indentation with an active enclosing block",
					Chain:  []string{"SPAN_PARSER"},
				}
			}
		}

		spans = append(spans, CodeSpan{
			LineNo: lineNo,
			Indent: indent,
			Raw:    rawLine,
			Stmt:   stmt,
		})
	}

	return spans, nil
}

func OptimizeBytecodeChunk(chunk *CompiledChunk) *CompiledChunk {
	if chunk == nil || len(chunk.Code) == 0 {
		return chunk
	}
	opt := NewCompiledChunk()
	opt.Constants = make([]Datum, len(chunk.Constants))
	for i, c := range chunk.Constants {
		opt.Constants[i] = c.DeepCopy()
	}
	opt.Symbols = make([]string, len(chunk.Symbols))
	copy(opt.Symbols, chunk.Symbols)

	idxMap := make([]int, len(chunk.Code))
	for i := range idxMap {
		idxMap[i] = -1
	}

	for oldPC, ins := range chunk.Code {
		if ins.Op == OpJump && ins.ArgA == oldPC+1 {
			continue
		}
		idxMap[oldPC] = len(opt.Code)
		opt.Code = append(opt.Code, ins)
	}

	remap := func(target int) int {
		if target >= 0 && target < len(idxMap) {
			for p := target; p < len(idxMap); p++ {
				if idxMap[p] != -1 {
					return idxMap[p]
				}
			}
		}
		return len(opt.Code)
	}

	for i := range opt.Code {
		switch opt.Code[i].Op {
		case OpJump, OpJumpIfFalse, OpJumpIfTrue:
			opt.Code[i].ArgA = remap(opt.Code[i].ArgA)
		}
	}

	opt.EntryPC = remap(chunk.EntryPC)
	return opt
}

type AetherEngine struct {
	Kernel   *AetherKernel
	VM       *AetherVM
	Compiler *BytecodeCompiler
	LastProg *AetherProgram
}

func NewAetherEngine() *AetherEngine {
	k := NewAetherKernel()
	vm := NewAetherVM(k)
	comp := NewBytecodeCompiler(k, vm)
	return &AetherEngine{
		Kernel:   k,
		VM:       vm,
		Compiler: comp,
		LastProg: nil,
	}
}

func (eng *AetherEngine) BuildProgram(source string, path string) (*AetherProgram, error) {
	eng.Kernel.ResetCounter()
	spans, err := ParseAetherSpans(source)
	if err != nil {
		return nil, err
	}
	rawChunk, err := eng.Compiler.CompileProgram(spans)
	if err != nil {
		return nil, err
	}
	optChunk := OptimizeBytecodeChunk(rawChunk)
	prog := &AetherProgram{
		SourcePath: path,
		RawSource:  source,
		Spans:      spans,
		Chunk:      optChunk,
		Checksum:   ComputeSourceHash(source),
	}
	eng.LastProg = prog
	return prog, nil
}

func (eng *AetherEngine) EvalCode(source string) (Datum, error) {
	prog, err := eng.BuildProgram(source, "<memory>")
	if err != nil {
		return VoidDatum(), err
	}
	return eng.VM.ExecChunk(prog.Chunk)
}

func (eng *AetherEngine) EvalScriptFile(path string) (Datum, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return VoidDatum(), AetherDiag{
			Line:   1,
			Stage:  "LOADER",
			Code:   "ERR_FILE_IO",
			Detail: fmt.Sprintf("cannot read file '%s': %v", path, err),
			Hint:   "check file path and permissions",
			Chain:  []string{"FILE_RUNNER"},
		}
	}
	prog, err := eng.BuildProgram(string(data), path)
	if err != nil {
		return VoidDatum(), err
	}
	return eng.VM.ExecChunk(prog.Chunk)
}

func (eng *AetherEngine) ExpandDotAccessInExpr(expr string, line int) (string, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return trimmed, nil
	}
	if dotIdx := findTopLevelDot(trimmed); dotIdx > 0 && dotIdx < len(trimmed)-1 {
		leftPart := strings.TrimSpace(trimmed[:dotIdx])
		rightPart := strings.TrimSpace(trimmed[dotIdx+1:])
		if IsIdentToken(leftPart) && IsIdentToken(rightPart) {
			return fmt.Sprintf("%s[\"%s\"]", leftPart, rightPart), nil
		}
	}
	_ = line
	return trimmed, nil
}

func findTopLevelDot(s string) int {
	paren, bracket, brace := 0, 0, 0
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			paren++
		case '[':
			bracket++
		case '{':
			brace++
		case ')':
			if paren > 0 {
				paren--
			}
		case ']':
			if bracket > 0 {
				bracket--
			}
		case '}':
			if brace > 0 {
				brace--
			}
		case '.':
			if paren == 0 && bracket == 0 && brace == 0 {
				if i > 0 && i+1 < len(s) && unicode.IsDigit(rune(s[i-1])) && unicode.IsDigit(rune(s[i+1])) {
					continue
				}
				return i
			}
		}
	}
	return -1
}

func (eng *AetherEngine) DumpLastBytecode() string {
	if eng.LastProg == nil || eng.LastProg.Chunk == nil {
		return fmt.Sprintf("[%s] No compiled bytecode chunk available.", AetherName)
	}
	return eng.LastProg.Chunk.Disassemble()
}

func (eng *AetherEngine) ResetEngine() {
	eng.Kernel.PurgeAll()
	eng.VM.Stack = eng.VM.Stack[:0]
	eng.VM.CallInfo = eng.VM.CallInfo[:0]
	eng.VM.Units = make(map[string]UnitSchema)
	eng.VM.Protos = make(map[string]RoutineProto)
	eng.LastProg = nil
}

type FaultExplanation struct {
	Code         string
	Line         int
	Subject      string
	WhyBroken    string
	RootCause    string
	LineContext  string
	StepByStep   []string
	SampleFix    string
}

func (fe FaultExplanation) FormatReport() string {
	var b strings.Builder
	b.WriteString("========================================================\n")
	b.WriteString("            AETHER-BMM DEEP FAULT EXPLANATION           \n")
	b.WriteString("========================================================\n")
	b.WriteString(fmt.Sprintf("Fault Code   : %s\n", fe.Code))
	b.WriteString(fmt.Sprintf("Line Number  : %d\n", fe.Line))
	if fe.Subject != "" {
		b.WriteString(fmt.Sprintf("Target       : %s\n", fe.Subject))
	}
	b.WriteString(fmt.Sprintf("Why Invalid  : %s\n", fe.WhyBroken))
	b.WriteString(fmt.Sprintf("Root Cause   : %s\n", fe.RootCause))
	if fe.LineContext != "" {
		b.WriteString(fmt.Sprintf("Line Snippet : %s\n", fe.LineContext))
	}
	if len(fe.StepByStep) > 0 {
		b.WriteString("Diagnosis Steps:\n")
		for idx, step := range fe.StepByStep {
			b.WriteString(fmt.Sprintf("  %d. %s\n", idx+1, step))
		}
	}
	if fe.SampleFix != "" {
		b.WriteString(fmt.Sprintf("Suggested Fix: %s\n", fe.SampleFix))
	}
	b.WriteString("========================================================")
	return b.String()
}

type FaultAuditor struct {
	SourceLines []string
	KnownSyms   map[string]DatumKind
	KnownFuncs  []string
}

func NewFaultAuditor(source string) *FaultAuditor {
	norm := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	return &FaultAuditor{
		SourceLines: lines,
		KnownSyms:   make(map[string]DatumKind),
		KnownFuncs: []string{
			"emit_bmm", "kind_bmm", "len_bmm", "cycle_bmm", "glyph_bmm", "state_bmm",
			"seq_bmm", "push_bmm", "pop_bmm", "slice_bmm", "keys_bmm", "has_bmm",
			"pulse_bmm", "read_pulse_bmm", "vault_put_bmm", "vault_get_bmm", "vault_del_bmm",
			"abs_bmm", "clamp_bmm", "min_bmm", "max_bmm", "upper_bmm", "lower_bmm",
			"trim_bmm", "split_bmm", "join_bmm",
		},
	}
}

func (fa *FaultAuditor) extractSnippet(line int) string {
	if line < 1 || line > len(fa.SourceLines) {
		return ""
	}
	start := line - 2
	if start < 0 {
		start = 0
	}
	end := line + 1
	if end > len(fa.SourceLines) {
		end = len(fa.SourceLines)
	}
	var out []string
	for i := start; i < end; i++ {
		marker := "  "
		if i+1 == line {
			marker = ">>"
		}
		out = append(out, fmt.Sprintf("%s %3d | %s", marker, i+1, fa.SourceLines[i]))
	}
	return strings.Join(out, "\n")
}

func (fa *FaultAuditor) InspectFault(err error, fallbackLine int) FaultExplanation {
	line := fallbackLine
	if line < 1 {
		line = 1
	}
	code := "ERR_RUNTIME_INSPECT"
	detail := ""
	hint := ""

	if diag, ok := err.(AetherDiag); ok {
		if diag.Line > 0 {
			line = diag.Line
		}
		code = diag.Code
		detail = diag.Detail
		hint = diag.Hint
	} else if fRep, ok := err.(FaultReport); ok {
		if fRep.Line > 0 {
			line = fRep.Line
		}
		code = fRep.Code
		detail = fRep.Cause
		hint = fRep.Remedy
	} else if err != nil {
		detail = err.Error()
	}

	rawLine := ""
	if line >= 1 && line <= len(fa.SourceLines) {
		rawLine = strings.TrimSpace(fa.SourceLines[line-1])
	}

	exp := FaultExplanation{
		Code:        code,
		Line:        line,
		Subject:     rawLine,
		WhyBroken:   detail,
		RootCause:   fa.classifyRootCause(code, detail, rawLine),
		LineContext: fa.extractSnippet(line),
		StepByStep:  fa.buildDiagnosisSteps(code, detail, rawLine),
		SampleFix:   fa.buildSampleFix(code, detail, rawLine, hint),
	}
	return exp
}

func (fa *FaultAuditor) classifyRootCause(code string, detail string, rawLine string) string {
	upperCode := strings.ToUpper(code)
	lowerDetail := strings.ToLower(detail)

	switch {
	case strings.Contains(upperCode, "UNBOUND") || strings.Contains(lowerDetail, "not bound"):
		return "A symbol was referenced during evaluation or bytecode load before being bound in any visible scope frame."
	case strings.Contains(upperCode, "ZERO_DIV") || strings.Contains(upperCode, "ZERO_MOD") || strings.Contains(lowerDetail, "division by zero"):
		return "An arithmetic division or modulo operation evaluated a right-hand divisor equal to cycle 0."
	case strings.Contains(upperCode, "TYPE") || strings.Contains(upperCode, "MISMATCH") || strings.Contains(lowerDetail, "requires"):
		return "An operator or built-in routine received a datum kind that does not match its expected type contract."
	case strings.Contains(upperCode, "BOUNDS") || strings.Contains(upperCode, "RANGE") || strings.Contains(lowerDetail, "out of"):
		return "An index expression evaluated to a position outside the valid length range of the target container."
	case strings.Contains(upperCode, "INDENT") || strings.Contains(upperCode, "SYNTAX") || strings.Contains(upperCode, "TOKEN"):
		return "Source formatting, block indentation, punctuation, or operator precedence did not form a valid construct."
	case strings.Contains(upperCode, "ARITY") || strings.Contains(lowerDetail, "arguments"):
		return "The number of arguments supplied to a routine call does not match its expected parameter count."
	default:
		if rawLine != "" {
			return fmt.Sprintf("Execution halted while evaluating statement '%s': %s", rawLine, detail)
		}
		return fmt.Sprintf("Execution halted with diagnostic '%s': %s", code, detail)
	}
}

func (fa *FaultAuditor) buildDiagnosisSteps(code string, detail string, rawLine string) []string {
	steps := []string{
		fmt.Sprintf("Locate line %d in the source script and examine the active expression or statement.", fa.clampLine(code)),
		fmt.Sprintf("Inspect diagnostic code [%s] and message: %s", code, detail),
	}
	if rawLine != "" {
		steps = append(steps, fmt.Sprintf("Verify every variable, literal, bracket pair, and operator inside: %s", rawLine))
		if strings.Contains(rawLine, "/") || strings.Contains(rawLine, "%") {
			steps = append(steps, "Check any variable or expression on the right side of '/' or '%' to guarantee it cannot be 0 at runtime.")
		}
		if strings.Contains(rawLine, "[") && strings.Contains(rawLine, "]") {
			steps = append(steps, "Check index bounds so 0 <= index < len_bmm(container) before reading or writing.")
		}
	}
	steps = append(steps, "Run with ':status' or inspect bytecode disassembly to trace symbol bindings and stack state.")
	return steps
}

func (fa *FaultAuditor) clampLine(code string) int {
	_ = code
	if len(fa.SourceLines) == 0 {
		return 1
	}
	return len(fa.SourceLines)
}

func (fa *FaultAuditor) buildSampleFix(code string, detail string, rawLine string, hint string) string {
	upperCode := strings.ToUpper(code)
	if hint != "" {
		return fmt.Sprintf("%s (Context: %s)", hint, rawLine)
	}
	if strings.Contains(upperCode, "UNBOUND") {
		return "Declare and initialize the symbol first, e.g.: let x = 0"
	}
	if strings.Contains(upperCode, "ZERO_DIV") || strings.Contains(upperCode, "ZERO_MOD") {
		return "Guard the divisor with 'when denom != 0:' before dividing."
	}
	if strings.Contains(upperCode, "BOUNDS") {
		return "Check 'when idx >= 0 && idx < len_bmm(items):' before indexing items[idx]."
	}
	if strings.Contains(upperCode, "TYPE") || strings.Contains(upperCode, "MISMATCH") {
		return fmt.Sprintf("Convert operands explicitly using cycle_bmm(...), glyph_bmm(...), or state_bmm(...) (%s).", detail)
	}
	return "Review the expression around the reported line and ensure all types and arguments match."
}

func AuditSourceStructure(source string) []string {
	norm := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	var findings []string

	for idx, raw := range lines {
		lineNo := idx + 1
		clean := strings.TrimSpace(StripInlineComment(raw))
		if clean == "" {
			continue
		}
		if strings.Count(clean, "\"")%2 != 0 {
			findings = append(findings, fmt.Sprintf("Line %d: unclosed double quote in '%s'", lineNo, clean))
		}
		if !balancedDelimiters(clean) {
			findings = append(findings, fmt.Sprintf("Line %d: unbalanced parentheses, brackets, or braces in '%s'", lineNo, clean))
		}
		if strings.HasSuffix(clean, ":") && len(strings.TrimSpace(strings.TrimSuffix(clean, ":"))) == 0 {
			findings = append(findings, fmt.Sprintf("Line %d: empty block header ':'", lineNo))
		}
	}

	if len(findings) == 0 {
		findings = append(findings, "Static structural audit passed with zero delimiter or quote anomalies.")
	}
	return findings
}

func balancedDelimiters(s string) bool {
	p, b, br := 0, 0, 0
	quote := byte(0)
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if quote != 0 {
			if ch == quote && !isEscapedChar(s, i) {
				quote = 0
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case '(':
			p++
		case ')':
			p--
			if p < 0 {
				return false
			}
		case '[':
			b++
		case ']':
			b--
			if b < 0 {
				return false
			}
		case '{':
			br++
		case '}':
			br--
			if br < 0 {
				return false
			}
		}
	}
	return quote == 0 && p == 0 && b == 0 && br == 0
}

func (eng *AetherEngine) EvalWithExplanation(source string, path string) (Datum, string, error) {
	prog, err := eng.BuildProgram(source, path)
	if err != nil {
		auditor := NewFaultAuditor(source)
		exp := auditor.InspectFault(err, 1)
		return VoidDatum(), exp.FormatReport(), err
	}
	val, err := eng.VM.ExecChunk(prog.Chunk)
	if err != nil {
		auditor := NewFaultAuditor(source)
		exp := auditor.InspectFault(err, 1)
		return VoidDatum(), exp.FormatReport(), err
	}
	return val, "", nil
}

type AetherStats struct {
	Count int64
	Sum   int64
	Min   int64
	Max   int64
	Avg   int64
}

func ComputeCycleStats(items []Datum) (AetherStats, error) {
	if len(items) == 0 {
		return AetherStats{}, fmt.Errorf("cannot compute cycle stats over an empty vector")
	}
	var sum int64
	var minV, maxV int64
	for i, it := range items {
		if it.Kind != DatumCycle {
			return AetherStats{}, fmt.Errorf("element at index %d is '%s', expected cycle", i, it.TypeName())
		}
		if i == 0 {
			minV = it.Cycle
			maxV = it.Cycle
		} else {
			if it.Cycle < minV {
				minV = it.Cycle
			}
			if it.Cycle > maxV {
				maxV = it.Cycle
			}
		}
		sum += it.Cycle
	}
	cnt := int64(len(items))
	return AetherStats{
		Count: cnt,
		Sum:   sum,
		Min:   minV,
		Max:   maxV,
		Avg:   sum / cnt,
	}, nil
}

func ReverseVecDatum(items []Datum) Datum {
	out := make([]Datum, len(items))
	for i := 0; i < len(items); i++ {
		out[i] = items[len(items)-1-i].DeepCopy()
	}
	return VecDatum(out)
}

func ContainsDatumInVec(items []Datum, target Datum) bool {
	for _, it := range items {
		if it.Equals(target) {
			return true
		}
	}
	return false
}

func IndexOfDatumInVec(items []Datum, target Datum) int64 {
	for i, it := range items {
		if it.Equals(target) {
			return int64(i)
		}
	}
	return -1
}

func MergeDictDatum(base map[string]Datum, patch map[string]Datum) Datum {
	out := make(map[string]Datum, len(base)+len(patch))
	for k, v := range base {
		out[k] = v.DeepCopy()
	}
	for k, v := range patch {
		out[k] = v.DeepCopy()
	}
	return DictDatum(out)
}

func PowCycleDatum(base int64, exp int64, line int, kernel *AetherKernel) (Datum, error) {
	if exp < 0 {
		return VoidDatum(), kernel.fault(line, "ERR_POW_NEGATIVE", "negative exponent is not supported for integer cycle", "use non-negative cycle exponent >= 0")
	}
	var res int64 = 1
	b := base
	e := exp
	for e > 0 {
		if err := kernel.AdvanceCycle(line); err != nil {
			return VoidDatum(), err
		}
		if e&1 == 1 {
			res *= b
		}
		e >>= 1
		if e > 0 {
			b *= b
		}
	}
	return CycleDatum(res), nil
}

func ModPowCycleDatum(base int64, exp int64, mod int64, line int, kernel *AetherKernel) (Datum, error) {
	if mod <= 0 {
		return VoidDatum(), kernel.fault(line, "ERR_MODPOW_MOD", "modulus must be positive cycle > 0", "supply positive modulus")
	}
	if exp < 0 {
		return VoidDatum(), kernel.fault(line, "ERR_MODPOW_EXP", "exponent must be non-negative cycle >= 0", "supply exponent >= 0")
	}
	var res int64 = 1 % mod
	b := ((base % mod) + mod) % mod
	e := exp
	for e > 0 {
		if err := kernel.AdvanceCycle(line); err != nil {
			return VoidDatum(), err
		}
		if e&1 == 1 {
			res = (res * b) % mod
		}
		e >>= 1
		if e > 0 {
			b = (b * b) % mod
		}
	}
	return CycleDatum(res), nil
}

func GcdCycleDatum(a int64, b int64) int64 {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func LcmCycleDatum(a int64, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	g := GcdCycleDatum(a, b)
	if g == 0 {
		return 0
	}
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	return (a / g) * b
}

func (vm *AetherVM) EvalExtendedBuiltin(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "pow_bmm":
		if len(args) != 2 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_POW_ARGS_BMM", "pow_bmm expects (base_cycle, exp_cycle)", "pass two cycle integers")
		}
		v, err := PowCycleDatum(args[0].Cycle, args[1].Cycle, line, vm.Kernel)
		return true, v, err

	case "modpow_bmm":
		if len(args) != 3 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle || args[2].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MODPOW_ARGS_BMM", "modpow_bmm expects (base, exp, mod) as cycle", "pass three cycle integers")
		}
		v, err := ModPowCycleDatum(args[0].Cycle, args[1].Cycle, args[2].Cycle, line, vm.Kernel)
		return true, v, err

	case "gcd_bmm":
		if len(args) != 2 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_GCD_ARGS_BMM", "gcd_bmm expects two cycle integers", "pass two cycle values")
		}
		return true, CycleDatum(GcdCycleDatum(args[0].Cycle, args[1].Cycle)), nil

	case "lcm_bmm":
		if len(args) != 2 || args[0].Kind != DatumCycle || args[1].Kind != DatumCycle {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_LCM_ARGS_BMM", "lcm_bmm expects two cycle integers", "pass two cycle values")
		}
		return true, CycleDatum(LcmCycleDatum(args[0].Cycle, args[1].Cycle)), nil

	case "rev_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_REV_ARGS_BMM", "rev_bmm expects a single vec", "pass a vec datum")
		}
		return true, ReverseVecDatum(args[0].Vec), nil

	case "has_item_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_HAS_ITEM_BMM", "has_item_bmm expects (vec, target_datum)", "pass a vec and a target datum")
		}
		return true, FlagDatum(ContainsDatumInVec(args[0].Vec, args[1])), nil

	case "index_of_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_IDX_OF_BMM", "index_of_bmm expects (vec, target_datum)", "pass a vec and a target datum")
		}
		return true, CycleDatum(IndexOfDatumInVec(args[0].Vec, args[1])), nil

	case "merge_dict_bmm":
		if len(args) != 2 || args[0].Kind != DatumDict || args[1].Kind != DatumDict {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_MERGE_DICT_BMM", "merge_dict_bmm expects two dict operands", "pass two dict datums")
		}
		return true, MergeDictDatum(args[0].Dict, args[1].Dict), nil

	case "stats_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_STATS_ARGS_BMM", "stats_bmm expects a non-empty cycle vec", "pass a vec of cycle numbers")
		}
		st, err := ComputeCycleStats(args[0].Vec)
		if err != nil {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_STATS_EVAL_BMM", err.Error(), "ensure vec is non-empty and contains cycle items")
		}
		return true, DictDatum(map[string]Datum{
			"count": CycleDatum(st.Count),
			"sum":   CycleDatum(st.Sum),
			"min":   CycleDatum(st.Min),
			"max":   CycleDatum(st.Max),
			"avg":   CycleDatum(st.Avg),
		}), nil

	case "contains_str_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_CONTAINS_STR_BMM", "contains_str_bmm expects (glyph, sub_glyph)", "pass two glyph strings")
		}
		return true, FlagDatum(strings.Contains(args[0].Glyph, args[1].Glyph)), nil

	case "replace_str_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_REPLACE_STR_BMM", "replace_str_bmm expects (src, old, new) as glyph", "pass three glyph strings")
		}
		return true, GlyphDatum(strings.ReplaceAll(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

	case "vault_keys_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_VAULT_KEYS_BMM", "vault_keys_bmm takes no arguments", "call vault_keys_bmm()")
		}
		out := make([]Datum, 0, len(vm.Kernel.vaults))
		for k := range vm.Kernel.vaults {
			out = append(out, GlyphDatum(k))
		}
		return true, VecDatum(out), nil

	case "pulse_bands_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), vm.Kernel.fault(line, "ERR_PULSE_BANDS_BMM", "pulse_bands_bmm takes no arguments", "call pulse_bands_bmm()")
		}
		out := make([]Datum, 0, len(vm.Kernel.carriers))
		for k := range vm.Kernel.carriers {
			out = append(out, GlyphDatum(k))
		}
		return true, VecDatum(out), nil
	}

	return false, VoidDatum(), nil
}

type AetherTokenKind int

const (
	TokEOF AetherTokenKind = iota
	TokNewline
	TokIndent
	TokDedent
	TokIdent
	TokCycle
	TokGlyph
	TokOp
	TokPunct
	TokComment
)

type AetherToken struct {
	Kind   AetherTokenKind
	Lexeme string
	Line   int
	Col    int
}

func TokenizeAetherSource(src string) []AetherToken {
	norm := strings.ReplaceAll(src, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	var tokens []AetherToken
	indentStack := []int{0}

	for lineIdx, raw := range lines {
		lineNo := lineIdx + 1
		stripped := StripInlineComment(raw)
		trimmed := strings.TrimSpace(stripped)
		if trimmed == "" {
			tokens = append(tokens, AetherToken{Kind: TokNewline, Lexeme: "\\n", Line: lineNo, Col: 1})
			continue
		}

		indent := CalcLeadingIndent(stripped)
		top := indentStack[len(indentStack)-1]
		if indent > top {
			indentStack = append(indentStack, indent)
			tokens = append(tokens, AetherToken{Kind: TokIndent, Lexeme: strconv.Itoa(indent), Line: lineNo, Col: 1})
		} else if indent < top {
			for len(indentStack) > 1 && indentStack[len(indentStack)-1] > indent {
				indentStack = indentStack[:len(indentStack)-1]
				tokens = append(tokens, AetherToken{Kind: TokDedent, Lexeme: strconv.Itoa(indent), Line: lineNo, Col: 1})
			}
		}

		col := indent + 1
		i := indent
		for i < len(stripped) {
			ch := stripped[i]
			if ch == ' ' || ch == '\t' {
				i++
				col++
				continue
			}

			if ch == '"' || ch == '\'' {
				start := i
				quote := ch
				i++
				for i < len(stripped) {
					if stripped[i] == quote && !isEscapedChar(stripped, i) {
						i++
						break
					}
					i++
				}
				tokens = append(tokens, AetherToken{Kind: TokGlyph, Lexeme: stripped[start:i], Line: lineNo, Col: col})
				col += i - start
				continue
			}

			if unicode.IsDigit(rune(ch)) {
				start := i
				for i < len(stripped) && unicode.IsDigit(rune(stripped[i])) {
					i++
				}
				tokens = append(tokens, AetherToken{Kind: TokCycle, Lexeme: stripped[start:i], Line: lineNo, Col: col})
				col += i - start
				continue
			}

			if unicode.IsLetter(rune(ch)) || ch == '_' {
				start := i
				for i < len(stripped) {
					r := rune(stripped[i])
					if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
						i++
					} else {
						break
					}
				}
				tokens = append(tokens, AetherToken{Kind: TokIdent, Lexeme: stripped[start:i], Line: lineNo, Col: col})
				col += i - start
				continue
			}

			if i+1 < len(stripped) {
				two := stripped[i : i+2]
				switch two {
				case "==", "!=", "<=", ">=", "&&", "||", "+=", "-=", "*=", "/=", "%=":
					tokens = append(tokens, AetherToken{Kind: TokOp, Lexeme: two, Line: lineNo, Col: col})
					i += 2
					col += 2
					continue
				}
			}

			switch ch {
			case '+', '-', '*', '/', '%', '=', '<', '>', '!':
				tokens = append(tokens, AetherToken{Kind: TokOp, Lexeme: string(ch), Line: lineNo, Col: col})
			default:
				tokens = append(tokens, AetherToken{Kind: TokPunct, Lexeme: string(ch), Line: lineNo, Col: col})
			}
			i++
			col++
		}

		tokens = append(tokens, AetherToken{Kind: TokNewline, Lexeme: "\\n", Line: lineNo, Col: col})
	}

	for len(indentStack) > 1 {
		indentStack = indentStack[:len(indentStack)-1]
		tokens = append(tokens, AetherToken{Kind: TokDedent, Lexeme: "0", Line: len(lines), Col: 1})
	}
	tokens = append(tokens, AetherToken{Kind: TokEOF, Lexeme: "EOF", Line: len(lines), Col: 1})
	return tokens
}

type AstStmtNode struct {
	Kind     string
	Line     int
	Indent   int
	Text     string
	Children []*AstStmtNode
}

func BuildAstTreeFromSpans(spans []CodeSpan) []*AstStmtNode {
	var build func(list []CodeSpan) []*AstStmtNode
	build = func(list []CodeSpan) []*AstStmtNode {
		var nodes []*AstStmtNode
		i := 0
		for i < len(list) {
			cur := list[i]
			base := cur.Indent
			j := i + 1
			for j < len(list) && list[j].Indent > base {
				j++
			}
			kind := "stmt"
			if strings.HasSuffix(cur.Stmt, ":") {
				kind = "block"
			}
			node := &AstStmtNode{
				Kind:     kind,
				Line:     cur.LineNo,
				Indent:   base,
				Text:     cur.Stmt,
				Children: build(list[i+1 : j]),
			}
			nodes = append(nodes, node)
			i = j
		}
		return nodes
	}
	return build(spans)
}

type RepairProposal struct {
	Line        int
	Original    string
	Repaired    string
	Explanation string
}

func ProposeSyntaxRepairs(source string) []RepairProposal {
	norm := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	var proposals []RepairProposal

	for idx, raw := range lines {
		lineNo := idx + 1
		trimmed := strings.TrimSpace(StripInlineComment(raw))
		if trimmed == "" {
			continue
		}

		if strings.Count(trimmed, "\"")%2 != 0 {
			fixed := raw + `"`
			proposals = append(proposals, RepairProposal{
				Line:        lineNo,
				Original:    raw,
				Repaired:    fixed,
				Explanation: "Unclosed double quote detected; closing quote added at line end.",
			})
		}

		openP := strings.Count(trimmed, "(")
		closeP := strings.Count(trimmed, ")")
		if openP > closeP {
			fixed := raw + strings.Repeat(")", openP-closeP)
			proposals = append(proposals, RepairProposal{
				Line:        lineNo,
				Original:    raw,
				Repaired:    fixed,
				Explanation: fmt.Sprintf("Missing %d closing parenthesis ')' balanced at line end.", openP-closeP),
			})
		}

		openB := strings.Count(trimmed, "[")
		closeB := strings.Count(trimmed, "]")
		if openB > closeB {
			fixed := raw + strings.Repeat("]", openB-closeB)
			proposals = append(proposals, RepairProposal{
				Line:        lineNo,
				Original:    raw,
				Repaired:    fixed,
				Explanation: fmt.Sprintf("Missing %d closing bracket ']' balanced at line end.", openB-closeB),
			})
		}
	}

	return proposals
}

func ApplyExplicitRepairs(source string) (string, []RepairProposal) {
	props := ProposeSyntaxRepairs(source)
	if len(props) == 0 {
		return source, nil
	}
	norm := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	for _, p := range props {
		if p.Line >= 1 && p.Line <= len(lines) {
			lines[p.Line-1] = p.Repaired
		}
	}
	return strings.Join(lines, "\n"), props
}

func (eng *AetherEngine) RunWithFullPipeline(source string, path string, autoRepair bool) (Datum, string, error) {
	activeSrc := source
	var reportParts []string

	if autoRepair {
		repaired, applied := ApplyExplicitRepairs(activeSrc)
		if len(applied) > 0 {
			activeSrc = repaired
			for _, ap := range applied {
				reportParts = append(reportParts, fmt.Sprintf("[REPAIR @ line %d] %s -> %s (%s)", ap.Line, strings.TrimSpace(ap.Original), strings.TrimSpace(ap.Repaired), ap.Explanation))
			}
		}
	}

	val, diagReport, err := eng.EvalWithExplanation(activeSrc, path)
	if diagReport != "" {
		reportParts = append(reportParts, diagReport)
	}
	return val, strings.Join(reportParts, "\n"), err
}

func (eng *AetherEngine) InspectTokenStream(source string) string {
	toks := TokenizeAetherSource(source)
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== AETHER LEXER TOKEN STREAM (%d tokens) ===\n", len(toks)))
	for i, t := range toks {
		if t.Kind == TokEOF {
			b.WriteString(fmt.Sprintf("%04d [L%-3d C%-3d] %-10s %s\n", i, t.Line, t.Col, "EOF", t.Lexeme))
			break
		}
		var kName string
		switch t.Kind {
		case TokNewline:
			kName = "NEWLINE"
		case TokIndent:
			kName = "INDENT"
		case TokDedent:
			kName = "DEDENT"
		case TokIdent:
			kName = "IDENT"
		case TokCycle:
			kName = "CYCLE"
		case TokGlyph:
			kName = "GLYPH"
		case TokOp:
			kName = "OP"
		case TokPunct:
			kName = "PUNCT"
		default:
			kName = "OTHER"
		}
		b.WriteString(fmt.Sprintf("%04d [L%-3d C%-3d] %-10s %s\n", i, t.Line, t.Col, kName, t.Lexeme))
	}
	b.WriteString("==============================================")
	return b.String()
}

type ExecutionProfile struct {
	InstructionsExecuted int64
	PeakStackDepth       int
	FinalStackDepth      int
	OpcodeHistogram      map[string]int64
	ElapsedCycles        int64
}

func (vm *AetherVM) ExecChunkWithProfile(chunk *CompiledChunk) (Datum, ExecutionProfile, error) {
	prof := ExecutionProfile{
		OpcodeHistogram: make(map[string]int64),
	}
	if chunk == nil {
		return VoidDatum(), prof, vm.Kernel.fault(1, "ERR_NIL_CHUNK", "bytecode chunk is nil", "compile source into chunk before execution")
	}

	pc := chunk.EntryPC
	lastVal := VoidDatum()
	baseStackLen := len(vm.Stack)

	for pc >= 0 && pc < len(chunk.Code) {
		ins := chunk.Code[pc]
		if err := vm.Kernel.AdvanceCycle(ins.Line); err != nil {
			prof.FinalStackDepth = len(vm.Stack)
			return VoidDatum(), prof, err
		}
		prof.InstructionsExecuted++
		prof.OpcodeHistogram[ins.Op.Mnemonic()]++

		beforeLen := len(vm.Stack)
		stepChunk := &CompiledChunk{
			Code:      []Instruction{ins},
			Constants: chunk.Constants,
			Symbols:   chunk.Symbols,
			EntryPC:   0,
		}

		if ins.Op == OpJump {
			pc = ins.ArgA
			if len(vm.Stack) > prof.PeakStackDepth {
				prof.PeakStackDepth = len(vm.Stack)
			}
			continue
		}
		if ins.Op == OpJumpIfFalse {
			cond, err := vm.Pop(ins.Line)
			if err != nil {
				prof.FinalStackDepth = len(vm.Stack)
				return VoidDatum(), prof, err
			}
			if !cond.Truthy() {
				pc = ins.ArgA
			} else {
				pc++
			}
			if len(vm.Stack) > prof.PeakStackDepth {
				prof.PeakStackDepth = len(vm.Stack)
			}
			continue
		}
		if ins.Op == OpJumpIfTrue {
			cond, err := vm.Pop(ins.Line)
			if err != nil {
				prof.FinalStackDepth = len(vm.Stack)
				return VoidDatum(), prof, err
			}
			if cond.Truthy() {
				pc = ins.ArgA
			} else {
				pc++
			}
			if len(vm.Stack) > prof.PeakStackDepth {
				prof.PeakStackDepth = len(vm.Stack)
			}
			continue
		}
		if ins.Op == OpCall {
			if ins.ArgA >= 0 && ins.ArgA < len(chunk.Symbols) {
				fnName := chunk.Symbols[ins.ArgA]
				if _, isProto := vm.Protos[fnName]; isProto {
					argCount := ins.ArgB
					if argCount < 0 || len(vm.Stack) < argCount {
						prof.FinalStackDepth = len(vm.Stack)
						return VoidDatum(), prof, vm.Kernel.fault(ins.Line, "ERR_CALL_STACK", "insufficient arguments on stack", "check routine call arguments")
					}
					args := make([]Datum, argCount)
					start := len(vm.Stack) - argCount
					for i := 0; i < argCount; i++ {
						args[i] = vm.Stack[start+i].DeepCopy()
					}
					vm.Stack = vm.Stack[:start]
					proto := vm.Protos[fnName]
					if len(args) != len(proto.Params) {
						prof.FinalStackDepth = len(vm.Stack)
						return VoidDatum(), prof, vm.Kernel.fault(ins.Line, "ERR_ROUTINE_ARITY", fmt.Sprintf("routine '%s' expects %d args, got %d", fnName, len(proto.Params), len(args)), "match argument count")
					}
					if err := vm.Kernel.OpenFrame(fnName); err != nil {
						prof.FinalStackDepth = len(vm.Stack)
						return VoidDatum(), prof, err
					}
					for i, p := range proto.Params {
						vm.Kernel.StoreLocalSymbol(p, args[i])
					}
					vm.CallInfo = append(vm.CallInfo, CallFrame{
						ReturnPC:  pc + 1,
						BaseDepth: len(vm.Stack),
						Name:      fnName,
					})
					pc = proto.Entry
					if len(vm.Stack) > prof.PeakStackDepth {
						prof.PeakStackDepth = len(vm.Stack)
					}
					continue
				}
				if ins.ArgB >= 0 && len(vm.Stack) >= ins.ArgB {
					extArgs := make([]Datum, ins.ArgB)
					st := len(vm.Stack) - ins.ArgB
					for i := 0; i < ins.ArgB; i++ {
						extArgs[i] = vm.Stack[st+i].DeepCopy()
					}
					if handled, extVal, extErr := vm.EvalExtendedBuiltin(fnName, extArgs, ins.Line); handled {
						if extErr != nil {
							prof.FinalStackDepth = len(vm.Stack)
							return VoidDatum(), prof, extErr
						}
						vm.Stack = vm.Stack[:st]
						if err := vm.Push(extVal, ins.Line); err != nil {
							prof.FinalStackDepth = len(vm.Stack)
							return VoidDatum(), prof, err
						}
						pc++
						if len(vm.Stack) > prof.PeakStackDepth {
							prof.PeakStackDepth = len(vm.Stack)
						}
						continue
					}
				}
			}
		}

		res, err := vm.ExecChunk(stepChunk)
		if err != nil {
			prof.FinalStackDepth = len(vm.Stack)
			return VoidDatum(), prof, err
		}
		lastVal = res

		if ins.Op == OpReturn {
			if len(vm.CallInfo) > 0 {
				fr := vm.CallInfo[len(vm.CallInfo)-1]
				vm.CallInfo = vm.CallInfo[:len(vm.CallInfo)-1]
				if len(vm.Stack) > fr.BaseDepth {
					vm.Stack = vm.Stack[:fr.BaseDepth]
				}
				vm.Kernel.CloseFrame()
				if err := vm.Push(lastVal, ins.Line); err != nil {
					prof.FinalStackDepth = len(vm.Stack)
					return VoidDatum(), prof, err
				}
				pc = fr.ReturnPC
				if len(vm.Stack) > prof.PeakStackDepth {
					prof.PeakStackDepth = len(vm.Stack)
				}
				continue
			}
			break
		}

		if len(vm.Stack) > prof.PeakStackDepth {
			prof.PeakStackDepth = len(vm.Stack)
		}
		_ = beforeLen
		pc++
	}

	if len(vm.Stack) > baseStackLen {
		lastVal = vm.Stack[len(vm.Stack)-1].DeepCopy()
	}
	prof.FinalStackDepth = len(vm.Stack)
	prof.ElapsedCycles = vm.Kernel.stepUsed
	return lastVal, prof, nil
}

func ValidateSourceSemantics(source string) []string {
	norm := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(norm, "\n")
	var notes []string

	for idx, raw := range lines {
		lineNo := idx + 1
		clean := strings.TrimSpace(StripInlineComment(raw))
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, "for ") && !strings.Contains(clean, " in ") {
			notes = append(notes, fmt.Sprintf("Line %d: 'for' loop missing 'in' keyword -> expected 'for item in collection:'", lineNo))
		}
		if (strings.HasPrefix(clean, "let ") || strings.HasPrefix(clean, "bind ") || strings.HasPrefix(clean, "var ")) && !strings.Contains(clean, "=") {
			notes = append(notes, fmt.Sprintf("Line %d: declaration missing '=' assignment -> expected 'let x = value'", lineNo))
		}
		if strings.Contains(clean, "==") == false && strings.Contains(clean, " = ") && (strings.HasPrefix(clean, "when ") || strings.HasPrefix(clean, "if ") || strings.HasPrefix(clean, "loop ") || strings.HasPrefix(clean, "while ")) {
			notes = append(notes, fmt.Sprintf("Line %d: condition uses single '=' instead of '==' comparison", lineNo))
		}
	}

	if len(notes) == 0 {
		notes = append(notes, "Semantic check passed: no obvious loop, binding, or comparison anomalies found.")
	}
	return notes
}

func (eng *AetherEngine) FormatFullTelemetry() string {
	var b strings.Builder
	b.WriteString("========================================================\n")
	b.WriteString("           AETHER-BMM FULL RUNTIME TELEMETRY            \n")
	b.WriteString("========================================================\n")
	b.WriteString(fmt.Sprintf("Engine Name  : %s\n", AetherName))
	b.WriteString(fmt.Sprintf("Release      : %s\n", AetherRelease))
	b.WriteString(fmt.Sprintf("SubSystem    : %s\n", AetherSubSystem))
	b.WriteString(fmt.Sprintf("Frames       : %d (Call Stack: %d)\n", len(eng.Kernel.frames), len(eng.Kernel.callStack)))
	b.WriteString(fmt.Sprintf("VM Stack     : %d items\n", len(eng.VM.Stack)))
	b.WriteString(fmt.Sprintf("Routines     : %d\n", len(eng.VM.Protos)))
	b.WriteString(fmt.Sprintf("Unit Schemas : %d\n", len(eng.VM.Units)))
	b.WriteString(fmt.Sprintf("Pulse Bands  : %d\n", len(eng.Kernel.carriers)))
	b.WriteString(fmt.Sprintf("Vault Slots  : %d\n", len(eng.Kernel.vaults)))
	b.WriteString(fmt.Sprintf("Cycle Quota  : %d / %d\n", eng.Kernel.stepUsed, eng.Kernel.stepBudget))
	b.WriteString("========================================================")
	return b.String()
}

func RenderAetherGuide() string {
	return `========================================================
               AETHER-BMM COMPLETE LANGUAGE GUIDE
========================================================
DATUM TYPES:
  cycle    - 64-bit integer (e.g., 10, -5)
  glyph    - UTF-8 string (e.g., "aether_core")
  flag     - Boolean state (TRUE / FALSE)
  vec      - Ordered sequence ([1, 2, 3])
  dict     - Key-value map ({host: "grid", port: 8080})
  pulse    - Signal channel (<pulse:band@hzHz>)
  unit     - Structured schema instance via unit Name:
  coderef  - Routine bytecode entry reference
  void     - Empty datum (VOID)

KEYWORDS & CONSTRUCTS:
  let / bind / var        - Bind symbol in active scope
  when / elif / else      - Conditional branching
  loop / while            - Conditional cycle loop
  for x in collection     - Iterate vec, glyph, or dict
  fn / def / routine      - Define callable routine
  yield / ret / return    - Return value from routine or block
  unit / struct           - Declare field schema
  guard ... rescue        - Trap errors and bind last_fault
  assert cond [, msg]     - Assert condition is truthy
  raise_fault expr        - Raise structured fault

BUILT-IN (_bmm) ROUTINES:
  emit_bmm(...), kind_bmm(x), len_bmm(x)
  cycle_bmm(x), glyph_bmm(x), state_bmm(x)
  seq_bmm(a [, b [, c]]), push_bmm(v, x), pop_bmm(v), slice_bmm(v, a, b)
  keys_bmm(d), has_bmm(d, k), merge_dict_bmm(d1, d2)
  rev_bmm(v), has_item_bmm(v, x), index_of_bmm(v, x), stats_bmm(v)
  pulse_bmm(band, hz), read_pulse_bmm(band), pulse_bands_bmm()
  vault_put_bmm(k, v), vault_get_bmm(k), vault_del_bmm(k), vault_keys_bmm()
  abs_bmm(x), clamp_bmm(x, lo, hi), min_bmm(a, b), max_bmm(a, b)
  pow_bmm(a, b), modpow_bmm(a, b, m), gcd_bmm(a, b), lcm_bmm(a, b)
  upper_bmm(s), lower_bmm(s), trim_bmm(s), split_bmm(s, sep), join_bmm(v, sep)
  contains_str_bmm(s, sub), replace_str_bmm(s, old, new)

INTERACTIVE SHELL COMMANDS:
  :help      Show this guide
  :status    Show engine and VM telemetry
  :tokens    Print lexer tokens for last source
  :disasm    Print compiled bytecode instructions
  :audit     Run structural and semantic checks
  :repair    Apply explicit delimiter/quote repairs to input
  :budget N  Set max execution cycles
  :purge     Reset engine state and memory
  :exit      Quit shell
========================================================`
}

func (eng *AetherEngine) RunInteractiveShell() {
	fmt.Printf("[%s :: %s :: %s] Ready.\n", AetherName, AetherRelease, AetherSubSystem)
	fmt.Println("Type :help for commands or :exit to quit.")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 64*1024), 8*1024*1024)

	var blockBuf []string
	var blockIndent int
	inBlock := false
	lineNum := 0

	for {
		lineNum++
		if inBlock {
			fmt.Print("..> ")
		} else {
			fmt.Print("aether> ")
		}

		if !scanner.Scan() {
			fmt.Println()
			break
		}

		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)

		if !inBlock {
			if trimmed == "" {
				continue
			}

			switch {
			case trimmed == ":exit" || trimmed == "exit" || trimmed == "quit":
				fmt.Printf("[%s] Session ended.\n", AetherName)
				return
			case trimmed == ":help":
				fmt.Println(RenderAetherGuide())
				continue
			case trimmed == ":status":
				fmt.Println(eng.FormatFullTelemetry())
				continue
			case trimmed == ":purge":
				eng.ResetEngine()
				fmt.Printf("[%s] Engine state reset.\n", AetherName)
				continue
			case trimmed == ":tokens":
				if eng.LastProg == nil {
					fmt.Println("No source compiled yet.")
				} else {
					fmt.Println(eng.InspectTokenStream(eng.LastProg.RawSource))
				}
				continue
			case trimmed == ":disasm":
				fmt.Println(eng.DumpLastBytecode())
				continue
			case trimmed == ":audit":
				if eng.LastProg == nil {
					fmt.Println("No source compiled yet.")
				} else {
					for _, f := range AuditSourceStructure(eng.LastProg.RawSource) {
						fmt.Println("[STRUCT]", f)
					}
					for _, s := range ValidateSourceSemantics(eng.LastProg.RawSource) {
						fmt.Println("[SEMANTIC]", s)
					}
				}
				continue
			case strings.HasPrefix(trimmed, ":repair "):
				targetCode := strings.TrimSpace(strings.TrimPrefix(trimmed, ":repair "))
				repaired, props := ApplyExplicitRepairs(targetCode)
				for _, p := range props {
					fmt.Printf("[REPAIR L%d] %s -> %s (%s)\n", p.Line, strings.TrimSpace(p.Original), strings.TrimSpace(p.Repaired), p.Explanation)
				}
				val, diagTxt, err := eng.EvalWithExplanation(repaired, "<shell-repair>")
				if diagTxt != "" {
					fmt.Println(diagTxt)
				}
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
				} else if val.Kind != DatumVoid {
					fmt.Println(val.Stringify())
				}
				continue
			case strings.HasPrefix(trimmed, ":budget "):
				rawB := strings.TrimSpace(strings.TrimPrefix(trimmed, ":budget "))
				bVal, err := strconv.ParseInt(rawB, 10, 64)
				if err != nil || bVal < 1 {
					fmt.Fprintf(os.Stderr, "AETHER-SHELL [ERR_BUDGET @ line %d]: budget must be integer >= 1\n", lineNum)
					continue
				}
				eng.Kernel.stepBudget = bVal
				fmt.Printf("[%s] Cycle budget set to %d.\n", AetherName, bVal)
				continue
			}

			clean := strings.TrimSpace(StripInlineComment(raw))
			ind := CalcLeadingIndent(StripInlineComment(raw))
			if clean != "" && eng.Compiler.isBlockStart(clean) {
				inBlock = true
				blockIndent = ind
				blockBuf = []string{raw}
				continue
			}

			val, diagTxt, err := eng.EvalWithExplanation(raw, "<shell>")
			if diagTxt != "" {
				fmt.Println(diagTxt)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			if val.Kind != DatumVoid {
				fmt.Println(val.Stringify())
			}
			continue
		}

		if trimmed == "" {
			script := strings.Join(blockBuf, "\n")
			blockBuf = nil
			inBlock = false
			blockIndent = 0

			val, diagTxt, err := eng.EvalWithExplanation(script, "<shell-block>")
			if diagTxt != "" {
				fmt.Println(diagTxt)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}
			if val.Kind != DatumVoid {
				fmt.Println(val.Stringify())
			}
			continue
		}

		stripped := StripInlineComment(raw)
		curInd := CalcLeadingIndent(stripped)
		curTrim := strings.TrimSpace(stripped)

		if curInd <= blockIndent && len(blockBuf) > 0 && !eng.Compiler.isElseStart(curTrim) && !eng.Compiler.hasKw(curTrim, "elif") && !eng.Compiler.hasKw(curTrim, "else if") && !strings.HasPrefix(curTrim, "rescue") && !strings.HasPrefix(curTrim, "catch") {
			script := strings.Join(blockBuf, "\n")
			blockBuf = nil
			inBlock = false
			blockIndent = 0

			val, diagTxt, err := eng.EvalWithExplanation(script, "<shell-block>")
			if diagTxt != "" {
				fmt.Println(diagTxt)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
			} else if val.Kind != DatumVoid {
				fmt.Println(val.Stringify())
			}

			if curTrim != "" && eng.Compiler.isBlockStart(curTrim) {
				inBlock = true
				blockIndent = curInd
				blockBuf = []string{raw}
			} else if curTrim != "" {
				v2, d2, e2 := eng.EvalWithExplanation(raw, "<shell>")
				if d2 != "" {
					fmt.Println(d2)
				}
				if e2 != nil {
					fmt.Fprintln(os.Stderr, e2.Error())
				} else if v2.Kind != DatumVoid {
					fmt.Println(v2.Stringify())
				}
			}
			continue
		}

		blockBuf = append(blockBuf, raw)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "AETHER-SHELL [ERR_STDIN @ line %d]: %v\n", lineNum, err)
	}
}

func main() {
	engine := NewAetherEngine()

	if len(os.Args) == 1 {
		engine.RunInteractiveShell()
		return
	}

	switch os.Args[1] {
	case "-v", "--version", "version":
		fmt.Printf("%s %s [%s]\n", AetherName, AetherRelease, AetherSubSystem)
	case "-h", "--help", "help":
		fmt.Println(RenderAetherGuide())
	case "status", "--status":
		fmt.Println(engine.FormatFullTelemetry())
	case "-e", "--eval":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "AETHER-CLI [ERR_NO_CODE @ line 1]: missing expression after -e")
			os.Exit(1)
		}
		src := strings.Join(os.Args[2:], "\n")
		val, diagTxt, err := engine.EvalWithExplanation(src, "<cli-eval>")
		if diagTxt != "" {
			fmt.Fprintln(os.Stderr, diagTxt)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if val.Kind != DatumVoid {
			fmt.Println(val.Stringify())
		}
	case "--disasm":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "AETHER-CLI [ERR_NO_FILE @ line 1]: missing file path after --disasm")
			os.Exit(1)
		}
		raw, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if _, err := engine.BuildProgram(string(raw), os.Args[2]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(engine.DumpLastBytecode())
	case "--tokens":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "AETHER-CLI [ERR_NO_FILE @ line 1]: missing file path after --tokens")
			os.Exit(1)
		}
		raw, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(engine.InspectTokenStream(string(raw)))
	default:
		val, diagTxt, err := engine.EvalWithExplanation("", os.Args[1])
		_ = val
		_ = diagTxt
		_ = err
		fVal, fDiag, fErr := func() (Datum, string, error) {
			content, readErr := os.ReadFile(os.Args[1])
			if readErr != nil {
				return VoidDatum(), "", AetherDiag{
					Line:   1,
					Stage:  "LOADER",
					Code:   "ERR_FILE_IO",
					Detail: fmt.Sprintf("cannot read script '%s': %v", os.Args[1], readErr),
					Hint:   "verify file path and permissions",
				}
			}
			return engine.EvalWithExplanation(string(content), os.Args[1])
		}()
		if fDiag != "" {
			fmt.Fprintln(os.Stderr, fDiag)
		}
		if fErr != nil {
			fmt.Fprintln(os.Stderr, fErr.Error())
			os.Exit(1)
		}
		if fVal.Kind != DatumVoid {
			fmt.Println(fVal.Stringify())
		}
	}
}


type SubRoutine struct {
Ident  string
Params []string
Block  []string
}

type AetherKernel struct {
frames     []*ScopeFrame
callStack  []string
routines   map[string]SubRoutine
carriers   map[string]Datum
vaults     map[string]Datum
stepUsed   int64
stepBudget int64
}

func NewAetherKernel() *AetherKernel {
k := &AetherKernel{
frames:     make([]*ScopeFrame, 0, 32),
callStack:  make([]string, 0, 64),
routines:   make(map[string]SubRoutine),
carriers:   make(map[string]Datum),
vaults:     make(map[string]Datum),
stepUsed:   0,
stepBudget: MaxStepQuota,
}
k.frames = append(k.frames, NewScopeFrame("root_kernel"))
return k
}

func (k *AetherKernel) ResetCounter() {
k.stepUsed = 0
}

func (k *AetherKernel) PurgeAll() {
k.frames = []*ScopeFrame{NewScopeFrame("root_kernel")}
k.callStack = k.callStack[:0]
k.routines = make(map[string]SubRoutine)
k.carriers = make(map[string]Datum)
k.vaults = make(map[string]Datum)
k.stepUsed = 0
}

func (k *AetherKernel) AdvanceCycle(line int) error {
k.stepUsed++
if k.stepBudget > 0 && k.stepUsed > k.stepBudget {
return k.fault(line, "ERR_CYCLE_QUOTA", fmt.Sprintf("execution exceeded cycle budget (%d)", k.stepBudget), "increase cycle budget with :budget N or optimize loop bounds")
}
return nil
}

func (k *AetherKernel) OpenFrame(label string) error {
if len(k.frames) >= MaxCallFrames {
return k.fault(1, "ERR_FRAME_OVERFLOW", "scope frame stack limit reached", "check recursive routine depth")
}
k.frames = append(k.frames, NewScopeFrame(label))
k.callStack = append(k.callStack, label)
return nil
}

func (k *AetherKernel) CloseFrame() {
if len(k.frames) > 1 {
k.frames = k.frames[:len(k.frames)-1]
}
if len(k.callStack) > 0 {
k.callStack = k.callStack[:len(k.callStack)-1]
}
}

func (k *AetherKernel) LoadSymbol(name string) (Datum, bool) {
for i := len(k.frames) - 1; i >= 0; i-- {
if val, ok := k.frames[i].Slots[name]; ok {
return val, true
}
}
return VoidDatum(), false
}

func (k *AetherKernel) StoreSymbol(name string, val Datum) {
for i := len(k.frames) - 1; i >= 0; i-- {
if _, exists := k.frames[i].Slots[name]; exists {
k.frames[i].Slots[name] = val.DeepCopy()
return
}
}
if len(k.frames) == 0 {
k.frames = append(k.frames, NewScopeFrame("root_kernel"))
}
k.frames[len(k.frames)-1].Slots[name] = val.DeepCopy()
}

func (k *AetherKernel) StoreLocalSymbol(name string, val Datum) {
if len(k.frames) == 0 {
k.frames = append(k.frames, NewScopeFrame("root_kernel"))
}
k.frames[len(k.frames)-1].Slots[name] = val.DeepCopy()
}

func (k *AetherKernel) fault(line int, code string, detail string, hint string) AetherDiag {
chain := make([]string, len(k.callStack))
copy(chain, k.callStack)
return AetherDiag{
Line:   line,
Stage:  "KERNEL",
Code:   code,
Detail: detail,
Hint:   hint,
Chain:  chain,
}
}

type FaultReport struct {
Line   int
Code   string
Cause  string
Remedy string
Trace  []string
}

func (fr FaultReport) Error() string {
tr := ""
if len(fr.Trace) > 0 {
tr = " [" + strings.Join(fr.Trace, " -> ") + "]"
}
if fr.Remedy != "" {
return fmt.Sprintf("AETHER-FAULT [%s @ line %d]: %s | Remedy: %s%s", fr.Code, fr.Line, fr.Cause, fr.Remedy, tr)
}
return fmt.Sprintf("AETHER-FAULT [%s @ line %d]: %s%s", fr.Code, fr.Line, fr.Cause, tr)
}
