package seedflag

import (
	"errors"
	"flag"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type FlagDefinition interface {
	Usage() string
	flag.Value
}

type FlagItem struct {
	usage string
}

func (f *FlagItem) Usage() string {
	return f.usage
}

type BoolFlag struct {
	FlagItem
	value bool
}

func (f *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return seederr.Wrap(err)
	}
	f.value = v
	return nil
}

func (f *BoolFlag) Get() bool {
	return f.value
}

func (f *BoolFlag) String() string {
	return strconv.FormatBool(f.value)
}

func (f *BoolFlag) IsBoolFlag() bool {
	return true
}

type StringFlag struct {
	FlagItem
	value string
}

func (f *StringFlag) Set(s string) error {
	f.value = s
	return nil
}

func (f *StringFlag) Get() string {
	return f.value
}

func (f *StringFlag) String() string {
	return f.value
}

// IsUnknownFlag reports whether err is the stdlib "flag provided but not
// defined" error. There's no exported sentinel, so we match on the message.
func IsUnknownFlag(err error) bool {
	return strings.HasPrefix(err.Error(), "flag provided but not defined")
}

type parseOptions struct {
	anywhereFlag bool
	unknownFlag  bool
}

type parseOption interface {
	applyParseOption(*parseOptions)
}

type globalParseOptions struct {
	parseOptions []parseOption

	envPrefix         string
	fallbackEnvPrefix string
}

type globalParseOption interface {
	applyGlobalParseOption(*globalParseOptions)
}

type withAnywhereFlag struct {
	value bool
}

func (w *withAnywhereFlag) applyParseOption(o *parseOptions) {
	o.anywhereFlag = w.value
}

func (w *withAnywhereFlag) applyGlobalParseOption(o *globalParseOptions) {
	o.parseOptions = append(o.parseOptions, w)
}

func WithAnywhereFlag(value bool) *withAnywhereFlag {
	return &withAnywhereFlag{value: value}
}

// WithUnknownFlag tells Parse to leave unknown flags in the returned args
// instead of treating them as an error, so a downstream subcommand parser
// can consume them. Only use this when a subcommand layer will run next; the
// caller must invoke Finalize on whatever args remain after subcommand parsing
// to surface any flags that nothing claimed.
type withUnknownFlag struct {
	value bool
}

func (w *withUnknownFlag) applyParseOption(o *parseOptions) {
	o.unknownFlag = w.value
}

func (w *withUnknownFlag) applyGlobalParseOption(o *globalParseOptions) {
	o.parseOptions = append(o.parseOptions, w)
}

func WithUnknownFlag(value bool) *withUnknownFlag {
	return &withUnknownFlag{value: value}
}

type CommandFlags struct {
	command string
	flags   map[string]FlagDefinition
}

func (cf *CommandFlags) DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	item := &BoolFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	cf.flags[name] = item
	return item
}

func (cf *CommandFlags) DefineString(name string, defaultValue string, usage string) *StringFlag {
	item := &StringFlag{
		FlagItem: FlagItem{
			usage: usage,
		},
		value: defaultValue,
	}
	cf.flags[name] = item
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

func (cf *CommandFlags) Parse(args []string, opts ...parseOption) ([]string, error) {
	o := &parseOptions{}
	for _, opt := range opts {
		opt.applyParseOption(o)
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
			if o.unknownFlag && IsUnknownFlag(err) {
				n := len(remainArgs) - len(s.Args())
				finalArgs = append(finalArgs, remainArgs[n-1])
				remainArgs = s.Args()
				continue
			}
			return nil, seederr.Wrap(err)
		}
		if o.anywhereFlag && len(s.Args()) > 0 {
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
	if o.anywhereFlag && len(remainArgs) > 0 {
		for {
			err := s.Parse(remainArgs)
			if err != nil {
				if !o.unknownFlag {
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

	return finalArgs, nil
}

func NewCommandFlags(command string) *CommandFlags {
	return &CommandFlags{
		command: command,
		flags:   map[string]FlagDefinition{},
	}
}

var globalFlags = NewCommandFlags(os.Args[0])

func DefineBool(name string, defaultValue bool, usage string) *BoolFlag {
	return globalFlags.DefineBool(name, defaultValue, usage)
}

func DefineString(name string, defaultValue string, usage string) *StringFlag {
	return globalFlags.DefineString(name, defaultValue, usage)
}

type withEnvPrefix struct {
	prefix string
}

func (w *withEnvPrefix) applyGlobalParseOption(o *globalParseOptions) {
	o.envPrefix = w.prefix
}

func WithEnvPrefix(prefix string) globalParseOption {
	return &withEnvPrefix{prefix: prefix}
}

type withFallbackEnvPrefix struct {
	prefix string
}

func (w *withFallbackEnvPrefix) applyGlobalParseOption(o *globalParseOptions) {
	o.fallbackEnvPrefix = w.prefix
}

func WithFallbackEnvPrefix(prefix string) globalParseOption {
	return &withFallbackEnvPrefix{prefix: prefix}
}

func Parse(opts ...globalParseOption) ([]string, error) {
	o := &globalParseOptions{}
	for _, opt := range opts {
		opt.applyGlobalParseOption(o)
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if o.envPrefix != "" && strings.HasPrefix(parts[0], o.envPrefix) {
			pure := strings.TrimPrefix(parts[0], o.envPrefix)
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
		} else if o.fallbackEnvPrefix != "" && strings.HasPrefix(parts[0], o.fallbackEnvPrefix) {
			pure := strings.TrimPrefix(parts[0], o.fallbackEnvPrefix)
			name := strings.ToLower(strings.TrimSpace(pure))
			value := strings.TrimSpace(parts[1])
			item, ok := globalFlags.flags[name]
			if !ok {
				continue
			}
			seedlog.Infof("Loaded fallback env flag: %s=%s", parts[0], value)
			err := item.Set(value)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		}
	}

	finalArgs, err := globalFlags.Parse(os.Args[1:], o.parseOptions...)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Infof("Global flags: %+v", globalFlags.flags)
	return finalArgs, nil
}

// Finalize must be called after subcommand parsing has consumed everything it
// recognizes from the args returned by a Parse(WithUnknownFlag()) call. Any
// remaining token that still looks like a flag is by definition unclaimed by
// either the global or the subcommand layer; Finalize prints usage and exits
// non-zero in that case.
func Finalize(args []string) {
	for _, arg := range args {
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
