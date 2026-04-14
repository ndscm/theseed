// Package seedshell provides helpers for running shell commands with support
// for dry-run mode and shell-eval mode.
//
// Commands are split into "pure" and "impure" variants. Pure commands have no
// external side effects (e.g. read-only queries) and always execute. Impure
// commands may mutate external state and are skipped when dry mode is enabled.
//
// When shell-eval mode is enabled, command output is captured and logged at
// debug level instead of being printed to stdout.
package seedshell

import (
	"os"
	"os/exec"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagDry = seedflag.DefineBool("dry", false, "make no external side effect")
var flagShellEval = seedflag.DefineBool("shell-eval", false, "only output shell command")

// Dry reports whether dry-run mode is enabled via the --dry flag.
func Dry() bool {
	return flagDry.Get()
}

// ShellEval reports whether shell-eval mode is enabled via the --shell-eval flag.
func ShellEval() bool {
	return flagShellEval.Get()
}

// PureRun executes a command that has no external side effects.
// It always runs regardless of dry mode. Stderr is forwarded to os.Stderr.
// In shell-eval mode, stdout is captured and logged; otherwise it is forwarded
// to os.Stdout.
func PureRun(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	if flagShellEval.Get() {
		outputBytes, err := cmd.Output()
		if err != nil {
			return seederr.Wrap(err)
		}
		seedlog.Debugf("Command: %v %v\n  output: %v", name, arg, string(outputBytes))
	} else {
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

// ImpureRun executes a command that may have external side effects.
// In dry mode the command is skipped and a log message is emitted instead.
// Stderr is forwarded to os.Stderr. In shell-eval mode, stdout is captured
// and logged; otherwise it is forwarded to os.Stdout.
func ImpureRun(name string, arg ...string) error {
	if flagDry.Get() {
		seedlog.Infof("Dry mode skip: %v %v", name, arg)
		return nil
	}
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	if flagShellEval.Get() {
		outputBytes, err := cmd.Output()
		if err != nil {
			return seederr.Wrap(err)
		}
		seedlog.Debugf("Command: %v %v\n  output: %v", name, arg, string(outputBytes))
	} else {
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

// PureOutput executes a command that has no external side effects and returns
// its stdout as a byte slice. It always runs regardless of dry mode. Stderr is
// forwarded to os.Stderr.
func PureOutput(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	outputBytes, err := cmd.Output()
	if err != nil {
		return outputBytes, seederr.Wrap(err)
	}
	return outputBytes, nil
}

// ImpureOutput executes a command that may have external side effects and
// returns its stdout as a byte slice. In dry mode the command is skipped and
// an empty byte slice is returned. Stderr is forwarded to os.Stderr.
func ImpureOutput(name string, arg ...string) ([]byte, error) {
	if flagDry.Get() {
		seedlog.Infof("Dry mode skip: %v %v", name, arg)
		return []byte{}, nil
	}
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	outputBytes, err := cmd.Output()
	if err != nil {
		return outputBytes, seederr.Wrap(err)
	}
	return outputBytes, nil
}
