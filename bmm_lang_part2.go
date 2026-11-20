package main

import (
"fmt"
"strings"
"time"
)

// bmmPart2Translate maps futuristic, cyber, and bmm pattern tokens into structured runtime signals
func bmmPart2Translate(rawInput string) string {
trimmed := strings.TrimSpace(rawInput)
if trimmed == "" {
return "FUTURA_IDLE[0]"
}
lower := strings.ToLower(trimmed)
tokens := strings.Fields(lower)

var tags []string
hasBMM := strings.Contains(lower, "bmm")
hasCyber := strings.Contains(lower, "cyber")
hasPP := strings.Contains(lower, "pp") || strings.Contains(lower, "bb")
hasMM := strings.Contains(lower, "mm")
hasP := strings.Contains(lower, "p") || strings.Contains(lower, "b")
hasM := strings.Contains(lower, "m")

if hasBMM {
tags = append(tags, "CORE_BMM_VECTOR")
}
if hasCyber {
tags = append(tags, "CYBER_MATRIX_LINK")
}
if hasPP {
tags = append(tags, "DOUBLE_PULSE_PP")
} else if hasP {
tags = append(tags, "SINGLE_PULSE_P")
}
if hasMM {
tags = append(tags, "DOUBLE_MEM_MM")
} else if hasM {
tags = append(tags, "SINGLE_MEM_M")
}

for _, tok := range tokens {
switch tok {
case "nx", "nexus":
tags = append(tags, "NEXUS_MESH")
case "syn", "sync":
tags = append(tags, "QUANTUM_SYNC")
case "vox", "voice":
tags = append(tags, "VOX_STREAM")
case "grid", "grd":
tags = append(tags, "CYBER_GRID")
case "core", "cr":
tags = append(tags, "AETHER_CORE")
case "warp", "wrp":
tags = append(tags, "TIME_WARP")
case "opt", "optimize":
tags = append(tags, "AUTO_OPT")
case "sec", "shd", "shield":
tags = append(tags, "CYBER_SHIELD")
case "neu", "neuro":
tags = append(tags, "NEURO_SPIKE")
case "fut", "future":
tags = append(tags, "FUTURA_PRIME")
}
}

if len(tags) == 0 {
tags = append(tags, "RAW_LEXEME")
}

return fmt.Sprintf("BMM_P2[%s] :: INPUT=%q => MODES=[%s]",
time.Now().UTC().Format("150405"), trimmed, strings.Join(tags, " | "))
}

// bmmPart2Alias expands short futuristic aliases into canonical bmm-lang expressions
func bmmPart2Alias(aliasCmd string) string {
clean := strings.TrimSpace(aliasCmd)
key := strings.ToLower(clean)
switch key {
case "f.sys", "sys.f":
return nexusSysLink("/data/data/com.termux/files/usr/bin/bmm_lang")
case "f.mesh", "mesh.f":
return nexusMeshMap("/data/data/com.termux/files/home/bmm_engine")
case "f.cpu", "cpu.f":
return lib51HwCpuInfo()
case "f.cyber", "cyber.scan":
return bmmPart2Translate("cyber bmm pp mm future grid sec")
default:
return bmmPart2Translate(clean)
}
}
