package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type calcTok struct {
	kind int
	val  float64
	op   rune
}

const (
	tokNum = iota + 1
	tokOp
	tokLParen
	tokRParen
)

func tokenizeCalc(expr string) ([]calcTok, error) {
	var out []calcTok
	i := 0
	n := len(expr)
	for i < n {
		ch := rune(expr[i])
		if unicode.IsSpace(ch) {
			i++
			continue
		}
		if unicode.IsDigit(ch) || ch == '.' {
			start := i
			dots := 0
			for i < n && (unicode.IsDigit(rune(expr[i])) || expr[i] == '.') {
				if expr[i] == '.' {
					dots++
					if dots > 1 {
						return nil, errors.New("invalid number format")
					}
				}
				i++
			}
			v, err := strconv.ParseFloat(expr[start:i], 64)
			if err != nil {
				return nil, err
			}
			out = append(out, calcTok{kind: tokNum, val: v})
			continue
		}
		if ch == '(' {
			out = append(out, calcTok{kind: tokLParen})
			i++
			continue
		}
		if ch == ')' {
			out = append(out, calcTok{kind: tokRParen})
			i++
			continue
		}
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '%' {
			out = append(out, calcTok{kind: tokOp, op: ch})
			i++
			continue
		}
		return nil, fmt.Errorf("unexpected token: %c", ch)
	}
	return out, nil
}

func calcPrecedence(op rune) int {
	switch op {
	case '+', '-':
		return 1
	case '*', '/', '%':
		return 2
	default:
		return 0
	}
}

func applyCalcOp(op rune, a, b float64) (float64, error) {
	switch op {
	case '+':
		return a + b, nil
	case '-':
		return a - b, nil
	case '*':
		return a * b, nil
	case '/':
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	case '%':
		if b == 0 {
			return 0, errors.New("modulo by zero")
		}
		bi := int64(b)
		if bi == 0 {
			return 0, errors.New("modulo by zero")
		}
		return float64(int64(a) % bi), nil
	default:
		return 0, errors.New("unsupported operator")
	}
}

func evalMathExpr(expr string) (float64, error) {
	tokens, err := tokenizeCalc(expr)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, errors.New("empty expression")
	}
	var vals []float64
	var ops []calcTok
	expectUnary := true
	for _, tk := range tokens {
		switch tk.kind {
		case tokNum:
			vals = append(vals, tk.val)
			expectUnary = false
		case tokLParen:
			ops = append(ops, tk)
			expectUnary = true
		case tokRParen:
			found := false
			for len(ops) > 0 {
				top := ops[len(ops)-1]
				ops = ops[:len(ops)-1]
				if top.kind == tokLParen {
					found = true
					break
				}
				if len(vals) < 2 {
					return 0, errors.New("malformed expression")
				}
				b := vals[len(vals)-1]
				a := vals[len(vals)-2]
				vals = vals[:len(vals)-2]
				res, err := applyCalcOp(top.op, a, b)
				if err != nil {
					return 0, err
				}
				vals = append(vals, res)
			}
			if !found {
				return 0, errors.New("mismatched parentheses")
			}
			expectUnary = false
		case tokOp:
			if expectUnary && (tk.op == '+' || tk.op == '-') {
				vals = append(vals, 0)
			}
			for len(ops) > 0 {
				top := ops[len(ops)-1]
				if top.kind != tokOp || calcPrecedence(top.op) < calcPrecedence(tk.op) {
					break
				}
				ops = ops[:len(ops)-1]
				if len(vals) < 2 {
					return 0, errors.New("malformed expression")
				}
				b := vals[len(vals)-1]
				a := vals[len(vals)-2]
				vals = vals[:len(vals)-2]
				res, err := applyCalcOp(top.op, a, b)
				if err != nil {
					return 0, err
				}
				vals = append(vals, res)
			}
			ops = append(ops, tk)
			expectUnary = true
		}
	}
	for len(ops) > 0 {
		top := ops[len(ops)-1]
		ops = ops[:len(ops)-1]
		if top.kind == tokLParen {
			return 0, errors.New("unclosed parenthesis")
		}
		if len(vals) < 2 {
			return 0, errors.New("malformed expression")
		}
		b := vals[len(vals)-1]
		a := vals[len(vals)-2]
		vals = vals[:len(vals)-2]
		res, err := applyCalcOp(top.op, a, b)
		if err != nil {
			return 0, err
		}
		vals = append(vals, res)
	}
	if len(vals) != 1 {
		return 0, errors.New("invalid expression result")
	}
	return vals[0], nil
}

