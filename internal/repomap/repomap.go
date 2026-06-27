package repomap

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type SymbolKind string

const (
	SymbolFunction SymbolKind = "function"
	SymbolMethod   SymbolKind = "method"
	SymbolClass    SymbolKind = "class"
	SymbolType     SymbolKind = "type"
	SymbolInterface SymbolKind = "interface"
	SymbolStruct   SymbolKind = "struct"
	SymbolImport   SymbolKind = "import"
	SymbolVariable SymbolKind = "variable"
	SymbolConst    SymbolKind = "constant"
	SymbolEnum     SymbolKind = "enum"
)

type Symbol struct {
	Name     string     `json:"name"`
	Kind     SymbolKind `json:"kind"`
	Line     int        `json:"line"`
	Exported bool       `json:"exported"`
}

type FileEntry struct {
	Path    string   `json:"path"`
	Symbols []Symbol `json:"symbols"`
	SizeKB  int      `json:"size_kb"`
	Lang    string   `json:"lang"`
}

type RepoMap struct {
	Root        string      `json:"root"`
	Files       []FileEntry `json:"files"`
	TotalFiles  int         `json:"total_files"`
	TotalSymbols int         `json:"total_symbols"`
}

type Builder struct {
	mu          sync.Mutex
	excludeDirs []string
	includeExts map[string]bool
}

func NewBuilder() *Builder {
	return &Builder{
		excludeDirs: []string{
			".git", "node_modules", "vendor", ".next", "dist", "build",
			"target", "__pycache__", ".venv", "venv", ".egg-info",
			".cache", ".bin", ".out", "coverage", ".terraform",
		},
		includeExts: map[string]bool{
			".go": true, ".ts": true, ".tsx": true, ".js": true, ".jsx": true,
			".py": true, ".rs": true, ".java": true, ".rb": true, ".cs": true,
			".cpp": true, ".cc": true, ".c": true, ".h": true, ".hpp": true,
			".swift": true, ".kt": true, ".scala": true, ".php": true,
			".vue": true, ".svelte": true, ".mjs": true, ".cjs": true,
			".dart": true, ".lua": true, ".r": true, ".pl": true, ".pm": true,
			".ex": true, ".exs": true, ".erl": true, ".hrl": true,
			".zig": true, ".nim": true, ".jl": true, ".sol": true,
			".hs": true, ".lhs": true, ".groovy": true, ".gradle": true,
			".clj": true, ".cljs": true, ".cljc": true, ".edn": true,
			".fs": true, ".fsx": true, ".ml": true, ".mli": true,
			".f": true, ".f90": true, ".f95": true, ".for": true,
			".m": true, ".mm": true, ".ada": true, ".adb": true, ".ads": true,
			".cr": true, ".proto": true,
			".sh": true, ".bash": true, ".zsh": true, ".ps1": true,
			".sql": true, ".html": true, ".xml": true, ".css": true,
			".scss": true, ".less": true, ".md": true, ".yaml": true, ".yml": true,
			".toml": true, ".json": true, ".tex": true, ".asm": true,
		},
	}
}

func (b *Builder) Build(root string) (*RepoMap, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path: %w", err)
	}

	rm := &RepoMap{
		Root:  root,
		Files: []FileEntry{},
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()
			for _, exclude := range b.excludeDirs {
				if name == exclude || strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !b.includeExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		entry, err := b.parseFile(path, relPath, ext)
		if err != nil {
			return nil
		}
		if entry != nil {
			rm.Files = append(rm.Files, *entry)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk root: %w", err)
	}

	rm.TotalFiles = len(rm.Files)
	for _, f := range rm.Files {
		rm.TotalSymbols += len(f.Symbols)
	}

	return rm, nil
}

func (b *Builder) parseFile(path, relPath, ext string) (*FileEntry, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	entry := &FileEntry{
		Path:   relPath,
		SizeKB: int(fi.Size() / 1024),
		Lang:   ext,
	}

	var symbols []Symbol
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		text := scanner.Text()
		trimmed := strings.TrimSpace(text)

		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		symbols = append(symbols, b.extractSymbols(trimmed, lineNum, ext)...)
	}

	entry.Symbols = symbols
	return entry, scanner.Err()
}

