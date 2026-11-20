package main

import (
"encoding/csv"
"os"
"path/filepath"
"strings"
)

func EvalDataLibs(name string, args []Datum, line int) (bool, Datum, error) {
switch name {
case "env_get_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_ENVG_ARG", Detail: "env_get_bmm(key) requires glyph key", Hint: "pass environment variable name"}
}
return true, GlyphDatum(os.Getenv(args[0].Glyph)), nil

case "env_set_bmm":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_ENVS_ARG", Detail: "env_set_bmm(key, val) requires two glyphs", Hint: "pass key and value strings"}
}
if err := os.Setenv(args[0].Glyph, args[1].Glyph); err != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_ENVS_ERR", Detail: err.Error(), Hint: "check environment variable name"}
}
return true, FlagDatum(true), nil

case "glob_files_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_GLOB_ARG", Detail: "glob_files_bmm(pattern) requires glyph pattern", Hint: "pass glob pattern like *.go"}
}
matches, err := filepath.Glob(args[0].Glyph)
if err != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_GLOB_ERR", Detail: err.Error(), Hint: "check glob pattern syntax"}
}
out := make([]Datum, 0, len(matches))
for _, m := range matches {
out = append(out, GlyphDatum(m))
}
return true, VecDatum(out), nil

case "csv_row_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_CSV_ARG", Detail: "csv_row_bmm(line) requires glyph CSV line", Hint: "pass single CSV row string"}
}
r := csv.NewReader(strings.NewReader(args[0].Glyph))
fields, err := r.Read()
if err != nil {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_CSV_ERR", Detail: err.Error(), Hint: "check CSV formatting"}
}
out := make([]Datum, 0, len(fields))
for _, f := range fields {
out = append(out, GlyphDatum(f))
}
return true, VecDatum(out), nil

case "str_split_bmm":
if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_SPL_ARG", Detail: "str_split_bmm(str, sep) requires two glyphs", Hint: "pass string and delimiter"}
}
parts := strings.Split(args[0].Glyph, args[1].Glyph)
out := make([]Datum, 0, len(parts))
for _, p := range parts {
out = append(out, GlyphDatum(p))
}
return true, VecDatum(out), nil

case "str_join_bmm":
if len(args) != 2 || args[0].Kind != DatumVec || args[1].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_JLN_ARG", Detail: "str_join_bmm(vec, sep) requires (vec, glyph)", Hint: "pass list and separator"}
}
parts := make([]string, 0, len(args[0].Vec))
for _, it := range args[0].Vec {
parts = append(parts, it.Stringify())
}
return true, GlyphDatum(strings.Join(parts, args[1].Glyph)), nil

case "str_trim_bmm":
if len(args) != 1 || args[0].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_TRM_ARG", Detail: "str_trim_bmm(str) requires glyph", Hint: "pass string to trim"}
}
return true, GlyphDatum(strings.TrimSpace(args[0].Glyph)), nil

case "str_replace_bmm":
if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
return true, VoidDatum(), AetherDiag{Line: line, Stage: "DATA_LIB", Code: "ERR_RPL_ARG", Detail: "str_replace_bmm(str, old, new) requires three glyphs", Hint: "pass source string, target, and replacement"}
}
return true, GlyphDatum(strings.ReplaceAll(args[0].Glyph, args[1].Glyph, args[2].Glyph)), nil
}

return false, VoidDatum(), nil
}
