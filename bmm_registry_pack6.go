package main

import (
"fmt"
"strings"
)

type bmmPkgMeta struct {
Name    string
Domain  string
Range   string
Symbols []string
}

func bmmCatalog() []bmmPkgMeta {
return []bmmPkgMeta{
{Name: "bmm_sys", Domain: "System and Files", Range: "1-5", Symbols: []string{"sys_info_lib", "fs_read_lib", "fs_write_lib"}},
{Name: "bmm_str", Domain: "String Processing", Range: "11-15", Symbols: []string{"lib11_case", "lib12_trim", "lib13_find", "lib14_replace", "lib15_split_join"}},
{Name: "bmm_math", Domain: "Math and Stats", Range: "16-20", Symbols: []string{"lib16_math", "lib17_round", "lib18_stat", "lib19_rnd", "lib20_conv"}},
{Name: "bmm_ui", Domain: "Terminal UI", Range: "21-30", Symbols: []string{"lib21_red_num", "lib22_blue_item", "lib23_yellow_table", "lib24_banner", "lib25_badge", "lib26_progress", "lib27_box", "lib28_divider", "lib29_kv", "lib30_ui_tag"}},
{Name: "bmm_net", Domain: "Network Diagnostics", Range: "31-32,49", Symbols: []string{"lib31_net_chk", "lib32_net_ip", "lib49_net_sec"}},
{Name: "bmm_io", Domain: "Shell, Logs and Config", Range: "33-35", Symbols: []string{"lib33_sh", "lib34_log", "lib35_cfg"}},
{Name: "bmm_list", Domain: "List Operations", Range: "36-40", Symbols: []string{"lib36_lst_len", "lib37_lst_get", "lib38_lst_has", "lib39_list_asc", "lib40_list_once"}},
{Name: "bmm_sec", Domain: "Cybersecurity", Range: "41-46", Symbols: []string{"lib41_sec_hash", "lib42_sec_key", "lib43_sec_mask", "lib44_sec_perm", "lib45_sec_scan", "lib46_sys_sec"}},
{Name: "bmm_cps", Domain: "Cyber-Physical and Robotics", Range: "47-48", Symbols: []string{"lib47_cps_guard", "lib48_robo_guard"}},
{Name: "bmm_neuro", Domain: "Neuromorphic Stability", Range: "50", Symbols: []string{"lib50_neuro_sec"}},
}
}

// bmm_pkg_list: return all official bmm_* package names
func libBmmPkgList() string {
var names []string
for _, p := range bmmCatalog() {
names = append(names, p.Name)
}
return strings.Join(names, ", ")
}

// bmm_pkg_info: inspect a specific bmm_* package
func libBmmPkgInfo(pkgName string) string {
target := strings.ToLower(strings.TrimSpace(pkgName))
for _, p := range bmmCatalog() {
if strings.ToLower(p.Name) == target {
return fmt.Sprintf("PKG: %s | DOMAIN: %s | LIBS: %s | CALLS: %s",
p.Name, p.Domain, p.Range, strings.Join(p.Symbols, ","))
}
}
return "UNKNOWN_BMM_PKG: " + strings.TrimSpace(pkgName)
}