var (
	goFuncRe      = regexp.MustCompile(`^func\s+(?:\s*\([^)]*\)\s+)?([A-Za-z_]\w*)`)
	goTypeRe      = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)
	goStructRe    = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s+struct`)
	goInterfaceRe = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)\s+interface`)

	tsFuncRe      = regexp.MustCompile(`^(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_$]\w*)`)
	tsClassRe     = regexp.MustCompile(`^(?:export\s+)?class\s+([A-Za-z_$]\w*)`)
	tsInterfaceRe = regexp.MustCompile(`^(?:export\s+)?interface\s+([A-Za-z_$]\w*)`)
	tsTypeRe      = regexp.MustCompile(`^(?:export\s+)?type\s+([A-Za-z_$]\w*)`)
	tsEnumRe      = regexp.MustCompile(`^(?:export\s+)?enum\s+([A-Za-z_$]\w*)`)
	tsConstRe     = regexp.MustCompile(`^(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$]\w*)`)
	tsArrowRe     = regexp.MustCompile(`^(?:export\s+)?(?:const|let|var)\s+([A-Za-z_$]\w*)\s*[:=].*=>`)

	pyFuncRe      = regexp.MustCompile(`^(?:async\s+)?def\s+([A-Za-z_]\w*)\s*\(`)
	pyClassRe     = regexp.MustCompile(`^class\s+([A-Za-z_]\w*)`)

	javaClassRe   = regexp.MustCompile(`^(?:public\s+)?(?:abstract\s+|final\s+)?(?:class|interface|enum|@interface|record)\s+([A-Za-z_]\w*)`)
	javaMethodRe  = regexp.MustCompile(`(?:public|private|protected|static|final|abstract|synchronized|native)\s+(?:(?:<[^>]+>\s+)?[A-Za-z_][\w.<>[\]]+)\s+([a-z_]\w*)\s*\(`)
	javaFieldRe   = regexp.MustCompile(`(?:public|private|protected|static|final)\s+(?:[A-Za-z_][\w<>\[\]]+)\s+([a-z_]\w*)\s*(?:=|;)`)

	rsFuncRe      = regexp.MustCompile(`^(?:pub\s+)?(?:unsafe\s+)?fn\s+([A-Za-z_]\w*)`)
	rsStructRe    = regexp.MustCompile(`^(?:pub\s+)?struct\s+([A-Za-z_]\w*)`)
	rsEnumRe      = regexp.MustCompile(`^(?:pub\s+)?enum\s+([A-Za-z_]\w*)`)
	rsTraitRe     = regexp.MustCompile(`^(?:pub\s+)?trait\s+([A-Za-z_]\w*)`)
	rsTypeRe      = regexp.MustCompile(`^(?:pub\s+)?type\s+([A-Za-z_]\w*)`)
	rsModRe       = regexp.MustCompile(`^(?:pub\s+)?mod\s+([A-Za-z_]\w*)`)

	csClassRe     = regexp.MustCompile(`^(?:public|private|internal|protected|static|abstract|sealed|partial|readonly)\s+(?:class|struct|interface|record|enum)\s+([A-Za-z_]\w*)`)
	csMethodRe    = regexp.MustCompile(`(?:public|private|internal|protected|static|virtual|override|abstract|async|unsafe)\s+(?:[A-Za-z_][\w<>\[\]?,]+)\s+([A-Za-z_]\w*)\s*\(`)
	csPropRe      = regexp.MustCompile(`(?:public|private|internal|protected|static|readonly)\s+(?:[A-Za-z_][\w<>\[\]?]+)\s+([A-Za-z_]\w*)\s*\{`)

	ktFuncRe      = regexp.MustCompile(`^(?:private|public|internal|protected|inline|suspend|tailrec|operator|infix)\s+(?:fun\s+)?([A-Za-z_]\w*)\s*\(`)
	ktClassRe     = regexp.MustCompile(`^(?:private|public|internal|data|sealed|open|abstract)\s+(?:class|interface|object|enum class|annotation class)\s+([A-Za-z_]\w*)`)
	ktValRe       = regexp.MustCompile(`^(?:private|public|internal|protected|lateinit|inline)?\s*(?:val|var)\s+([A-Za-z_]\w*)`)

	swiftFuncRe   = regexp.MustCompile(`^(?:private|public|internal|fileprivate|open|static|class|override|mutating|nonmutating|@objc)?\s*(?:func)\s+([A-Za-z_]\w*)`)
	swiftClassRe  = regexp.MustCompile(`^(?:private|public|internal|fileprivate|open|final|@objc)?\s*(?:class|struct|enum|protocol|extension)\s+([A-Za-z_]\w*)`)
	swiftTypeRe   = regexp.MustCompile(`^(?:private|public|internal|fileprivate)?\s*typealias\s+([A-Za-z_]\w*)`)

	phpFuncRe     = regexp.MustCompile(`^(?:public|private|protected|static|abstract|final)?\s*function\s+([A-Za-z_]\w*)\s*\(`)
	phpClassRe    = regexp.MustCompile(`^(?:abstract|final)?\s*class\s+([A-Za-z_]\w*)`)
	phpInterfaceRe = regexp.MustCompile(`^interface\s+([A-Za-z_]\w*)`)
	phpTraitRe    = regexp.MustCompile(`^trait\s+([A-Za-z_]\w*)`)
	phpEnumRe     = regexp.MustCompile(`^enum\s+([A-Za-z_]\w*)`)

	dartClassRe   = regexp.MustCompile(`^(?:abstract|base|interface|sealed|final|mixin)?\s*(?:class|mixin|enum|extension)\s+([A-Za-z_]\w*)`)
	dartFuncRe    = regexp.MustCompile(`^(?:[A-Za-z_]\w*\s+)?([A-Za-z_]\w*)\s*\([^)]*\)\s*(?:async|sync|\{)`)
	dartTypeRe    = regexp.MustCompile(`^typedef\s+([A-Za-z_]\w*)`)

	scalaClassRe  = regexp.MustCompile(`^(?:private|protected|sealed|abstract|open|case)?\s*(?:class|object|trait|enum|case class|case object)\s+([A-Za-z_]\w*)`)
	scalaDefRe    = regexp.MustCompile(`^(?:private|protected|override|implicit|lazy|inline|final)?\s*def\s+([A-Za-z_]\w*)`)
	scalaTypeRe   = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)
	scalaValRe    = regexp.MustCompile(`^(?:private|protected|lazy|override|implicit)?\s*(?:val|var)\s+([A-Za-z_]\w*)`)

	luaFuncRe     = regexp.MustCompile(`^(?:(?:local\s+)?function\s+([A-Za-z_]\w*))`)
	luaMethodRe   = regexp.MustCompile(`^(?:function\s+[A-Za-z_]\w*\.)?([A-Za-z_]\w*)\s*=\s*function`)
	luaTableRe    = regexp.MustCompile(`^([A-Za-z_]\w*)\s*=\s*\{`)

	rFuncRe       = regexp.MustCompile(`^([A-Za-z_]\w*)\s*<-\s*function`)
	rClassRe      = regexp.MustCompile(`^setClass\s*\(\s*\"([A-Za-z_]\w*)\"`)

	hsFuncRe      = regexp.MustCompile(`^([A-Za-z_]\w*)\s*::`)
	hsDataRe      = regexp.MustCompile(`^data\s+([A-Za-z_]\w*)`)
	hsTypeRe      = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)
	hsClassRe     = regexp.MustCompile(`^class\s+([A-Za-z_]\w*)`)

	elixirDefRe   = regexp.MustCompile(`^def\s+(?:pipeline\s+)?([A-Za-z_]\w*)\s*\(`)
	elixirDefpRe  = regexp.MustCompile(`^defp\s+([A-Za-z_]\w*)\s*\(`)
	elixirModuleRe = regexp.MustCompile(`^defmodule\s+([A-Za-z_]\w*)`)

	erlFuncRe     = regexp.MustCompile(`^([a-z_]\w*)\s*\(`)
	erlModuleRe   = regexp.MustCompile(`^-module\(([A-Za-z_]\w*)\)`)
	erlRecordRe   = regexp.MustCompile(`^-record\(([A-Za-z_]\w*)`)

	groovyClassRe = regexp.MustCompile(`^(?:public|private|protected|abstract|final|static)?\s*(?:class|interface|enum|trait|@interface)\s+([A-Za-z_]\w*)`)
	groovyDefRe   = regexp.MustCompile(`^(?:public|private|protected|static|final|abstract|def)\s+(?:def|void|[A-Za-z_][\w<>[\]]+)\s+([A-Za-z_]\w*)\s*\(`)

	cljDefRe      = regexp.MustCompile(`^\(def\s+([A-Za-z_][\w-]*)`)
	cljDefnRe     = regexp.MustCompile(`^\(defn\s+([A-Za-z_][\w-]*)`)
	cljDefmacroRe = regexp.MustCompile(`^\(defmacro\s+([A-Za-z_][\w-]*)`)
	cljProtocolRe = regexp.MustCompile(`^\(defprotocol\s+([A-Za-z_][\w-]*)`)
	cljRecordRe   = regexp.MustCompile(`^\(defrecord\s+([A-Za-z_][\w-]*)`)
	cljNsRe       = regexp.MustCompile(`^\(ns\s+([A-Za-z_][\w.-]*)`)

	fsModuleRe    = regexp.MustCompile(`^module\s+([A-Za-z_]\w*)`)
	fsLetRe       = regexp.MustCompile(`^let\s+([A-Za-z_]\w*)\s*=`)
	fsTypeRe      = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)

	ocamlLetRe    = regexp.MustCompile(`^let\s+(?:rec\s+)?([a-z_]\w*)`)
	ocamlTypeRe   = regexp.MustCompile(`^type\s+([a-z_]\w*)`)
	ocamlModuleRe = regexp.MustCompile(`^module\s+(?:type\s+)?([A-Za-z_]\w*)`)

	fortranFuncRe = regexp.MustCompile(`^(?:RECURSIVE\s+)?(?:FUNCTION|SUBROUTINE)\s+([A-Za-z_]\w*)`)
	fortranProgRe = regexp.MustCompile(`^PROGRAM\s+([A-Za-z_]\w*)`)
	fortranModRe  = regexp.MustCompile(`^MODULE\s+([A-Za-z_]\w*)`)

	adaFuncRe     = regexp.MustCompile(`^(?:procedure|function)\s+([A-Za-z_]\w*)`)
	adaPackageRe  = regexp.MustCompile(`^package\s+(?:body\s+)?([A-Za-z_]\w*)`)
	adaTaskRe     = regexp.MustCompile(`^task\s+(?:type\s+)?([A-Za-z_]\w*)`)
	adaTypeRe     = regexp.MustCompile(`^type\s+([A-Za-z_]\w*)`)

	protoMsgRe    = regexp.MustCompile(`^message\s+([A-Za-z_]\w*)`)
	protoEnumRe   = regexp.MustCompile(`^enum\s+([A-Za-z_]\w*)`)
	protoServiceRe = regexp.MustCompile(`^service\s+([A-Za-z_]\w*)`)
	protoRpcRe    = regexp.MustCompile(`^\s*rpc\s+([A-Za-z_]\w*)`)

	sqlCreateRe   = regexp.MustCompile(`(?i)^\s*create\s+(?:table|view|index|function|procedure|trigger)\s+(?:if\s+not\s+exists\s+)?(?:` + "`" + `)?(\w+)(?:` + "`" + `)?`)

	shFuncRe      = regexp.MustCompile(`^\s*(?:function\s+)?([A-Za-z_]\w*)\s*\(\s*\)`)
	shVarRe       = regexp.MustCompile(`^\s*([A-Z_][A-Z_0-9]*)=`)
)

