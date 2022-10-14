package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
	"github.com/BlueAlder/secret-santa-allocator/pkg/utils"
	"golang.org/x/sync/errgroup"
)

var namesFile = flag.String("namesFile", "", "name of the file containing the list of names")
var passwordsFile = flag.String("passwordsFile", "", "name of the file containing the list of name")

func validateFlags() error {
	if *namesFile == "" {
		return errors.New("must provide a namelist filename")
	}

	if *passwordsFile == "" {
		return errors.New("must provide a passwords filename")
	}

	return nil
}

func loadNamesAndPasswords(ctx context.Context) ([]string, []string, error) {
	var names []string
	var passwords []string

	errs, _ := errgroup.WithContext(ctx)

	errs.Go(func() error { return utils.ReadFileIntoSlice(*namesFile, &names) })
	errs.Go(func() error { return utils.ReadFileIntoSlice(*passwordsFile, &passwords) })
	err := errs.Wait()

	return names, passwords, err
}

func main() {
	flag.Parse()
	err := validateFlags()

	if err != nil {
		flag.Usage()
		log.Fatalf("Got error while parsing flags:%v", err)
	}

	names, passwords, err := loadNamesAndPasswords(context.Background())
	if err != nil {
		log.Fatalf("Error while trying to load names and passwords from files: %v", err)
	}

	a := &allocator.Allocator{
		Names:     names,
		Passwords: passwords,
	}

	alcc, err := a.Allocate()
	if err != nil {
		log.Fatalf("unable to allocate names, got error: %v", err)
	}

	alcc.PrintNameToPassword()

}
