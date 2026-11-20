package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func EvalNetPlusLibs(name string, args []Datum, line int) (bool, Datum, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	switch name {
	case "http_get_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HGET_ARG", Detail: "http_get_bmm(url) requires glyph url", Hint: "pass target http/https url"}
		}
		resp, err := client.Get(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HGET_REQ", Detail: err.Error(), Hint: "check url and network connection"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HGET_IO", Detail: err.Error(), Hint: "failed reading response body"}
		}
		return true, GlyphDatum(string(body)), nil

	case "http_post_bmm":
		if len(args) != 3 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HPOST_ARG", Detail: "http_post_bmm(url, contentType, body) requires three glyphs", Hint: "pass url, mime type, and payload"}
		}
		resp, err := client.Post(args[0].Glyph, args[1].Glyph, bytes.NewBufferString(args[2].Glyph))
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HPOST_REQ", Detail: err.Error(), Hint: "check endpoint and network connection"}
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HPOST_IO", Detail: err.Error(), Hint: "failed reading response body"}
		}
		return true, GlyphDatum(string(body)), nil

	case "http_status_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HSTAT_ARG", Detail: "http_status_bmm(url) requires glyph url", Hint: "pass target http/https url"}
		}
		resp, err := client.Get(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_HSTAT_REQ", Detail: err.Error(), Hint: "check url and network connection"}
		}
		defer resp.Body.Close()
		return true, GlyphDatum(strconv.Itoa(resp.StatusCode)), nil

	case "url_encode_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_UENC_ARG", Detail: "url_encode_bmm(text) requires glyph", Hint: "pass text to encode"}
		}
		return true, GlyphDatum(url.QueryEscape(args[0].Glyph)), nil

	case "url_decode_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_UDEC_ARG", Detail: "url_decode_bmm(text) requires glyph", Hint: "pass url-encoded text"}
		}
		dec, err := url.QueryUnescape(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "NET_PLUS_LIB", Code: "ERR_UDEC_ERR", Detail: err.Error(), Hint: "check url encoding format"}
		}
		return true, GlyphDatum(dec), nil
	}

	return false, VoidDatum(), nil
}

