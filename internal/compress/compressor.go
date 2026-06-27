package compress

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yoosuf/Slash/internal/store"
)

type Compressor struct {
	cache  *store.CCRCache
	codeSk *CodeSkeletonizer
}

func NewCompressor(cache *store.CCRCache) *Compressor {
	return &Compressor{
		cache:  cache,
		codeSk: NewCodeSkeletonizer(),
	}
}

func (c *Compressor) Compress(output interface{}, filename string) (interface{}, string) {
	contentType := detectType(output)

	switch contentType {
	case "json":
		return c.compressJSON(output)
	case "code":
		return c.compressCode(output, filename)
	case "logs":
		return c.compressLogs(output)
	case "text":
		return c.compressText(output)
	default:
		return output, ""
	}
}

func (c *Compressor) CompressReRead(output string, previousContent string) (interface{}, string) {
	if previousContent == output {
		handle, err := c.cache.Insert([]byte(output), "read", "", "re_read_unchanged", int64(len(output)))
		if err != nil {
			return output, ""
		}
		result := fmt.Sprintf(`[content unchanged since last read — %d bytes, retrieve(%s) for full content]`, len(output), handle)
		return result, "[slash: content unchanged, retrieve handle provided]"
	}

	diffs := DiffLines(previousContent, output)

	var changedLines int
	for _, d := range diffs {
		if d.Type != DiffEqual {
			changedLines++
		}
	}

	totalChanged := changedLines
	changeRatio := float64(totalChanged) / float64(len(diffs))

	if changeRatio > 0.3 {
		d := ComputeDiffHunks(previousContent, output, 3)
		handle, err := c.cache.Insert([]byte(output), "read", "", "re_read_diff", int64(len(d)))
		if err != nil {
			return output, ""
		}
		result := fmt.Sprintf("%s\n[retrieve(%s) for full content]", d, handle)
		savings := 100 * (len(output) - len(d)) / len(output)
		meta := fmt.Sprintf("[slash: re-read diff, %d%% reduction, %d lines changed]", savings, totalChanged)
		return result, meta
	}

	summary := CompactDiffSummary(previousContent, output)
	handle, err := c.cache.Insert([]byte(output), "read", "", "re_read_changed", int64(len(summary)))
	if err != nil {
		return output, ""
	}
	result := fmt.Sprintf("%s\n[retrieve(%s) for full content]", summary, handle)
	savings := 100 * (len(output) - len(summary)) / len(output)
	meta := fmt.Sprintf("[slash: re-read compact diff, %d%% reduction, %d lines changed]", savings, totalChanged)
	return result, meta
}

func detectType(output interface{}) string {
	s, ok := output.(string)
	if !ok {
		return "unknown"
	}

	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[") {
		if isJSON(s) {
			return "json"
		}
	}

	if isCode(s) {
		return "code"
	}

	if isLogs(s) {
		return "logs"
	}

	return "text"
}

func isJSON(s string) bool {
	var any interface{}
	return json.Unmarshal([]byte(s), &any) == nil
}

func isCode(s string) bool {
	codePatterns := []string{
		"func ", "def ", "class ", "import ", "package ",
		"const ", "let ", "var ", "function ",
		"public ", "private ", "protected ",
		"interface ", "type ", "struct ", "enum ",
		"impl ", "trait ", "fn ",
		"require(",
	}

	for _, pattern := range codePatterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}

	return strings.Contains(s, "()") || strings.Contains(s, "{}") || strings.Contains(s, "[]")
}

func isLogs(s string) bool {
	logPatterns := []string{
		"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]",
		"ERROR:", "WARNING:", "INFO:", "DEBUG:",
		"ERROR |", "WARN  |",
		"TRACE:", "FATAL:",
		"2024-", "2025-", "2026-",
		"T00:", "T01:", "T02:", "T03:", "T04:", "T05:",
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 3 {
		return false
	}

	matchCount := 0
	for _, line := range lines {
		for _, pattern := range logPatterns {
			if strings.Contains(line, pattern) {
				matchCount++
				break
			}
		}
	}

	return float64(matchCount)/float64(len(lines)) > 0.2
}

func (c *Compressor) compressJSON(output interface{}) (interface{}, string) {
	s, ok := output.(string)
	if !ok {
		return output, ""
	}

	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return output, ""
	}

	skeleton := skeletonizeJSON(data)
	skeletonBytes, _ := json.Marshal(skeleton)
	skeletonStr := string(skeletonBytes)

	if len(skeletonStr) >= len(s)*9/10 {
		return output, ""
	}

	originalBytes := []byte(s)
	handle, err := c.cache.Insert(originalBytes, "json", "", "json_skeleton", int64(len(skeletonStr)))
	if err != nil {
		return output, ""
	}

	savings := 0
	if len(s) > 0 {
		savings = 100 * (len(s) - len(skeletonStr)) / len(s)
	}

	result := fmt.Sprintf(`%s
[retrieve(%s) for full content]`, skeletonStr, handle)

	meta := fmt.Sprintf("[slash: JSON skeleton, %d%% reduction]", savings)
	return result, meta
}

func skeletonizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		skeleton := make(map[string]interface{})
		for k, val := range v {
			skeleton[k] = skeletonizeJSON(val)
		}
		return skeleton
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		if len(v) <= 3 {
			items := make([]interface{}, len(v))
			for i, item := range v {
				items[i] = skeletonizeJSON(item)
			}
			return items
		}
		return fmt.Sprintf("[array of %d items]", len(v))
	case string:
		if len(v) > 30 {
			if len(v) > 100 {
				return fmt.Sprintf("<string: %d chars>", len(v))
			}
			return v[:20] + "..."
		}
		return v
	case float64:
		return "<number>"
	case bool:
		return "<boolean>"
	case nil:
		return nil
	default:
		return fmt.Sprintf("<%T>", v)
	}
}

func (c *Compressor) compressCode(output interface{}, filename string) (interface{}, string) {
	s, ok := output.(string)
	if !ok {
		return output, ""
	}

	if len(s) == 0 {
		return output, ""
	}

	skeleton := c.codeSk.Skeletonize(s, filename)

	if skeleton == s || len(skeleton) >= len(s)*9/10 {
		return output, ""
	}

	originalBytes := []byte(s)
	handle, err := c.cache.Insert(originalBytes, "code", "", "code_skeleton", int64(len(skeleton)))
	if err != nil {
		return output, ""
	}

	result := fmt.Sprintf(`%s
[retrieve(%s) for full code]`, skeleton, handle)

	origTokens := len(s) / 4
	compTokens := len(skeleton) / 4
	savings := 0
	if origTokens > 0 {
		savings = 100 * (origTokens - compTokens) / origTokens
	}

	meta := fmt.Sprintf("[slash: code skeleton, %d%% reduction]", savings)
	return result, meta
}

func (c *Compressor) compressLogs(output interface{}) (interface{}, string) {
	s, ok := output.(string)
	if !ok {
		return output, ""
	}

	lines := strings.Split(s, "\n")
	if len(lines) < 5 {
		return output, ""
	}

	seen := make(map[string]int)
	var compressed []string
	var order []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; !exists {
			order = append(order, trimmed)
		}
		seen[trimmed]++
	}

	for _, line := range order {
		count := seen[line]
		if count == 1 {
			compressed = append(compressed, line)
		} else if count <= 3 {
			for i := 0; i < count; i++ {
				compressed = append(compressed, line)
			}
		} else {
			compressed = append(compressed, line)
			compressed = append(compressed, fmt.Sprintf("  [... repeated %d more times]", count-1))
		}
	}

	compressedStr := strings.Join(compressed, "\n")

	if len(compressedStr) >= len(s)*9/10 {
		return output, ""
	}

	originalBytes := []byte(s)
	handle, err := c.cache.Insert(originalBytes, "logs", "", "log_dedup", int64(len(compressedStr)))
	if err != nil {
		return output, ""
	}

	result := fmt.Sprintf(`%s
[retrieve(%s) for full logs]`, compressedStr, handle)

	savings := 0
	if len(s) > 0 {
		savings = 100 * (len(s) - len(compressedStr)) / len(s)
	}
	meta := fmt.Sprintf("[slash: log dedup, %d%% reduction]", savings)
	return result, meta
}

func (c *Compressor) compressText(output interface{}) (interface{}, string) {
	s, ok := output.(string)
	if !ok {
		return output, ""
	}

	maxChars := 5000
	if len(s) <= maxChars {
		return output, ""
	}

	headLen := maxChars * 2 / 3
	tailLen := maxChars / 3

	head := s[:headLen]
	tail := s[len(s)-tailLen:]

	truncated := head + "\n...\n[truncated: " + fmt.Sprintf("%d bytes omitted", len(s)-maxChars) + "]\n...\n" + tail
	originalBytes := []byte(s)
	handle, err := c.cache.Insert(originalBytes, "text", "", "text_truncate", int64(len(truncated)))
	if err != nil {
		return output, ""
	}

	result := fmt.Sprintf(`%s
[retrieve(%s) for full content]`, truncated, handle)

	savings := 0
	if len(s) > 0 {
		savings = 100 * (len(s) - len(truncated)) / len(s)
	}
	meta := fmt.Sprintf("[slash: text truncated, %d%% reduction]", savings)
	return result, meta
}