func computeHMAC(secret, msg string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func EvalRuntimeEngineLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "atomic_write_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_ARG", Detail: "atomic_write_bmm(path, content) requires two glyphs", Hint: "pass target file path and content"}
		}
		dir := filepath.Dir(args[0].Glyph)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_DIR", Detail: err.Error(), Hint: "check parent directory"}
		}
		tmpFile, err := os.CreateTemp(dir, ".bmm_atomic_*")
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_TMP", Detail: err.Error(), Hint: "check write permissions in directory"}
		}
		tmpName := tmpFile.Name()
		if _, err := tmpFile.WriteString(args[1].Glyph); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpName)
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_W", Detail: err.Error(), Hint: "failed writing temporary file"}
		}
		if err := tmpFile.Sync(); err != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpName)
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_SYNC", Detail: err.Error(), Hint: "failed syncing file to disk"}
		}
		if err := tmpFile.Close(); err != nil {
			_ = os.Remove(tmpName)
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_CLS", Detail: err.Error(), Hint: "failed closing temporary file"}
		}
		if err := os.Chmod(tmpName, 0644); err != nil {
			_ = os.Remove(tmpName)
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_MOD", Detail: err.Error(), Hint: "failed setting file permissions"}
		}
		if err := os.Rename(tmpName, args[0].Glyph); err != nil {
			_ = os.Remove(tmpName)
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ATOMWR_RNM", Detail: err.Error(), Hint: "failed replacing target file"}
		}
		return true, FlagDatum(true), nil

	case "dotenv_load_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ENVLD_ARG", Detail: "dotenv_load_bmm(path) requires glyph", Hint: "pass path to .env file"}
		}
		raw, err := os.ReadFile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ENVLD_IO", Detail: err.Error(), Hint: "check .env file path"}
		}
		count := 0
		bmmStoreMu.Lock()
		for _, ln := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(ln)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			if strings.HasPrefix(trimmed, "export ") {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "export "))
			}
			eqIdx := strings.IndexByte(trimmed, '=')
			if eqIdx <= 0 {
				continue
			}
			k := strings.TrimSpace(trimmed[:eqIdx])
			v := strings.TrimSpace(trimmed[eqIdx+1:])
			if len(v) >= 2 && ((strings.HasPrefix(v, "") && strings.HasSuffix(v, "")) || (strings.HasPrefix(v, "'") && strings.HasSuffix(v, "'"))) {
				v = v[1 : len(v)-1]
			}
			bmmStoreData[k] = v
			count++
		}
		bmmStoreMu.Unlock()
		return true, GlyphDatum(strconv.Itoa(count)), nil

	case "dotenv_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_ENVG_ARG", Detail: "dotenv_get_bmm(key) requires glyph", Hint: "pass key name"}
		}
		bmmStoreMu.RLock()
		v := bmmStoreData[args[0].Glyph]
		bmmStoreMu.RUnlock()
		return true, GlyphDatum(v), nil

	case "csv_count_rows_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVC_ARG", Detail: "csv_count_rows_bmm(csv_text) requires glyph", Hint: "pass csv string"}
		}
		r := csv.NewReader(strings.NewReader(args[0].Glyph))
		rows, err := r.ReadAll()
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVC_PRS", Detail: err.Error(), Hint: "check valid csv formatting"}
		}
		return true, GlyphDatum(strconv.Itoa(len(rows))), nil

	case "csv_row_col_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVRC_ARG", Detail: "csv_row_col_bmm(csv_text, row_idx, col_idx) requires three glyphs", Hint: "pass csv text and 0-based row and column indices"}
		}
		rowIdx, err := strconv.Atoi(args[1].Glyph)
		if err != nil || rowIdx < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVRC_ROW", Detail: "invalid row index", Hint: "pass non-negative integer string"}
		}
		colIdx, err := strconv.Atoi(args[2].Glyph)
		if err != nil || colIdx < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVRC_COL", Detail: "invalid col index", Hint: "pass non-negative integer string"}
		}
		r := csv.NewReader(strings.NewReader(args[0].Glyph))
		rows, err := r.ReadAll()
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CSVRC_PRS", Detail: err.Error(), Hint: "check valid csv formatting"}
		}
		if rowIdx >= len(rows) || colIdx >= len(rows[rowIdx]) {
			return true, GlyphDatum(""), nil
		}
		return true, GlyphDatum(rows[rowIdx][colIdx]), nil

	case "kv_pick_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_KVP_ARG", Detail: "kv_pick_bmm(kv_text, key) requires two glyphs", Hint: "pass lines of key=value and target key"}
		}
		targetKey := strings.TrimSpace(args[1].Glyph)
		for _, ln := range strings.Split(args[0].Glyph, "\n") {
			trimmed := strings.TrimSpace(ln)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			eqIdx := strings.IndexByte(trimmed, '=')
			if eqIdx <= 0 {
				continue
			}
			k := strings.TrimSpace(trimmed[:eqIdx])
			if k == targetKey {
				return true, GlyphDatum(strings.TrimSpace(trimmed[eqIdx+1:])), nil
			}
		}
		return true, GlyphDatum(""), nil

	case "hmac_sha256_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_HMAC_ARG", Detail: "hmac_sha256_bmm(secret, message) requires two glyphs", Hint: "pass secret key and message"}
		}
		return true, GlyphDatum(computeHMAC(args[0].Glyph, args[1].Glyph)), nil

	case "hmac_verify_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_HMACV_ARG", Detail: "hmac_verify_bmm(secret, message, sig_hex) requires three glyphs", Hint: "pass secret, message, and expected hex signature"}
		}
		actual := computeHMAC(args[0].Glyph, args[1].Glyph)
		expected := strings.ToLower(strings.TrimSpace(args[2].Glyph))
		if len(actual) != len(expected) {
			return true, FlagDatum(false), nil
		}
		ok := subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
		return true, FlagDatum(ok), nil

	case "ram_bmm":
