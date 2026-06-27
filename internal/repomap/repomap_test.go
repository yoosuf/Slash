package repomap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	b := NewBuilder()
	if b == nil {
		t.Fatal("expected non-nil builder")
	}
}

func TestBuildRepoMap_GoFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	goFile := filepath.Join(dir, "main.go")
	if err := os.WriteFile(goFile, []byte(`package main

import "fmt"

type User struct {
	Name string
}

type Config struct {
	Port int
}

func main() {
	fmt.Println("hello")
}

func helper() {
	fmt.Println("helper")
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 1 {
		t.Fatalf("expected 1 file, got %d", rm.TotalFiles)
	}

	if rm.TotalSymbols != 4 {
		t.Fatalf("expected 4 symbols (User, Config, main, helper), got %d: %+v", rm.TotalSymbols, rm.Files[0].Symbols)
	}

	foundMain := false
	foundHelper := false
	foundUser := false
	foundConfig := false
	for _, sym := range rm.Files[0].Symbols {
		switch sym.Name {
		case "main":
			foundMain = true
			if sym.Exported {
				t.Error("main should not be exported")
			}
		case "helper":
			foundHelper = true
			if sym.Exported {
				t.Error("helper should not be exported")
			}
		case "User":
			foundUser = true
			if !sym.Exported {
				t.Error("User should be exported")
			}
		case "Config":
			foundConfig = true
			if !sym.Exported {
				t.Error("Config should be exported")
			}
		}
	}
	if !foundMain || !foundHelper || !foundUser || !foundConfig {
		t.Error("expected to find all symbols")
	}
}

func TestBuildRepoMap_PythonFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	pyFile := filepath.Join(dir, "app.py")
	if err := os.WriteFile(pyFile, []byte(`class MyApp:
    def run(self):
        pass

def helper():
    return True
`), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundMyApp := false
	foundHelper := false
	for _, f := range rm.Files {
		for _, sym := range f.Symbols {
			switch sym.Name {
			case "MyApp":
				foundMyApp = true
				if sym.Kind != SymbolClass {
					t.Errorf("expected SymbolClass, got %v", sym.Kind)
				}
			case "helper":
				foundHelper = true
				if sym.Kind != SymbolFunction {
					t.Errorf("expected SymbolFunction, got %v", sym.Kind)
				}
			}
		}
	}
	if !foundMyApp || !foundHelper {
		t.Error("expected to find MyApp and helper symbols")
	}
}

func TestBuildRepoMap_TypeScriptFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	tsFile := filepath.Join(dir, "handler.ts")
	if err := os.WriteFile(tsFile, []byte(`export class RequestHandler {
    handle(req: Request): Response {
        return new Response("ok");
    }
}

export interface Config {
    port: number;
}

export type Result = string;

const VERSION = "1.0";

function internalHelper() {
    return 42;
}
`), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 1 {
		t.Fatalf("expected 1 file, got %d", rm.TotalFiles)
	}

	expected := map[string]SymbolKind{
		"RequestHandler": SymbolClass,
		"Config":         SymbolInterface,
		"Result":         SymbolType,
		"VERSION":        SymbolVariable,
		"internalHelper": SymbolFunction,
	}

	for _, f := range rm.Files {
		for _, sym := range f.Symbols {
			expectedKind, ok := expected[sym.Name]
			if !ok {
				continue
			}
			if sym.Kind != expectedKind {
				t.Errorf("symbol %s: expected kind %v, got %v", sym.Name, expectedKind, sym.Kind)
			}
			delete(expected, sym.Name)
		}
	}

	if len(expected) > 0 {
		t.Errorf("expected symbols not found: %v", expected)
	}
}

func TestBuildRepoMap_MultipleFiles(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte("package lib\nfunc A() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.go"), []byte("package lib\nfunc B() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 2 {
		t.Fatalf("expected 2 files, got %d", rm.TotalFiles)
	}
	if rm.TotalSymbols != 2 {
		t.Fatalf("expected 2 symbols, got %d", rm.TotalSymbols)
	}
}

func TestBuildRepoMap_EmptyDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 0 {
		t.Fatalf("expected 0 files for empty dir, got %d", rm.TotalFiles)
	}
}

func TestBuildRepoMap_IgnoresNodeModules(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "big.js"), []byte("function wasted() {}"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 0 {
		t.Fatalf("expected node_modules to be skipped, got %d files", rm.TotalFiles)
	}
}

func TestBuildRepoMap_NestedDir(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.MkdirAll(filepath.Join(dir, "api", "v1"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "api", "v1", "handler.go"), []byte("package v1\nfunc Handle() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if rm.TotalFiles != 1 {
		t.Fatalf("expected 1 file in nested dir, got %d", rm.TotalFiles)
	}
	if !strings.HasSuffix(rm.Files[0].Path, "handler.go") {
		t.Fatalf("expected handler.go, got %s", rm.Files[0].Path)
	}
}

func TestFindSymbol(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("func TargetFunc() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	results := rm.FindSymbol("TargetFunc")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "TargetFunc" {
		t.Fatalf("expected TargetFunc, got %s", results[0].Name)
	}
}

func TestToJSON(t *testing.T) {
	dir, err := os.MkdirTemp("", "repomap-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder()
	rm, err := b.Build(dir)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	json, err := rm.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if !strings.Contains(json, "root") {
		t.Fatal("JSON should contain root field")
	}
	if !strings.Contains(json, "main") {
		t.Fatal("JSON should contain symbol 'main'")
	}
}

func TestBuildRepoMap_NonExistentPath(t *testing.T) {
	b := NewBuilder()
	rm, err := b.Build("/tmp/__nonexistent_slash_test__")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rm.TotalFiles != 0 {
		t.Fatalf("expected 0 files for nonexistent dir, got %d", rm.TotalFiles)
	}
}
