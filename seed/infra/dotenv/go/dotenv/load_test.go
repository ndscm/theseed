package dotenv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEnvFile(t *testing.T) {
	t.Run("valid_file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		os.WriteFile(path, []byte("KEY=value\n"), 0644)

		result, err := ReadEnvFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if result["KEY"] != "value" {
			t.Fatalf("got %q, want 'value'", result["KEY"])
		}
	})

	t.Run("missing_file", func(t *testing.T) {
		_, err := ReadEnvFile("/nonexistent/.env")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid_content", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		os.WriteFile(path, []byte("INVALID_LINE\n"), 0644)

		_, err := ReadEnvFile(path)
		if err == nil {
			t.Fatal("expected error for invalid content")
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("sets_env", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		os.WriteFile(path, []byte("DOTENV_TEST_A=hello\n"), 0644)
		t.Cleanup(func() { os.Unsetenv("DOTENV_TEST_A") })

		err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		got := os.Getenv("DOTENV_TEST_A")
		if got != "hello" {
			t.Fatalf("got %q, want 'hello'", got)
		}
	})

	t.Run("override_existing", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, ".env")
		os.WriteFile(path, []byte("DOTENV_TEST_B=fromfile\n"), 0644)
		os.Setenv("DOTENV_TEST_B", "existing")
		t.Cleanup(func() { os.Unsetenv("DOTENV_TEST_B") })

		err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		got := os.Getenv("DOTENV_TEST_B")
		if got != "fromfile" {
			t.Fatalf("got %q, want 'fromfile'", got)
		}
	})

	t.Run("multiple_files_later_wins", func(t *testing.T) {
		dir := t.TempDir()
		path1 := filepath.Join(dir, "one.env")
		path2 := filepath.Join(dir, "two.env")
		os.WriteFile(path1, []byte("DOTENV_TEST_C=first\n"), 0644)
		os.WriteFile(path2, []byte("DOTENV_TEST_C=second\n"), 0644)
		t.Cleanup(func() { os.Unsetenv("DOTENV_TEST_C") })

		err := Load(path1, path2)
		if err != nil {
			t.Fatal(err)
		}
		got := os.Getenv("DOTENV_TEST_C")
		if got != "second" {
			t.Fatalf("got %q, want 'second'", got)
		}
	})

	t.Run("missing_file", func(t *testing.T) {
		err := Load("/nonexistent/.env")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("no_files", func(t *testing.T) {
		err := Load()
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestLoadAncestor(t *testing.T) {
	t.Run("finds_parent_env", func(t *testing.T) {
		root := t.TempDir()
		child := filepath.Join(root, "a", "b")
		os.MkdirAll(child, 0755)
		os.WriteFile(filepath.Join(root, ".env"), []byte("DOTENV_TEST_D=parent\n"), 0644)
		t.Cleanup(func() { os.Unsetenv("DOTENV_TEST_D") })

		err := LoadAncestor(child, ".env")
		if err != nil {
			t.Fatal(err)
		}
		got := os.Getenv("DOTENV_TEST_D")
		if got != "parent" {
			t.Fatalf("got %q, want 'parent'", got)
		}
	})

	t.Run("child_overrides_parent", func(t *testing.T) {
		root := t.TempDir()
		child := filepath.Join(root, "a")
		os.MkdirAll(child, 0755)
		os.WriteFile(filepath.Join(root, ".env"), []byte("DOTENV_TEST_E=parent\n"), 0644)
		os.WriteFile(filepath.Join(child, ".env"), []byte("DOTENV_TEST_E=child\n"), 0644)
		t.Cleanup(func() { os.Unsetenv("DOTENV_TEST_E") })

		err := LoadAncestor(child, ".env")
		if err != nil {
			t.Fatal(err)
		}
		got := os.Getenv("DOTENV_TEST_E")
		if got != "child" {
			t.Fatalf("got %q, want 'child'", got)
		}
	})

	t.Run("no_env_found", func(t *testing.T) {
		dir := t.TempDir()
		err := LoadAncestor(dir, ".env.nonexistent")
		if err != nil {
			t.Fatal(err)
		}
	})
}