return true, GlyphDatum(getSystemRAM()), nil

case "sys_bmm":
return true, GlyphDatum(getSysArchInfo()), nil

case "read_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_READ_ARG", Detail: "read_bmm(path) requires a file path"}
}
return true, GlyphDatum(readFileSafe(args[0].Glyph)), nil

case "write_bmm":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_WRITE_ARG", Detail: "write_bmm(path, text) requires two strings"}
}
return true, GlyphDatum(writeFileSafe(args[0].Glyph, args[1].Glyph)), nil

case "exists_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_EXISTS_ARG", Detail: "exists_bmm(path) requires a file path"}
}
return true, FlagDatum(fileExistsSafe(args[0].Glyph)), nil

case "lib1_ram":
return true, GlyphDatum(lib1RamInfo()), nil

case "lib2_cpu":
return true, GlyphDatum(lib2CpuInfo()), nil

case "lib3_battery":
return true, GlyphDatum(lib3BatteryInfo()), nil

case "lib4_disk":
return true, GlyphDatum(lib4DiskInfo()), nil

case "lib5_env":
key := ""
if len(args) == 1 && args[0].Kind == DatumGlyph {
key = args[0].Glyph
}
return true, GlyphDatum(lib5EnvInfo(key)), nil

case "lib6_read":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB6_ARG", Detail: "lib6_read(path) requires a string path"}
}
return true, GlyphDatum(lib6FsRead(args[0].Glyph)), nil

case "lib7_write":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB7_ARG", Detail: "lib7_write(path, text) requires two strings"}
}
return true, GlyphDatum(lib7FsWrite(args[0].Glyph, args[1].Glyph, false)), nil

case "lib7_append":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB7A_ARG", Detail: "lib7_append(path, text) requires two strings"}
}
return true, GlyphDatum(lib7FsWrite(args[0].Glyph, args[1].Glyph, true)), nil

case "lib8_fs":
if len(args) < 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB8_ARG", Detail: "lib8_fs(action, src, [dst]) requires strings"}
}
dst := ""
if len(args) >= 3 && args[2].Kind == DatumGlyph {
dst = args[2].Glyph
}
return true, GlyphDatum(lib8FsManage(args[0].Glyph, args[1].Glyph, dst)), nil

case "lib9_dir":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB9_ARG", Detail: "lib9_dir(action, path) requires two strings"}
}
return true, GlyphDatum(lib9FsDir(args[0].Glyph, args[1].Glyph)), nil

case "lib10_meta":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB10_ARG", Detail: "lib10_meta(path) requires a string path"}
}
return true, GlyphDatum(lib10FsMeta(args[0].Glyph)), nil

