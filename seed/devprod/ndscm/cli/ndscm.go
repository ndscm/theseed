package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func parseFlags() ([]string, error) {
	remainArgs := os.Args
	finalArgs := []string{}
	for len(remainArgs) > 0 {
		finalArgs = append(finalArgs, remainArgs[0])
		err := flag.CommandLine.Parse(remainArgs[1:])
		if err != nil {
			return nil, common.WrapTrace(err)
		}
		remainArgs = flag.CommandLine.Args()
	}
	return finalArgs[1:], nil
}

func main() {
	args, err := parseFlags()
	if err != nil {
		log.Fatalf("\x1b[31mERROR: parse flags failed: %v\x1b[0m", err)
	}
	ndConfig, err := common.LoadConfig()
	if err != nil {
		log.Fatalf("\x1b[31mERROR: load config failed: %v\x1b[0m", err)
	}
	if len(args) < 1 {
		fmt.Printf("ndscm is not-distributed source code manager\n")
		return
	}
	switch args[0] {
	case "cut":
		err := NdCut(args, ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "dev":
		err := NdDev(args, ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "setup":
		err := NdSetup(args, ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "shell":
		err := NdShell(args, ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "sync":
		err := NdSync(args, ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	default:
		log.Fatalf("\x1b[31mERROR: Unknown command %v\x1b[0m", args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
}
