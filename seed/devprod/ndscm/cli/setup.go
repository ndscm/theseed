package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func injectShrcLine(dry bool, shrcPath string, line string) error {
	shrcBytes, err := os.ReadFile(shrcPath)
	if err != nil && !os.IsNotExist(err) {
		return common.WrapTrace(err)
	}
	if os.IsNotExist(err) {
		if dry {
			log.Printf("Dry mode skip: create shrc file at %v", shrcPath)
			return nil
		}
		err := os.WriteFile(shrcPath, []byte(line+"\n"), 0666)
		if err != nil {
			return common.WrapTrace(err)
		}
		return nil
	}
	shrcLines := strings.Split(string(shrcBytes), "\n")
	for _, shrcLine := range shrcLines {
		if shrcLine == line {
			log.Printf("Already injected to %v", shrcPath)
			return nil
		}
	}
	if dry {
		log.Printf("Dry mode skip: update shrc file at %v", shrcPath)
		return nil
	}
	shrcFile, err := os.OpenFile(shrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return common.WrapTrace(err)
	}
	_, err = shrcFile.WriteString("\n" + line + "\n")
	if err != nil {
		return common.WrapTrace(err)
	}
	return nil
}

func NdSetup(args []string, ndConfig *common.NdConfig) error {
	if ndConfig.ShellEval {
		return common.WrapTrace(fmt.Errorf("nd-setup should not run with --shell-eval"))
	}
	if len(args) != 1 {
		return common.WrapTrace(fmt.Errorf("nd-setup usage: ndscm setup"))
	}
	injectionCommand := "eval \"$(ndscm --shell-eval shell)\""
	userHome, err := os.UserHomeDir()
	if err != nil {
		return common.WrapTrace(err)
	}
	err = injectShrcLine(ndConfig.Dry, filepath.Join(userHome, ".bashrc"), injectionCommand)
	if err != nil {
		return err
	}
	err = injectShrcLine(ndConfig.Dry, filepath.Join(userHome, ".zshrc"), injectionCommand)
	if err != nil {
		return err
	}
	return nil
}
