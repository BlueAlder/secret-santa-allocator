package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
)

// FlagVars
var (
	configFileName string
	outputFile     string
	outputFormat   string
)

// Flags

var (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
)

func validateFlags() error {
	if configFileName == "" {
		return errors.New("must provide a allocation configuration file")
	}

	// We are writing out to a file
	if outputFile != "" {
		if outputFormat != "json" && outputFormat != "yaml" {
			return fmt.Errorf("invalid outputFormat, got: %s", outputFormat)
		}
	} else {
		fmt.Println("no outputFile flag detected. will instead print the results to stdout")
	}

	return nil
}

// logFatal prints a message in red and then runs os.Exit(1)
func logFatal(s string, a ...any) {
	fmt.Printf(colorRed+s, a...)
	os.Exit(1)
}

func main() {
	printBanner()

	flag.StringVar(&configFileName, "configFile", "", "name of the yaml file which contains the allocation configuration")
	flag.StringVar(&configFileName, "c", "", "name of the yaml file which contains the allocation configuration (shorthand)")

	flag.StringVar(&outputFile, "outputFile", "", "file to write allocation to")
	flag.StringVar(&outputFile, "o", "", "file to write allocation to (shorthand)")

	flag.StringVar(&outputFormat, "outputFormat", "json", "format to write output file to ")
	flag.StringVar(&outputFormat, "f", "json", "format to write output file to (shorthand)")

	flag.Parse()

	err := validateFlags()

	if err != nil {
		flag.Usage()
		logFatal("Got error while parsing flags: %v\n", err)
	}

	configFile, err := os.ReadFile(configFileName)
	if err != nil {
		logFatal("unable to read config yaml file: %v\n", err)
	}

	c, err := allocator.LoadConfigFromYaml(configFile)
	if err != nil {
		logFatal("error while loading config file: %v\n", err)
	}
	fmt.Printf("loaded config file %s\n", configFileName)

	a, err := allocator.New(c)
	if err != nil {
		logFatal("error while creating allocator: %v\n", err)
	}

	fmt.Println("trying to find a suitable allocation...")
	alcc, err := a.Allocate()
	if err != nil {
		logFatal("unable to allocate names, got error: %v\n", err)
	}

	if outputFile != "" {
		fmt.Printf("printing resulting allocation to file: %s in %s format\n", outputFile, outputFormat)
		err := alcc.OutputToFile(outputFile, outputFormat)
		if err != nil {
			fmt.Printf("there was an error writing to the file got: %v", err)
		}
		fmt.Printf(colorGreen+"successfully wrote to %s!\n", outputFile)
	} else {
		fmt.Println()
		alcc.PrintAliases()
		fmt.Println()
		alcc.PrintNameToName()
		fmt.Println()
		alcc.PrintNameToPassword()
	}

}
