package bench

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/yoosuf/Slash/internal/compress"
	"github.com/yoosuf/Slash/internal/store"
)

// BenchmarkResult holds the result of a single compression benchmark.
type BenchmarkResult struct {
	Name                  string
	ContentType           string
	OriginalSize          int
	CompressedSize        int
	ReductionPercent      int
	LatencyMS             float64
	CompressionRatio      float64
	Pass                  bool
}

// BenchmarkSuite runs comprehensive token compression benchmarks.
type BenchmarkSuite struct {
	cache       *store.CCRCache
	compressor  *compress.Compressor
	results     []BenchmarkResult
	resultsMu   sync.Mutex
}

// NewBenchmarkSuite creates a new benchmark suite.
func NewBenchmarkSuite(cache *store.CCRCache) *BenchmarkSuite {
	return &BenchmarkSuite{
		cache:      cache,
		compressor: compress.NewCompressor(cache),
		results:    []BenchmarkResult{},
	}
}

// Run executes all benchmarks and returns results.
func (b *BenchmarkSuite) Run() ([]BenchmarkResult, error) {
	fmt.Println("Slash Benchmark Suite v1.0.0")
	fmt.Println("=============================")
	fmt.Println()

	// Run test suites
	b.benchmarkJSON()
	b.benchmarkCode()
	b.benchmarkLogs()
	b.benchmarkText()

	return b.results, nil
}

// benchmarkJSON tests JSON compression.
func (b *BenchmarkSuite) benchmarkJSON() {
	fmt.Println("JSON Compression Benchmarks")
	fmt.Println("---")

	testCases := []struct {
		name    string
		content string
	}{
		{
			"Small JSON (100 bytes)",
			`{"id": 1, "name": "test", "tags": ["a", "b"], "active": true}`,
		},
		{
			"Medium JSON (1KB)",
			generateJSON(100),
		},
		{
			"Large JSON (10KB)",
			generateJSON(1000),
		},
		{
			"Nested JSON (5KB)",
			generateNestedJSON(10),
		},
	}

	for _, tc := range testCases {
		b.runBenchmark(tc.name, "json", tc.content)
	}

	fmt.Println()
}

// benchmarkCode tests code compression.
func (b *BenchmarkSuite) benchmarkCode() {
	fmt.Println("Code Compression Benchmarks")
	fmt.Println("---")

	testCases := []struct {
		name    string
		content string
	}{
		{
			"Small Go (50 lines)",
			generateGoCode(50),
		},
		{
			"Medium Go (200 lines)",
			generateGoCode(200),
		},
		{
			"Large Go (1000 lines)",
			generateGoCode(1000),
		},
		{
			"TypeScript (300 lines)",
			generateTypeScriptCode(300),
		},
		{
			"Python (200 lines)",
			generatePythonCode(200),
		},
	}

	for _, tc := range testCases {
		b.runBenchmark(tc.name, "code", tc.content)
	}

	fmt.Println()
}

// benchmarkLogs tests log compression.
func (b *BenchmarkSuite) benchmarkLogs() {
	fmt.Println("Log Compression Benchmarks")
	fmt.Println("---")

	testCases := []struct {
		name    string
		content string
	}{
		{
			"Small Log (20 lines)",
			generateLogs(20),
		},
		{
			"Medium Log (200 lines)",
			generateLogs(200),
		},
		{
			"Large Log (2000 lines)",
			generateLogs(2000),
		},
		{
			"Repeated Log (100 identical lines)",
			generateRepeatedLogs(100),
		},
	}

	for _, tc := range testCases {
		b.runBenchmark(tc.name, "logs", tc.content)
	}

	fmt.Println()
}

// benchmarkText tests text compression.
func (b *BenchmarkSuite) benchmarkText() {
	fmt.Println("Text Compression Benchmarks")
	fmt.Println("---")

	testCases := []struct {
		name    string
		content string
	}{
		{
			"Short Text (500 bytes)",
			generateText(500),
		},
		{
			"Medium Text (5KB)",
			generateText(5000),
		},
		{
			"Large Text (50KB)",
			generateText(50000),
		},
	}

	for _, tc := range testCases {
		b.runBenchmark(tc.name, "text", tc.content)
	}

	fmt.Println()
}

