package compress

import (
	"fmt"
	"strings"
)

type Lang int

const (
	LangUnknown Lang = iota
	LangGo
	LangTypeScript
	LangJavaScript
	LangPython
	LangRust
	LangJava
	LangCSharp
	LangRuby
	LangCPP
	LangC
	LangPHP
	LangDart
	LangKotlin
	LangScala
	LangSwift
	LangShell
	LangSQL
	LangHTML
	LangCSS
	LangMarkdown
	LangYAML
	LangTOML
	LangJSON
	LangVue
	LangSvelte
	LangHaskell
	LangLua
	LangR
	LangPerl
	LangElixir
	LangErlang
	LangZig
	LangNim
	LangJulia
	LangSolidity
	LangMakefile
	LangDockerfile
	LangGraphQL
	LangXML
	LangTeX
	LangPowerShell
	LangAssembly
	LangObjectiveC
	LangGroovy
	LangClojure
	LangFSharp
	LangOCaml
	LangFortran
	LangMATLAB
	LangAda
	LangCrystal
	LangZsh
	LangProtocolBuffers
)

type CodeSkeletonizer struct {
	MaxBodyLines int
}

func NewCodeSkeletonizer() *CodeSkeletonizer {
	return &CodeSkeletonizer{MaxBodyLines: 5}
}

