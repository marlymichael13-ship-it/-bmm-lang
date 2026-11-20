package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func computeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFileSimple(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func EvalPkgReleaseLibs(name string, args []Datum, line int) (bool, Datum, error) {
	switch name {
	case "file_sha256_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_FSHA_ARG", Detail: "file_sha256_bmm(file_path) requires glyph", Hint: "pass path to file"}
		}
		sum, err := computeFileSHA256(args[0].Glyph)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_FSHA_IO", Detail: err.Error(), Hint: "check file path and read permissions"}
		}
		return true, GlyphDatum(sum), nil

	case "pkg_stage_bmm":
		if len(args) != 2 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGS_ARG", Detail: "pkg_stage_bmm(bin_path, stage_dir) requires two glyphs", Hint: "pass binary path and staging directory"}
		}
		binPath := args[0].Glyph
		stageDir := args[1].Glyph
		if err := os.MkdirAll(filepath.Join(stageDir, "usr", "bin"), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGS_DIR", Detail: err.Error(), Hint: "failed creating staging bin directory"}
		}
		if err := os.MkdirAll(filepath.Join(stageDir, "share", "doc", "aether-bmm"), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGS_DOC", Detail: err.Error(), Hint: "failed creating staging doc directory"}
		}
		dstBin := filepath.Join(stageDir, "usr", "bin", filepath.Base(binPath))
		if err := copyFileSimple(binPath, dstBin, 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGS_CP", Detail: err.Error(), Hint: "check binary path and permissions"}
		}
		return true, GlyphDatum(dstBin), nil

	case "pkgbuild_gen_bmm":
		if len(args) != 4 || args[0].Kind != DatumGlyph || args[1].Kind != DatumGlyph || args[2].Kind != DatumGlyph || args[3].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGB_ARG", Detail: "pkgbuild_gen_bmm(out_path, version, pkg_name, artifact_path) requires four glyphs", Hint: "pass output PKGBUILD path, version, package name, and artifact path"}
		}
		outPath := args[0].Glyph
		ver := strings.TrimSpace(args[1].Glyph)
		pkgName := strings.TrimSpace(args[2].Glyph)
		artifactPath := strings.TrimSpace(args[3].Glyph)
		sum := "SKIP"
		if s, err := computeFileSHA256(artifactPath); err == nil {
			sum = s
		}
		arch := runtime.GOARCH
		if arch == "arm64" {
			arch = "aarch64"
		}
		content := "# Maintainer: AETHER-BMM Contributors\n" +
			"pkgname=" + pkgName + "\n" +
			"pkgver=" + ver + "\n" +
			"pkgrel=1\n" +
			"pkgdesc=\"AETHER-BMM self-contained runtime and CLI engine\"\n" +
			"arch=('" + arch + "')\n" +
			"url=\"https://github.com/aether-bmm/aether-bmm\"\n" +
			"license=('MIT')\n" +
			"sha256sums=('" + sum + "')\n\n" +
			"package() {\n" +
			"  install -Dm755 \"${srcdir}/bmm_lang\" \"${pkgdir}/usr/bin/bmm_lang\"\n" +
			"}\n"
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGB_DIR", Detail: err.Error(), Hint: "check output directory"}
		}
		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_PKGB_IO", Detail: err.Error(), Hint: "failed writing PKGBUILD"}
		}
		return true, FlagDatum(true), nil

	case "gitignore_gen_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_GIGN_ARG", Detail: "gitignore_gen_bmm(out_path) requires glyph", Hint: "pass path to .gitignore file"}
		}
		gi := "*.deb\n*.pkg.tar.*\n*.tar.gz\n*.zip\nbmm_lang\n.bmm_atomic_*\n__pycache__/\n*.pyc\n"
		if err := os.MkdirAll(filepath.Dir(args[0].Glyph), 0755); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_GIGN_DIR", Detail: err.Error(), Hint: "check output directory"}
		}
		if err := os.WriteFile(args[0].Glyph, []byte(gi), 0644); err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_GIGN_IO", Detail: err.Error(), Hint: "failed writing .gitignore"}
		}
		return true, FlagDatum(true), nil

	case "release_audit_bmm":
		if len(args) != 1 || args[0].Kind != DatumGlyph {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_AUDT_ARG", Detail: "release_audit_bmm(repo_dir) requires glyph", Hint: "pass project directory path"}
		}
		dir := args[0].Glyph
		checks := []string{
			"bmm.go",
			"bmm_lang",
			"bmm_all_in_one_libs.go",
			"bmm_full_stack_libs.go",
			"bmm_core_pro_libs.go",
			"bmm_runtime_engine_libs.go",
			"bmm_pkg_release_libs.go",
		}
		status := make(map[string]bool, len(checks))
		allOK := true
		for _, item := range checks {
			_, err := os.Stat(filepath.Join(dir, item))
			ok := err == nil
			status[item] = ok
			if !ok {
				allOK = false
			}
		}
		report := map[string]any{
			"ready": allOK,
			"files": status,
		}
		b, err := json.Marshal(report)
		if err != nil {
			return true, VoidDatum(), AetherDiag{Line: line, Stage: "PKG_RELEASE", Code: "ERR_AUDT_ENC", Detail: err.Error(), Hint: "failed encoding audit report"}
		}
		return true, GlyphDatum(string(b)), nil
	}

	return false, VoidDatum(), nil
}
