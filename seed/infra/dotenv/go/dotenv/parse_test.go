package dotenv

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		key, value, err := ParseLine("FOO=bar")
		if err != nil {
			t.Fatal(err)
		}
		if key != "FOO" || value != "bar" {
			t.Fatalf("got %q=%q, want FOO=bar", key, value)
		}
	})

	t.Run("empty", func(t *testing.T) {
		key, value, err := ParseLine("")
		if err != nil {
			t.Fatal(err)
		}
		if key != "" || value != "" {
			t.Fatalf("expected empty key/value for blank line, got %q=%q", key, value)
		}
	})

	t.Run("comment", func(t *testing.T) {
		key, value, err := ParseLine("# this is a comment")
		if err != nil {
			t.Fatal(err)
		}
		if key != "" || value != "" {
			t.Fatalf("expected empty key/value for comment, got %q=%q", key, value)
		}
	})

	t.Run("export_prefix", func(t *testing.T) {
		key, value, err := ParseLine("export DB_HOST=localhost")
		if err != nil {
			t.Fatal(err)
		}
		if key != "DB_HOST" || value != "localhost" {
			t.Fatalf("got %q=%q, want DB_HOST=localhost", key, value)
		}
	})

	t.Run("single_quoted", func(t *testing.T) {
		key, value, err := ParseLine("KEY='hello world'")
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "hello world" {
			t.Fatalf("got %q=%q, want KEY='hello world'", key, value)
		}
	})

	t.Run("single_quoted_no_escape", func(t *testing.T) {
		key, value, err := ParseLine(`KEY='hello\nworld'`)
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != `hello\nworld` {
			t.Fatalf("got %q=%q, want literal backslash-n", key, value)
		}
	})

	t.Run("double_quoted", func(t *testing.T) {
		key, value, err := ParseLine(`KEY="hello world"`)
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "hello world" {
			t.Fatalf("got %q=%q", key, value)
		}
	})

	t.Run("double_quoted_escapes", func(t *testing.T) {
		key, value, err := ParseLine(`KEY="line1\nline2\ttab\\"`)
		if err != nil {
			t.Fatal(err)
		}
		want := "line1\nline2\ttab\\"
		if key != "KEY" || value != want {
			t.Fatalf("got %q=%q, want %q", key, value, want)
		}
	})

	t.Run("inline_comment", func(t *testing.T) {
		key, value, err := ParseLine("KEY=value # comment")
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "value" {
			t.Fatalf("got %q=%q, want KEY=value", key, value)
		}
	})

	t.Run("no_equals", func(t *testing.T) {
		_, _, err := ParseLine("INVALID_LINE")
		if err == nil {
			t.Fatal("expected error for line without =")
		}
	})

	t.Run("whitespace", func(t *testing.T) {
		key, value, err := ParseLine("  KEY  =  value  ")
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "value" {
			t.Fatalf("got %q=%q, want KEY=value", key, value)
		}
	})

	t.Run("crlf", func(t *testing.T) {
		key, value, err := ParseLine("KEY=value\r\n")
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "value" {
			t.Fatalf("got %q=%q, want KEY=value", key, value)
		}
	})

	t.Run("empty_value", func(t *testing.T) {
		key, value, err := ParseLine("KEY=")
		if err != nil {
			t.Fatal(err)
		}
		if key != "KEY" || value != "" {
			t.Fatalf("got %q=%q, want KEY=(empty)", key, value)
		}
	})
}

func TestParse(t *testing.T) {
	t.Run("full", func(t *testing.T) {
		input := `# database config
DB_HOST=localhost
DB_PORT=5432
DB_NAME="my_app"
DB_PASS='s3cret'
export DB_USER=admin

# blank lines are ok

EMPTY=
COMMENTED=value # inline comment
`
		result, err := Parse(input)
		if err != nil {
			t.Fatal(err)
		}

		expected := map[string]string{
			"DB_HOST":   "localhost",
			"DB_PORT":   "5432",
			"DB_NAME":   "my_app",
			"DB_PASS":   "s3cret",
			"DB_USER":   "admin",
			"EMPTY":     "",
			"COMMENTED": "value",
		}

		if len(result) != len(expected) {
			t.Fatalf("got %d entries, want %d", len(result), len(expected))
		}

		for k, want := range expected {
			got, ok := result[k]
			if !ok {
				t.Errorf("missing key %q", k)
			} else if got != want {
				t.Errorf("key %q: got %q, want %q", k, got, want)
			}
		}
	})

	t.Run("duplicate_keys_last_wins", func(t *testing.T) {
		input := "KEY=first\nKEY=second\n"
		result, err := Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		if result["KEY"] != "second" {
			t.Fatalf("got %q, want 'second'", result["KEY"])
		}
	})

	t.Run("error_stops_early", func(t *testing.T) {
		input := "GOOD=value\nBAD_LINE\nANOTHER=ok\n"
		_, err := Parse(input)
		if err == nil {
			t.Fatal("expected error for invalid line")
		}
	})
}
