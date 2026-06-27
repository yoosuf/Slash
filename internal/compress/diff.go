package compress

import (
	"fmt"
	"strings"
)

type DiffType int

const (
	DiffEqual DiffType = iota
	DiffInsert
	DiffDelete
)

type DiffLine struct {
	Type      DiffType
	Text      string
	OldLineNo int
	NewLineNo int
}

type DiffHunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []DiffLine
}

func DiffLines(a, b string) []DiffLine {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	m, n := len(linesA), len(linesB)
	lcs := make([][]int, m+1)
	for i := range lcs {
		lcs[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if linesA[i-1] == linesB[j-1] {
				lcs[i][j] = lcs[i-1][j-1] + 1
			} else if lcs[i-1][j] > lcs[i][j-1] {
				lcs[i][j] = lcs[i-1][j]
			} else {
				lcs[i][j] = lcs[i][j-1]
			}
		}
	}

	var result []DiffLine
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && linesA[i-1] == linesB[j-1] {
			result = append(result, DiffLine{Type: DiffEqual, Text: linesA[i-1], OldLineNo: i, NewLineNo: j})
			i--
			j--
		} else if j > 0 && (i == 0 || lcs[i][j-1] >= lcs[i-1][j]) {
			result = append(result, DiffLine{Type: DiffInsert, Text: linesB[j-1], NewLineNo: j})
			j--
		} else {
			result = append(result, DiffLine{Type: DiffDelete, Text: linesA[i-1], OldLineNo: i})
			i--
		}
	}

	for left, right := 0, len(result)-1; left < right; left, right = left+1, right-1 {
		result[left], result[right] = result[right], result[left]
	}

	return result
}

func ComputeDiffHunks(oldContent, newContent string, contextLines int) string {
	diffs := DiffLines(oldContent, newContent)
	if len(diffs) == 0 {
		return "[diff: content unchanged]"
	}

	onlyEqual := true
	for _, d := range diffs {
		if d.Type != DiffEqual {
			onlyEqual = false
			break
		}
	}
	if onlyEqual {
		return "[diff: content unchanged]"
	}

	hunks := groupIntoHunks(diffs, contextLines)
	if len(hunks) == 0 {
		return "[diff: no significant changes]"
	}

	var b strings.Builder
	b.WriteString("```diff\n")
	for _, hunk := range hunks {
		b.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", hunk.OldStart, hunk.OldCount, hunk.NewStart, hunk.NewCount))
		for _, line := range hunk.Lines {
			switch line.Type {
			case DiffEqual:
				b.WriteString(" " + line.Text + "\n")
			case DiffInsert:
				b.WriteString("+" + line.Text + "\n")
			case DiffDelete:
				b.WriteString("-" + line.Text + "\n")
			}
		}
	}
	b.WriteString("```\n")
	return b.String()
}

func groupIntoHunks(diffs []DiffLine, context int) []DiffHunk {
	if len(diffs) == 0 {
		return nil
	}

	var hunks []DiffHunk
	var current DiffHunk
	inHunk := false
	contextAfter := 0

	for i, d := range diffs {
		if d.Type != DiffEqual {
			if !inHunk {
				start := i - context
				if start < 0 {
					start = 0
				}
				current = DiffHunk{
					OldStart: diffs[start].OldLineNo,
					NewStart: diffs[start].NewLineNo,
				}
				for k := start; k < i; k++ {
					current.Lines = append(current.Lines, diffs[k])
				}
				inHunk = true
			}
			current.Lines = append(current.Lines, d)
			current.OldCount = (d.OldLineNo - current.OldStart + 1)
			current.NewCount = (d.NewLineNo - current.NewStart + 1)
			contextAfter = 0
		} else if inHunk {
			current.Lines = append(current.Lines, d)
			if d.OldLineNo > 0 {
				current.OldCount = d.OldLineNo - current.OldStart + 1
			}
			if d.NewLineNo > 0 {
				current.NewCount = d.NewLineNo - current.NewStart + 1
			}
			contextAfter++

			if contextAfter >= context {
				end := i + 1
				nearEnd := end + context
				hasChangeBeforeEnd := false
				for k := i + 1; k < len(diffs) && k < nearEnd; k++ {
					if diffs[k].Type != DiffEqual {
						hasChangeBeforeEnd = true
						break
					}
				}
				if !hasChangeBeforeEnd {
					if contextAfter > context {
						trim := contextAfter - context
						current.Lines = current.Lines[:len(current.Lines)-trim]
					}
					hunks = append(hunks, current)
					current = DiffHunk{}
					inHunk = false
					contextAfter = 0
				}
			}
		}
	}

	if inHunk {
		hunks = append(hunks, current)
	}

	return hunks
}

func CompactDiffSummary(oldContent, newContent string) string {
	diffs := DiffLines(oldContent, newContent)

	var added, deleted, changed int
	addedLines := make(map[string]int)
	deletedLines := make(map[string]int)

	for _, d := range diffs {
		switch d.Type {
		case DiffInsert:
			added++
			addedLines[d.Text]++
		case DiffDelete:
			deleted++
			deletedLines[d.Text]++
		case DiffEqual:
			changed = max(changed-1, 0)
		}
	}

	var parts []string
	if added > 0 {
		parts = append(parts, fmt.Sprintf("+%d added", added))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("-%d removed", deleted))
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("[diff: %s]", strings.Join(parts, ", ")))

	if len(addedLines) <= 3 && len(addedLines) > 0 {
		for line := range addedLines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 80 {
				trimmed = trimmed[:80] + "..."
			}
			b.WriteString(fmt.Sprintf("\n  + %s", trimmed))
		}
	}
	if len(deletedLines) <= 3 && len(deletedLines) > 0 {
		for line := range deletedLines {
			trimmed := strings.TrimSpace(line)
			if len(trimmed) > 80 {
				trimmed = trimmed[:80] + "..."
			}
			b.WriteString(fmt.Sprintf("\n  - %s", trimmed))
		}
	}
	if len(addedLines) > 3 || len(deletedLines) > 3 {
		b.WriteString("\n  [retrieve handle for full diff]")
	}

	return b.String()
}