case "lib11_case":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB11_ARG", Detail: "lib11_case(mode, text) requires two strings"}
}
return true, GlyphDatum(lib11StrCase(args[0].Glyph, args[1].Glyph)), nil

case "lib12_trim":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB12_ARG", Detail: "lib12_trim(text, [cutset]) requires a string"}
}
cut := ""
if len(args) >= 2 && args[1].Kind == DatumGlyph {
cut = args[1].Glyph
}
return true, GlyphDatum(lib12StrTrim(args[0].Glyph, cut)), nil

case "lib13_find":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB13_ARG", Detail: "lib13_find(mode, text, sub) requires three strings"}
}
return true, GlyphDatum(lib13StrFind(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "lib14_replace":
if len(args) < 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB14_ARG", Detail: "lib14_replace(text, old, new, [all]) requires strings"}
}
allFlag := false
if len(args) >= 4 && args[3].Kind == DatumGlyph && strings.EqualFold(args[3].Glyph, "all") {
allFlag = true
}
return true, GlyphDatum(lib14StrReplace(args[0].Glyph, args[1].Glyph, args[2].Glyph, allFlag)), nil

case "lib15_split_join":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB15_ARG", Detail: "lib15_split_join(text, sep, joinSep) requires three strings"}
}
return true, GlyphDatum(lib15StrSplitJoin(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "lib16_math":
if len(args) < 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB16_ARG", Detail: "lib16_math(op, a, [b]) requires op and numbers"}
}
a, errA := strconv.ParseFloat(strings.TrimSpace(args[1].Glyph), 64)
if errA != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB16_NUM", Detail: "invalid first number"}
}
b := 0.0
if len(args) >= 3 && args[2].Kind == DatumGlyph {
b, _ = strconv.ParseFloat(strings.TrimSpace(args[2].Glyph), 64)
}
return true, GlyphDatum(lib16MathBasic(args[0].Glyph, a, b)), nil

case "lib17_round":
if len(args) < 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB17_ARG", Detail: "lib17_round(mode, val, [places]) requires strings"}
}
val, errV := strconv.ParseFloat(strings.TrimSpace(args[1].Glyph), 64)
if errV != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB17_NUM", Detail: "invalid float value"}
}
places := 0
if len(args) >= 3 && args[2].Kind == DatumGlyph {
places, _ = strconv.Atoi(strings.TrimSpace(args[2].Glyph))
}
return true, GlyphDatum(lib17MathRound(args[0].Glyph, val, places)), nil

case "lib18_stat":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB18_ARG", Detail: "lib18_stat(mode, csvNums) requires two strings"}
}
return true, GlyphDatum(lib18MathStat(args[0].Glyph, args[1].Glyph)), nil

case "lib19_rnd":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB19_ARG", Detail: "lib19_rand(min, max) requires two strings"}
}
minV, err1 := strconv.Atoi(strings.TrimSpace(args[0].Glyph))
maxV, err2 := strconv.Atoi(strings.TrimSpace(args[1].Glyph))
if err1 != nil || err2 != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB19_NUM", Detail: "min and max must be integers"}
}
return true, GlyphDatum(lib19MathRand(minV, maxV)), nil

case "lib20_conv":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB20_ARG", Detail: "lib20_conv(num, base) requires two strings"}
}
nVal, errN := strconv.ParseInt(strings.TrimSpace(args[0].Glyph), 10, 64)
baseVal, errB := strconv.Atoi(strings.TrimSpace(args[1].Glyph))
if errN != nil || errB != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB20_NUM", Detail: "num and base must be integers"}
}
return true, GlyphDatum(lib20MathConv(nVal, baseVal)), nil

case "lib21_red_num":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB21_ARG", Detail: "lib21_red_num(val) requires a string"}
}
return true, GlyphDatum(lib21UiNumRed(args[0].Glyph)), nil

case "lib22_blue_item":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB22_ARG", Detail: "lib22_blue_item(val) requires a string"}
}
return true, GlyphDatum(lib22UiItemBlue(args[0].Glyph)), nil

case "lib23_yellow_table":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB23_ARG", Detail: "lib23_yellow_table(title, rows) requires two strings"}
}
return true, GlyphDatum(lib23UiTableYellow(args[0].Glyph, args[1].Glyph)), nil

