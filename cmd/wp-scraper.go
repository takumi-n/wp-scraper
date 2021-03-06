package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/takumi-n/wp-scraper"
	"os"
)

const (
	returnCodeFail = 1
)

var (
	testOpt    = flag.Bool("t", false, "Enable test mode")
	limitOpt   = flag.Int("l", -1, "Acquire up to this limit articles")
	verboseOpt = flag.Bool("v", false, "Make the operation more talkative")
)

func main() {
	flag.Parse()
	configFile := flag.Arg(0)
	if configFile == "" {
		exitWithError(errors.New("specify config file"))
	}
	if *verboseOpt {
		fmt.Printf("loading config file %v ...\n", configFile)
	}

	config, err := scraper.ReadConfig(configFile)
	if err != nil {
		exitWithError(err)
	}

	// If test mode is enabled, verbose mode will be enabled automatically
	*verboseOpt = *verboseOpt || *testOpt

	scraper := scraper.NewScraper(*config, *verboseOpt)
	err = scraper.Scrape(*limitOpt)

	if err != nil {
		exitWithError(err)
	}

	if *testOpt {
		fmt.Println("Quit because test mode is enabled")
		return
	}

	url, err := scraper.SendToServer()

	if err != nil {
		exitWithError(err)
	}

	fmt.Println("Successfully created demo site: ")
	fmt.Println(url)
}

func exitWithError(err error) {
	fmt.Println(err)
	os.Exit(returnCodeFail)
}
