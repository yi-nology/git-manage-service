package stats

import (
	"testing"
)

func TestGetLanguageConfig_Go(t *testing.T) {
	cfg := GetLanguageConfig("main.go")
	if cfg == nil {
		t.Fatal("expected config for .go")
	}
	if cfg.Language != "Go" {
		t.Errorf("expected Go, got %s", cfg.Language)
	}
	if cfg.SingleLine != "//" {
		t.Errorf("expected //, got %s", cfg.SingleLine)
	}
}

func TestGetLanguageConfig_Python(t *testing.T) {
	cfg := GetLanguageConfig("script.py")
	if cfg == nil {
		t.Fatal("expected config for .py")
	}
	if cfg.Language != "Python" {
		t.Errorf("expected Python, got %s", cfg.Language)
	}
	if cfg.SingleLine != "#" {
		t.Errorf("expected #, got %s", cfg.SingleLine)
	}
}

func TestGetLanguageConfig_JavaScript(t *testing.T) {
	cfg := GetLanguageConfig("app.js")
	if cfg == nil {
		t.Fatal("expected config for .js")
	}
	if cfg.Language != "JavaScript" {
		t.Errorf("expected JavaScript, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_TypeScript(t *testing.T) {
	cfg := GetLanguageConfig("app.tsx")
	if cfg == nil {
		t.Fatal("expected config for .tsx")
	}
	if cfg.Language != "TypeScript" {
		t.Errorf("expected TypeScript, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Java(t *testing.T) {
	cfg := GetLanguageConfig("Main.java")
	if cfg == nil {
		t.Fatal("expected config for .java")
	}
	if cfg.Language != "Java" {
		t.Errorf("expected Java, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_C(t *testing.T) {
	cfg := GetLanguageConfig("main.c")
	if cfg == nil {
		t.Fatal("expected config for .c")
	}
	if cfg.Language != "C/C++" {
		t.Errorf("expected C/C++, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Rust(t *testing.T) {
	cfg := GetLanguageConfig("main.rs")
	if cfg == nil {
		t.Fatal("expected config for .rs")
	}
	if cfg.Language != "Rust" {
		t.Errorf("expected Rust, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Shell(t *testing.T) {
	cfg := GetLanguageConfig("run.sh")
	if cfg == nil {
		t.Fatal("expected config for .sh")
	}
	if cfg.Language != "Shell" {
		t.Errorf("expected Shell, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_YAML(t *testing.T) {
	cfg := GetLanguageConfig("config.yaml")
	if cfg == nil {
		t.Fatal("expected config for .yaml")
	}
	if cfg.Language != "YAML" {
		t.Errorf("expected YAML, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_SQL(t *testing.T) {
	cfg := GetLanguageConfig("query.sql")
	if cfg == nil {
		t.Fatal("expected config for .sql")
	}
	if cfg.Language != "SQL" {
		t.Errorf("expected SQL, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Dockerfile(t *testing.T) {
	cfg := GetLanguageConfig("Dockerfile")
	if cfg == nil {
		t.Fatal("expected config for Dockerfile")
	}
	if cfg.Language != "Dockerfile" {
		t.Errorf("expected Dockerfile, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Makefile(t *testing.T) {
	cfg := GetLanguageConfig("Makefile")
	if cfg == nil {
		t.Fatal("expected config for Makefile")
	}
	if cfg.Language != "Makefile" {
		t.Errorf("expected Makefile, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_MakefileMK(t *testing.T) {
	cfg := GetLanguageConfig("rules.mk")
	if cfg == nil {
		t.Fatal("expected config for .mk")
	}
	if cfg.Language != "Makefile" {
		t.Errorf("expected Makefile, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Vue(t *testing.T) {
	cfg := GetLanguageConfig("App.vue")
	if cfg == nil {
		t.Fatal("expected config for .vue")
	}
	if cfg.Language != "Vue" {
		t.Errorf("expected Vue, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_Unknown(t *testing.T) {
	cfg := GetLanguageConfig("file.xyz")
	if cfg != nil {
		t.Errorf("expected nil for unknown extension, got %+v", cfg)
	}
}

func TestGetLanguageConfig_CaseInsensitive(t *testing.T) {
	cfg := GetLanguageConfig("Main.GO")
	if cfg == nil {
		t.Fatal("expected case-insensitive match")
	}
	if cfg.Language != "Go" {
		t.Errorf("expected Go, got %s", cfg.Language)
	}
}

func TestGetLanguageConfig_PathWithDir(t *testing.T) {
	cfg := GetLanguageConfig("src/pkg/main.go")
	if cfg == nil {
		t.Fatal("expected config for .go in subdirectory")
	}
}

func TestGetSupportedExtensions_NonEmpty(t *testing.T) {
	exts := GetSupportedExtensions()
	if len(exts) == 0 {
		t.Error("expected non-empty extensions list")
	}
}

func TestGetSupportedExtensions_ContainsCommon(t *testing.T) {
	exts := GetSupportedExtensions()
	extMap := map[string]bool{}
	for _, e := range exts {
		extMap[e] = true
	}
	common := []string{".go", ".py", ".js", ".ts", ".java", ".rs", ".sql"}
	for _, c := range common {
		if !extMap[c] {
			t.Errorf("expected extension %s in supported list", c)
		}
	}
}

func TestLanguageConfigs_AllHaveLanguage(t *testing.T) {
	for _, cfg := range LanguageConfigs {
		if cfg.Language == "" {
			t.Error("expected all LanguageConfigs to have a Language")
		}
		if len(cfg.Extensions) == 0 {
			t.Errorf("Language %s has no extensions", cfg.Language)
		}
	}
}