case "lib24_banner":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB24_ARG", Detail: "lib24_banner(title, sub) requires two strings"}
}
return true, GlyphDatum(lib24UiBanner(args[0].Glyph, args[1].Glyph)), nil

case "lib25_badge":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB25_ARG", Detail: "lib25_badge(kind, msg) requires two strings"}
}
return true, GlyphDatum(lib25UiBadge(args[0].Glyph, args[1].Glyph)), nil

case "lib26_progress":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB26_ARG", Detail: "lib26_progress(label, pct) requires two strings"}
}
pct, errP := strconv.Atoi(strings.TrimSpace(args[1].Glyph))
if errP != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB26_NUM", Detail: "pct must be an integer"}
}
return true, GlyphDatum(lib26UiProgress(args[0].Glyph, pct)), nil

case "lib27_box":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB27_ARG", Detail: "lib27_box(title, body) requires two strings"}
}
return true, GlyphDatum(lib27UiBox(args[0].Glyph, args[1].Glyph)), nil

case "lib28_divider":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB28_ARG", Detail: "lib28_divider(ch, count) requires two strings"}
}
cnt, errC := strconv.Atoi(strings.TrimSpace(args[1].Glyph))
if errC != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB28_NUM", Detail: "count must be an integer"}
}
return true, GlyphDatum(lib28UiDivider(args[0].Glyph, cnt)), nil

case "lib29_kv":
if len(args) < 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB29_ARG", Detail: "lib29_kv(key, val, [red]) requires strings"}
}
redFlag := false
if len(args) >= 3 && args[2].Kind == DatumGlyph && strings.EqualFold(strings.TrimSpace(args[2].Glyph), "red") {
redFlag = true
}
return true, GlyphDatum(lib29UiKv(args[0].Glyph, args[1].Glyph, redFlag)), nil

case "lib30_ui_tag":
mode := "tag"
if len(args) == 1 && args[0].Kind == DatumGlyph {
mode = args[0].Glyph
}
return true, GlyphDatum(lib30UiClear(mode)), nil

case "lib31_net_chk":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB31_ARG", Detail: "lib31_net_chk(target, [timeoutMs]) requires a string"}
}
tMs := 2500
if len(args) >= 2 && args[1].Kind == DatumGlyph {
if v, err := strconv.Atoi(strings.TrimSpace(args[1].Glyph)); err == nil {
tMs = v
}
}
return true, GlyphDatum(lib31NetCheck(args[0].Glyph, tMs)), nil

case "lib32_net_ip":
return true, GlyphDatum(lib32NetLocalIPs()), nil

case "lib33_sh":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB33_ARG", Detail: "lib33_sh(cmd) requires a string command"}
}
return true, GlyphDatum(lib33ShellExec(args[0].Glyph)), nil

case "lib34_log":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB34_ARG", Detail: "lib34_log(path, level, msg) requires three strings"}
}
return true, GlyphDatum(lib34DiagLog(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "lib35_cfg":
if len(args) < 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB35_ARG", Detail: "lib35_cfg(mode, path, key, [val]) requires strings"}
}
val := ""
if len(args) >= 4 && args[3].Kind == DatumGlyph {
val = args[3].Glyph
}
return true, GlyphDatum(lib35Config(args[0].Glyph, args[1].Glyph, args[2].Glyph, val)), nil

case "lib36_lst_len":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB36_ARG", Detail: "lib36_lst_len(list, [sep]) requires a string"}
}
sep := ","
if len(args) >= 2 && args[1].Kind == DatumGlyph {
sep = args[1].Glyph
}
return true, GlyphDatum(lib36ListLen(args[0].Glyph, sep)), nil

case "lib37_lst_get":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB37_ARG", Detail: "lib37_lst_get(list, sep, idx) requires three strings"}
}
idx, errI := strconv.Atoi(strings.TrimSpace(args[2].Glyph))
if errI != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB37_NUM", Detail: "idx must be an integer"}
}
return true, GlyphDatum(lib37ListGet(args[0].Glyph, args[1].Glyph, idx)), nil

