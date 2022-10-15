package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
)

// var namesFile = flag.String("namesFile", "", "name of the file containing the list of names")
// var passwordsFile = flag.String("passwordsFile", "", "name of the file containing the list of name")
var configFileName = flag.String("configFile", "", "name of the yaml file which contains the allocation configuration")
var outputFile = flag.String("outputFile", "", "file to write allocation to")
var outputFormat = flag.String("ouputFormat", "json", "format to write output file to ")

type OutputFormat string

func validateFlags() error {
	if *configFileName == "" {
		return errors.New("must provide a allocation configuration file")
	}

	// We are writing out to a file
	if *outputFile != "" {
		if *outputFormat != "json" && *outputFile != "yaml" {
			return fmt.Errorf("invalid outputFormat, got: %s", *outputFormat)
		}
	}

	return nil
}

// func writeToFile()

func main() {
	printBanner()
	flag.Parse()
	err := validateFlags()

	if err != nil {
		flag.Usage()
		log.Fatalf("Got error while parsing flags: %v", err)
	}

	configFile, err := os.ReadFile(*configFileName)
	if err != nil {
		log.Fatalf("unable to read config yaml file: %v", err)
	}
	c, err := allocator.LoadConfigFromYaml(configFile)
	if err != nil {
		log.Fatalf("error while loading config file: %v", err)
	}
	log.Printf("loaded config file %s", *configFileName)

	a, err := allocator.New(c)
	if err != nil {
		log.Fatalf("error while creating allocator: %v", err)
	}

	log.Println("trying to find a suitable allocation...")
	alcc, err := a.Allocate()
	if err != nil {
		log.Fatalf("unable to allocate names, got error: %v", err)
	}

	alcc.PrintNameToName()
	fmt.Println()
	alcc.PrintNameToPassword()
	fmt.Println()
	alcc.PrintAliases()
	fmt.Println()

	// fmt.Println(alcc)

}