func DetectLanguage(source, filename string) Lang {
	if filename != "" {
		switch {
		case strings.HasSuffix(filename, ".go"):
			return LangGo
		case strings.HasSuffix(filename, ".ts"), strings.HasSuffix(filename, ".tsx"), strings.HasSuffix(filename, ".mts"), strings.HasSuffix(filename, ".d.ts"):
			return LangTypeScript
		case strings.HasSuffix(filename, ".js"), strings.HasSuffix(filename, ".jsx"), strings.HasSuffix(filename, ".mjs"), strings.HasSuffix(filename, ".cjs"):
			return LangJavaScript
		case strings.HasSuffix(filename, ".vue"):
			return LangVue
		case strings.HasSuffix(filename, ".svelte"):
			return LangSvelte
		case strings.HasSuffix(filename, ".py"), strings.HasSuffix(filename, ".pyw"), strings.HasSuffix(filename, ".pyx"), strings.HasSuffix(filename, ".pxd"):
			return LangPython
		case strings.HasSuffix(filename, ".rs"):
			return LangRust
		case strings.HasSuffix(filename, ".java"), strings.HasSuffix(filename, ".jsp"), strings.HasSuffix(filename, ".class"):
			return LangJava
		case strings.HasSuffix(filename, ".cs"):
			return LangCSharp
		case strings.HasSuffix(filename, ".rb"), strings.HasSuffix(filename, ".rake"), strings.HasSuffix(filename, ".gemspec"), strings.HasSuffix(filename, ".erb"):
			return LangRuby
		case strings.HasSuffix(filename, ".cpp"), strings.HasSuffix(filename, ".cc"), strings.HasSuffix(filename, ".cxx"), strings.HasSuffix(filename, ".hpp"):
			return LangCPP
		case strings.HasSuffix(filename, ".c"), strings.HasSuffix(filename, ".h"):
			return LangC
		case strings.HasSuffix(filename, ".php"), strings.HasSuffix(filename, ".phtml"), strings.HasSuffix(filename, ".php3"), strings.HasSuffix(filename, ".php4"), strings.HasSuffix(filename, ".php5"):
			return LangPHP
		case strings.HasSuffix(filename, ".dart"):
			return LangDart
		case strings.HasSuffix(filename, ".kt"), strings.HasSuffix(filename, ".kts"):
			return LangKotlin
		case strings.HasSuffix(filename, ".scala"), strings.HasSuffix(filename, ".sc"):
			return LangScala
		case strings.HasSuffix(filename, ".swift"):
			return LangSwift
		case strings.HasSuffix(filename, ".sh"), strings.HasSuffix(filename, ".bash"), strings.HasSuffix(filename, ".zsh"), strings.HasSuffix(filename, ".fish"):
			return LangShell
		case strings.HasSuffix(filename, ".sql"):
			return LangSQL
		case strings.HasSuffix(filename, ".html"), strings.HasSuffix(filename, ".htm"), strings.HasSuffix(filename, ".xhtml"):
			return LangHTML
		case strings.HasSuffix(filename, ".css"), strings.HasSuffix(filename, ".scss"), strings.HasSuffix(filename, ".sass"), strings.HasSuffix(filename, ".less"), strings.HasSuffix(filename, ".styl"):
			return LangCSS
		case strings.HasSuffix(filename, ".md"), strings.HasSuffix(filename, ".mdx"):
			return LangMarkdown
		case strings.HasSuffix(filename, ".yaml"), strings.HasSuffix(filename, ".yml"):
			return LangYAML
		case strings.HasSuffix(filename, ".toml"):
			return LangTOML
		case strings.HasSuffix(filename, ".json"), strings.HasSuffix(filename, ".jsonc"):
			return LangJSON
		case strings.HasSuffix(filename, ".hs"), strings.HasSuffix(filename, ".lhs"):
			return LangHaskell
		case strings.HasSuffix(filename, ".lua"):
			return LangLua
		case strings.HasSuffix(filename, ".r"), strings.HasSuffix(filename, ".R"), strings.HasSuffix(filename, ".rmd"):
			return LangR
		case strings.HasSuffix(filename, ".pl"), strings.HasSuffix(filename, ".pm"), strings.HasSuffix(filename, ".t"):
			return LangPerl
		case strings.HasSuffix(filename, ".ex"), strings.HasSuffix(filename, ".exs"):
			return LangElixir
		case strings.HasSuffix(filename, ".erl"), strings.HasSuffix(filename, ".hrl"):
			return LangErlang
		case strings.HasSuffix(filename, ".zig"):
			return LangZig
		case strings.HasSuffix(filename, ".nim"):
			return LangNim
		case strings.HasSuffix(filename, ".jl"):
			return LangJulia
		case strings.HasSuffix(filename, ".sol"):
			return LangSolidity
		case strings.HasSuffix(filename, "Makefile"), strings.HasSuffix(filename, "makefile"), strings.HasSuffix(filename, "GNUmakefile"):
			return LangMakefile
		case strings.HasSuffix(filename, "Dockerfile"), strings.HasSuffix(filename, "dockerfile"):
			return LangDockerfile
		case strings.HasSuffix(filename, ".graphql"), strings.HasSuffix(filename, ".gql"):
			return LangGraphQL
		case strings.HasSuffix(filename, ".xml"), strings.HasSuffix(filename, ".svg"), strings.HasSuffix(filename, ".xsd"), strings.HasSuffix(filename, ".xslt"), strings.HasSuffix(filename, ".plist"), strings.HasSuffix(filename, ".xsl"):
			return LangXML
		case strings.HasSuffix(filename, ".tex"), strings.HasSuffix(filename, ".sty"), strings.HasSuffix(filename, ".cls"), strings.HasSuffix(filename, ".ltx"), strings.HasSuffix(filename, ".bib"):
			return LangTeX
		case strings.HasSuffix(filename, ".ps1"), strings.HasSuffix(filename, ".psm1"), strings.HasSuffix(filename, ".psd1"):
			return LangPowerShell
		case strings.HasSuffix(filename, ".asm"), strings.HasSuffix(filename, ".s"), strings.HasSuffix(filename, ".inc"):
			return LangAssembly
		case strings.HasSuffix(filename, ".m"), strings.HasSuffix(filename, ".mm"):
			return LangObjectiveC
		case strings.HasSuffix(filename, ".groovy"), strings.HasSuffix(filename, ".gvy"), strings.HasSuffix(filename, ".gant"), strings.HasSuffix(filename, ".gradle"):
			return LangGroovy
		case strings.HasSuffix(filename, ".clj"), strings.HasSuffix(filename, ".cljs"), strings.HasSuffix(filename, ".cljc"), strings.HasSuffix(filename, ".edn"):
			return LangClojure
		case strings.HasSuffix(filename, ".fs"), strings.HasSuffix(filename, ".fsx"), strings.HasSuffix(filename, ".fsscript"):
			return LangFSharp
		case strings.HasSuffix(filename, ".ml"), strings.HasSuffix(filename, ".mli"):
			return LangOCaml
		case strings.HasSuffix(filename, ".f"), strings.HasSuffix(filename, ".f90"), strings.HasSuffix(filename, ".f95"), strings.HasSuffix(filename, ".f03"), strings.HasSuffix(filename, ".for"):
			return LangFortran
		case strings.HasSuffix(filename, ".mat"):
			return LangMATLAB
		case strings.HasSuffix(filename, ".ada"), strings.HasSuffix(filename, ".adb"), strings.HasSuffix(filename, ".ads"):
			return LangAda
		case strings.HasSuffix(filename, ".cr"):
			return LangCrystal
		case strings.HasSuffix(filename, ".proto"):
			return LangProtocolBuffers
		}
	}

	if source != "" {
		firstLine := source
		if idx := strings.Index(source, "\n"); idx >= 0 {
			firstLine = source[:idx]
		}
		trimmed := strings.TrimSpace(firstLine)

		if strings.HasPrefix(trimmed, "#!/bin/sh") || strings.HasPrefix(trimmed, "#!/bin/bash") || strings.HasPrefix(trimmed, "#!/usr/bin/env bash") || strings.HasPrefix(trimmed, "#!/bin/zsh") {
			return LangShell
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env python") || strings.HasPrefix(trimmed, "#!/usr/bin/python") {
			return LangPython
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env ruby") || strings.HasPrefix(trimmed, "#!/usr/bin/ruby") {
			return LangRuby
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env node") || strings.HasPrefix(trimmed, "#!/usr/bin/node") {
			return LangJavaScript
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env perl") || strings.HasPrefix(trimmed, "#!/usr/bin/perl") {
			return LangPerl
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env groovy") || strings.HasPrefix(trimmed, "#!/usr/bin/groovy") {
			return LangGroovy
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env pwsh") || strings.HasPrefix(trimmed, "#!/usr/bin/pwsh") {
			return LangPowerShell
		}
		if strings.HasPrefix(trimmed, "#!/usr/bin/env lua") || strings.HasPrefix(trimmed, "#!/usr/bin/lua") {
			return LangLua
		}
		if strings.HasPrefix(trimmed, "<?php") {
			return LangPHP
		}
		if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<svg") {
			return LangXML
		}
		if strings.HasPrefix(trimmed, "<template") || strings.HasPrefix(trimmed, "<script") {
			return LangVue
		}
		if strings.HasPrefix(trimmed, "---") && strings.Contains(source, ":") {
			return LangYAML
		}
		if strings.HasPrefix(trimmed, "\\documentclass") || strings.HasPrefix(trimmed, "\\begin{") {
			return LangTeX
		}
	}

	sourceTrim := strings.TrimSpace(source)
	if strings.Contains(source, "package ") && (strings.Contains(source, "import (") || strings.Contains(source, "func ")) {
		return LangGo
	}
	if strings.Contains(source, "useEffect") || strings.Contains(source, "useState") || strings.Contains(source, "import React") || strings.Contains(source, "export default") {
		if strings.Contains(source, ":") && !strings.Contains(source, "def ") {
			return LangTypeScript
		}
		return LangJavaScript
	}
	if strings.Contains(source, "fn ") && strings.Contains(source, "->") {
		return LangRust
	}
	if (strings.Contains(source, "public class") || strings.Contains(source, "public static void")) && strings.Contains(source, "{") {
		return LangJava
	}
	if strings.Contains(source, "using System") || strings.Contains(source, "namespace ") {
		return LangCSharp
	}
	if strings.Contains(source, "<?php") {
		return LangPHP
	}
	if strings.Contains(source, "fun ") && strings.Contains(source, ":") && !strings.Contains(source, "def ") {
		return LangKotlin
	}
	if strings.Contains(source, "def ") && strings.Contains(source, "end") {
		return LangRuby
	}
	if strings.Contains(source, "fn ") && strings.Contains(source, "=>") && !strings.Contains(source, "def ") {
		return LangElixir
	}
	if strings.Contains(source, "CREATE TABLE") || strings.Contains(source, "SELECT ") {
		return LangSQL
	}
	if strings.Contains(sourceTrim, "edn/") || (strings.Contains(source, "(") && strings.Contains(source, "defn ")) {
		return LangClojure
	}
	if strings.Contains(source, "module ") && strings.Contains(source, "where") && (strings.Contains(source, "data ") || strings.Contains(source, "type ")) {
		return LangHaskell
	}
	if strings.Contains(source, "fun main") || strings.Contains(source, "println(") {
		return LangKotlin
	}
	if strings.Contains(source, "function ") && strings.Contains(source, "=>") {
		return LangTypeScript
	}
	if strings.Contains(source, "function ") {
		return LangJavaScript
	}
	if strings.Contains(source, "def ") {
		return LangPython
	}

	return LangUnknown
}

