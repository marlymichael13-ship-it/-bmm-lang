package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var bmmBenchStart = make(map[string]time.Time)

func EvalAllInOneLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "sys_exec_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_SYSEXEC_ARG", Detail: "sys_exec_bmm(cmd) requires glyph command", Hint: "pass shell command string"}
		}
		cmd := exec.Command("sh", "-c", args[0].Glyph)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return true, GlyphDatum(strings.TrimSpace(string(out) + "\nERR:" + err.Error())), nil
		}
		return true, GlyphDatum(strings.TrimSpace(string(out))), nil

	case "env_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ENVGET_ARG", Detail: "env_get_bmm(key) requires glyph key", Hint: "pass env var name"}
		}
		return true, GlyphDatum(os.Getenv(args[0].Glyph)), nil

	case "env_set_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ENVSET_ARG", Detail: "env_set_bmm(key, val) requires two glyphs", Hint: "pass key and value"}
		}
		if err := os.Setenv(args[0].Glyph, args[1].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ENVSET_IO", Detail: err.Error(), Hint: "check env key"}
		}
		return true, FlagDatum(true), nil

	case "loop_run_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOOP_ARG", Detail: "loop_run_bmm(cmd, count, sleep_ms) requires three glyphs", Hint: "pass command, count, and sleep_ms"}
		}
		cnt, err := strconv.Atoi(args[1].Glyph)
		if err != nil || cnt < 1 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOOP_CNT", Detail: "invalid count", Hint: "pass positive integer string"}
		}
		ms, err := strconv.ParseInt(args[2].Glyph, 10, 64)
		if err != nil || ms < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOOP_MS", Detail: "invalid sleep_ms", Hint: "pass non-negative integer string"}
		}
		var results []string
		for i := 0; i < cnt; i++ {
			out, _ := exec.Command("sh", "-c", args[0].Glyph).CombinedOutput()
			results = append(results, strings.TrimSpace(string(out)))
			if i < cnt-1 && ms > 0 {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		}
		return true, GlyphDatum(strings.Join(results, " | ")), nil

	case "json_stringify_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_JSONSTR_ARG", Detail: "json_stringify_bmm(text) requires glyph", Hint: "pass raw string"}
		}
		b, err := json.Marshal(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_JSONSTR_ENC", Detail: err.Error(), Hint: "check input string"}
		}
		return true, GlyphDatum(string(b)), nil

	case "json_parse_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_JSONPRS_ARG", Detail: "json_parse_bmm(raw) requires glyph", Hint: "pass json string"}
		}
		var v any
		if err := json.Unmarshal([]byte(args[0].Glyph), &v); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_JSONPRS_ERR", Detail: err.Error(), Hint: "check valid json syntax"}
		}
		b, err := json.Marshal(v)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_JSONPRS_OUT", Detail: err.Error(), Hint: "failed formatting json"}
		}
		return true, GlyphDatum(string(b)), nil

	case "pick_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_PICK_ARG", Detail: "pick_bmm(json_str, key) requires two glyphs", Hint: "pass json object string and key"}
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(args[0].Glyph), &obj); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_PICK_JSON", Detail: err.Error(), Hint: "pass valid json object"}
		}
		val, ok := obj[args[1].Glyph]
		if !ok {
			return true, GlyphDatum(""), nil
		}
		switch t := val.(type) {
		case string:
			return true, GlyphDatum(t), nil
		default:
			b, _ := json.Marshal(t)
			return true, GlyphDatum(string(b)), nil
		}

	case "count_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_CNT_ARG", Detail: "count_bmm(text, sub) requires two glyphs", Hint: "pass text and substring"}
		}
		if args[1].Glyph == "" {
			return true, GlyphDatum("0"), nil
		}
		return true, GlyphDatum(strconv.Itoa(strings.Count(args[0].Glyph, args[1].Glyph))), nil

	case "cli_args_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ARGS_ARG", Detail: "cli_args_bmm() takes no arguments", Hint: "call cli_args_bmm()"}
		}
		return true, GlyphDatum(strings.Join(os.Args[1:], " ")), nil

	case "ansi_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ANSI_ARG", Detail: "ansi_bmm(color, text) requires two glyphs", Hint: "color: green, red, yellow, cyan, bold"}
		}
		var code string
		switch strings.ToLower(args[0].Glyph) {
		case "green":
			code = "\033[32m"
		case "red":
			code = "\033[31m"
		case "yellow":
			code = "\033[33m"
		case "cyan":
			code = "\033[36m"
		case "bold":
			code = "\033[1m"
		default:
			code = ""
		}
		if code == "" {
			return true, GlyphDatum(args[1].Glyph), nil
		}
		return true, GlyphDatum(code + args[1].Glyph + "\033[0m"), nil

	case "prompt_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_PRMT_ARG", Detail: "prompt_bmm(label) requires glyph label", Hint: "pass prompt text"}
		}
		fmt.Print(args[0].Glyph)
		reader := bufio.NewReader(os.Stdin)
		lineStr, err := reader.ReadString('\n')
		if err != nil && len(lineStr) == 0 {
			return true, GlyphDatum(""), nil
		}
		return true, GlyphDatum(strings.TrimRight(lineStr, "\r\n")), nil

	case "assert_eq_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ASSETEQ_ARG", Detail: "assert_eq_bmm(actual, expected) requires two glyphs", Hint: "pass actual and expected strings"}
		}
		if args[0].Glyph != args[1].Glyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ASSERT_FAIL", Detail: "actual=" + args[0].Glyph + " expected=" + args[1].Glyph, Hint: "values do not match"}
		}
		return true, FlagDatum(true), nil

	case "assert_true_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ASSERTTR_ARG", Detail: "assert_true_bmm(val) requires glyph", Hint: "pass truthy string like TRUE or 1"}
		}
		v := strings.ToLower(strings.TrimSpace(args[0].Glyph))
		ok := v == "true" || v == "1" || v == "yes" || v == "ok"
		if !ok {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ASSERT_FALSE", Detail: "value is not truthy: " + args[0].Glyph, Hint: "check condition"}
		}
		return true, FlagDatum(true), nil

	case "bench_start_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_BENCH_ARG", Detail: "bench_start_bmm(tag) requires glyph tag", Hint: "pass benchmark label"}
		}
		bmmBenchStart[args[0].Glyph] = time.Now()
		return true, FlagDatum(true), nil

	case "bench_elapsed_ms_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_ELAPSED_ARG", Detail: "bench_elapsed_ms_bmm(tag) requires glyph tag", Hint: "pass benchmark label"}
		}
		st, ok := bmmBenchStart[args[0].Glyph]
		if !ok {
			return true, GlyphDatum("0"), nil
		}
		ms := time.Since(st).Milliseconds()
		return true, GlyphDatum(strconv.FormatInt(ms, 10)), nil

	case "log_write_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOG_ARG", Detail: "log_write_bmm(path, level, msg) requires three glyphs", Hint: "pass log file path, level, and message"}
		}
		f, err := os.OpenFile(args[0].Glyph, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOG_IO", Detail: err.Error(), Hint: "check log file path"}
		}
		defer f.Close()
		entry := time.Now().Format(time.RFC3339) + " [" + strings.ToUpper(args[1].Glyph) + "] " + args[2].Glyph + "\n"
		if _, err := f.WriteString(entry); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "ALL_IN_ONE", Code: "ERR_LOG_WRITE", Detail: err.Error(), Hint: "failed writing log entry"}
		}
		return true, FlagDatum(true), nil
	}

	return false, VoidDatum(), nil
}
