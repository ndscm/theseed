package seedflag

import (
	"errors"
	"flag"
	"io"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// IsUnknownFlag reports whether err is the stdlib "flag provided but not
// defined" error. There's no exported sentinel, so we match on the message.
func IsUnknownFlag(err error) bool {
	return strings.HasPrefix(err.Error(), "flag provided but not defined")
}

type ParseOptions struct {
	AnywhereFlag bool

	UnknownFlag bool

	LogFlags bool
}

type GlobalParseOptions struct {
	ParseOptions

	EnvPrefix string

	FallbackEnvPrefixes []string
}

type CommandFlags struct {
	command string
	flags   map[string]FlagDefinition
}

func (cf *CommandFlags) String() string {
	s := strings.Builder{}
	s.WriteString(cf.command)
	s.WriteString(":")
	for name, item := range cf.flags {
		s.WriteString("\n  ")
		s.WriteString(name)
		s.WriteString(": ")
		s.WriteString(item.String())
	}
	return s.String()
}

func (cf *CommandFlags) DefineFlag(name string, flagDefinition FlagDefinition) FlagDefinition {
	cf.flags[name] = flagDefinition
	return flagDefinition
}

func (cf *CommandFlags) DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	item := NewBoolFlag(defaultValue, usage)
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineString(name string, defaultValue string, usage string) *StringFlag {
	item := NewStringFlag(defaultValue, usage)
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineStringList(name string, defaultValues []string, usage string) *StringListFlag {
	item := NewStringListFlag(defaultValues, usage)
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineFile(name string, defaultValue string, usage string) *FileFlag {
	item := NewFileFlag(defaultValue, usage)
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineSecret(name string, usage string) *SecretFlag {
	item := NewSecretFlag(usage)
	// TODO(nagi): support susceptible secret flag for loading secret from cli or env.
	cf.flags[name+"_file"] = item.fileFlag
	return item
}

func (cf *CommandFlags) newFlagSet(output io.Writer) *flag.FlagSet {
	s := flag.NewFlagSet(cf.command, flag.ContinueOnError)
	s.SetOutput(output)
	for name, item := range cf.flags {
		s.Var(item, name, item.Usage())
	}
	return s
}

func (cf *CommandFlags) PrintUsage() {
	s := cf.newFlagSet(os.Stderr)
	s.Usage()
}

func (cf *CommandFlags) Parse(args []string, opts *ParseOptions) ([]string, error) {
	if opts == nil {
		opts = &ParseOptions{}
	}

	// A literal "--" terminates flag parsing: everything from the first "--"
	// onward is appended verbatim and never treated as a flag, even in
	// AnywhereFlag mode. The "--" is kept rather than dropped so that re-parsing
	// the same args in a later layer (global, then subcommand) stops at the same
	// boundary; the terminal consumer drops it.
	var passthrough []string
	for i, arg := range args {
		if arg == "--" {
			passthrough = args[i:]
			args = args[:i]
			break
		}
	}

	s := cf.newFlagSet(io.Discard)
	finalArgs := []string{}
	remainArgs := args

	// First round: -h/--help always triggers usage. Loops while unknown
	// flags get captured; exits past the first positional or once args run
	// out. Cases that don't need a second round leave remainArgs empty so
	// the if-guard below skips it.
	for {
		err := s.Parse(remainArgs)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				cf.PrintUsage()
				// Exit here to mirror the stdlib flag.CommandLine behavior,
				// which calls os.Exit(0) on -h/--help via flag.ExitOnError.
				// Keeping the same contract avoids surprising callers that
				// migrated from the stdlib package.
				os.Exit(0)
			}
			if opts.UnknownFlag && IsUnknownFlag(err) {
				n := len(remainArgs) - len(s.Args())
				finalArgs = append(finalArgs, remainArgs[n-1])
				remainArgs = s.Args()
				continue
			}
			return nil, seederr.Wrap(err)
		}
		if opts.AnywhereFlag && len(s.Args()) > 0 {
			finalArgs = append(finalArgs, s.Args()[0])
			remainArgs = s.Args()[1:]
			break
		}
		finalArgs = append(finalArgs, s.Args()...)
		remainArgs = nil
		break
	}

	// Subsequent rounds: only reachable when anywhereFlag is set and the
	// first round left args after a positional. -h/--help is passed through
	// to the downstream parser when unknownFlag is set; otherwise it still
	// triggers usage like any -h/--help would.
	if opts.AnywhereFlag && len(remainArgs) > 0 {
		for {
			err := s.Parse(remainArgs)
			if err != nil {
				if !opts.UnknownFlag {
					if errors.Is(err, flag.ErrHelp) {
						cf.PrintUsage()
						// Same as the first round: match stdlib
						// flag.CommandLine, which exits 0 on -h/--help.
						os.Exit(0)
					}
				} else {
					if IsUnknownFlag(err) || errors.Is(err, flag.ErrHelp) {
						n := len(remainArgs) - len(s.Args())
						finalArgs = append(finalArgs, remainArgs[n-1])
						remainArgs = s.Args()
						continue
					}
				}
				return nil, seederr.Wrap(err)
			}
			if len(s.Args()) == 0 {
				break
			}
			finalArgs = append(finalArgs, s.Args()[0])
			remainArgs = s.Args()[1:]
		}
	}

	// Must be read after parsing is done
	if flagLogFlags.Get() {
		opts.LogFlags = true
	}

	if opts.LogFlags || flagLogFlags.Get() {
		seedlog.Infof("%s", cf.String())
	} else {
		seedlog.Debugf("%s", cf.String())
	}

	finalArgs = append(finalArgs, passthrough...)
	return finalArgs, nil
}

func NewCommandFlags(command string) *CommandFlags {
	return &CommandFlags{
		command: command,
		flags:   map[string]FlagDefinition{},
	}
}

var globalFlags = NewCommandFlags(os.Args[0])

func DefineFlag(name string, flagDefinition FlagDefinition) FlagDefinition {
	return globalFlags.DefineFlag(name, flagDefinition)
}

func DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	return globalFlags.DefineBool(name, defaultValue, usage)
}

func DefineString(name string, defaultValue string, usage string) *StringFlag {
	return globalFlags.DefineString(name, defaultValue, usage)
}

func DefineStringList(name string, defaultValues []string, usage string) *StringListFlag {
	return globalFlags.DefineStringList(name, defaultValues, usage)
}

func DefineFile(name string, defaultValue string, usage string) *FileFlag {
	return globalFlags.DefineFile(name, defaultValue, usage)
}

func DefineSecret(name string, usage string) *SecretFlag {
	return globalFlags.DefineSecret(name, usage)
}

var flagLogFlags = DefineBool("log_flags", false, "Print all flags and their values")

func Parse(opts *GlobalParseOptions) ([]string, error) {
	if opts == nil {
		opts = &GlobalParseOptions{}
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if opts.EnvPrefix != "" && strings.HasPrefix(parts[0], opts.EnvPrefix) {
			pure := strings.TrimPrefix(parts[0], opts.EnvPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := globalFlags.flags[name]
			if !ok {
				seedlog.Warnf("Unknown env flag: %s=%s", parts[0], value)
				continue
			}
			err := item.Set(value)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		} else {
			for _, fallbackPrefix := range opts.FallbackEnvPrefixes {
				if fallbackPrefix == "" || !strings.HasPrefix(parts[0], fallbackPrefix) {
					continue
				}
				pure := strings.TrimPrefix(parts[0], fallbackPrefix)
				name := strings.ToLower(strings.TrimSpace(pure))
				value := strings.TrimSpace(parts[1])
				item, ok := globalFlags.flags[name]
				if !ok {
					continue
				}
				err := item.Set(value)
				if err != nil {
					return nil, seederr.Wrap(err)
				}
				seedlog.Infof("Loaded fallback env flag: %s=%s", parts[0], item.String())
				break
			}
		}
	}

	finalArgs, err := globalFlags.Parse(os.Args[1:], &opts.ParseOptions)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return finalArgs, nil
}

// Finalize must be called after subcommand parsing has consumed everything it
// recognizes from the args returned by a Parse(WithUnknownFlag()) call. Any
// remaining token that still looks like a flag is by definition unclaimed by
// either the global or the subcommand layer; Finalize prints usage and exits
// non-zero in that case.
//
// A "--" marks end-of-flags, so only tokens before the first "--" are checked.
func Finalize(args []string) {
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if strings.HasPrefix(arg, "-") {
			seedlog.Errorf("Unknown flag: %s", arg)
			globalFlags.PrintUsage()
			// Exit here to mirror stdlib flag.CommandLine, which exits 2
			// on an unknown flag under flag.ExitOnError. We use 1 to stay
			// consistent with the rest of this codebase's error exits, but
			// the intent—terminate the process on an unrecognized flag—is
			// the same contract callers get from the stdlib package.
			os.Exit(1)
		}
	}
}
