package compress

import (
	"strings"
	"testing"
)

func TestDiffLines_Identical(t *testing.T) {
	a := "line1\nline2\nline3"
	b := "line1\nline2\nline3"
	diffs := DiffLines(a, b)

	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d", len(diffs))
	}
	for _, d := range diffs {
		if d.Type != DiffEqual {
			t.Fatalf("expected all equal, got %v", d.Type)
		}
	}
}

func TestDiffLines_Insert(t *testing.T) {
	a := "line1\nline3"
	b := "line1\nline2\nline3"
	diffs := DiffLines(a, b)

	found := false
	for _, d := range diffs {
		if d.Type == DiffInsert && d.Text == "line2" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected insert of 'line2'")
	}
}

func TestDiffLines_Delete(t *testing.T) {
	a := "line1\nline2\nline3"
	b := "line1\nline3"
	diffs := DiffLines(a, b)

	found := false
	for _, d := range diffs {
		if d.Type == DiffDelete && d.Text == "line2" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected delete of 'line2'")
	}
}

func TestDiffLines_Replace(t *testing.T) {
	a := "line1\nold\nline3"
	b := "line1\nnew\nline3"
	diffs := DiffLines(a, b)

	hasDelete := false
	hasInsert := false
	for _, d := range diffs {
		if d.Type == DiffDelete && d.Text == "old" {
			hasDelete = true
		}
		if d.Type == DiffInsert && d.Text == "new" {
			hasInsert = true
		}
	}
	if !hasDelete || !hasInsert {
		t.Fatal("expected delete of 'old' and insert of 'new'")
	}
}

func TestDiffLines_Empty(t *testing.T) {
	diffs := DiffLines("", "")
	if len(diffs) == 0 {
		return
	}
	for _, d := range diffs {
		if d.Type != DiffEqual || d.Text != "" {
			t.Fatalf("expected single equal empty diff, got type=%d text=%q", d.Type, d.Text)
		}
	}
}

func TestDiffLines_NewFile(t *testing.T) {
	diffs := DiffLines("", "new content")
	
	hasInsert := false
	for _, d := range diffs {
		if d.Type == DiffInsert {
			hasInsert = true
			break
		}
	}
	if !hasInsert {
		t.Fatal("expected inserts for new content")
	}
}

func TestDiffLines_DeletedFile(t *testing.T) {
	diffs := DiffLines("old content", "")
	
	hasDelete := false
	for _, d := range diffs {
		if d.Type == DiffDelete {
			hasDelete = true
			break
		}
	}
	if !hasDelete {
		t.Fatal("expected deletes for removed content")
	}
}

func TestComputeDiffHunks_Context(t *testing.T) {
	oldContent := "a\nb\nc\nd\ne\nf\ng"
	newContent := "a\nb\nX\nY\ne\nf\ng"
	
	result := ComputeDiffHunks(oldContent, newContent, 1)
	
	if !strings.HasPrefix(result, "```diff\n") {
		t.Fatal("expected diff format with ```diff marker")
	}
	if !strings.Contains(result, "-c") || !strings.Contains(result, "-d") {
		t.Fatal("expected deletions of 'c' and 'd'")
	}
	if !strings.Contains(result, "+X") || !strings.Contains(result, "+Y") {
		t.Fatal("expected insertions of 'X' and 'Y'")
	}
}

func TestComputeDiffHunks_Unchanged(t *testing.T) {
	result := ComputeDiffHunks("same", "same", 3)
	if !strings.Contains(result, "unchanged") {
		t.Fatalf("expected 'unchanged' for identical content, got: %s", result)
	}
}

func TestComputeDiffHunks_EmptyOld(t *testing.T) {
	result := ComputeDiffHunks("", "new\ncontent", 3)
	if !strings.Contains(result, "+new") {
		t.Fatalf("expected insertions in diff, got: %s", result)
	}
}

func TestCompactDiffSummary(t *testing.T) {
	oldContent := "a\nb\nc"
	newContent := "a\nX\nc"
	
	summary := CompactDiffSummary(oldContent, newContent)
	
	if !strings.Contains(summary, "+1 added") || !strings.Contains(summary, "-1 removed") {
		t.Fatalf("expected +/- counts in summary, got: %s", summary)
	}
}

func TestCompactDiffSummary_Unchanged(t *testing.T) {
	result := CompactDiffSummary("same", "same")
	if !strings.Contains(result, "diff") {
		t.Fatal("expected diff marker for unchanged content")
	}
}

func TestCompactDiffSummary_LargeChanges(t *testing.T) {
	var a, b strings.Builder
	for i := 0; i < 100; i++ {
		a.WriteString("line\n")
		b.WriteString("line\n")
	}
	b.WriteString("extra1\nextra2\nextra3\nextra4\n")

	summary := CompactDiffSummary(a.String(), b.String())
	if !strings.Contains(summary, "+") {
		t.Fatal("expected additions in summary for large change")
	}
}

func TestDiff_HashMap(t *testing.T) {
	a := "key=value1\nkey=value2\nkey=value3"
	b := "key=value1\nkey=value2\nkey=value2"
	diffs := DiffLines(a, b)

	changeCount := 0
	for _, d := range diffs {
		if d.Type != DiffEqual {
			changeCount++
		}
	}
	if changeCount == 0 {
		t.Fatal("expected at least one change")
	}
}
