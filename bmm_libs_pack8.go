package main

import (
"encoding/base64"
"encoding/hex"
"encoding/json"
"fmt"
"regexp"
"strings"
)

// 61. bmm_data: extract top-level or dot-path string/number value from JSON text
func lib61JsonGet(rawJSON, keyPath string) string {
var root any
if err := json.Unmarshal([]byte(strings.TrimSpace(rawJSON)), &root); err != nil {
return "ERR_JSON_PARSE: " + err.Error()
}
parts := strings.Split(strings.TrimSpace(keyPath), ".")
cur := root
for _, p := range parts {
if p == "" {
continue
}
m, ok := cur.(map[string]any)
if !ok {
return "JSON_NOT_FOUND: " + keyPath
}
nxt, exists := m[p]
if !exists {
return "JSON_NOT_FOUND: " + keyPath
}
cur = nxt
}
if cur == nil {
return "null"
}
switch v := cur.(type) {
case string:
return v
default:
b, err := json.Marshal(v)
if err != nil {
return fmt.Sprint(v)
}
return string(b)
}
}

// 62. bmm_data: base64 encode or decode
func lib62Base64(mode, input string) string {
m := strings.ToLower(strings.TrimSpace(mode))
switch m {
case "enc", "encode":
return base64.StdEncoding.EncodeToString([]byte(input))
case "dec", "decode":
raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(input))
if err != nil {
return "ERR_B64_DECODE: " + err.Error()
}
return string(raw)
default:
return "ERR_B64_MODE"
}
}

// 63. bmm_data: hex encode or decode
func lib63HexCodec(mode, input string) string {
m := strings.ToLower(strings.TrimSpace(mode))
switch m {
case "enc", "encode":
return hex.EncodeToString([]byte(input))
case "dec", "decode":
raw, err := hex.DecodeString(strings.TrimSpace(input))
if err != nil {
return "ERR_HEX_DECODE: " + err.Error()
}
return string(raw)
default:
return "ERR_HEX_MODE"
}
}

// 64. bmm_str: regex check or first match extraction
func lib64RegexMatch(mode, pattern, text string) string {
re, err := regexp.Compile(pattern)
if err != nil {
return "ERR_REGEX: " + err.Error()
}
m := strings.ToLower(strings.TrimSpace(mode))
if m == "find" || m == "first" {
loc := re.FindString(text)
if loc == "" {
return "NO_MATCH"
}
return loc
}
if re.MatchString(text) {
return "true"
}
return "false"
}