func (b *Builder) extractSymbols(line string, lineNum int, ext string) []Symbol {
	switch ext {
	case ".go":
		return b.extractGoSymbols(line, lineNum)
	case ".ts", ".tsx", ".js", ".jsx", ".mjs", ".cjs":
		return b.extractTSSymbols(line, lineNum)
	case ".py":
		return b.extractPythonSymbols(line, lineNum)
	case ".java", ".jsp":
		return b.extractJavaSymbols(line, lineNum)
	case ".rs":
		return b.extractRustSymbols(line, lineNum)
	case ".cs":
		return b.extractCSharpSymbols(line, lineNum)
	case ".kt", ".kts":
		return b.extractKotlinSymbols(line, lineNum)
	case ".swift":
		return b.extractSwiftSymbols(line, lineNum)
	case ".php", ".phtml":
		return b.extractPHPSymbols(line, lineNum)
	case ".dart":
		return b.extractDartSymbols(line, lineNum)
	case ".scala", ".sc":
		return b.extractScalaSymbols(line, lineNum)
	case ".rb", ".rake":
		return b.extractRubySymbols(line, lineNum)
	case ".lua":
		return b.extractLuaSymbols(line, lineNum)
	case ".r", ".R":
		return b.extractRSymbols(line, lineNum)
	case ".hs", ".lhs":
		return b.extractHaskellSymbols(line, lineNum)
	case ".ex", ".exs":
		return b.extractElixirSymbols(line, lineNum)
	case ".erl", ".hrl":
		return b.extractErlangSymbols(line, lineNum)
	case ".groovy", ".gradle":
		return b.extractGroovySymbols(line, lineNum)
	case ".clj", ".cljs", ".cljc", ".edn":
		return b.extractClojureSymbols(line, lineNum)
	case ".fs", ".fsx":
		return b.extractFSharpSymbols(line, lineNum)
	case ".ml", ".mli":
		return b.extractOCamlSymbols(line, lineNum)
	case ".f", ".f90", ".f95", ".f03", ".for":
		return b.extractFortranSymbols(line, lineNum)
	case ".ada", ".adb", ".ads":
		return b.extractAdaSymbols(line, lineNum)
	case ".proto":
		return b.extractProtoSymbols(line, lineNum)
	case ".sql":
		return b.extractSQLSymbols(line, lineNum)
	case ".sh", ".bash", ".zsh":
		return b.extractShellSymbols(line, lineNum)
	case ".cr":
		return b.extractCrystalSymbols(line, lineNum)
	default:
		return nil
	}
}

