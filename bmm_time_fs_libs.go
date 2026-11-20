package main

import (
	"os"
	"strconv"
	"time"
)

func EvalTimeFSLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "now_iso_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_NISO_ARG", Detail: "now_iso_bmm() takes no arguments", Hint: "call now_iso_bmm()"}
		}
		return true, GlyphDatum(time.Now().Format(time.RFC3339)), nil

	case "now_unix_bmm":
		if len(args) != 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_NUNX_ARG", Detail: "now_unix_bmm() takes no arguments", Hint: "call now_unix_bmm()"}
		}
		return true, GlyphDatum(strconv.FormatInt(time.Now().Unix(), 10)), nil

	case "sleep_ms_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_SLPM_ARG", Detail: "sleep_ms_bmm(ms) requires glyph milliseconds", Hint: "pass integer milliseconds as string like \"250\""}
		}
		ms, err := strconv.ParseInt(args[0].Glyph, 10, 64)
		if err != nil || ms < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_SLPM_VAL", Detail: "invalid milliseconds value", Hint: "pass non-negative integer string"}
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return true, FlagDatum(true), nil

	case "file_read_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FRD_ARG", Detail: "file_read_bmm(path) requires glyph path", Hint: "pass file path"}
		}
		data, err := os.ReadFile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FRD_IO", Detail: err.Error(), Hint: "check file path and read permissions"}
		}
		return true, GlyphDatum(string(data)), nil

	case "file_write_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FWR_ARG", Detail: "file_write_bmm(path, content) requires two glyphs", Hint: "pass file path and content string"}
		}
		if err := os.WriteFile(args[0].Glyph, []byte(args[1].Glyph), 0644); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FWR_IO", Detail: err.Error(), Hint: "check directory and write permissions"}
		}
		return true, FlagDatum(true), nil

	case "file_append_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FAP_ARG", Detail: "file_append_bmm(path, content) requires two glyphs", Hint: "pass file path and content to append"}
		}
		f, err := os.OpenFile(args[0].Glyph, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FAP_OPEN", Detail: err.Error(), Hint: "check file path and permissions"}
		}
		defer f.Close()
		if _, err := f.WriteString(args[1].Glyph); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FAP_IO", Detail: err.Error(), Hint: "failed appending content"}
		}
		return true, FlagDatum(true), nil

	case "file_exists_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "TIME_FS_LIB", Code: "ERR_FEX_ARG", Detail: "file_exists_bmm(path) requires glyph path", Hint: "pass file or folder path"}
		}
		_, err := os.Stat(args[0].Glyph)
		return true, FlagDatum(err == nil), nil
	}

	return false, VoidDatum(), nil
}

