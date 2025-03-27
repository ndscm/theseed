package main

import (
	"flag"
	"log"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func main() {
	flag.Parse()
	ndConfig, err := common.LoadConfig()
	if err != nil {
		log.Fatalf("\x1b[31mERROR: load config failed: %v\x1b[0m", err)
	}
	switch flag.Arg(0) {
	case "cut":
		err := NdCut(flag.Args(), ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "dev":
		err := NdDev(flag.Args(), ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "shell":
		err := NdShell(flag.Args(), ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	case "sync":
		err := NdSync(flag.Args(), ndConfig)
		if err != nil {
			log.Fatalf("\x1b[31mERROR: %v\x1b[0m", err)
		}
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}