func (b *Builder) extractGoSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol

	if matches := goStructRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolStruct, Line: lineNum, Exported: isExported(matches[1])})
		return symbols
	}
	if matches := goInterfaceRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: isExported(matches[1])})
		return symbols
	}
	if matches := goTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: isExported(matches[1])})
		return symbols
	}
	if matches := goFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: isExported(matches[1])})
		return symbols
	}

	return symbols
}

func (b *Builder) extractTSSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol

	if matches := tsClassRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsInterfaceRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsEnumRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolEnum, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsArrowRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := tsConstRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: true})
		return symbols
	}

	return symbols
}

func (b *Builder) extractPythonSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol

	if matches := pyClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.HasPrefix(matches[1], "_")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := pyFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.HasPrefix(matches[1], "_")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}

	return symbols
}

func (b *Builder) extractJavaSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := javaClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || len(matches[1]) > 0 && matches[1][0] >= 'A' && matches[1][0] <= 'Z'
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := javaMethodRe.FindStringSubmatch(line); len(matches) > 1 && !strings.Contains(line, "import ") {
		exported := strings.Contains(line, "public")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolMethod, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := javaFieldRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractRustSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := rsStructRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolStruct, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := rsEnumRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolEnum, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := rsTraitRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := rsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := rsTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := rsModRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.HasPrefix(strings.TrimSpace(line), "pub")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractCSharpSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := csClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := csMethodRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolMethod, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := csPropRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractKotlinSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := ktClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := ktFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := ktValRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractSwiftSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := swiftClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private") && !strings.Contains(line, "fileprivate")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := swiftFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private") && !strings.Contains(line, "fileprivate")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := swiftTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractPHPSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := phpClassRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := phpInterfaceRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := phpTraitRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := phpEnumRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolEnum, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := phpFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractDartSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := dartClassRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := dartFuncRe.FindStringSubmatch(line); len(matches) > 1 && !strings.Contains(line, "=>") {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := dartTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractScalaSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := scalaClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := scalaDefRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := scalaTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := scalaValRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractRubySymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := pyClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.HasPrefix(matches[1], "_")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := pyFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := !strings.HasPrefix(matches[1], "_")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	if strings.HasPrefix(strings.TrimSpace(line), "module ") {
		name := strings.TrimSpace(line)[7:]
		exported := !strings.HasPrefix(name, "_")
		symbols = append(symbols, Symbol{Name: name, Kind: SymbolType, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractLuaSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := luaFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := luaMethodRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := luaTableRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractRSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := rFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := rClassRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractHaskellSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := hsDataRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := hsTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := hsClassRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := hsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractElixirSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := elixirModuleRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := elixirDefRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := elixirDefpRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: false})
		return symbols
	}
	return symbols
}

