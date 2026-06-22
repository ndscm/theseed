package scalpel

import (
	"regexp"
	"testing"
)

func TestNeedsQuoting(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		if needsQuoting("plain/ascii.txt") {
			t.Error("plain ASCII path should not be quoted")
		}
	})

	t.Run("dash_dot", func(t *testing.T) {
		if needsQuoting("with-dash_and.dot") {
			t.Error("dash/dot path should not be quoted")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if needsQuoting("") {
			t.Error("empty path should not be quoted")
		}
	})

	t.Run("space", func(t *testing.T) {
		// Space is 0x20, not below it, so it is not quoted.
		if needsQuoting("has space.txt") {
			t.Error("space path should not be quoted")
		}
	})

	t.Run("tab", func(t *testing.T) {
		if !needsQuoting("tab\there") { // 0x09 control char
			t.Error("tab path should be quoted")
		}
	})

	t.Run("newline", func(t *testing.T) {
		if !needsQuoting("newline\n") { // 0x0a control char
			t.Error("newline path should be quoted")
		}
	})

	t.Run("quote", func(t *testing.T) {
		if !needsQuoting(`quote"inside`) { // 0x22
			t.Error("quote path should be quoted")
		}
	})

	t.Run("backslash", func(t *testing.T) {
		if !needsQuoting(`back\slash`) { // 0x5c
			t.Error("backslash path should be quoted")
		}
	})

	t.Run("del", func(t *testing.T) {
		if !needsQuoting("del\x7f") { // 0x7f
			t.Error("DEL path should be quoted")
		}
	})

	t.Run("non_ascii", func(t *testing.T) {
		if !needsQuoting("café.txt") { // high-bit (non-ASCII)
			t.Error("non-ASCII path should be quoted")
		}
	})
}

func TestStripPath(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		got, err := stripPath("a/", "a/foo.txt")
		if err != nil {
			t.Fatal(err)
		}
		if got != "foo.txt" {
			t.Fatalf("got %q, want foo.txt", got)
		}
	})

	t.Run("quoted", func(t *testing.T) {
		got, err := stripPath("b/", `"b/some path.txt"`)
		if err != nil {
			t.Fatal(err)
		}
		if got != "some path.txt" {
			t.Fatalf("got %q, want 'some path.txt'", got)
		}
	})

	t.Run("wrong_prefix", func(t *testing.T) {
		_, err := stripPath("a/", "b/foo.txt")
		if err == nil {
			t.Fatal("expected error for mismatched prefix")
		}
	})
}

func TestFormatPath(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		got := formatPath("a/", "foo.txt")
		if got != "a/foo.txt" {
			t.Fatalf("got %q, want a/foo.txt", got)
		}
	})

	t.Run("needs_quoting", func(t *testing.T) {
		got := formatPath("b/", "café.txt")
		if got != `"b/café.txt"` {
			t.Fatalf("got %q, want quoted", got)
		}
	})

	t.Run("empty_prefix", func(t *testing.T) {
		got := formatPath("", "rename target.txt")
		if got != "rename target.txt" {
			t.Fatalf("got %q", got)
		}
	})
}

func TestParseDiffLine(t *testing.T) {
	t.Run("plain", func(t *testing.T) {
		a, b, err := parseDiffLine("diff --git a/foo.txt b/foo.txt")
		if err != nil {
			t.Fatal(err)
		}
		if a != "foo.txt" || b != "foo.txt" {
			t.Fatalf("got a=%q b=%q", a, b)
		}
	})

	t.Run("both_quoted", func(t *testing.T) {
		a, b, err := parseDiffLine(`diff --git "a/old name.txt" "b/new name.txt"`)
		if err != nil {
			t.Fatal(err)
		}
		if a != "old name.txt" || b != "new name.txt" {
			t.Fatalf("got a=%q b=%q", a, b)
		}
	})

	t.Run("a_quoted_only", func(t *testing.T) {
		a, b, err := parseDiffLine(`diff --git "a/old name.txt" b/new.txt`)
		if err != nil {
			t.Fatal(err)
		}
		if a != "old name.txt" || b != "new.txt" {
			t.Fatalf("got a=%q b=%q", a, b)
		}
	})

	t.Run("no_header", func(t *testing.T) {
		_, _, err := parseDiffLine("index 1234..5678 100644")
		if err == nil {
			t.Fatal("expected error for missing header")
		}
	})

	t.Run("unterminated_quote", func(t *testing.T) {
		_, _, err := parseDiffLine(`diff --git "a/foo.txt b/foo.txt`)
		if err == nil {
			t.Fatal("expected error for unterminated quote")
		}
	})

	t.Run("no_space", func(t *testing.T) {
		_, _, err := parseDiffLine("diff --git a/foo.txt")
		if err == nil {
			t.Fatal("expected error for missing space separator")
		}
	})
}

