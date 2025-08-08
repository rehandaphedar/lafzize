package main

import (
	"fmt"
	"log"
	"os"

	"git.sr.ht/~rehandaphedar/lafzize/v3/pkg/api"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
	}

	command := os.Args[1]

	switch command {
	case "api":
		api.RunApiCommand(os.Args[2:])
	case "server":
		runServerCommand(os.Args[2:])
	default:
		printHelp()
	}
}

func printHelp() {
	log.Println("Invalid command")
	fmt.Println(`Usage: lafzize [subcommand] [flags]
Subcommands:
- api
- server`)
	os.Exit(1)
}