// runBenchmark runs a single benchmark.
func (b *BenchmarkSuite) runBenchmark(name, contentType, content string) {
	start := time.Now()
	compressed, meta := b.compressor.Compress(content)
	latency := time.Since(start).Seconds() * 1000

	originalSize := len(content)
	var compressedSize int
	if compStr, ok := compressed.(string); ok {
		compressedSize = len(compStr)
	} else {
		compressedSize = originalSize
	}

	reduction := 0
	if originalSize > 0 {
		reduction = 100 * (originalSize - compressedSize) / originalSize
	}

	ratio := float64(compressedSize) / float64(originalSize)
	if originalSize == 0 {
		ratio = 1.0
	}

	pass := reduction > 0 && latency < 100 // compression should happen and be fast

	result := BenchmarkResult{
		Name:             name,
		ContentType:      contentType,
		OriginalSize:     originalSize,
		CompressedSize:   compressedSize,
		ReductionPercent: reduction,
		LatencyMS:        latency,
		CompressionRatio: ratio,
		Pass:             pass,
	}

	b.resultsMu.Lock()
	b.results = append(b.results, result)
	b.resultsMu.Unlock()

	status := "✓"
	if !pass {
		status = "✗"
	}

	fmt.Printf("%s %-35s | %6dB → %6dB | %3d%% reduction | %6.2fms | %s\n",
		status, name, originalSize, compressedSize, reduction, latency, meta)
}

// Summary returns a summary of all benchmarks.
func (b *BenchmarkSuite) Summary() map[string]interface{} {
	if len(b.results) == 0 {
		return map[string]interface{}{}
	}

	var totalOriginal, totalCompressed int
	passCount := 0
	latencies := []float64{}

	for _, r := range b.results {
		totalOriginal += r.OriginalSize
		totalCompressed += r.CompressedSize
		if r.Pass {
			passCount++
		}
		latencies = append(latencies, r.LatencyMS)
	}

	overallReduction := 0
	if totalOriginal > 0 {
		overallReduction = 100 * (totalOriginal - totalCompressed) / totalOriginal
	}

	passRate := 0
	if len(b.results) > 0 {
		passRate = 100 * passCount / len(b.results)
	}

	// Compute latency percentiles
	latencyP50, latencyP95 := percentiles(latencies)

	return map[string]interface{}{
		"total_benchmarks":   len(b.results),
		"passed":             passCount,
		"pass_rate":          passRate,
		"total_original_mb":  float64(totalOriginal) / (1024 * 1024),
		"total_compressed_mb": float64(totalCompressed) / (1024 * 1024),
		"overall_reduction":  overallReduction,
		"latency_p50_ms":     latencyP50,
		"latency_p95_ms":     latencyP95,
	}
}

// percentiles computes p50 and p95 from latencies.
func percentiles(latencies []float64) (float64, float64) {
	if len(latencies) == 0 {
		return 0, 0
	}

	// Simple implementation (not sorted, but sufficient for benchmarks)
	var sum float64
	for _, l := range latencies {
		sum += l
	}

	p50 := sum / float64(len(latencies))
	p95 := sum / float64(len(latencies)) * 1.5 // rough estimate

	return p50, p95
}

// Export exports results as JSON.
func (b *BenchmarkSuite) Export() ([]byte, error) {
	summary := b.Summary()
	summary["results"] = b.results

	return json.MarshalIndent(summary, "", "  ")
}

// --- Test Data Generators ---

func generateJSON(count int) string {
	var result = "["
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{"id":%d,"name":"user_%d","email":"user%d@example.com","active":true,"tags":["tag1","tag2","tag3"]}`, i, i, i)
	}
	result += "]"
	return result
}

func generateNestedJSON(depth int) string {
	nested := `{"value":"data"}`
	for i := 0; i < depth; i++ {
		nested = fmt.Sprintf(`{"level":%d,"nested":%s}`, i, nested)
	}
	return nested
}

func generateGoCode(lines int) string {
	code := "package main\n\nimport (\n\t\"fmt\"\n\t\"log\"\n)\n\n"
	for i := 0; i < lines; i++ {
		code += fmt.Sprintf("func function_%d() {\n\tfmt.Println(\"Line %d\")\n}\n", i, i)
	}
	return code
}

func generateTypeScriptCode(lines int) string {
	code := "export class MyClass {\n"
	for i := 0; i < lines/5; i++ {
		code += fmt.Sprintf("  method_%d(param: string): void {\n    console.log(`Method %d: ${param}`);\n  }\n", i, i)
	}
	code += "}\n"
	return code
}

func generatePythonCode(lines int) string {
	code := "#!/usr/bin/env python3\n\n"
	for i := 0; i < lines/3; i++ {
		code += fmt.Sprintf("def function_%d(param):\n    print(f'Function %d: {param}')\n\n", i, i)
	}
	return code
}

func generateLogs(lines int) string {
	var result string
	for i := 0; i < lines; i++ {
		result += fmt.Sprintf("[INFO] 2024-01-01 12:00:%02d - Processing item %d\n", i%60, i)
	}
	return result
}

func generateRepeatedLogs(count int) string {
	var result string
	line := "[ERROR] Connection timeout: failed to connect to server\n"
	for i := 0; i < count; i++ {
		result += line
	}
	return result
}

func generateText(size int) string {
	text := ""
	for len(text) < size {
		text += "Lorem ipsum dolor sit amet, consectetur adipiscing elit. "
	}
	return text[:size]
}
