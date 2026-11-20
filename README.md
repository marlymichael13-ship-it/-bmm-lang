# AETHER-BMM (`bmm-lang`) — v13.0.0

Self-contained programming language and CLI runtime in Go for Termux (`aarch64`).

## Install `.deb`

```bash
dpkg -i bmm-lang_13.0.0_aarch64.deb
```

## Build from Source

```bash
git clone [https://github.com/marlymichaell3-ship-it/-bmm-lang.git](https://github.com/marlymichaell3-ship-it/-bmm-lang.git)
cd -bmm-lang
go build -o bmm_lang bmm.go bmm_libs.go bmm_pro_libs.go bmm_extra_4.go bmm_regex_libs.go bmm_data_libs.go bmm_net_plus_libs.go bmm_time_fs_libs.go bmm_all_in_one_libs.go bmm_full_stack_libs.go bmm_core_pro_libs.go bmm_runtime_engine_libs.go bmm_pkg_release_libs.go
./bmm_lang -e 'emit_bmm(calc_bmm("(12 + 8) * 3"))'
```
