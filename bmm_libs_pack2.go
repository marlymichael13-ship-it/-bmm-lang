package main

import (
"fmt"
"math"
"math/rand"
"strconv"
"strings"
"time"
)

// 11. str_case_lib: convert string case (upper, lower, title)
func lib11StrCase(mode, s string) string {
switch strings.ToLower(strings.TrimSpace(mode)) {
case "upper":
return strings.ToUpper(s)
case "lower":
return strings.ToLower(s)
case "title":
if s == "" {
return ""
}
parts := strings.Fields(s)
for i, p := range parts {
parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
}
return strings.Join(parts, " ")
default:
return s
}
}

// 12. str_trim_lib: trim whitespace or specified cutset
func lib12StrTrim(s, cutset string) string {
if cutset == "" {
return strings.TrimSpace(s)
}
return strings.Trim(s, cutset)
}

// 13. str_find_lib: check contains, prefix, suffix, or count occurrences
func lib13StrFind(mode, s, sub string) string {
switch strings.ToLower(strings.TrimSpace(mode)) {
case "contains":
return strconv.FormatBool(strings.Contains(s, sub))
case "prefix":
return strconv.FormatBool(strings.HasPrefix(s, sub))
case "suffix":
return strconv.FormatBool(strings.HasSuffix(s, sub))
case "count":
return strconv.Itoa(strings.Count(s, sub))
default:
return strconv.Itoa(strings.Index(s, sub))
}
}

// 14. str_replace_lib: replace first occurrence or all occurrences
func lib14StrReplace(s, oldStr, newStr string, all bool) string {
if all {
return strings.ReplaceAll(s, oldStr, newStr)
}
return strings.Replace(s, oldStr, newStr, 1)
}

// 15. str_split_join_lib: split string by sep and join with newSep
func lib15StrSplitJoin(s, sep, joinSep string) string {
parts := strings.Split(s, sep)
return strings.Join(parts, joinSep)
}

// 16. math_basic_lib: power, sqrt, abs, mod
func lib16MathBasic(op string, a, b float64) string {
var res float64
switch strings.ToLower(strings.TrimSpace(op)) {
case "pow":
res = math.Pow(a, b)
case "sqrt":
if a < 0 {
return "ERR_SQRT_NEG"
}
res = math.Sqrt(a)
case "abs":
res = math.Abs(a)
case "mod":
if b == 0 {
return "ERR_MOD_ZERO"
}
res = math.Mod(a, b)
default:
return "ERR_UNKNOWN_MATH_OP"
}
return strconv.FormatFloat(res, 'f', -1, 64)
}

// 17. math_round_lib: ceil, floor, round to decimal places
func lib17MathRound(mode string, val float64, places int) string {
var res float64
switch strings.ToLower(strings.TrimSpace(mode)) {
case "ceil":
res = math.Ceil(val)
case "floor":
res = math.Floor(val)
case "round":
factor := math.Pow(10, float64(places))
res = math.Round(val*factor) / factor
default:
res = math.Round(val)
}
return strconv.FormatFloat(res, 'f', -1, 64)
}

// 18. math_stat_lib: sum, avg, min, max for comma-separated numbers
func lib18MathStat(mode, csvNums string) string {
raw := strings.Split(csvNums, ",")
var nums []float64
for _, r := range raw {
t := strings.TrimSpace(r)
if t == "" {
continue
}
v, err := strconv.ParseFloat(t, 64)
if err != nil {
return "ERR_PARSE_NUM: " + t
}
nums = append(nums, v)
}
if len(nums) == 0 {
return "ERR_EMPTY_LIST"
}
sum := 0.0
minVal := nums[0]
maxVal := nums[0]
for _, n := range nums {
sum += n
if n < minVal {
minVal = n
}
if n > maxVal {
maxVal = n
}
}
switch strings.ToLower(strings.TrimSpace(mode)) {
case "sum":
return strconv.FormatFloat(sum, 'f', -1, 64)
case "avg":
return strconv.FormatFloat(sum/float64(len(nums)), 'f', -1, 64)
case "min":
return strconv.FormatFloat(minVal, 'f', -1, 64)
case "max":
return strconv.FormatFloat(maxVal, 'f', -1, 64)
default:
return fmt.Sprintf("Sum: %g | Avg: %g | Min: %g | Max: %g", sum, sum/float64(len(nums)), minVal, maxVal)
}
}

// 19. math_rand_lib: random int in [minVal, maxVal]
func lib19MathRand(minVal, maxVal int) string {
if maxVal < minVal {
minVal, maxVal = maxVal, minVal
}
rng := rand.New(rand.NewSource(time.Now().UnixNano()))
n := rng.Intn(maxVal-minVal+1) + minVal
return strconv.Itoa(n)
}

// 20. math_conv_lib: convert base-10 integer to bin, oct, hex
func lib20MathConv(val int64, base int) string {
switch base {
case 2:
return strconv.FormatInt(val, 2)
case 8:
return strconv.FormatInt(val, 8)
case 16:
return strings.ToUpper(strconv.FormatInt(val, 16))
default:
return strconv.FormatInt(val, 10)
}
}
