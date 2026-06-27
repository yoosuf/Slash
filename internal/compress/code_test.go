package compress

import (
	"fmt"
	"strings"
	"testing"
)

func TestDetectLanguage_Go(t *testing.T) {
	source := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`
	lang := DetectLanguage(source, "main.go")
	if lang != LangGo {
		t.Fatalf("expected LangGo, got %v", lang)
	}
}

func TestDetectLanguage_Python(t *testing.T) {
	source := `def hello():
    print("hello")`
	lang := DetectLanguage(source, "test.py")
	if lang != LangPython {
		t.Fatalf("expected LangPython, got %v", lang)
	}
}

func TestDetectLanguage_TypeScript(t *testing.T) {
	source := `export class MyClass {
	method(): void {}`
	lang := DetectLanguage(source, "test.ts")
	if lang != LangTypeScript {
		t.Fatalf("expected LangTypeScript, got %v", lang)
	}
}

func TestDetectLanguage_Unknown(t *testing.T) {
	source := `some random text`
	lang := DetectLanguage(source, "readme.txt")
	if lang != LangUnknown {
		t.Fatalf("expected LangUnknown, got %v", lang)
	}
}

func TestSkeletonize_SmallCode(t *testing.T) {
	sk := NewCodeSkeletonizer()
	source := "package main\n\nfunc main() {\n\tprintln(\"hi\")\n}"
	result := sk.Skeletonize(source, "")
	if result != source {
		t.Fatalf("small code should not be skeletonized, got different output")
	}
}

func TestSkeletonize_LargeGoFunction(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("package main\n\nfunc longFunc() {\n")
	for i := 0; i < 20; i++ {
		b.WriteString("\tfmt.Println(\"line\")\n")
	}
	b.WriteString("}\n")

	source := b.String()
	result := sk.Skeletonize(source, "")

	if result == source {
		t.Fatal("expected large function body to be collapsed")
	}
	if !strings.Contains(result, "lines omitted") {
		t.Fatal("expected omission message in result")
	}
}

func TestSkeletonize_GoMultipleFunctions(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("package main\n")
	for i := 0; i < 5; i++ {
		b.WriteString("func shortFunc() {\n\tfmt.Println(\"hi\")\n}\n")
	}
	source := b.String()
	result := sk.Skeletonize(source, "")

	if result != source {
		t.Fatalf("short functions should not be collapsed\nsource=%q\nresult=%q", source, result)
	}
}

func TestSkeletonize_PythonClass(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("class MyClass:\n")
	for i := 0; i < 3; i++ {
		b.WriteString("\tdef method(self):\n\t\tpass\n\n")
	}
	source := b.String()
	result := sk.Skeletonize(source, "")

	if !strings.Contains(result, "class MyClass") {
		t.Fatal("class definition should be kept")
	}
}

func TestSkeletonize_LargePythonMethod(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("class MyClass:\n\tdef long_method(self):\n")
	for i := 0; i < 30; i++ {
		b.WriteString("\t\tprint(\"line\")\n")
	}
	source := b.String()
	result := sk.Skeletonize(source, "")

	if result == source {
		t.Fatal("expected long Python method body to be collapsed")
	}
	if !strings.Contains(result, "lines omitted") {
		t.Fatal("expected omission message in result")
	}
}

func TestSkeletonize_GoNestedBlocks(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("package main\n\nfunc process() {\n")
	for i := 0; i < 3; i++ {
		b.WriteString("\tif true {\n\t\tfmt.Println(\"nested\")\n\t}\n")
	}
	b.WriteString("\tfmt.Println(\"done\")\n}\n")
	source := b.String()
	result := sk.Skeletonize(source, "")

	if result != source {
		t.Fatal("short nested blocks should not be collapsed")
	}
	if !strings.Contains(result, "if true") {
		t.Fatal("if statements should be visible")
	}
}

func TestSkeletonize_EmptySource(t *testing.T) {
	sk := NewCodeSkeletonizer()
	result := sk.Skeletonize("", "")
	if result != "" {
		t.Fatalf("expected empty result for empty source")
	}
}

func TestSkeletonize_PreserveComments(t *testing.T) {
	sk := NewCodeSkeletonizer()
	source := "package main\n\n// This is a comment\n/* block comment */\nfunc foo() {\n\t// inline comment\n\tdoWork()\n}"
	result := sk.Skeletonize(source, "")
	if !strings.Contains(result, "// This is a comment") {
		t.Fatal("comments should be preserved")
	}
}

func TestSkeletonize_PythonDocstring(t *testing.T) {
	sk := NewCodeSkeletonizer()
	source := "def foo():\n\t\"\"\"docstring\"\"\"\n\tpass"
	result := sk.Skeletonize(source, "")
	if !strings.Contains(result, "def foo()") {
		t.Fatal("function def should be preserved")
	}
}

func TestSkeletonize_TypeScript(t *testing.T) {
	sk := NewCodeSkeletonizer()
	source := "export class Handler {\n\texecute(param: string): void {\n\t\tconsole.log(param);\n\t}\n}"
	result := sk.Skeletonize(source, "test.ts")
	if !strings.Contains(result, "export class Handler") {
		t.Fatal("class should be preserved")
	}
}

func TestSkeletonize_BracedBlocksOnly(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("var data = {\n")
	for i := 0; i < 10; i++ {
		b.WriteString("\t\"key\": \"value\",\n")
	}
	b.WriteString("}\n")
	source := b.String()
	result := sk.Skeletonize(source, "")

	if result != source {
		t.Fatal("object literal should not be collapsed as code skeleton")
	}
}

func TestDetectLanguage_XML(t *testing.T) {
	tests := []struct {
		filename string
		ext      string
	}{
		{"test.xml", ".xml"},
		{"icon.svg", ".svg"},
		{"schema.xsd", ".xsd"},
		{"transform.xslt", ".xslt"},
	}
	for _, tt := range tests {
		lang := DetectLanguage("", tt.filename)
		if lang != LangXML {
			t.Fatalf("expected LangXML for %s, got %v", tt.filename, lang)
		}
	}
}

func TestDetectLanguage_XMLByContent(t *testing.T) {
	lang := DetectLanguage(`<?xml version="1.0"?><root><item/></root>`, "")
	if lang != LangXML {
		t.Fatalf("expected LangXML from content, got %v", lang)
	}
}

func TestDetectLanguage_JavaExtensions(t *testing.T) {
	lang := DetectLanguage("", "Main.java")
	if lang != LangJava {
		t.Fatalf("expected LangJava for Main.java, got %v", lang)
	}
}

func TestDetectLanguage_JavaByContent(t *testing.T) {
	lang := DetectLanguage(`public class HelloWorld { public static void main(String[] args) {} }`, "")
	if lang != LangJava {
		t.Fatalf("expected LangJava from content, got %v", lang)
	}
}

func TestDetectLanguage_TeX(t *testing.T) {
	tests := []string{".tex", ".sty", ".cls", ".ltx", ".bib"}
	for _, ext := range tests {
		lang := DetectLanguage("", "file"+ext)
		if lang != LangTeX {
			t.Fatalf("expected LangTeX for %s, got %v", ext, lang)
		}
	}
}

func TestDetectLanguage_TeXByContent(t *testing.T) {
	lang := DetectLanguage(`\documentclass{article}`, "")
	if lang != LangTeX {
		t.Fatalf("expected LangTeX from content, got %v", lang)
	}
}

func TestDetectLanguage_PowerShell(t *testing.T) {
	lang := DetectLanguage("", "script.ps1")
	if lang != LangPowerShell {
		t.Fatalf("expected LangPowerShell, got %v", lang)
	}
}

func TestDetectLanguage_Assembly(t *testing.T) {
	lang := DetectLanguage("", "code.asm")
	if lang != LangAssembly {
		t.Fatalf("expected LangAssembly, got %v", lang)
	}
}

func TestDetectLanguage_ObjectiveC(t *testing.T) {
	lang := DetectLanguage("", "impl.m")
	if lang != LangObjectiveC {
		t.Fatalf("expected LangObjectiveC, got %v", lang)
	}
}

func TestDetectLanguage_Groovy(t *testing.T) {
	lang := DetectLanguage("", "build.gradle")
	if lang != LangGroovy {
		t.Fatalf("expected LangGroovy, got %v", lang)
	}
}

func TestDetectLanguage_Clojure(t *testing.T) {
	lang := DetectLanguage("", "core.clj")
	if lang != LangClojure {
		t.Fatalf("expected LangClojure, got %v", lang)
	}
}

func TestDetectLanguage_FSharp(t *testing.T) {
	lang := DetectLanguage("", "lib.fs")
	if lang != LangFSharp {
		t.Fatalf("expected LangFSharp, got %v", lang)
	}
}

func TestDetectLanguage_OCaml(t *testing.T) {
	lang := DetectLanguage("", "lib.ml")
	if lang != LangOCaml {
		t.Fatalf("expected LangOCaml, got %v", lang)
	}
}

func TestDetectLanguage_Fortran(t *testing.T) {
	tests := []string{".f90", ".f95", ".f", ".for"}
	for _, ext := range tests {
		lang := DetectLanguage("", "code"+ext)
		if lang != LangFortran {
			t.Fatalf("expected LangFortran for %s, got %v", ext, lang)
		}
	}
}

func TestDetectLanguage_MATLAB(t *testing.T) {
	lang := DetectLanguage("", "matrix.mat")
	if lang != LangMATLAB {
		t.Fatalf("expected LangMATLAB, got %v", lang)
	}
}

func TestDetectLanguage_Ada(t *testing.T) {
	tests := []string{".ada", ".adb", ".ads"}
	for _, ext := range tests {
		lang := DetectLanguage("", "pkg"+ext)
		if lang != LangAda {
			t.Fatalf("expected LangAda for %s, got %v", ext, lang)
		}
	}
}

func TestDetectLanguage_Crystal(t *testing.T) {
	lang := DetectLanguage("", "app.cr")
	if lang != LangCrystal {
		t.Fatalf("expected LangCrystal, got %v", lang)
	}
}

func TestDetectLanguage_Protobuf(t *testing.T) {
	lang := DetectLanguage("", "message.proto")
	if lang != LangProtocolBuffers {
		t.Fatalf("expected LangProtocolBuffers, got %v", lang)
	}
}

func TestDetectLanguage_ShebangPowerShell(t *testing.T) {
	lang := DetectLanguage("#!/usr/bin/env pwsh\nGet-ChildItem\n", "")
	if lang != LangPowerShell {
		t.Fatalf("expected LangPowerShell from shebang, got %v", lang)
	}
}

func TestDetectLanguage_ShebangGroovy(t *testing.T) {
	lang := DetectLanguage("#!/usr/bin/env groovy\nprintln 'hello'", "")
	if lang != LangGroovy {
		t.Fatalf("expected LangGroovy from shebang, got %v", lang)
	}
}

func TestDetectLanguage_SCSS(t *testing.T) {
	lang := DetectLanguage("", "style.scss")
	if lang != LangCSS {
		t.Fatalf("expected LangCSS for .scss, got %v", lang)
	}
}

func TestDetectLanguage_Less(t *testing.T) {
	lang := DetectLanguage("", "style.less")
	if lang != LangCSS {
		t.Fatalf("expected LangCSS for .less, got %v", lang)
	}
}

func TestDetectLanguage_Stylus(t *testing.T) {
	lang := DetectLanguage("", "style.styl")
	if lang != LangCSS {
		t.Fatalf("expected LangCSS for .styl, got %v", lang)
	}
}

func TestSkeletonize_XML(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("<root>\n")
	for i := 0; i < 25; i++ {
		b.WriteString("\t<item>\n\t\t<name>value</name>\n\t</item>\n")
	}
	b.WriteString("</root>\n")
	source := b.String()
	result := sk.Skeletonize(source, "data.xml")

	if result == source {
		t.Fatal("expected large XML to be collapsed")
	}
	if !strings.Contains(result, "<root>") {
		t.Fatal("expected root tag preserved")
	}
	if !strings.Contains(result, "lines omitted") {
		t.Fatal("expected omission message")
	}
}

func TestSkeletonize_LargeJavaClass(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("public class LargeClass {\n")
	b.WriteString("\tprivate int counter;\n\n")
	b.WriteString("\tpublic void methodOne() {\n")
	for i := 0; i < 10; i++ {
		b.WriteString("\t\tSystem.out.println(\"line\");\n")
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tpublic String methodTwo() {\n")
	for i := 0; i < 8; i++ {
		b.WriteString("\t\treturn \"data\";\n")
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	source := b.String()
	result := sk.Skeletonize(source, "LargeClass.java")

	if result == source {
		t.Fatal("expected large Java class to be collapsed")
	}
	if !strings.Contains(result, "public class LargeClass") {
		t.Fatal("expected class declaration preserved")
	}
}

func TestDetectLanguage_RubyContent(t *testing.T) {
	lang := DetectLanguage(`class Foo
	def bar
		puts "hello"
	end
end`, "test.rb")
	if lang != LangRuby {
		t.Fatalf("expected LangRuby, got %v", lang)
	}
}

func TestDetectLanguage_CSharpContent(t *testing.T) {
	lang := DetectLanguage(`using System;
namespace App {
	class Program {
		static void Main() {}
	}
}`, "Program.cs")
	if lang != LangCSharp {
		t.Fatalf("expected LangCSharp, got %v", lang)
	}
}

func TestDetectLanguage_RustContent(t *testing.T) {
	lang := DetectLanguage(`fn main() { println!("hello"); }`, "main.rs")
	if lang != LangRust {
		t.Fatalf("expected LangRust, got %v", lang)
	}
}

func TestDetectLanguage_KotlinContent(t *testing.T) {
	lang := DetectLanguage(`fun main() { println("hello") }`, "main.kt")
	if lang != LangKotlin {
		t.Fatalf("expected LangKotlin, got %v", lang)
	}
}

func TestDetectLanguage_CppHeader(t *testing.T) {
	lang := DetectLanguage("", "header.hpp")
	if lang != LangCPP {
		t.Fatalf("expected LangCPP for .hpp, got %v", lang)
	}
}

func TestDetectLanguage_Swift(t *testing.T) {
	lang := DetectLanguage("", "App.swift")
	if lang != LangSwift {
		t.Fatalf("expected LangSwift, got %v", lang)
	}
}

func TestDetectLanguage_Dart(t *testing.T) {
	lang := DetectLanguage("", "app.dart")
	if lang != LangDart {
		t.Fatalf("expected LangDart, got %v", lang)
	}
}

func TestDetectLanguage_Scala(t *testing.T) {
	lang := DetectLanguage("", "App.scala")
	if lang != LangScala {
		t.Fatalf("expected LangScala, got %v", lang)
	}
}

func TestDetectLanguage_ElixirContent(t *testing.T) {
	lang := DetectLanguage(`defmodule App do
	fn name -> name end
end`, "app.ex")
	if lang != LangElixir {
		t.Fatalf("expected LangElixir, got %v", lang)
	}
}

func TestDetectLanguage_ShebangLua(t *testing.T) {
	lang := DetectLanguage("#!/usr/bin/env lua\nprint('hello')", "")
	if lang != LangLua {
		t.Fatalf("expected LangLua from shebang, got %v", lang)
	}
}

func TestDetectLanguage_HaskellContent(t *testing.T) {
	lang := DetectLanguage(`module Main where
	data Foo = Bar
	main :: IO ()
	main = putStrLn "hello"`, "Main.hs")
	if lang != LangHaskell {
		t.Fatalf("expected LangHaskell, got %v", lang)
	}
}

func TestDetectLanguage_Julia(t *testing.T) {
	lang := DetectLanguage("", "script.jl")
	if lang != LangJulia {
		t.Fatalf("expected LangJulia, got %v", lang)
	}
}

func TestSkeletonize_LargeMarkdown(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	for i := 0; i < 25; i++ {
		b.WriteString("Some paragraph text that goes on for a while.\n")
	}
	source := b.String()
	result := sk.Skeletonize(source, "doc.md")

	if result == source {
		t.Fatal("expected large markdown to be collapsed")
	}
	if !strings.Contains(result, "lines omitted") {
		t.Fatal("expected omission message")
	}
}

func TestSkeletonize_LargeYAML(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("root:\n")
	b.WriteString("  key: value\n")
	for i := 0; i < 20; i++ {
		b.WriteString("  item:\n")
		b.WriteString("    name: val\n")
		b.WriteString("    desc: something\n")
	}
	source := b.String()
	result := sk.Skeletonize(source, "config.yaml")

	if result == source {
		t.Fatal("expected large YAML to be collapsed")
	}
	if !strings.Contains(result, "root:") {
		t.Fatal("expected root key preserved")
	}
}

func TestSkeletonize_LargeSQL(t *testing.T) {
	sk := NewCodeSkeletonizer()
	var b strings.Builder
	b.WriteString("CREATE TABLE users (\n")
	b.WriteString("\tid INT PRIMARY KEY,\n")
	for i := 0; i < 20; i++ {
		b.WriteString("\tcolumn_" + fmt.Sprint(i) + " VARCHAR(255),\n")
	}
	b.WriteString(");\n")
	source := b.String()
	result := sk.Skeletonize(source, "schema.sql")

	if !strings.Contains(result, "CREATE TABLE") {
		t.Fatal("expected CREATE TABLE preserved")
	}
}

func TestSkeletonize_EmptyLines(t *testing.T) {
	sk := NewCodeSkeletonizer()
	source := "package main\n\nfunc foo() {\n\tx := 1\n\t\n\t\n\ty := 2\n\tz := 3\n}"
	result := sk.Skeletonize(source, "")
	if len(result) == 0 {
		t.Fatal("unexpected empty result")
	}
}

func TestDetectLanguage_VueContent(t *testing.T) {
	lang := DetectLanguage(`<template><div>hi</div></template><script>export default {}</script>`, "App.vue")
	if lang != LangVue {
		t.Fatalf("expected LangVue, got %v", lang)
	}
}

func TestDetectLanguage_Svelte(t *testing.T) {
	lang := DetectLanguage("", "App.svelte")
	if lang != LangSvelte {
		t.Fatalf("expected LangSvelte, got %v", lang)
	}
}
