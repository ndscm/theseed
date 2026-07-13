package playpen

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"google.golang.org/grpc/codes"
)

// maxFileSize bounds a file read or written in one call. The file system is
// reached one whole file at a time — an editor saves the buffer it holds, and
// reads the file it opens — so this is what stands between a stray `podman exec
// cat` of a disk image and the memory of the process running it.
const maxFileSize = 32 * 1024 * 1024

// FileType is what a path is. The values are VS Code's own — vscode.FileType,
// which is what a caller of this eventually hands its editor — so that nothing
// between here and there has to translate them.
//
// It is a bitmask, and reads as one: a symbolic link to a directory is
// `FileTypeDirectory | FileTypeSymbolicLink`, which says both what will be found
// there and that a link was followed to find it. A caller that only wants to
// know what it is looking at masks the link bit off; one that cares that it
// points asks for the bit.
type FileType int

// FileTypeUnknown is a path that is there but is neither of the two below: a
// socket, a device, a symbolic link pointing nowhere. There is nothing to do
// with one, and a caller shown it is being told only that the name is taken.
const FileTypeUnknown FileType = 0

// FileTypeFile is a path that can be read whole and written whole.
const FileTypeFile FileType = 1

// FileTypeDirectory is a path that can be listed.
const FileTypeDirectory FileType = 2

// FileTypeSymbolicLink says that the path is a symbolic link. It is a bit rather
// than a kind of its own: what the link points at is in the rest of the type, so
// a link to a directory is a directory that points, and a link pointing nowhere
// is FileTypeSymbolicLink and nothing else.
const FileTypeSymbolicLink FileType = 64

// FileStat is what is known about one path in the playpen. It is what `stat`
// says about it, and nothing is remembered between two of these: a path is
// asked about afresh every time.
type FileStat struct {
	// Type is what the path is, links followed and the link itself said in the
	// same word: see FileType.
	Type FileType

	// Size is how many bytes the file holds, of the target of a symbolic link
	// rather than of the link. A directory's is whatever the file system keeps
	// its entries in, and means nothing to a caller.
	Size int64

	// ModificationTimestampMs is when the file's contents were last written, in
	// milliseconds since the epoch.
	ModificationTimestampMs int64

	// ChangeTimestampMs is when the file's inode last changed, in milliseconds
	// since the epoch. It is not when the file was made and not when it was
	// written: it moves later than ModificationTimestampMs whenever a rename or a
	// chmod touches the file without touching what is in it.
	//
	// This is POSIX's `ctime`, which is not what VS Code means by `ctime` — the
	// editor means the birth time below. They are two different times, both kept
	// here, and a caller says which it means by naming it.
	ChangeTimestampMs int64

	// CreationTimestampMs is when the file was made, in milliseconds since the
	// epoch — the birth time, and the one thing here that never moves again.
	//
	// It is zero when the file system does not keep one. Not all of them do: the
	// birth time is a later addition to the inode, and `stat` reports 0 rather
	// than guess. A caller that must show something for a file that has no birth
	// time falls back to ChangeTimestampMs, which every file has.
	CreationTimestampMs int64
}

// classifyStderr reads a failed command's complaint and gives it a code the
// caller can act on. What the file system refused is in the words coreutils
// used, and nowhere else: `podman exec` reports only that the command exited
// non-zero.
//
// An error that says nothing recognizable is Internal, not NotFound. A file
// system that has broken in some new way must not be mistaken for an empty
// directory.
func classifyStderr(stderr string) codes.Code {
	switch {
	case strings.Contains(stderr, "No such file or directory"):
		return codes.NotFound
	case strings.Contains(stderr, "Permission denied"):
		return codes.PermissionDenied
	case strings.Contains(stderr, "File exists"):
		return codes.AlreadyExists
	case strings.Contains(stderr, "Not a directory"),
		strings.Contains(stderr, "Is a directory"),
		strings.Contains(stderr, "Directory not empty"):
		return codes.FailedPrecondition
	default:
		return codes.Internal
	}
}

// parseFileType reads one of `find`'s or `stat`'s type letters, of a path with
// its links already followed. `stat` is asked for `%F`, whose words are spelled
// out; `find` for `%Y`, a single letter. Both are answered here, so a caller of
// either has one thing to read.
//
// The link bit is not in what either of them says: whether the path that was
// followed was a link is a second question, asked separately, and its answer is
// OR'd onto this one by whoever asked it.
func parseFileType(fileType string) FileType {
	switch fileType {
	case "regular file", "regular empty file", "f":
		return FileTypeFile
	case "directory", "d":
		return FileTypeDirectory
	default:
		return FileTypeUnknown
	}
}