// A canonical patch as ParseFormatPatch -> Render would emit it: header, blank,
// message ending in "---", blank, then diff blocks. Used by round-trip and
// editor tests below.
const samplePatch = `From 1234567890abcdef Mon Sep 17 00:00:00 2001
From: Jane Doe <jane@example.com>
Subject: [PATCH] edit foo, add bar

Commit body line one.

---

diff --git a/foo.txt b/foo.txt
index 1111111..2222222 100644
--- a/foo.txt
+++ b/foo.txt
@@ -1 +1 @@
-old
+new
diff --git a/bar.txt b/bar.txt
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/bar.txt
@@ -0,0 +1 @@
+hello
`

func TestFormatPatchPickDiff(t *testing.T) {
	t.Run("keeps_matching", func(t *testing.T) {
		p, err := ParseFormatPatch("sample.patch", samplePatch)
		if err != nil {
			t.Fatal(err)
		}
		p.PickDiff(
			[]*regexp.Regexp{regexp.MustCompile(`foo\.txt`)},
			[]*regexp.Regexp{regexp.MustCompile(`.*`)},
		)
		if len(p.diffs) != 1 {
			t.Fatalf("got %d diffs, want 1", len(p.diffs))
		}
		if p.diffs[0].a != "foo.txt" {
			t.Fatalf("kept wrong diff: %q", p.diffs[0].a)
		}
	})
}

func TestFormatPatchDropDiff(t *testing.T) {
	t.Run("drops_matching", func(t *testing.T) {
		p, err := ParseFormatPatch("sample.patch", samplePatch)
		if err != nil {
			t.Fatal(err)
		}
		p.DropDiff(
			[]*regexp.Regexp{regexp.MustCompile(`foo\.txt`)},
			[]*regexp.Regexp{regexp.MustCompile(`.*`)},
		)
		if len(p.diffs) != 1 {
			t.Fatalf("got %d diffs, want 1", len(p.diffs))
		}
		if p.diffs[0].a != "bar.txt" {
			t.Fatalf("kept wrong diff: %q", p.diffs[0].a)
		}
	})
}