case "lib38_lst_has":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB38_ARG", Detail: "lib38_lst_has(list, sep, item) requires three strings"}
}
return true, GlyphDatum(lib38ListHas(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "lib39_list_asc":
if len(args) < 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB39_ARG", Detail: "lib39_lst_sort(list, sep, [order]) requires strings"}
}
ord := "asc"
if len(args) >= 3 && args[2].Kind == DatumGlyph {
ord = args[2].Glyph
}
return true, GlyphDatum(lib39ListSort(args[0].Glyph, args[1].Glyph, ord)), nil

case "lib40_list_once":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB40_ARG", Detail: "lib40_lst_uniq(list, [sep]) requires a string"}
}
sep := ","
if len(args) >= 2 && args[1].Kind == DatumGlyph {
sep = args[1].Glyph
}
return true, GlyphDatum(lib40ListUnique(args[0].Glyph, sep)), nil

case "lib41_sec_hash":
if len(args) < 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB41_ARG", Detail: "lib41_sec_hash(mode, algo, target, [expected]) requires strings"}
}
exp := ""
if len(args) >= 4 && args[3].Kind == DatumGlyph {
exp = args[3].Glyph
}
return true, GlyphDatum(lib41SecHash(args[0].Glyph, args[1].Glyph, args[2].Glyph, exp)), nil

case "lib42_sec_key":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB42_ARG", Detail: "lib42_sec_key(bytes) requires a string integer"}
}
n, errN := strconv.Atoi(strings.TrimSpace(args[0].Glyph))
if errN != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB42_NUM", Detail: "bytes must be an integer"}
}
return true, GlyphDatum(lib42SecToken(n)), nil

case "lib43_sec_mask":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB43_ARG", Detail: "lib43_sec_mask(secret, [keep]) requires a string"}
}
keep := 2
if len(args) >= 2 && args[1].Kind == DatumGlyph {
if k, err := strconv.Atoi(strings.TrimSpace(args[1].Glyph)); err == nil {
keep = k
}
}
return true, GlyphDatum(lib43SecMask(args[0].Glyph, keep)), nil

case "lib44_sec_perm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB44_ARG", Detail: "lib44_sec_perm(path) requires a string path"}
}
return true, GlyphDatum(lib44SecPermAudit(args[0].Glyph)), nil

case "lib45_sec_scan":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB45_ARG", Detail: "lib45_sec_scan(input) requires a string"}
}
return true, GlyphDatum(lib45SecPayloadScan(args[0].Glyph)), nil

case "lib46_sys_sec":
return true, GlyphDatum(lib46SysSecAudit()), nil

case "lib47_cps_guard":
if len(args) != 4 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph || args[3].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB47_ARG", Detail: "lib47_cps_guard(csv, min, max, maxStep) requires four strings"}
}
return true, GlyphDatum(lib47CpsSensorGuard(args[0].Glyph, args[1].Glyph, args[2].Glyph, args[3].Glyph)), nil

case "lib48_robo_guard":
if len(args) != 5 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph || args[3].Kind != DatumGlyph || args[4].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB48_ARG", Detail: "lib48_robo_guard(cmd, min, max, maxTemp, curTemp) requires five strings"}
}
return true, GlyphDatum(lib48RoboActuatorGuard(args[0].Glyph, args[1].Glyph, args[2].Glyph, args[3].Glyph, args[4].Glyph)), nil

case "lib49_net_sec":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB49_ARG", Detail: "lib49_net_sec(host:port, [timeoutMs]) requires a string"}
}
tMs := 2000
if len(args) >= 2 && args[1].Kind == DatumGlyph {
if v, err := strconv.Atoi(strings.TrimSpace(args[1].Glyph)); err == nil {
tMs = v
}
}
return true, GlyphDatum(lib49NetSecCheck(args[0].Glyph, tMs)), nil

case "lib50_neuro_sec":
if len(args) != 4 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph || args[3].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB50_ARG", Detail: "lib50_neuro_sec(ratesCsv, weightsCsv, maxRate, maxWeight) requires four strings"}
}
return true, GlyphDatum(lib50NeuroSecAudit(args[0].Glyph, args[1].Glyph, args[2].Glyph, args[3].Glyph)), nil

case "bmm_pkg_list":
return true, GlyphDatum(libBmmPkgList()), nil

case "bmm_pkg_info":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_PKG_INFO_ARG", Detail: "bmm_pkg_info(pkgName) requires a string"}
}
return true, GlyphDatum(libBmmPkgInfo(args[0].Glyph)), nil

case "lib51_cpu_info":
return true, GlyphDatum(lib51HwCpuInfo()), nil