// shellFlag renders a flag as the "1" the scripts test for.
func shellFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}

// SimpleFileSystem is the file system inside the playpen container, reached
// from outside it.
//
// The operations are named as VS Code names them — Stat, ReadDirectory,
// ReadFile, WriteFile, CreateDirectory, Delete, Rename — because a VS Code file
// system provider is what they are read through, and one name for a thing is
// better than two. FUSE's names would suggest something this does not do: a file
// is read whole and written whole here, as it is in a provider, where reading a
// part of one is a proposed API and not to be built on yet.
//
// Nothing is mounted, and nothing is held between two calls: each is a command
// run in the container, and what it says is what the caller is told.
//
// Every one of them runs as the person whose workstation this is. The
// container's own permissions are the authorization, and there is no way around
// them here: a path they may not read is refused inside, and the refusal is
// carried back out.
type SimpleFileSystem struct {
	containerName string
	userHandle    string
}

// exec runs a shell script inside the playpen container as the playpen user,
// with args as its positional parameters, and hands it stdin. It is what every
// operation below is made of, and everything true of all of them is true here.
//
// The paths a script acts on are passed as arguments rather than spliced into
// it: a file called `; rm -rf ~` is a file, and naming it must not run anything.
//
// Every command a script runs is named by its absolute path. A shell searches
// PATH, and PATH belongs to the person whose shell it is: a `stat` earlier on
// theirs than the real one is a `stat` this would run, as them, and believe. The
// container's own /usr/bin is what these mean, and saying so is what keeps the
// answer the file system's rather than something a login script put in front of
// it. `printf` and `test` are the shell's own and are not looked up at all.
//
// What a script refuses, it refuses in the words coreutils would have used,
// because that is what classifyStderr reads back: a script that checks a
// condition itself — that a directory is a directory, that a file is not already
// there — says so the same way the command it stands in for would have.
func (fs *SimpleFileSystem) exec(
	ctx context.Context, script string, stdin []byte, args ...string,
) ([]byte, error) {
	// The "sh" after the script is `$0`, and is there to be nothing: `sh -c` reads
	// its next operand as the shell's own name, and only the ones after that as
	// `$1`, `$2`. Without something in its place the first path would be taken for
	// the name, and every script's `"$1"` would quietly be the argument after the
	// one it means.
	execArgs := []string{
		"exec",
		"--interactive",
		"--user", fs.userHandle,
		fs.containerName,
		"/bin/sh", "-c", script, "sh",
	}
	execArgs = append(execArgs, args...)

	cmd := exec.CommandContext(ctx, "podman", execArgs...)
	cmd.Stdin = bytes.NewReader(stdin)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		exitErr, ok := errors.AsType[*exec.ExitError](err)
		if !ok {
			return nil, seederr.Wrap(err)
		}
		// The shell's complaint is the error: it says which path, and what
		// about it, in the words the code below is read back from.
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = exitErr.Error()
		}
		return nil, seederr.CodeErrorf(classifyStderr(message), "%v", message)
	}
	return stdout.Bytes(), nil
}

