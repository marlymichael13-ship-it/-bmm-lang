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