func (s *CodeSkeletonizer) Skeletonize(source string, filename string) string {
	if len(source) == 0 {
		return ""
	}

	lines := strings.Split(source, "\n")
	if len(lines) <= 20 {
		return source
	}

	lang := DetectLanguage(source, filename)
	stripTrailingEmptyLines(&lines)

	var result []string
	var skippedLines int

	switch lang {
	case LangGo, LangRust, LangJava, LangCSharp, LangKotlin, LangSwift, LangScala, LangGroovy:
		result = s.skeletonizeFuncBody(lines, &skippedLines)
	case LangTypeScript, LangJavaScript:
		result = s.skeletonizeFuncBody(lines, &skippedLines)
	case LangPython:
		result = s.skeletonizePython(lines, &skippedLines)
	case LangRuby:
		result = s.skeletonizeRuby(lines, &skippedLines)
	case LangShell:
		result = s.skeletonizeShell(lines, &skippedLines)
	case LangSQL:
		result = s.skeletonizeSQL(lines, &skippedLines)
	case LangHTML, LangVue, LangSvelte:
		result = s.skeletonizeTagBased(lines, &skippedLines)
	case LangXML:
		result = s.skeletonizeXML(lines, &skippedLines)
	case LangCSS:
		result = s.skeletonizeCSS(lines, &skippedLines)
	case LangJSON, LangYAML, LangTOML:
		result = s.skeletonizeStructuredData(lines, &skippedLines)
	case LangMarkdown:
		result = s.skeletonizeMarkdown(lines, &skippedLines)
	case LangMakefile, LangDockerfile:
		result = s.skeletonizeMakefile(lines, &skippedLines)
	case LangGraphQL:
		result = s.skeletonizeBraced(lines, lang, &skippedLines)
	default:
		result = s.skeletonizeBraced(lines, lang, &skippedLines)
	}

	if skippedLines > 0 {
		result = append(result, fmt.Sprintf("[skeleton: %d total lines, %d lines omitted]", len(lines), skippedLines))
	}

	return strings.Join(result, "\n")
}

func isCommentLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "<!--") ||
		strings.HasPrefix(trimmed, "%") || strings.HasPrefix(trimmed, "\"\"\"") ||
		strings.HasPrefix(trimmed, "'''")
}

// skeletonizeFuncBody handles languages where functions are delimited by braces.
// It tracks total lines shown per top-level block (not just consecutive non-brace lines),
// so functions with lots of control flow still get truncated once they're long enough.
func (s *CodeSkeletonizer) skeletonizeFuncBody(lines []string, skippedLines *int) []string {
	var result []string
	depth := 0
	blockLines := 0 // lines shown inside current depth-1 block
	omittedCount := 0
	omitting := false
	limit := s.MaxBodyLines * 3 // ~15 lines shown per function before omission kicks in

	flushOmit := func() {
		if omitting {
			result = append(result, fmt.Sprintf("    [%d lines omitted]", omittedCount))
			omitting = false
			omittedCount = 0
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		opens := strings.Count(trimmed, "{")
		closes := strings.Count(trimmed, "}")
		isComment := isCommentLine(trimmed)
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if depth <= 1 {
				flushOmit()
			}
			result = append(result, line)
			continue
		}

		prevDepth := depth
		depth += opens - closes
		if depth < 0 {
			depth = 0
		}

		// Returning to top level: flush any pending omission and show closing brace
		if prevDepth > 0 && depth == 0 {
			flushOmit()
			blockLines = 0
			result = append(result, line)
			continue
		}

		// At top level (depth == 0 before and after): always show (type defs, var, etc.)
		if prevDepth == 0 && depth == 0 {
			result = append(result, line)
			continue
		}

		// Opening a top-level block (e.g. func signature with {): show and reset counter
		if prevDepth == 0 && depth > 0 {
			result = append(result, line)
			blockLines = 0
			omitting = false
			omittedCount = 0
			continue
		}

		// Inside a block: count total lines shown and omit beyond limit
		if blockLines < limit {
			flushOmit()
			result = append(result, line)
			blockLines++
		} else {
			omitting = true
			omittedCount++
			*skippedLines++
		}
	}

	flushOmit()

	return result
}

