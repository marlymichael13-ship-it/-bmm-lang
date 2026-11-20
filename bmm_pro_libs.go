package main

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func EvalProLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "http_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_GET_ARG", Detail: "http_get_bmm(url) requires glyph url", Hint: "pass http or https URL"}
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_GET_REQ", Detail: err.Error(), Hint: "check network connection or URL"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_GET_READ", Detail: err.Error(), Hint: "failed reading HTTP response body"}
		}
		return true, GlyphDatum(string(body)), nil

	case "http_post_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_POST_ARG", Detail: "http_post_bmm(url, body, content_type) requires three glyphs", Hint: "pass URL, payload string, and MIME type"}
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Post(args[0].Glyph, args[2].Glyph, strings.NewReader(args[1].Glyph))
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_POST_REQ", Detail: err.Error(), Hint: "check endpoint URL and network"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_NET", Code: "ERR_HTTP_POST_READ", Detail: err.Error(), Hint: "failed reading response body"}
		}
		return true, GlyphDatum(string(body)), nil

	case "shell_exec_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_EXEC", Code: "ERR_SH_ARG", Detail: "shell_exec_bmm(cmd) requires glyph command", Hint: "pass shell command string"}
		}
		cmd := exec.Command("sh", "-c", args[0].Glyph)
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf
		err := cmd.Run()
		code := int64(0)
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				code = int64(exitErr.ExitCode())
			} else {
				code = 1
			}
		}
		resMap := map[string]Datum{
			"code": CycleDatum(code),
			"out":  GlyphDatum(strings.TrimSuffix(outBuf.String(), "\n")),
		}
		return true, DictDatum(resMap), nil

	case "md5_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_SEC", Code: "ERR_MD5_ARG", Detail: "md5_bmm(text) requires glyph", Hint: "pass text string"}
		}
		h := md5.Sum([]byte(args[0].Glyph))
		return true, GlyphDatum(hex.EncodeToString(h[:])), nil

	case "rand_hex_bmm":
		if len(args) != 1 || args[0].Kind != DatumCycle || args[0].Cycle <= 0 || args[0].Cycle > 4096 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_SEC", Code: "ERR_RAND_ARG", Detail: "rand_hex_bmm(n_bytes) requires cycle between 1 and 4096", Hint: "pass positive byte count"}
		}
		buf := make([]byte, int(args[0].Cycle))
		if _, err := rand.Read(buf); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_SEC", Code: "ERR_RAND_IO", Detail: err.Error(), Hint: "entropy source unavailable"}
		}
		return true, GlyphDatum(hex.EncodeToString(buf)), nil

	case "csv_parse_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_DATA", Code: "ERR_CSV_ARG", Detail: "csv_parse_bmm(raw, sep) requires two glyphs", Hint: "pass CSV text and separator like ','"}
		}
		sep := args[1].Glyph
		if sep == "" {
			sep = ","
		}
		rows := make([]Datum, 0)
		for _, ln := range strings.Split(strings.ReplaceAll(args[0].Glyph, "\r\n", "\n"), "\n") {
			if strings.TrimSpace(ln) == "" {
				continue
			}
			cols := strings.Split(ln, sep)
			rowVec := make([]Datum, 0, len(cols))
			for _, c := range cols {
				rowVec = append(rowVec, GlyphDatum(strings.TrimSpace(c)))
			}
			rows = append(rows, VecDatum(rowVec))
		}
		return true, VecDatum(rows), nil

	case "json_pretty_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_DATA", Code: "ERR_JSONP_ARG", Detail: "json_pretty_bmm(json_str) requires glyph", Hint: "pass valid JSON string"}
		}
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, []byte(args[0].Glyph), "", "  "); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_DATA", Code: "ERR_JSONP_ERR", Detail: err.Error(), Hint: "check JSON validity"}
		}
		return true, GlyphDatum(pretty.String()), nil

	case "min_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec || len(args[0].Vec) == 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_MATH", Code: "ERR_MIN_ARG", Detail: "min_vec_bmm(vec) requires non-empty cycle vec", Hint: "pass non-empty number vector"}
		}
		var minVal int64
		found := false
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				if !found || it.Cycle < minVal {
					minVal = it.Cycle
					found = true
				}
			}
		}
		if !found {
			return true, VoidDatum(), nil
		}
		return true, CycleDatum(minVal), nil

	case "max_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec || len(args[0].Vec) == 0 {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_MATH", Code: "ERR_MAX_ARG", Detail: "max_vec_bmm(vec) requires non-empty cycle vec", Hint: "pass non-empty number vector"}
		}
		var maxVal int64
		found := false
		for _, it := range args[0].Vec {
			if it.Kind == DatumCycle {
				if !found || it.Cycle > maxVal {
					maxVal = it.Cycle
					found = true
				}
			}
		}
		if !found {
			return true, VoidDatum(), nil
		}
		return true, CycleDatum(maxVal), nil

	case "unique_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_MATH", Code: "ERR_UNIQ_ARG", Detail: "unique_vec_bmm(vec) requires vec", Hint: "pass vector"}
		}
		seen := make(map[string]bool)
		out := make([]Datum, 0, len(args[0].Vec))
		for _, it := range args[0].Vec {
			key := fmt.Sprintf("%d", it.Kind) + ":" + it.Stringify()
			if !seen[key] {
				seen[key] = true
				out = append(out, it.DeepCopy())
			}
		}
		return true, VecDatum(out), nil

	case "reverse_vec_bmm":
		if len(args) != 1 || args[0].Kind != DatumVec {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PRO_MATH", Code: "ERR_REV_ARG", Detail: "reverse_vec_bmm(vec) requires vec", Hint: "pass vector"}
		}
		n := len(args[0].Vec)
		out := make([]Datum, n)
		for i := 0; i < n; i++ {
			out[i] = args[0].Vec[n-1-i].DeepCopy()
		}
		return true, VecDatum(out), nil
	}

	return false, VoidDatum(), nil
}

