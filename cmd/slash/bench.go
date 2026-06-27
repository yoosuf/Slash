package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yoosuf/Slash/internal/bench"
	"github.com/yoosuf/Slash/internal/store"
)

// cmdBench runs the benchmark suite.
func cmdBench(args []string) {
	fs := flag.NewFlagSet("bench", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: slash bench [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}

	output := fs.String("output", "", "output file for results (JSON)")
	saveCache := fs.Bool("keep-cache", false, "keep cache after benchmarks")

	if err := fs.Parse(args); err != nil {
		return
	}

	// Open cache for benchmarks
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		home, _ := os.UserHomeDir()
		xdgCache = filepath.Join(home, ".cache")
	}
	cacheDir := filepath.Join(xdgCache, "slash-bench-"+time.Now().Format("20060102150405"))

	cache, err := store.NewCCRCache(cacheDir, 24*time.Hour, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create cache: %v\n", err)
		os.Exit(1)
	}
	defer cache.Close()

	// Clean up cache directory if not saving
	defer func() {
		if !*saveCache {
			os.RemoveAll(cacheDir)
		}
	}()

	// Run benchmarks
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║             Slash Compression Benchmark Suite v1.0.0            ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║  Measuring token reduction, latency, and quality across        ║")
	fmt.Println("║  JSON, code, logs, and text compression.                       ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	suite := bench.NewBenchmarkSuite(cache)
	_, err = suite.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Benchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Summary")
	fmt.Println("=======")
	summary := suite.Summary()
	fmt.Printf("Total benchmarks run:        %v\n", summary["total_benchmarks"])
	fmt.Printf("Passed:                      %v/%v\n", summary["passed"], summary["total_benchmarks"])
	fmt.Printf("Pass rate:                   %v%%\n", summary["pass_rate"])
	fmt.Printf("Total data (original):       %.2f MB\n", summary["total_original_mb"])
	fmt.Printf("Total data (compressed):     %.2f MB\n", summary["total_compressed_mb"])
	fmt.Printf("Overall reduction:           %v%%\n", summary["overall_reduction"])
	fmt.Printf("Latency (p50):               %.2f ms\n", summary["latency_p50_ms"])
	fmt.Printf("Latency (p95):               %.2f ms\n", summary["latency_p95_ms"])
	fmt.Println()

	// Export to JSON if requested
	if *output != "" {
		data, _ := suite.Export()
		if err := os.WriteFile(*output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write results: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Results saved to: %s\n", *output)
	}

	if *saveCache {
		fmt.Printf("Cache saved to: %s\n", cacheDir)
	}

	fmt.Println()
	fmt.Println("✓ Benchmarks complete. Use --output <file> to save JSON results.")
}
