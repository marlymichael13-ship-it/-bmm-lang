package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func EvalCustomLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	// --- IO & FILESYSTEM ---
	case "read_file_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_READ_ARG", Detail: "read_file_bmm(path) requires glyph path", Hint: "pass file path string"}
		}
		b, err := os.ReadFile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_READ_IO", Detail: err.Error(), Hint: "check if file exists and is readable"}
		}
		return true, GlyphDatum(string(b)), nil

	case "write_file_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_WRITE_ARG", Detail: "write_file_bmm(path, text) requires two glyphs", Hint: "pass path and content"}
		}
		if err := os.WriteFile(args[0].Glyph, []byte(args[1].Glyph), 0644); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_WRITE_IO", Detail: err.Error(), Hint: "check folder permissions"}
		}
		return true, FlagDatum(true), nil

	case "append_file_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_APPEND_ARG", Detail: "append_file_bmm(path, text) requires two glyphs", Hint: "pass path and text"}
		}
		f, err := os.OpenFile(args[0].Glyph, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_APPEND_IO", Detail: err.Error(), Hint: "check file path"}
		}
		defer f.Close()
		if _, err := f.WriteString(args[1].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_APPEND_WRITE", Detail: err.Error(), Hint: "check storage space"}
		}
		return true, FlagDatum(true), nil

	case "file_exists_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_EXISTS_ARG", Detail: "file_exists_bmm(path) requires glyph path", Hint: "pass file path string"}
		}
		_, err := os.Stat(args[0].Glyph)
		return true, FlagDatum(err == nil), nil

	case "mkdir_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_MKDIR_ARG", Detail: "mkdir_bmm(path) requires glyph path", Hint: "pass directory path"}
		}
		if err := os.MkdirAll(args[0].Glyph, 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_MKDIR_IO", Detail: err.Error(), Hint: "check parent path permissions"}
		}
		return true, FlagDatum(true), nil

	case "remove_file_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_RM_ARG", Detail: "remove_file_bmm(path) requires glyph path", Hint: "pass target file path"}
		}
		if err := os.Remove(args[0].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_RM_IO", Detail: err.Error(), Hint: "check if file exists"}
		}
		return true, FlagDatum(true), nil

	case "list_dir_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_DIR_ARG", Detail: "list_dir_bmm(dir) requires glyph directory path", Hint: "pass directory path"}
		}
		entries, err := os.ReadDir(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_IO", Code: "ERR_DIR_IO", Detail: err.Error(), Hint: "check directory path"}
		}
		out := make([]Datum, 0, len(entries))
		for _, e := range entries {
			out = append(out, GlyphDatum(e.Name()))
		}
		return true, VecDatum(out), nil

	// --- TEXT & PARSING ---
	case "lines_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_LINES_ARG", Detail: "lines_bmm(str) requires glyph", Hint: "pass string"}
		}
		norm := strings.ReplaceAll(args[0].Glyph, "\r\n", "\n")
		parts := strings.Split(norm, "\n")
		out := make([]Datum, 0, len(parts))
		for _, p := range parts {
			out = append(out, GlyphDatum(p))
		}
		return true, VecDatum(out), nil

	case "trim_lines_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_TRIML_ARG", Detail: "trim_lines_bmm(vec) requires vec", Hint: "pass vector of glyphs"}
		}
		out := make([]Datum, 0, len(args[0].Vec))
		for _, it := range args[0].Vec {
			if it.Kind == DatumGlyph {
				out = append(out, GlyphDatum(strings.TrimSpace(it.Glyph)))
			} else {
				out = append(out, it.DeepCopy())
			}
		}
		return true, VecDatum(out), nil

	case "split_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_SPLIT_ARG", Detail: "split_bmm(s, sep) requires two glyphs", Hint: "pass string and delimiter"}
		}
		parts := strings.Split(args[0].Glyph, args[1].Glyph)
		out := make([]Datum, 0, len(parts))
		for _, p := range parts {
			out = append(out, GlyphDatum(p))
		}
		return true, VecDatum(out), nil

	case "join_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_JOIN_ARG", Detail: "join_bmm(vec, sep) requires (vec, glyph)", Hint: "pass vector and separator"}
		}
		parts := make([]string, 0, len(args[0].Vec))
		for _, it := range args[0].Vec {
			parts = append(parts, it.Stringify())
		}
		return true, GlyphDatum(strings.Join(parts, args[1].Glyph)), nil

	case "replace_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_REP_ARG", Detail: "replace_bmm(s, old, new) requires three glyphs", Hint: "pass text, target, and replacement"}
		}
		return true, GlyphDatum(strings.ReplaceAll(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

	case "contains_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_CNT_ARG", Detail: "contains_bmm(s, sub) requires two glyphs", Hint: "pass string and substring"}
		}
		return true, FlagDatum(strings.Contains(args[0].Glyph, args[1].Glyph)), nil

	case "starts_with_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_SW_ARG", Detail: "starts_with_bmm(s, prefix) requires two glyphs", Hint: "pass string and prefix"}
		}
		return true, FlagDatum(strings.HasPrefix(args[0].Glyph, args[1].Glyph)), nil

	case "ends_with_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_EW_ARG", Detail: "ends_with_bmm(s, suffix) requires two glyphs", Hint: "pass string and suffix"}
		}
		return true, FlagDatum(strings.HasSuffix(args[0].Glyph, args[1].Glyph)), nil

	case "substr_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumCycle || args[2].Kind != DatumCycle {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_TEXT", Code: "ERR_SUB_ARG", Detail: "substr_bmm(s, start, end) requires (glyph, cycle, cycle)", Hint: "pass string and start/end indices"}
		}
		runes := []rune(args[0].Glyph)
		a := int(args[1].Cycle)
		b := int(args[2].Cycle)
		if a < 0 {
			a = 0
		}
		if b > len(runes) {
			b = len(runes)
		}
		if a > b {
			return true, GlyphDatum(""), nil
		}
		return true, GlyphDatum(string(runes[a:b])), nil

	// --- JSON & KV ---
	case "json_get_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_JSON_GET", Detail: "json_get_bmm(json, key) requires two glyphs", Hint: "pass JSON string and key"}
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[0].Glyph), &obj); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_JSON_PARSE", Detail: err.Error(), Hint: "check JSON syntax"}
		}
		v, ok := obj[args[1].Glyph]
		if !ok {
			return true, VoidDatum(), nil
		}
		switch t := v.(type) {
		case string:
			return true, GlyphDatum(t), nil
		case bool:
			return true, FlagDatum(t), nil
		case float64:
			return true, CycleDatum(int64(t)), nil
		case nil:
			return true, VoidDatum(), nil
		default:
			raw, err := json.Marshal(t)
			if err != nil {
				return true, VoidDatum(), nil
			}
			return true, GlyphDatum(string(raw)), nil
		}

	case "json_set_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_JSON_SET", Detail: "json_set_bmm(json, key, val) requires (glyph, glyph, any)", Hint: "pass JSON object string, key, and value"}
		}
		obj := make(map[string]any)
		if strings.TrimSpace(args[0].Glyph) != "" {
			if err := json.Unmarshal([]byte(args[0].Glyph), &obj); err != nil {
				return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_JSON_PARSE", Detail: err.Error(), Hint: "pass valid JSON object string"}
			}
		}
		switch args[2].Kind {
		case DatumCycle:
			obj[args[1].Glyph] = args[2].Cycle
		case DatumFlag:
			obj[args[1].Glyph] = args[2].Flag
		case DatumVoid:
			obj[args[1].Glyph] = nil
		default:
			obj[args[1].Glyph] = args[2].Glyph
		}
		outBytes, err := json.Marshal(obj)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_JSON_ENC", Detail: err.Error(), Hint: "check value encoding"}
		}
		return true, GlyphDatum(string(outBytes)), nil

	case "kv_parse_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_DATA", Code: "ERR_KV_ARG", Detail: "kv_parse_bmm(raw) requires glyph text", Hint: "pass KEY=VALUE lines"}
		}
		m := make(map[string]Datum)
		for _, lineStr := range strings.Split(strings.ReplaceAll(args[0].Glyph, "\r\n", "\n"), "\n") {
			t := strings.TrimSpace(lineStr)
			if t == "" || strings.HasPrefix(t, "#") {
				continue
			}
			idx := strings.IndexByte(t, '=')
			if idx <= 0 {
				continue
			}
			k := strings.TrimSpace(t[:idx])
			v := strings.TrimSpace(t[idx+1:])
			if iv, err := strconv.ParseInt(v, 10, 64); err == nil {
				m[k] = CycleDatum(iv)
			} else if v == "true" || v == "TRUE" {
				m[k] = FlagDatum(true)
			} else if v == "false" || v == "FALSE" {
				m[k] = FlagDatum(false)
			} else {
				m[k] = GlyphDatum(strings.Trim(v, `"'`))
			}
		}
		return true, DictDatum(m), nil

	// --- MATH, VEC & SORTING ---
	case "sum_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_MATH", Code: "ERR_SUM_ARG", Detail: "sum_vec_bmm(vec) requires cycle vec", Hint: "pass vector of numbers"}
		}
		var total int64
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				total += it.Cycle
			}
		}
		return true, CycleDatum(total), nil

	case "avg_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec || len(args[0].Vec) == 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_MATH", Code: "ERR_AVG_ARG", Detail: "avg_vec_bmm(vec) requires non-empty cycle vec", Hint: "pass non-empty number vector"}
		}
		var total int64
		var cnt int64
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				total += it.Cycle
				cnt++
			}
		}
		if cnt == 0 {
			return true, CycleDatum(0), nil
		}
		return true, CycleDatum(total / cnt), nil

	case "filter_gt_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec || args[1].Kind != DatumCycle {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_MATH", Code: "ERR_FLT_ARG", Detail: "filter_gt_bmm(vec, threshold) requires (vec, cycle)", Hint: "pass number vector and threshold"}
		}
		out := make([]Datum, 0)
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle && it.Cycle > args[1].Cycle {
				out = append(out, CycleDatum(it.Cycle))
			}
		}
		return true, VecDatum(out), nil

	case "map_add_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec || args[1].Kind != DatumCycle {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_MATH", Code: "ERR_MAPA_ARG", Detail: "map_add_bmm(vec, delta) requires (vec, cycle)", Hint: "pass number vector and delta"}
		}
		out := make([]Datum, 0, len(args[0].Vec))
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				out = append(out, CycleDatum(it.Cycle+args[1].Cycle))
			} else {
				out = append(out, it.DeepCopy())
			}
		}
		return true, VecDatum(out), nil

	case "sort_asc_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_MATH", Code: "ERR_SORT_ARG", Detail: "sort_asc_bmm(vec) requires cycle vec", Hint: "pass vector of numbers"}
		}
		nums := make([]int64, 0, len(args[0].Vec))
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				nums = append(nums, it.Cycle)
			}
		}
		sort.Slice(nums, func(i, j int) bool { return nums[i] < nums[j] })
		out := make([]Datum, 0, len(nums))
		for _, n := range nums {
			out = append(out, CycleDatum(n))
		}
		return true, VecDatum(out), nil

	// --- CRYPTO, ENCODING & SYSTEM ---
	case "sha256_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SEC", Code: "ERR_SHA_ARG", Detail: "sha256_bmm(text) requires glyph", Hint: "pass text string"}
		}
		sum := sha256.Sum256([]byte(args[0].Glyph))
		return true, GlyphDatum(hex.EncodeToString(sum[:])), nil

	case "b64_enc_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SEC", Code: "ERR_B64E_ARG", Detail: "b64_enc_bmm(text) requires glyph", Hint: "pass text string"}
		}
		return true, GlyphDatum(base64.StdEncoding.EncodeToString([]byte(args[0].Glyph))), nil

	case "b64_dec_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SEC", Code: "ERR_B64D_ARG", Detail: "b64_dec_bmm(encoded) requires glyph", Hint: "pass base64 string"}
		}
		raw, err := base64.StdEncoding.DecodeString(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SEC", Code: "ERR_B64D_DEC", Detail: err.Error(), Hint: "ensure valid base64 input"}
		}
		return true, GlyphDatum(string(raw)), nil

	case "env_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SYS", Code: "ERR_ENV_ARG", Detail: "env_get_bmm(name) requires glyph", Hint: "pass environment variable name"}
		}
		return true, GlyphDatum(os.Getenv(args[0].Glyph)), nil

	case "now_unix_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SYS", Code: "ERR_NOW_ARG", Detail: "now_unix_bmm() takes 0 arguments", Hint: "call now_unix_bmm()"}
		}
		return true, CycleDatum(time.Now().Unix()), nil

	case "cwd_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SYS", Code: "ERR_CWD_ARG", Detail: "cwd_bmm() takes 0 arguments", Hint: "call cwd_bmm()"}
		}
		wd, err := os.Getwd()
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "LIB_SYS", Code: "ERR_CWD_IO", Detail: err.Error(), Hint: "check current directory"}
		}
		return true, GlyphDatum(filepath.Clean(wd)), nil
	}

	return false, VoidDatum(), nil
}

