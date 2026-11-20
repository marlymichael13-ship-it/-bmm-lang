package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func EvalExtra4Libs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "file_size_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_IO", Code: "ERR_FSIZE_ARG", Detail: "file_size_bmm(path) requires glyph path", Hint: "pass existing file path"}
		}
		info, err := os.Stat(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_IO", Code: "ERR_FSIZE_IO", Detail: err.Error(), Hint: "verify file exists"}
		}
		return true, CycleDatum(info.Size()), nil

	case "is_dir_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_IO", Code: "ERR_ISDIR_ARG", Detail: "is_dir_bmm(path) requires glyph path", Hint: "pass path"}
		}
		info, err := os.Stat(args[0].Glyph)
		if err != nil {
			return true, FlagDatum(false), nil
		}
		return true, FlagDatum(info.IsDir()), nil

	case "dict_has_bmm":
		if len(args) != 2 || args[0].Kind != DatumDict || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_DICT", Code: "ERR_DHAS_ARG", Detail: "dict_has_bmm(dict, key) requires (dict, glyph)", Hint: "pass dict and key"}
		}
		_, ok := args[0].Dict[args[1].Glyph]
		return true, FlagDatum(ok), nil

	case "dict_del_bmm":
		if len(args) != 2 || args[0].Kind != DatumDict || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_DICT", Code: "ERR_DDEL_ARG", Detail: "dict_del_bmm(dict, key) requires (dict, glyph)", Hint: "pass dict and key"}
		}
		m := make(map[string]Datum, len(args[0].Dict))
		for k, v := range args[0].Dict {
			if k != args[1].Glyph {
				m[k] = v.DeepCopy()
			}
		}
		return true, DictDatum(m), nil

	case "vec_has_bmm":
		if len(args) != 2 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_VEC", Code: "ERR_VHAS_ARG", Detail: "vec_has_bmm(vec, item) requires (vec, any)", Hint: "pass vector and value"}
		}
		targetKey := fmt.Sprintf("%d:%s", args[1].Kind, args[1].Stringify())
		for _, it := range args[0].Vec {
			if fmt.Sprintf("%d:%s", it.Kind, it.Stringify()) == targetKey {
				return true, FlagDatum(true), nil
			}
		}
		return true, FlagDatum(false), nil

	case "str_repeat_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumCycle || args[1].Cycle < 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_STR", Code: "ERR_REP_ARG", Detail: "str_repeat_bmm(s, count) requires (glyph, non-negative cycle)", Hint: "pass string and repeat count"}
		}
		return true, GlyphDatum(strings.Repeat(args[0].Glyph, int(args[1].Cycle))), nil

	case "str_pad_left_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumCycle || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_STR", Code: "ERR_PADL_ARG", Detail: "str_pad_left_bmm(s, width, pad) requires (glyph, cycle, glyph)", Hint: "pass string, target length, and pad char"}
		}
		padCh := args[2].Glyph
		if padCh == "" {
			padCh = " "
		}
		runes := []rune(args[0].Glyph)
		target := int(args[1].Cycle)
		if len(runes) >= target {
			return true, GlyphDatum(args[0].Glyph), nil
		}
		var b strings.Builder
		for len(runes)+b.Len() < target {
			b.WriteString(padCh)
		}
		b.WriteString(args[0].Glyph)
		return true, GlyphDatum(b.String()), nil

	case "sleep_ms_bmm":
		if len(args) != 1 || args[0].Kind != DatumCycle || args[0].Cycle < 0 || args[0].Cycle > 60000 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_TIME", Code: "ERR_SLP_ARG", Detail: "sleep_ms_bmm(ms) requires cycle between 0 and 60000", Hint: "pass milliseconds up to 60s"}
		}
		time.Sleep(time.Duration(args[0].Cycle) * time.Millisecond)
		return true, FlagDatum(true), nil

	case "path_abs_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_PATH", Code: "ERR_PABS_ARG", Detail: "path_abs_bmm(p) requires glyph path", Hint: "pass relative or absolute path"}
		}
		absP, err := filepath.Abs(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "EXTRA4_PATH", Code: "ERR_PABS_ERR", Detail: err.Error(), Hint: "check path string"}
		}
		return true, GlyphDatum(absP), nil
	}

	return false, VoidDatum(), nil
}

