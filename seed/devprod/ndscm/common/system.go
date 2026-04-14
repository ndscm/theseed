package common

import (
	"log"
	"os"
	"os/exec"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func ShellRun(dry bool, shellEval bool, name string, arg ...string) error {
	if dry {
		log.Printf("Dry mode skip: %v, %v", name, arg)
		return nil
	}
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	if shellEval {
		outputBytes, err := cmd.Output()
		if err != nil {
			return seederr.Wrap(err)
		}
		log.Printf("Command output: %v", string(outputBytes))
	} else {
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

func ShellOutput(dry bool, name string, arg ...string) ([]byte, error) {
	if dry {
		log.Printf("Dry mode skip: %v, %v", name, arg)
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