func (s *CodeSkeletonizer) skeletonizeBraced(lines []string, lang Lang, skippedLines *int) []string {
	var result []string
	depth := 0
	bodyLines := 0
	omitted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := isCommentLine(trimmed)
		isEmpty := trimmed == ""
		opens := strings.Count(trimmed, "{")
		closes := strings.Count(trimmed, "}")
		hasBraces := opens > 0 || closes > 0

		if depth == 0 || isComment || isEmpty || hasBraces {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}

			if !isEmpty || len(result) == 0 || strings.TrimSpace(result[len(result)-1]) != "" {
				result = append(result, line)
			}

			if opens > 0 {
				depth += opens
			}
			if closes > 0 {
				depth -= closes
			}
			bodyLines = 0
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizePython(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			if trimmed != "" || !omitted {
				result = append(result, line)
			}
			bodyLines = 0
			continue
		}

		lineIndent := len(line) - len(strings.TrimLeft(line, " \t"))
		isDefOrClass := strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ")
		isTopLevel := lineIndent == 0
		endsColon := strings.HasSuffix(trimmed, ":")

		if isDefOrClass || isTopLevel || (endsColon && isTopLevel) {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeRuby(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false

	keywordStack := make([]bool, 0)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		isDef := strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "module ") || strings.HasPrefix(trimmed, "if ") ||
			strings.HasPrefix(trimmed, "unless ") || strings.HasPrefix(trimmed, "while ") ||
			strings.HasPrefix(trimmed, "until ") || strings.HasPrefix(trimmed, "for ") ||
			strings.HasPrefix(trimmed, "begin ") || strings.HasPrefix(trimmed, "case ") ||
			strings.HasPrefix(trimmed, "do ") || strings.HasSuffix(trimmed, " do") ||
			strings.HasPrefix(trimmed, "private") || strings.HasPrefix(trimmed, "public") ||
			strings.HasPrefix(trimmed, "protected")

		isEnd := trimmed == "end" || strings.HasPrefix(trimmed, "end ")

		isElse := strings.HasPrefix(trimmed, "else") || strings.HasPrefix(trimmed, "elsif ") ||
			strings.HasPrefix(trimmed, "rescue ") || strings.HasPrefix(trimmed, "ensure ") ||
			strings.HasPrefix(trimmed, "when ")

		if isElse {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		if isDef {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			keywordStack = append(keywordStack, true)
			bodyLines = 0
			continue
		}

		if isEnd {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			if len(keywordStack) > 0 {
				keywordStack = keywordStack[:len(keywordStack)-1]
			}
			bodyLines = 0
			continue
		}

		if len(keywordStack) > 0 {
			bodyLines++
			if bodyLines <= s.MaxBodyLines {
				result = append(result, line)
			} else if !omitted {
				omitted = true
				*skippedLines++
			} else {
				*skippedLines++
			}
		} else {
			result = append(result, line)
			bodyLines = 0
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeShell(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false
	inFunc := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		isFuncDef := strings.HasSuffix(trimmed, "()") || strings.HasSuffix(trimmed, "() {") ||
			strings.HasPrefix(trimmed, "function ")
		isCase := strings.HasPrefix(trimmed, "case ") || strings.HasPrefix(trimmed, "esac") ||
			strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "fi") ||
			strings.HasPrefix(trimmed, "then") || strings.HasPrefix(trimmed, "else") ||
			strings.HasPrefix(trimmed, "elif ") || strings.HasPrefix(trimmed, "for ") ||
			strings.HasPrefix(trimmed, "while ") || strings.HasPrefix(trimmed, "until ") ||
			strings.HasPrefix(trimmed, "do") || strings.HasPrefix(trimmed, "done") ||
			strings.HasPrefix(trimmed, "select ") || strings.HasPrefix(trimmed, "in ")

		isStructural := isFuncDef || isCase || strings.HasPrefix(trimmed, "export ") ||
			strings.HasPrefix(trimmed, "alias ") || strings.HasPrefix(trimmed, "source ") ||
			strings.HasPrefix(trimmed, "trap ") || strings.HasPrefix(trimmed, "set ") ||
			strings.HasPrefix(trimmed, "local ") || strings.HasPrefix(trimmed, "readonly ")

		if isFuncDef {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			inFunc = true
			bodyLines = 0
			continue
		}

		if strings.HasPrefix(trimmed, "}") {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			inFunc = false
			bodyLines = 0
			continue
		}

		if inFunc {
			bodyLines++
			if bodyLines <= s.MaxBodyLines {
				result = append(result, line)
			} else if !omitted {
				omitted = true
				*skippedLines++
			} else {
				*skippedLines++
			}
		} else {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			if isStructural || bodyLines == 0 {
				result = append(result, line)
			} else {
				bodyLines++
				if bodyLines <= s.MaxBodyLines {
					result = append(result, line)
				} else if !omitted {
					omitted = true
					*skippedLines++
				} else {
					*skippedLines++
				}
			}
			if isStructural {
				bodyLines = 0
			}
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeSQL(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false

	statementKeywords := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "ALTER", "DROP",
		"TRUNCATE", "MERGE", "CALL", "EXEC", "DECLARE", "SET", "BEGIN",
		"COMMIT", "ROLLBACK", "SAVEPOINT", "GRANT", "REVOKE", "WITH",
		"TABLE", "VIEW", "INDEX", "FUNCTION", "PROCEDURE", "TRIGGER",
		"SCHEMA", "DATABASE", "USE",
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "#")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		upper := strings.ToUpper(trimmed)
		isStatement := false
		for _, kw := range statementKeywords {
			if strings.HasPrefix(upper, kw) {
				isStatement = true
				break
			}
		}

		if isStatement || strings.HasPrefix(trimmed, ")") {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeTagBased(lines []string, skippedLines *int) []string {
	var result []string
	depth := 0
	bodyLines := 0
	omitted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "<!--") || strings.HasPrefix(trimmed, "{#") || strings.HasPrefix(trimmed, "{/*")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		opens := strings.Count(trimmed, "<") - strings.Count(trimmed, "</")
		closes := strings.Count(trimmed, "</") + strings.Count(trimmed, "/>")
		hasTag := strings.HasPrefix(trimmed, "<") || strings.Contains(trimmed, "</") || strings.Contains(trimmed, "/>")

		if depth == 0 || hasTag || trimmed == "" {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			if opens > 0 {
				depth += 1
			}
			if closes > 0 {
				depth -= closes
			}
			if depth < 0 {
				depth = 0
			}
			bodyLines = 0
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeCSS(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false
	inSelector := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		isAtRule := strings.HasPrefix(trimmed, "@")

		if isAtRule || strings.HasSuffix(trimmed, "{") || strings.HasPrefix(trimmed, "}") {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			inSelector = strings.HasSuffix(trimmed, "{")
			bodyLines = 0
			continue
		}

		if inSelector {
			bodyLines++
			if bodyLines <= s.MaxBodyLines {
				result = append(result, line)
			} else if !omitted {
				omitted = true
				*skippedLines++
			} else {
				*skippedLines++
			}
		} else {
			result = append(result, line)
			bodyLines = 0
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeXML(lines []string, skippedLines *int) []string {
	var result []string
	omitted := false
	depth := 0
	depth1shown := 0
	depth1omitted := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "<!--") || strings.HasPrefix(trimmed, "<?")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", depth1omitted))
				omitted = false
			}
			result = append(result, line)
			continue
		}

		netOpen := strings.Count(trimmed, "<") - strings.Count(trimmed, "</") - strings.Count(trimmed, "/>")
		netClose := strings.Count(trimmed, "</") + strings.Count(trimmed, "/>")
		isClose := strings.HasPrefix(trimmed, "</")

		if isClose {
			depth -= netClose
			if depth < 0 {
				depth = 0
			}
			if omitted && depth == 0 {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", depth1omitted))
				omitted = false
				depth1omitted = 0
			}
			if !omitted || depth == 0 {
				result = append(result, line)
			}
			continue
		}

		netDepth := netOpen - netClose
		prevDepth := depth
		depth += netDepth
		if depth < 0 {
			depth = 0
		}

		if omitted {
			if depth == 0 {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", depth1omitted))
				omitted = false
				depth1omitted = 0
				result = append(result, line)
				continue
			}
			if prevDepth == 1 && depth > prevDepth {
				depth1omitted++
			}
			*skippedLines++
			continue
		}

		if prevDepth == 1 && depth > prevDepth {
			depth1shown++
			if depth1shown > s.MaxBodyLines {
				omitted = true
				depth1omitted = 1
				*skippedLines++
				continue
			}
		}

		result = append(result, line)
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", depth1omitted))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeStructuredData(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false
	depth := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "//")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		isKey := strings.Contains(trimmed, ":") || strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "-")
		isClosing := strings.HasPrefix(trimmed, "]") || strings.HasPrefix(trimmed, "}")

		if indent == 0 && depth <= 1 || isKey && indent <= depth || isClosing {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			if strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") {
				depth++
			}
			if strings.HasPrefix(trimmed, "]") || strings.HasPrefix(trimmed, "}") {
				depth--
			}
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeMarkdown(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isHeading := strings.HasPrefix(trimmed, "#")
		isCodeFence := strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
		isList := strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") ||
			strings.HasPrefix(trimmed, "1.") || strings.HasPrefix(trimmed, "1)")
		isBlockquote := strings.HasPrefix(trimmed, "> ")
		isThematic := trimmed == "---" || trimmed == "***" || trimmed == "___"

		isStructural := isHeading || isCodeFence || isList || isBlockquote || isThematic ||
			strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "![") ||
			strings.HasPrefix(trimmed, "|")

		if isStructural || trimmed == "" {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		bodyLines++
		if bodyLines <= s.MaxBodyLines {
			result = append(result, line)
		} else if !omitted {
			omitted = true
			*skippedLines++
		} else {
			*skippedLines++
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func (s *CodeSkeletonizer) skeletonizeMakefile(lines []string, skippedLines *int) []string {
	var result []string
	bodyLines := 0
	omitted := false
	inRecipe := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(trimmed, "#")
		isEmpty := trimmed == ""

		if isComment || isEmpty {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		isTarget := strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "\t") && !strings.HasPrefix(trimmed, " ")
		isVariable := strings.Contains(trimmed, ":=") || strings.Contains(trimmed, "?=") || strings.Contains(trimmed, "+=")
		isDirective := strings.HasPrefix(trimmed, "include ") || strings.HasPrefix(trimmed, "export ") ||
			strings.HasPrefix(trimmed, "ifdef") || strings.HasPrefix(trimmed, "ifndef") ||
			strings.HasPrefix(trimmed, "ifeq") || strings.HasPrefix(trimmed, "ifneq") ||
			strings.HasPrefix(trimmed, "else") || strings.HasPrefix(trimmed, "endif")

		if isTarget {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			inRecipe = true
			bodyLines = 0
			continue
		}

		if isVariable || isDirective {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			bodyLines = 0
			continue
		}

		if inRecipe && (strings.HasPrefix(line, "\t") || strings.HasPrefix(line, " ")) {
			bodyLines++
			if bodyLines <= s.MaxBodyLines {
				result = append(result, line)
			} else if !omitted {
				omitted = true
				*skippedLines++
			} else {
				*skippedLines++
			}
		} else {
			if omitted {
				result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
				omitted = false
			}
			result = append(result, line)
			inRecipe = false
			bodyLines = 0
		}
	}

	if omitted {
		result = append(result, fmt.Sprintf("    [%d lines of implementation omitted]", bodyLines))
	}

	return result
}

func stripTrailingEmptyLines(lines *[]string) {
	for len(*lines) > 0 && strings.TrimSpace((*lines)[len(*lines)-1]) == "" {
		*lines = (*lines)[:len(*lines)-1]
	}
}
