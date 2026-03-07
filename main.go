package main

import (
	"os"

	"golang-blockchain/cli"
)

func main() {

	// Handle the exit signal
	// This will allow the program to exit gracefully
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()
}


