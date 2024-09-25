package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	switch os.Args[1] {
	case "server":
		runServer()
	case "fetch":
		fetchVerseText()
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println(`Usage:
- lafzize server [port]
- lafzize fetch`)
}
