package utils

import (
	"bufio"
	"math/rand"
	"os"
	"time"
)

func ReadFileIntoSlice(filename string, sliceToLoad *[]string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		*sliceToLoad = append(*sliceToLoad, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Takes a slice and an index and returns a new slice
// with the element at s[index] removed
func RemoveIndex[T any](s []T, index int) []T {
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	ret = append(ret, s[index+1:]...)
	return ret
}

// Selects a random element from a slice and returns
// it and it's index
func RandomElement[T any](s []T) (T, int) {
	rand.Seed(time.Now().Unix())
	randomChoice := rand.Intn(len(s))
	return s[randomChoice], randomChoice
}
