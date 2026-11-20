package main

import (
	"regexp"
)

func EvalRegexLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "regex_match_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RMATCH_ARG", Detail: "regex_match_bmm(pattern, text) requires two glyphs", Hint: "pass regex pattern and target text"}
		}
		re, err := regexp.Compile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RMATCH_PAT", Detail: err.Error(), Hint: "check regular expression syntax"}
		}
		return true, FlagDatum(re.MatchString(args[1].Glyph)), nil

	case "regex_find_all_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RFIND_ARG", Detail: "regex_find_all_bmm(pattern, text) requires two glyphs", Hint: "pass regex pattern and target text"}
		}
		re, err := regexp.Compile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RFIND_PAT", Detail: err.Error(), Hint: "check regular expression syntax"}
		}
		matches := re.FindAllString(args[1].Glyph, -1)
		out := make([]Datum, 0, len(matches))
		for _, m := range matches {
			out = append(out, GlyphDatum(m))
		}
		return true, VecDatum(out), nil

	case "regex_replace_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RREP_ARG", Detail: "regex_replace_bmm(pattern, text, repl) requires three glyphs", Hint: "pass pattern, text, and replacement"}
		}
		re, err := regexp.Compile(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "REGEX_LIB", Code: "ERR_RREP_PAT", Detail: err.Error(), Hint: "check regular expression syntax"}
		}
		return true, GlyphDatum(re.ReplaceAllString(args[1].Glyph, args[2].Glyph)), nil
	}

	return false, VoidDatum(), nil
}

