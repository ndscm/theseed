package playpen

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"google.golang.org/grpc/codes"
)

// withPodmanShim puts a `podman` on the PATH that runs the command on this
// machine instead of in a container: it drops the
// `exec --interactive --user <handle> playpen` in front and runs the rest.
//
// What is being tested is the part that can be wrong — the scripts, and the
// reading of what they print. `podman exec` carrying argv into a container and
// bytes back out is not, and standing a container up to watch it do so would
// test podman rather than this.
func withPodmanShim(t *testing.T) {
	binDir := t.TempDir()
	shim := filepath.Join(binDir, "podman")
	err := os.WriteFile(shim, []byte("#!/bin/sh\nshift 5\nexec \"$@\"\n"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))
}

// errorCode is the code the file system failed with, which is the whole of what
// an editor acts on: a path that is not there is a different thing from one it
// may not read.
func errorCode(err error) codes.Code {
	seedErr := &seederr.SeedError{}
	if !errors.As(err, &seedErr) {
		return codes.OK
	}
	return codes.Code(seedErr.Code())
}

// TestPlaypenFileSystem drives every operation against a real file system
// reached through the shim. Each case makes its own tree in its own t.TempDir
// and touches nothing another case made, so none is the setup for the next: a
// case reads the same whichever order the cases run in, and a failure is the
// fault of the one case it is reported against.
func TestPlaypenFileSystem(t *testing.T) {
	withPodmanShim(t)
	fileSystem := WrapSimpleFileSystem(playpenContainerName, "christina")
	ctx := context.Background()

	t.Run("stat a file", func(t *testing.T) {
		home := t.TempDir()
		path := filepath.Join(home, "hello.txt")
		err := fileSystem.WriteFile(ctx, path, []byte("hello world\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		info, err := fileSystem.Stat(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Type != FileTypeFile {
			t.Errorf("type is %v, want %v", info.Type, FileTypeFile)
		}
		if info.Size != 12 {
			t.Errorf("size is %v, want 12", info.Size)
		}
		if info.ModificationTimestampMs < 1_600_000_000_000 {
			t.Errorf("modified at %v, which is not a time in milliseconds", info.ModificationTimestampMs)
		}
		if info.ChangeTimestampMs < 1_600_000_000_000 {
			t.Errorf("changed at %v, which is not a time in milliseconds", info.ChangeTimestampMs)
		}
		// The birth time is 0 on a file system that keeps none, and this test runs
		// on whichever one the runner happens to have. So a file just made either
		// was made now, or was made never — and never is 0, not some other year.
		if info.CreationTimestampMs != 0 && info.CreationTimestampMs < 1_600_000_000_000 {
			t.Errorf("created at %v, which is neither a time in milliseconds nor 0",
				info.CreationTimestampMs)
		}
	})

	t.Run("stat a symbolic link to a directory", func(t *testing.T) {
		home := t.TempDir()
		target := filepath.Join(home, "notes")
		err := fileSystem.CreateDirectory(ctx, target)
		if err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(target, filepath.Join(home, "link"))
		if err != nil {
			t.Fatal(err)
		}

		// The link is followed for what it is — a caller opening it will find a
		// directory — and reported as a link all the same: both are in the one type.
		info, err := fileSystem.Stat(ctx, filepath.Join(home, "link"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Type != FileTypeDirectory|FileTypeSymbolicLink {
			t.Errorf("type is %v, want %v", info.Type, FileTypeDirectory|FileTypeSymbolicLink)
		}
	})

	t.Run("stat a path that is not there", func(t *testing.T) {
		home := t.TempDir()
		_, err := fileSystem.Stat(ctx, filepath.Join(home, "ghost"))
		if errorCode(err) != codes.NotFound {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.NotFound)
		}
	})

	t.Run("read a directory", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.CreateDirectory(ctx, filepath.Join(home, "notes"))
		if err != nil {
			t.Fatal(err)
		}
		err = os.Symlink(filepath.Join(home, "notes"), filepath.Join(home, "link"))
		if err != nil {
			t.Fatal(err)
		}

		entries, err := fileSystem.ReadDirectory(ctx, home)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Fatalf("listed %v, want the directory and the link to it", entries)
		}
		if entries["notes"] != FileTypeDirectory {
			t.Errorf("notes is listed as %v", entries["notes"])
		}
		if entries["link"] != FileTypeDirectory|FileTypeSymbolicLink {
			t.Errorf("link is listed as %v", entries["link"])
		}
	})

	t.Run("read an empty directory", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.CreateDirectory(ctx, filepath.Join(home, "empty"))
		if err != nil {
			t.Fatal(err)
		}

		entries, err := fileSystem.ReadDirectory(ctx, filepath.Join(home, "empty"))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Errorf("an empty directory listed %v", entries)
		}
	})

	t.Run("read a file as a directory", func(t *testing.T) {
		home := t.TempDir()
		path := filepath.Join(home, "kept.txt")
		err := fileSystem.WriteFile(ctx, path, []byte("kept\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fileSystem.ReadDirectory(ctx, path)
		if errorCode(err) != codes.FailedPrecondition {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.FailedPrecondition)
		}
	})

	t.Run("create a directory that is already there", func(t *testing.T) {
		home := t.TempDir()
		dir := filepath.Join(home, "empty")
		err := fileSystem.CreateDirectory(ctx, dir)
		if err != nil {
			t.Fatal(err)
		}

		err = fileSystem.CreateDirectory(ctx, dir)
		if errorCode(err) != codes.AlreadyExists {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.AlreadyExists)
		}
	})

	t.Run("read a directory as a file", func(t *testing.T) {
		home := t.TempDir()
		dir := filepath.Join(home, "empty")
		err := fileSystem.CreateDirectory(ctx, dir)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fileSystem.ReadFile(ctx, dir)
		if errorCode(err) != codes.FailedPrecondition {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.FailedPrecondition)
		}
	})

	t.Run("read a path shaped like a shell command is a path", func(t *testing.T) {
		home := t.TempDir()
		// The scripts take the path as an argument and use it quoted, so a name that
		// reads like a command is a name and runs nothing. Were it spliced in
		// unquoted, `cat >…/x; touch pwned` would redirect into `x` and then run
		// `touch`; here it writes one file whose name is the whole string.
		path := filepath.Join(home, "x; touch pwned")
		err := fileSystem.WriteFile(ctx, path, []byte("safe\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		content, err := fileSystem.ReadFile(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "safe\n" {
			t.Errorf("read back %q, want %q", content, "safe\n")
		}

		// Had the write been split on the `;`, the redirect would have made this
		// file: an absolute path, so it lands where the test can see it.
		split := filepath.Join(home, "x")
		_, err = os.Stat(split)
		if !os.IsNotExist(err) {
			t.Errorf("%q exists, so the path was split and run as a command", split)
		}
	})

	t.Run("read a special file that streams without end is capped", func(t *testing.T) {
		// A stat can only answer the size guard for a regular file; /dev/zero stats
		// as a zero-byte character device, so it passes that guard and then yields
		// bytes without end. The read caps itself at maxFileSize rather than trust
		// the stat, so this is refused rather than drawn into memory without bound —
		// the very `cat /dev/zero` the cap stands against.
		_, err := fileSystem.ReadFile(ctx, "/dev/zero")
		if errorCode(err) != codes.FailedPrecondition {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.FailedPrecondition)
		}
	})

	t.Run("write a file and read it back", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.CreateDirectory(ctx, filepath.Join(home, "notes"))
		if err != nil {
			t.Fatal(err)
		}

		path := filepath.Join(home, "notes", "hello.txt")
		err = fileSystem.WriteFile(ctx, path, []byte("hello world\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		content, err := fileSystem.ReadFile(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "hello world\n" {
			t.Errorf("read back %q, want %q", content, "hello world\n")
		}
	})

	t.Run("write a file that is not there without leave to create it", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.WriteFile(ctx, filepath.Join(home, "ghost.txt"), []byte("x"), false, false)
		if errorCode(err) != codes.NotFound {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.NotFound)
		}
	})

	t.Run("write over a file without leave to overwrite it", func(t *testing.T) {
		home := t.TempDir()
		path := filepath.Join(home, "kept.txt")
		err := fileSystem.WriteFile(ctx, path, []byte("kept\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		err = fileSystem.WriteFile(ctx, path, []byte("lost\n"), true, false)
		if errorCode(err) != codes.AlreadyExists {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.AlreadyExists)
		}

		content, err := fileSystem.ReadFile(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != "kept\n" {
			t.Errorf("the file is now %q, so the refused write went through anyway", content)
		}
	})

	t.Run("delete a directory and everything in it", func(t *testing.T) {
		home := t.TempDir()
		notes := filepath.Join(home, "notes")
		err := fileSystem.CreateDirectory(ctx, notes)
		if err != nil {
			t.Fatal(err)
		}
		child := filepath.Join(notes, "hello.txt")
		err = fileSystem.WriteFile(ctx, child, []byte("hello world\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		err = fileSystem.Delete(ctx, notes, true)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fileSystem.Stat(ctx, child)
		if errorCode(err) != codes.NotFound {
			t.Errorf("a file under the deleted directory stats as %v", errorCode(err))
		}
	})

	t.Run("delete a path that is not there", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.Delete(ctx, filepath.Join(home, "ghost"), false)
		if errorCode(err) != codes.NotFound {
			t.Errorf("failed with %v, want %v", errorCode(err), codes.NotFound)
		}
	})

	t.Run("rename a file", func(t *testing.T) {
		home := t.TempDir()
		err := fileSystem.CreateDirectory(ctx, filepath.Join(home, "notes"))
		if err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(home, "notes", "hello.txt")
		err = fileSystem.WriteFile(ctx, source, []byte("hello world\n"), true, false)
		if err != nil {
			t.Fatal(err)
		}

		destination := filepath.Join(home, "notes", "renamed.txt")
		err = fileSystem.Rename(ctx, source, destination, false)
		if err != nil {
			t.Fatal(err)
		}

		_, err = fileSystem.Stat(ctx, source)
		if errorCode(err) != codes.NotFound {
			t.Errorf("the file it was renamed from stats as %v", errorCode(err))
		}
		_, err = fileSystem.Stat(ctx, destination)
		if err != nil {
			t.Errorf("the file it was renamed to is not there: %v", err)
		}
	})
}
