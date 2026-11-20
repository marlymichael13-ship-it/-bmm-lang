package main

import "strings"

// Part4Bridge provides a clean, isolated resolution helper for part3/part2 symbols
// without cluttering core VM execution or causing build conflicts.
type Part4Bridge struct {
Known map[string]string
}

func NewPart4Bridge() *Part4Bridge {
return &Part4Bridge{
Known: map[string]string{
"p3_color":   "part3.color",
"p3_lexicon": "part3.lexicon",
"p3_compile": "part3.compile",
"p3_eval":    "part3.eval",
"p2_futura":  "part2.futura",
"p2_alias":   "part2.alias",
},
}
}

func (b *Part4Bridge) Resolve(sym string) (string, bool) {
key := strings.TrimSpace(sym)
val, ok := b.Known[key]
return val, ok
}
