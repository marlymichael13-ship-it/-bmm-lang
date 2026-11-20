package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func EvalFullStackLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "http_get_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HGET_ARG", Detail: "http_get_bmm(url, timeout_sec) requires two glyphs", Hint: "pass url and timeout in seconds"}
		}
		sec, err := strconv.Atoi(args[1].Glyph)
		if err != nil || sec < 1 {
			sec = 10
		}
		client := &http.Client{Timeout: time.Duration(sec) * time.Second}
		resp, err := client.Get(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HGET_REQ", Detail: err.Error(), Hint: "check network connection and url"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HGET_BODY", Detail: err.Error(), Hint: "failed reading response body"}
		}
		return true, GlyphDatum(string(body)), nil

	case "http_post_bmm":
		if len(args) != 4 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph || args[3].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HPOST_ARG", Detail: "http_post_bmm(url, content_type, payload, timeout_sec) requires four glyphs", Hint: "pass url, content type, payload, and timeout"}
		}
		sec, err := strconv.Atoi(args[3].Glyph)
		if err != nil || sec < 1 {
			sec = 10
		}
		client := &http.Client{Timeout: time.Duration(sec) * time.Second}
		resp, err := client.Post(args[0].Glyph, args[1].Glyph, bytes.NewBufferString(args[2].Glyph))
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HPOST_REQ", Detail: err.Error(), Hint: "check endpoint and network"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_HPOST_BODY", Detail: err.Error(), Hint: "failed reading response"}
		}
		return true, GlyphDatum(string(body)), nil

	case "sha256_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_SHA_ARG", Detail: "sha256_bmm(text) requires glyph", Hint: "pass string to hash"}
		}
		sum := sha256.Sum256([]byte(args[0].Glyph))
		return true, GlyphDatum(hex.EncodeToString(sum[:])), nil

	case "b64_enc_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_B64E_ARG", Detail: "b64_enc_bmm(text) requires glyph", Hint: "pass plain text"}
		}
		return true, GlyphDatum(base64.StdEncoding.EncodeToString([]byte(args[0].Glyph))), nil

	case "b64_dec_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_B64D_ARG", Detail: "b64_dec_bmm(b64_text) requires glyph", Hint: "pass base64 string"}
		}
		raw, err := base64.StdEncoding.DecodeString(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_B64D_ERR", Detail: err.Error(), Hint: "invalid base64 input"}
		}
		return true, GlyphDatum(string(raw)), nil

	case "path_join_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_PJOIN_ARG", Detail: "path_join_bmm(dir, name) requires two glyphs", Hint: "pass base path and child name"}
		}
		return true, GlyphDatum(filepath.Join(args[0].Glyph, args[1].Glyph)), nil

	case "path_base_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_PBASE_ARG", Detail: "path_base_bmm(path) requires glyph", Hint: "pass file or folder path"}
		}
		return true, GlyphDatum(filepath.Base(args[0].Glyph)), nil

	case "json_len_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JLEN_ARG", Detail: "json_len_bmm(json_arr) requires glyph", Hint: "pass json array string"}
		}
		var arr []any
		if err := json.Unmarshal([]byte(args[0].Glyph), &arr); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JLEN_PRS", Detail: err.Error(), Hint: "pass valid json array"}
		}
		return true, GlyphDatum(strconv.Itoa(len(arr))), nil

	case "json_at_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JAT_ARG", Detail: "json_at_bmm(json_arr, idx) requires two glyphs", Hint: "pass json array and index"}
		}
		idx, err := strconv.Atoi(args[1].Glyph)
		if err != nil || idx < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JAT_IDX", Detail: "invalid index", Hint: "pass non-negative integer index"}
		}
		var arr []any
		if err := json.Unmarshal([]byte(args[0].Glyph), &arr); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JAT_PRS", Detail: err.Error(), Hint: "pass valid json array"}
		}
		if idx >= len(arr) {
			return true, GlyphDatum(""), nil
		}
		switch v := arr[idx].(type) {
		case string:
			return true, GlyphDatum(v), nil
		default:
			b, _ := json.Marshal(v)
			return true, GlyphDatum(string(b)), nil
		}

	case "json_keys_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JKEYS_ARG", Detail: "json_keys_bmm(json_obj) requires glyph", Hint: "pass json object string"}
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[0].Glyph), &obj); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_JKEYS_PRS", Detail: err.Error(), Hint: "pass valid json object"}
		}
		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return true, GlyphDatum(strings.Join(keys, ",")), nil

	case "list_dir_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_LDIR_ARG", Detail: "list_dir_bmm(dir_path) requires glyph", Hint: "pass directory path"}
		}
		entries, err := os.ReadDir(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_LDIR_IO", Detail: err.Error(), Hint: "check directory path and permissions"}
		}
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return true, GlyphDatum(strings.Join(names, "\n")), nil

	case "run_script_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_RUNSCR_ARG", Detail: "run_script_bmm(script_path) requires glyph", Hint: "pass path to .bmm script file"}
		}
		binPath, err := os.Executable()
		if err != nil {
			binPath = "bmm_lang"
		}
		out, err := exec.Command(binPath, args[0].Glyph).CombinedOutput()
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "FULL_STACK", Code: "ERR_RUNSCR_EXEC", Detail: strings.TrimSpace(string(out) + " " + err.Error()), Hint: "check script syntax and path"}
		}
		return true, GlyphDatum(strings.TrimSpace(string(out))), nil
	}

	return false, VoidDatum(), nil
}
