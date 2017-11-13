package main

import (
    "flag"
    "os"
    "errors"
    "fmt"
)

const (
    returnCodeFail = 1
)

var (
    testOpt  = flag.Bool("b", false, "Enable test mode")
    limitOpt = flag.Int("l", -1, "Acquire up to this limit articles")
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

    config, err := ReadConfig(configFile)
    if err != nil {
        exitWithError(err)
    }

    scraper := NewScraper(*config, *verboseOpt)
    err = scraper.Scrape(*limitOpt)

    if err != nil {
        exitWithError(err)
    }

    if *testOpt {
        fmt.Println("Quit because test mode is enabled")
        return
    }

    scraper.SendToServer()
}

func exitWithError(err error) {
    fmt.Println(err)
    os.Exit(returnCodeFail)
}
