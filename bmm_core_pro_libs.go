package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	bmmStoreMu   sync.RWMutex
	bmmStoreData = make(map[string]string)
)

func bmmCopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func EvalCoreProLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "mkdir_all_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_MKDIR_ARG", Detail: "mkdir_all_bmm(dir_path) requires glyph", Hint: "pass target directory path"}
		}
		if err := os.MkdirAll(args[0].Glyph, 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_MKDIR_IO", Detail: err.Error(), Hint: "check parent path and permissions"}
		}
		return true, FlagDatum(true), nil

	case "file_copy_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FCP_ARG", Detail: "file_copy_bmm(src, dst) requires two glyphs", Hint: "pass source and destination paths"}
		}
		if err := bmmCopyFile(args[0].Glyph, args[1].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FCP_IO", Detail: err.Error(), Hint: "check source file and destination permissions"}
		}
		return true, FlagDatum(true), nil

	case "file_move_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FMV_ARG", Detail: "file_move_bmm(src, dst) requires two glyphs", Hint: "pass source and destination paths"}
		}
		if err := os.MkdirAll(filepath.Dir(args[1].Glyph), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FMV_DIR", Detail: err.Error(), Hint: "check destination directory"}
		}
		if err := os.Rename(args[0].Glyph, args[1].Glyph); err != nil {
			if copyErr := bmmCopyFile(args[0].Glyph, args[1].Glyph); copyErr != nil {
				return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FMV_IO", Detail: err.Error() + "; " + copyErr.Error(), Hint: "check paths and permissions"}
			}
			if rmErr := os.Remove(args[0].Glyph); rmErr != nil {
				return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FMV_RM", Detail: rmErr.Error(), Hint: "copied file but failed removing source"}
			}
		}
		return true, FlagDatum(true), nil

	case "file_remove_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FRM_ARG", Detail: "file_remove_bmm(path) requires glyph", Hint: "pass file or directory path"}
		}
		if err := os.RemoveAll(args[0].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FRM_IO", Detail: err.Error(), Hint: "check path and permissions"}
		}
		return true, FlagDatum(true), nil

	case "file_size_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FSZ_ARG", Detail: "file_size_bmm(path) requires glyph", Hint: "pass file path"}
		}
		info, err := os.Stat(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_FSZ_IO", Detail: err.Error(), Hint: "check if file exists"}
		}
		return true, GlyphDatum(strconv.FormatInt(info.Size(), 10)), nil

	case "path_ext_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_PEXT_ARG", Detail: "path_ext_bmm(path) requires glyph", Hint: "pass file path"}
		}
		return true, GlyphDatum(filepath.Ext(args[0].Glyph)), nil

	case "read_lines_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RLN_ARG", Detail: "read_lines_bmm(path, start_line, max_lines) requires three glyphs", Hint: "pass path, start_line (1-based), and max_lines"}
		}
		startLine, err := strconv.Atoi(args[1].Glyph)
		if err != nil || startLine < 1 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RLN_START", Detail: "start_line must be >= 1", Hint: "pass positive integer string"}
		}
		maxLines, err := strconv.Atoi(args[2].Glyph)
		if err != nil || maxLines < 1 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RLN_MAX", Detail: "max_lines must be >= 1", Hint: "pass positive integer string"}
		}
		f, err := os.Open(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RLN_IO", Detail: err.Error(), Hint: "check file path and read permissions"}
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
		var out []string
		cur := 0
		for scanner.Scan() {
			cur++
			if cur < startLine {
				continue
			}
			out = append(out, scanner.Text())
			if len(out) >= maxLines {
				break
			}
		}
		if err := scanner.Err(); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RLN_SCAN", Detail: err.Error(), Hint: "failed scanning file lines"}
		}
		return true, GlyphDatum(strings.Join(out, "\n")), nil

	case "str_replace_all_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SRPL_ARG", Detail: "str_replace_all_bmm(text, old, new) requires three glyphs", Hint: "pass source text, target substring, and replacement"}
		}
		return true, GlyphDatum(strings.ReplaceAll(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil

	case "str_split_at_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SSPL_ARG", Detail: "str_split_at_bmm(text, sep, idx) requires three glyphs", Hint: "pass text, separator, and 0-based index"}
		}
		idx, err := strconv.Atoi(args[2].Glyph)
		if err != nil || idx < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SSPL_IDX", Detail: "invalid index", Hint: "pass non-negative integer string"}
		}
		parts := strings.Split(args[0].Glyph, args[1].Glyph)
		if idx >= len(parts) {
			return true, GlyphDatum(""), nil
		}
		return true, GlyphDatum(parts[idx]), nil

	case "str_join_sep_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SJN_ARG", Detail: "str_join_sep_bmm(a, b, sep) requires three glyphs", Hint: "pass first string, second string, and separator"}
		}
		return true, GlyphDatum(args[0].Glyph + args[2].Glyph + args[1].Glyph), nil

	case "rand_hex_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RHEX_ARG", Detail: "rand_hex_bmm(num_bytes) requires glyph", Hint: "pass byte count between 1 and 1024"}
		}
		nBytes, err := strconv.Atoi(args[0].Glyph)
		if err != nil || nBytes < 1 || nBytes > 1024 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RHEX_LEN", Detail: "num_bytes must be between 1 and 1024", Hint: "pass valid integer byte count"}
		}
		buf := make([]byte, nBytes)
		if _, err := rand.Read(buf); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_RHEX_RNG", Detail: err.Error(), Hint: "crypto random source unavailable"}
		}
		return true, GlyphDatum(hex.EncodeToString(buf)), nil

	case "uuid_v4_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_UUID_ARG", Detail: "uuid_v4_bmm() takes no arguments", Hint: "call uuid_v4_bmm()"}
		}
		var b [16]byte
		if _, err := rand.Read(b[:]); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_UUID_RNG", Detail: err.Error(), Hint: "crypto random source unavailable"}
		}
		b[6] = (b[6] & 0x0f) | 0x40
		b[8] = (b[8] & 0x3f) | 0x80
		u := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
		return true, GlyphDatum(u), nil

	case "sha256_verify_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SHAV_ARG", Detail: "sha256_verify_bmm(text, expected_hex) requires two glyphs", Hint: "pass plain text and expected hex digest"}
		}
		sum := sha256.Sum256([]byte(args[0].Glyph))
		actualHex := hex.EncodeToString(sum[:])
		expectedHex := strings.ToLower(strings.TrimSpace(args[1].Glyph))
		if len(actualHex) != len(expectedHex) {
			return true, FlagDatum(false), nil
		}
		ok := subtle.ConstantTimeCompare([]byte(actualHex), []byte(expectedHex)) == 1
		return true, FlagDatum(ok), nil

	case "json_is_valid_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_JVAL_ARG", Detail: "json_is_valid_bmm(raw) requires glyph", Hint: "pass json string"}
		}
		return true, FlagDatum(json.Valid([]byte(args[0].Glyph))), nil

	case "exec_timeout_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_EXTO_ARG", Detail: "exec_timeout_bmm(cmd, timeout_sec) requires two glyphs", Hint: "pass shell command and timeout in seconds"}
		}
		sec, err := strconv.Atoi(args[1].Glyph)
		if err != nil || sec < 1 {
			sec = 5
		}
		cmd := exec.Command("sh", "-c", args[0].Glyph)
		if err := cmd.Start(); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_EXTO_START", Detail: err.Error(), Hint: "failed starting command"}
		}
		done := make(chan error, 1)
		var outBuf strings.Builder
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf
		go func() {
			done <- cmd.Run()
		}()
		select {
		case err := <-done:
			if err != nil {
				return true, GlyphDatum(strings.TrimSpace(outBuf.String() + "\nERR:" + err.Error())), nil
			}
			return true, GlyphDatum(strings.TrimSpace(outBuf.String())), nil
		case <-time.After(time.Duration(sec) * time.Second):
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_EXTO_TIMEOUT", Detail: "command exceeded timeout of " + strconv.Itoa(sec) + "s", Hint: "increase timeout or optimize command"}
		}

	case "store_set_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STSET_ARG", Detail: "store_set_bmm(key, val) requires two glyphs", Hint: "pass key and value"}
		}
		bmmStoreMu.Lock()
		bmmStoreData[args[0].Glyph] = args[1].Glyph
		bmmStoreMu.Unlock()
		return true, FlagDatum(true), nil

	case "store_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STGET_ARG", Detail: "store_get_bmm(key) requires glyph", Hint: "pass key name"}
		}
		bmmStoreMu.RLock()
		v := bmmStoreData[args[0].Glyph]
		bmmStoreMu.RUnlock()
		return true, GlyphDatum(v), nil

	case "store_del_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STDEL_ARG", Detail: "store_del_bmm(key) requires glyph", Hint: "pass key name"}
		}
		bmmStoreMu.Lock()
		delete(bmmStoreData, args[0].Glyph)
		bmmStoreMu.Unlock()
		return true, FlagDatum(true), nil

	case "store_keys_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STKEYS_ARG", Detail: "store_keys_bmm() takes no arguments", Hint: "call store_keys_bmm()"}
		}
		bmmStoreMu.RLock()
		keys := make([]string, 0, len(bmmStoreData))
		for k := range bmmStoreData {
			keys = append(keys, k)
		}
		bmmStoreMu.RUnlock()
		sort.Strings(keys)
		return true, GlyphDatum(strings.Join(keys, ",")), nil

	case "store_save_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STSAV_ARG", Detail: "store_save_bmm(file_path) requires glyph", Hint: "pass path to json store file"}
		}
		bmmStoreMu.RLock()
		snap := make(map[string]string, len(bmmStoreData))
		for k, v := range bmmStoreData {
			snap[k] = v
		}
		bmmStoreMu.RUnlock()
		raw, err := json.MarshalIndent(snap, "", "  ")
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STSAV_ENC", Detail: err.Error(), Hint: "failed encoding store data"}
		}
		if err := os.MkdirAll(filepath.Dir(args[0].Glyph), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STSAV_DIR", Detail: err.Error(), Hint: "check target directory"}
		}
		if err := os.WriteFile(args[0].Glyph, raw, 0644); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STSAV_IO", Detail: err.Error(), Hint: "check write permissions"}
		}
		return true, FlagDatum(true), nil

	case "store_load_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STLD_ARG", Detail: "store_load_bmm(file_path) requires glyph", Hint: "pass path to json store file"}
		}
		raw, err := os.ReadFile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STLD_IO", Detail: err.Error(), Hint: "check store file path"}
		}
		var loaded map[string]string
		if err := json.Unmarshal(raw, &loaded); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_STLD_JSON", Detail: err.Error(), Hint: "invalid json store format"}
		}
		bmmStoreMu.Lock()
		bmmStoreData = loaded
		bmmStoreMu.Unlock()
		return true, FlagDatum(true), nil

	case "sys_info_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SYSINF_ARG", Detail: "sys_info_bmm() takes no arguments", Hint: "call sys_info_bmm()"}
		}
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		host, _ := os.Hostname()
		info := map[string]any{
			"os":           runtime.GOOS,
			"arch":         runtime.GOARCH,
			"go_version":   runtime.Version(),
			"num_cpu":      runtime.NumCPU(),
			"goroutines":   runtime.NumGoroutine(),
			"heap_alloc_b": ms.HeapAlloc,
			"hostname":     host,
		}
		b, err := json.Marshal(info)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_SYSINF_ENC", Detail: err.Error(), Hint: "failed encoding system info"}
		}
		return true, GlyphDatum(string(b)), nil

	case "cwd_get_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_CWDG_ARG", Detail: "cwd_get_bmm() takes no arguments", Hint: "call cwd_get_bmm()"}
		}
		wd, err := os.Getwd()
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_CWDG_IO", Detail: err.Error(), Hint: "failed reading working directory"}
		}
		return true, GlyphDatum(wd), nil

	case "cwd_set_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_CWDS_ARG", Detail: "cwd_set_bmm(dir_path) requires glyph", Hint: "pass target directory path"}
		}
		if err := os.Chdir(args[0].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "CORE_PRO", Code: "ERR_CWDS_IO", Detail: err.Error(), Hint: "check target directory path"}
		}
		return true, FlagDatum(true), nil
	}

	return false, VoidDatum(), nil
}