// Stat reports what a path is. A path that is not there is NotFound, which is
// how a caller asks whether it exists.
func (fs *SimpleFileSystem) Stat(ctx context.Context, path string) (*FileStat, error) {
	// The path is stat'ed twice over: first as it stands, so a symbolic link can
	// be seen to be one, and then with links followed, which is what a caller
	// opening it will find. `stat` fails on a path that is not there, which is
	// how the caller hears that.
	//
	// Both stats are a `--printf` rather than a `--format`, because only the
	// former reads the `\n` in the format as a newline; `--format` would print it
	// as the two characters it is written with, and hand back one long line. It
	// prints no trailing newline of its own either, which is why every field ends
	// with one.
	//
	// `%Z` is the inode's change time, which every file has. `%W` is the birth
	// time, and is 0 on a file system that keeps none. Both are asked for and both
	// are reported: they are different times, and which one a caller wants is
	// theirs to say.
	output, err := fs.exec(ctx, `
set -e
/usr/bin/stat --printf '%F\n' -- "$1"
/usr/bin/stat --dereference --printf '%F\n%s\n%Y\n%Z\n%W\n' -- "$1"
`, nil, path)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// Six lines: the type of the path as it stands, then the type, size,
	// modification time, change time and birth time of what it points at.
	lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")
	if len(lines) != 6 {
		return nil, seederr.CodeErrorf(codes.Internal,
			"stat of %q reported %d lines, want 6", path, len(lines))
	}

	size, err := strconv.ParseInt(lines[2], 10, 64)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	modificationTimestampS, err := strconv.ParseInt(lines[3], 10, 64)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	changeTimestampS, err := strconv.ParseInt(lines[4], 10, 64)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	creationTimestampS, err := strconv.ParseInt(lines[5], 10, 64)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// The first line is the path as it stands, and is how a link is seen to be
	// one; the second is what it points at, which is what the rest of the type
	// says. A link to a directory comes back as both.
	fileType := parseFileType(lines[1])
	if lines[0] == "symbolic link" {
		fileType |= FileTypeSymbolicLink
	}

	// A birth time of 0 is `stat` saying it has none to give, and stays 0: a file
	// made at the epoch and a file whose file system does not remember are told
	// apart by nobody, and multiplying either by a thousand is still nothing.
	info := &FileStat{
		Type: fileType,
		Size: size,

		ModificationTimestampMs: modificationTimestampS * 1000,
		ChangeTimestampMs:       changeTimestampS * 1000,
		CreationTimestampMs:     creationTimestampS * 1000,
	}
	return info, nil
}

// ReadDirectory lists what is in a directory: each name in it, and what that
// name is. The names are the directory's own and nothing else — the directory
// they are in is the one the caller asked about, and is not repeated on each of
// them — and no two of them are the same, which is what makes the listing a map.
//
// It does not recurse: a caller walking a tree asks about each directory in
// turn.
func (fs *SimpleFileSystem) ReadDirectory(
	ctx context.Context, path string,
) (map[string]FileType, error) {
	// One entry per record: the type of the entry as it stands and with links
	// followed, then the name. Records are terminated by a NUL and the two fields
	// split on a tab, because a file name may contain anything else, newlines
	// included.
	//
	// `find` walks silently into a path that is not a directory, so the script
	// checks first: a caller listing a file has made a mistake, and must hear
	// about it rather than be handed nothing.
	output, err := fs.exec(ctx, `
if [ ! -e "$1" ]; then
  printf '%s: No such file or directory\n' "$1" >&2
  exit 1
fi
if [ ! -d "$1" ]; then
  printf '%s: Not a directory\n' "$1" >&2
  exit 1
fi
exec /usr/bin/find "$1" -mindepth 1 -maxdepth 1 -printf '%y%Y\t%f\0'
`, nil, path)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	entries := map[string]FileType{}
	for _, record := range strings.Split(string(output), "\x00") {
		if record == "" {
			// The last record is terminated too, so the split leaves an empty
			// tail — and an empty directory is nothing but that tail.
			continue
		}
		letters, name, found := strings.Cut(record, "\t")
		// `%y%Y` is the type of the entry followed by the type of what it points
		// at, which are the same two letters unless the entry is a link.
		if !found || len(letters) != 2 {
			return nil, seederr.CodeErrorf(codes.Internal,
				"listing of %q produced an unreadable entry %q", path, record)
		}

		fileType := parseFileType(letters[1:])
		if letters[0] == 'l' {
			fileType |= FileTypeSymbolicLink
		}

		entries[name] = fileType
	}
	return entries, nil
}

// CreateDirectory creates one directory. Its parent must be there already:
// nothing here makes a chain of them, because a caller that meant to create one
// directory and misspelled its parent must not end up with two.
func (fs *SimpleFileSystem) CreateDirectory(ctx context.Context, path string) error {
	_, err := fs.exec(ctx, `
exec /usr/bin/mkdir -- "$1"
`, nil, path)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// ReadFile reads a whole file: there are no handles and no offsets here, and a
// file is read all at once or not at all. One larger than maxFileSize is refused
// rather than truncated — half a file is not the file, and an editor handed the
// beginning of one would save it back over the rest.
func (fs *SimpleFileSystem) ReadFile(ctx context.Context, path string) ([]byte, error) {
	info, err := fs.Stat(ctx, path)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if info.Type&FileTypeDirectory != 0 {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "%v: Is a directory", path)
	}
	if info.Size > maxFileSize {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition,
			"%v is %d bytes, over the %d that may be read at once", path, info.Size, maxFileSize)
	}

	// The stat above is a claim, not the read. Its size can grow before the read
	// reaches the file, and a fifo, a device, or a file still being appended never
	// had a truthful size to give: each of those passes the guard above and would
	// then stream without end into a buffer this process holds — a `cat /dev/zero`
	// spent as the memory of amadeus itself, not the container it serves. So the
	// read caps itself rather than trusting the stat: `head -c` stops after
	// maxFileSize+1 bytes however much more the file would yield, and a read that
	// comes back that one byte over the limit is one that hit the cap and is
	// refused, the same as an honest large file is above.
	content, err := fs.exec(ctx, `
exec /usr/bin/head -c "$2" -- "$1"
`, nil, path, strconv.FormatInt(maxFileSize+1, 10))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if int64(len(content)) > maxFileSize {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition,
			"%v is over the %d bytes that may be read at once", path, maxFileSize)
	}
	return content, nil
}