func TestFormatPatchMoveDiff(t *testing.T) {
	t.Run("rewrites_paths", func(t *testing.T) {
		p, err := ParseFormatPatch("sample.patch", samplePatch)
		if err != nil {
			t.Fatal(err)
		}
		p.MoveDiff(
			regexp.MustCompile(`^foo\.txt$`),
			regexp.MustCompile(`^foo\.txt$`),
			"src/foo.txt",
			"src/foo.txt",
		)
		d := p.diffs[0]
		if d.a != "src/foo.txt" || d.b != "src/foo.txt" {
			t.Fatalf("got a=%q b=%q", d.a, d.b)
		}
		if d.diffLine != "diff --git a/src/foo.txt b/src/foo.txt" {
			t.Fatalf("diffLine not rewritten: %q", d.diffLine)
		}
		if d.minusLine != "--- a/src/foo.txt" {
			t.Fatalf("minusLine not rewritten: %q", d.minusLine)
		}
		if d.plusLine != "+++ b/src/foo.txt" {
			t.Fatalf("plusLine not rewritten: %q", d.plusLine)
		}
		// The non-matching diff (bar.txt) is untouched.
		if p.diffs[1].a != "bar.txt" {
			t.Fatalf("bar.txt should be untouched, got %q", p.diffs[1].a)
		}
	})

	t.Run("rewrites_rename_mode_lines", func(t *testing.T) {
		text := `Subject: [PATCH] rename

---

diff --git a/old.txt b/new.txt
similarity index 100%
rename from old.txt
rename to new.txt
`
		p, err := ParseFormatPatch("rename.patch", text)
		if err != nil {
			t.Fatal(err)
		}
		p.MoveDiff(
			regexp.MustCompile(`^old\.txt$`),
			regexp.MustCompile(`^new\.txt$`),
			"dir/old.txt",
			"dir/new.txt",
		)
		d := p.diffs[0]
		foundFrom := false
		foundTo := false
		for _, ml := range d.modeLines {
			if ml == "rename from dir/old.txt" {
				foundFrom = true
			}
			if ml == "rename to dir/new.txt" {
				foundTo = true
			}
		}
		if !foundFrom || !foundTo {
			t.Fatalf("rename mode lines not rewritten: %v", d.modeLines)
		}
	})

	t.Run("no_match_no_change", func(t *testing.T) {
		p, err := ParseFormatPatch("sample.patch", samplePatch)
		if err != nil {
			t.Fatal(err)
		}
		before := p.Render()
		p.MoveDiff(
			regexp.MustCompile(`nonexistent`),
			regexp.MustCompile(`.*`),
			"x",
			"y",
		)
		if p.Render() != before {
			t.Fatal("non-matching MoveDiff should not change anything")
		}
	})
}

func TestParseFormatPatch(t *testing.T) {
	t.Run("round_trip", func(t *testing.T) {
		p, err := ParseFormatPatch("sample.patch", samplePatch)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.diffs) != 2 {
			t.Fatalf("got %d diffs, want 2", len(p.diffs))
		}
		if p.diffs[0].a != "foo.txt" || p.diffs[1].b != "bar.txt" {
			t.Fatalf("unexpected paths: %q, %q", p.diffs[0].a, p.diffs[1].b)
		}
		got := p.Render()
		if got != samplePatch {
			t.Fatalf("round-trip mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, samplePatch)
		}
	})

	t.Run("message_only", func(t *testing.T) {
		// An empty commit: header + message, no diffs.
		text := "From abc Mon Sep 17 00:00:00 2001\nSubject: [PATCH] empty\n\njust a message\n"
		p, err := ParseFormatPatch("empty.patch", text)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.diffs) != 0 {
			t.Fatalf("got %d diffs, want 0", len(p.diffs))
		}
		if len(p.messageLines) != 1 || p.messageLines[0] != "just a message" {
			t.Fatalf("unexpected message: %v", p.messageLines)
		}
	})

	t.Run("binary", func(t *testing.T) {
		text := `From abc Mon Sep 17 00:00:00 2001
Subject: [PATCH] add image

---

diff --git a/img.png b/img.png
new file mode 100644
index 0000000..1111111
GIT binary patch
literal 8
PcmZ1234567890abcdef

literal 0
HcmV?d00001
`
		p, err := ParseFormatPatch("bin.patch", text)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.diffs) != 1 {
			t.Fatalf("got %d diffs, want 1", len(p.diffs))
		}
		d := p.diffs[0]
		if !d.binary {
			t.Fatal("expected binary diff")
		}
		if d.minusLine != "" || d.plusLine != "" || len(d.hunkLines) != 0 {
			t.Fatal("binary diff should not populate text fields")
		}
		if p.Render() != text {
			t.Fatalf("binary round-trip mismatch:\n%s", p.Render())
		}
	})

	t.Run("path_mismatch", func(t *testing.T) {
		text := `Subject: [PATCH] x

---

diff --git a/foo.txt b/foo.txt
--- a/other.txt
+++ b/foo.txt
@@ -1 +1 @@
-a
+b
`
		_, err := ParseFormatPatch("bad.patch", text)
		if err == nil {
			t.Fatal("expected error for --- path mismatch")
		}
	})

	t.Run("bad_diff_line", func(t *testing.T) {
		text := `Subject: [PATCH] x

---

diff --git a/foo.txt
`
		_, err := ParseFormatPatch("bad.patch", text)
		if err == nil {
			t.Fatal("expected error for malformed diff --git line")
		}
	})
}
