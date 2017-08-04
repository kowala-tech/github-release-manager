package main

import (
	"flag"
	"fmt"
	"os"
)

var requestedAsset = flag.String("asset", "", "Asset name to download from the release")
var requestedOutput = flag.String("output", "", "Path to write the downloaded asset")

func init() {
	flag.StringVar(requestedAsset, "a", "", "Asset name to download from the release")
	flag.StringVar(requestedOutput, "o", "", "Path to write the downloaded asset")
}

func main() {

	flag.Parse()

	if len(flag.Args()) < 1 {
		printSyntax()
	}

	app := newApplication(flag.Args()[0], *requestedAsset, *requestedOutput)

	if err := app.Fetch(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func printSyntax() {
	fmt.Printf("Syntax: %s [flags] <owner/repo>\n\nFlags:\n\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}