// Write replaces a file's contents, all of them: as with Read, there is no
// offset to write at, and the buffer an editor holds is the file. create says a
// file that is not there may be made, overwrite that one that is there may be
// replaced; a write allowed to do neither is refused.
func (fs *SimpleFileSystem) WriteFile(
	ctx context.Context, path string, content []byte, create bool, overwrite bool,
) error {
	if len(content) > maxFileSize {
		return seederr.CodeErrorf(codes.FailedPrecondition,
			"%v is %d bytes, over the %d that may be written at once", path, len(content), maxFileSize)
	}

	// Whether the file may be created, and whether one that is there may be
	// replaced, are the flags in $2 and $3: a write allowed to do neither is
	// refused rather than quietly doing nothing.
	_, err := fs.exec(ctx, `
if [ -e "$1" ]; then
  if [ "$3" != 1 ]; then
    printf '%s: File exists\n' "$1" >&2
    exit 1
  fi
elif [ "$2" != 1 ]; then
  printf '%s: No such file or directory\n' "$1" >&2
  exit 1
fi
exec /usr/bin/cat >"$1"
`, content, path, shellFlag(create), shellFlag(overwrite))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Delete removes a path, of whatever kind: a file system provider deletes by
// naming a path, and working out what is there is this side's business rather
// than the editor's. recursive is what says a directory may go with everything
// in it; without it, one that is not empty is refused.
func (fs *SimpleFileSystem) Delete(ctx context.Context, path string, recursive bool) error {
	// A directory goes with everything in it only when $2 says so; otherwise
	// `rmdir` refuses one that is not empty. `rm -f` would call a path that was
	// never there a success, so the script looks first — a caller deleting
	// something that is gone has lost track of it, and an editor shows that as an
	// error rather than a no-op.
	_, err := fs.exec(ctx, `
if [ ! -e "$1" ] && [ ! -L "$1" ]; then
  printf '%s: No such file or directory\n' "$1" >&2
  exit 1
fi
if [ "$2" = 1 ]; then
  exec /usr/bin/rm -rf -- "$1"
fi
if [ -d "$1" ] && [ ! -L "$1" ]; then
  exec /usr/bin/rmdir -- "$1"
fi
exec /usr/bin/rm -- "$1"
`, nil, path, shellFlag(recursive))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Rename moves a path onto another, which is also how one is renamed in place. A
// destination that exists is replaced only when overwrite says so.
func (fs *SimpleFileSystem) Rename(
	ctx context.Context, sourcePath string, destinationPath string, overwrite bool,
) error {
	// `mv` would replace the destination without being asked, so the script asks:
	// an overwrite the caller did not allow is refused.
	_, err := fs.exec(ctx, `
if [ ! -e "$1" ] && [ ! -L "$1" ]; then
  printf '%s: No such file or directory\n' "$1" >&2
  exit 1
fi
if { [ -e "$2" ] || [ -L "$2" ]; } && [ "$3" != 1 ]; then
  printf '%s: File exists\n' "$2" >&2
  exit 1
fi
exec /usr/bin/mv -f -- "$1" "$2"
`, nil, sourcePath, destinationPath, shellFlag(overwrite))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// WrapSimpleFileSystem reaches the file system of the named container, as the
// named user. Both are the playpen's — PlaypenController.FileSystem is what
// hands them over, and there is no other file system to reach.
func WrapSimpleFileSystem(containerName string, userHandle string) *SimpleFileSystem {
	return &SimpleFileSystem{
		containerName: containerName,
		userHandle:    userHandle,
	}
}