case "lib52_cpu_bind":
t := "0"
if len(args) >= 1 && args[0].Kind == DatumGlyph {
t = args[0].Glyph
}
return true, GlyphDatum(lib52HwCpuBind(t)), nil

case "lib53_to_asm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB53_ARG", Detail: "lib53_to_asm(spec) requires a string"}
}
return true, GlyphDatum(lib53HwToAsm(args[0].Glyph)), nil

case "lib54_net_mod":
act := "list"
iface := ""
if len(args) >= 1 && args[0].Kind == DatumGlyph {
act = args[0].Glyph
}
if len(args) >= 2 && args[1].Kind == DatumGlyph {
iface = args[1].Glyph
}
return true, GlyphDatum(lib54HwNetMod(act, iface)), nil

case "lib55_bt_mod":
act := "status"
if len(args) >= 1 && args[0].Kind == DatumGlyph {
act = args[0].Glyph
}
return true, GlyphDatum(lib55HwBtMod(act)), nil

case "lib56_wifi_mod":
act := "status"
dev := "wlan0"
if len(args) >= 1 && args[0].Kind == DatumGlyph {
act = args[0].Glyph
}
if len(args) >= 2 && args[1].Kind == DatumGlyph {
dev = args[1].Glyph
}
return true, GlyphDatum(lib56HwWifiMod(act, dev)), nil

case "lib57_chip_ctrl":
bus := "all"
if len(args) >= 1 && args[0].Kind == DatumGlyph {
bus = args[0].Glyph
}
return true, GlyphDatum(lib57HwChipCtrl(bus)), nil

case "lib58_sock_ctrl":
mode := "tcp_count"
addr := ""
if len(args) >= 1 && args[0].Kind == DatumGlyph {
mode = args[0].Glyph
}
if len(args) >= 2 && args[1].Kind == DatumGlyph {
addr = args[1].Glyph
}
return true, GlyphDatum(lib58HwSockCtrl(mode, addr)), nil

case "lib59_board_info":
return true, GlyphDatum(lib59HwBoardInfo()), nil

case "lib60_mod_map":
return true, GlyphDatum(lib60HwModMap()), nil

case "lib61_json_get":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB61_ARG", Detail: "lib61_json_get(jsonStr, keyPath) requires two strings"}
}
return true, GlyphDatum(lib61JsonGet(args[0].Glyph, args[1].Glyph)), nil

case "lib62_b64":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB62_ARG", Detail: "lib62_b64(mode, input) requires two strings"}
}
return true, GlyphDatum(lib62Base64(args[0].Glyph, args[1].Glyph)), nil

case "lib63_hex_codec":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB63_ARG", Detail: "lib63_hex_codec(mode, input) requires two strings"}
}
return true, GlyphDatum(lib63HexCodec(args[0].Glyph, args[1].Glyph)), nil

case "lib64_regex_match":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB64_ARG", Detail: "lib64_regex_match(mode, pattern, text) requires three strings"}
}
return true, GlyphDatum(lib64RegexMatch(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "lib65_fs_snap":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB65_ARG", Detail: "lib65_fs_snap(path, [prevSnap]) requires a string path"}
}
prev := ""
if len(args) >= 2 && args[1].Kind == DatumGlyph {
prev = args[1].Glyph
}
return true, GlyphDatum(lib65FsWatchSnap(args[0].Glyph, prev)), nil

case "lib66_fs_find":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB66_ARG", Detail: "lib66_fs_find(dir, [suffix]) requires a string dir"}
}
suf := ""
if len(args) >= 2 && args[1].Kind == DatumGlyph {
suf = args[1].Glyph
}
return true, GlyphDatum(lib66FsFind(args[0].Glyph, suf)), nil

case "lib67_net_ping":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB67_ARG", Detail: "lib67_net_ping(host:port, [timeoutMs]) requires a string"}
}
tMs := "2000"
if len(args) >= 2 && args[1].Kind == DatumGlyph {
tMs = args[1].Glyph
}
return true, GlyphDatum(lib67NetPingMs(args[0].Glyph, tMs)), nil

case "lib68_http_probe":
if len(args) < 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB68_ARG", Detail: "lib68_http_probe(url, [timeoutSec]) requires a string"}
}
sec := "4"
if len(args) >= 2 && args[1].Kind == DatumGlyph {
sec = args[1].Glyph
}
return true, GlyphDatum(lib68HttpProbe(args[0].Glyph, sec)), nil