func (b *Builder) extractErlangSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := erlModuleRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := erlFuncRe.FindStringSubmatch(line); len(matches) > 1 && !strings.Contains(line, "-") {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := erlRecordRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractGroovySymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := groovyClassRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: exported})
		return symbols
	}
	if matches := groovyDefRe.FindStringSubmatch(line); len(matches) > 1 {
		exported := strings.Contains(line, "public") || !strings.Contains(line, "private")
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: exported})
		return symbols
	}
	return symbols
}

func (b *Builder) extractClojureSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := cljDefnRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := cljDefmacroRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := cljDefRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := cljProtocolRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := cljRecordRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := cljNsRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractFSharpSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := fsModuleRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := fsLetRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := fsTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractOCamlSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := ocamlLetRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := ocamlTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := ocamlModuleRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractFortranSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := fortranFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := fortranModRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := fortranProgRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractAdaSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := adaPackageRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := adaFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := adaTaskRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := adaTypeRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractProtoSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := protoMsgRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := protoEnumRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolEnum, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := protoServiceRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolInterface, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := protoRpcRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractSQLSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := sqlCreateRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolType, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractShellSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := shFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := shVarRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolVariable, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func (b *Builder) extractCrystalSymbols(line string, lineNum int) []Symbol {
	var symbols []Symbol
	if matches := rsFuncRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolFunction, Line: lineNum, Exported: true})
		return symbols
	}
	if matches := rsStructRe.FindStringSubmatch(line); len(matches) > 1 {
		symbols = append(symbols, Symbol{Name: matches[1], Kind: SymbolClass, Line: lineNum, Exported: true})
		return symbols
	}
	return symbols
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func (rm *RepoMap) ToJSON() (string, error) {
	data, err := json.MarshalIndent(rm, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (rm *RepoMap) ToCompactJSON() (string, error) {
	data, err := json.Marshal(rm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (rm *RepoMap) FindSymbol(name string) []Symbol {
	var results []Symbol
	for _, file := range rm.Files {
		for _, sym := range file.Symbols {
			if sym.Name == name {
				results = append(results, sym)
			}
		}
	}
	return results
}
