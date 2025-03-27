package common

import (
	"log"
	"os"
	"os/exec"
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
			return WrapTrace(err)
		}
		log.Printf("Command output: %v", string(outputBytes))
	} else {
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return WrapTrace(err)
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
		return outputBytes, WrapTrace(err)
	}
	return outputBytes, nil
}