case "lib69_hmac":
if len(args) < 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB69_ARG", Detail: "lib69_hmac(mode, secret, msg, [expectedHex]) requires at least three strings"}
}
exp := ""
if len(args) >= 4 && args[3].Kind == DatumGlyph {
exp = args[3].Glyph
}
return true, GlyphDatum(lib69HmacSha256(args[0].Glyph, args[1].Glyph, args[2].Glyph, exp)), nil

case "lib70_spike_rate":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_LIB70_ARG", Detail: "lib70_spike_rate(spikesCsv, windowSec, maxHz) requires three strings"}
}
return true, GlyphDatum(lib70SpikeRateCheck(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

case "nexus_mesh":
d := "/data/data/com.termux/files/home/bmm_engine"
if len(args) >= 1 && args[0].Kind == DatumGlyph && args[0].Glyph != "" {
d = args[0].Glyph
}
return true, GlyphDatum(nexusMeshMap(d)), nil

case "nexus_bridge":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_NEXUS_BRIDGE_ARG", Detail: "nexus_bridge(spec) requires a string"}
}
return true, GlyphDatum(nexusLibBridge(args[0].Glyph)), nil

case "nexus_sys":
b := "/data/data/com.termux/files/usr/bin/bmm_lang"
if len(args) >= 1 && args[0].Kind == DatumGlyph && args[0].Glyph != "" {
b = args[0].Glyph
}
return true, GlyphDatum(nexusSysLink(b)), nil

case "p2_futura":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P2_FUTURA_ARG", Detail: "p2_futura(expr) requires a string"}
}
return true, GlyphDatum(bmmPart2Translate(args[0].Glyph)), nil

case "p2_alias":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P2_ALIAS_ARG", Detail: "p2_alias(cmd) requires a string"}
}
return true, GlyphDatum(bmmPart2Alias(args[0].Glyph)), nil

case "p3_eval":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P3_EVAL_ARG", Detail: "p3_eval(code) requires a string"}
}
return true, GlyphDatum(p3Eval(args[0].Glyph)), nil

case "p3_compile":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P3_COMPILE_ARG", Detail: "p3_compile(code) requires a string"}
}
return true, GlyphDatum(p3Compile(args[0].Glyph)), nil

case "p3_lexicon":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P3_LEX_ARG", Detail: "p3_lexicon(text) requires a string"}
}
return true, GlyphDatum(p3LexiconScan(args[0].Glyph)), nil

case "p3_color":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_P3_COLOR_ARG", Detail: "p3_color(text) requires a string"}
}
return true, GlyphDatum(p3ColorizeDynamic(args[0].Glyph)), nil

case "calc_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CALC_ARG", Detail: "calc_bmm(expr) requires glyph expression", Hint: "pass math expression like (10 + 5) * 2"}
		}
		res, err := evalMathExpr(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_CALC_EVAL", Detail: err.Error(), Hint: "check arithmetic operators and parentheses"}
		}
		fmtStr := strconv.FormatFloat(res, 'f', -1, 64)
		return true, GlyphDatum(fmtStr), nil

	case "rx_match_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_RXM_ARG", Detail: "rx_match_bmm(pattern, text) requires two glyphs", Hint: "pass regex pattern and text"}
		}
		re, err := regexp.Compile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_RXM_PAT", Detail: err.Error(), Hint: "invalid regular expression pattern"}
		}
		return true, FlagDatum(re.MatchString(args[1].Glyph)), nil

	case "rx_find_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_RXF_ARG", Detail: "rx_find_bmm(pattern, text) requires two glyphs", Hint: "pass regex pattern and text"}
		}
		re, err := regexp.Compile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_RXF_PAT", Detail: err.Error(), Hint: "invalid regular expression pattern"}
		}
		return true, GlyphDatum(re.FindString(args[1].Glyph)), nil

	case "str_trim_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_STRM_ARG", Detail: "str_trim_bmm(text) requires glyph", Hint: "pass string to trim"}
		}
		return true, GlyphDatum(strings.TrimSpace(args[0].Glyph)), nil

	case "str_has_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "RUNTIME_ENGINE", Code: "ERR_STRH_ARG", Detail: "str_has_bmm(text, sub) requires two glyphs", Hint: "pass text and substring"}
		}
		return true, FlagDatum(strings.Contains(args[0].Glyph, args[1].Glyph)), nil
	}

	return false, VoidDatum(), nil
}
