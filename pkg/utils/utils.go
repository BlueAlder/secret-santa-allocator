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
// with the element at s[index] removed, also returns the removed element
func RemoveIndex[T any](s []T, index int) ([]T, T) {
	removed := s[index]
	ret := make([]T, 0)
	ret = append(ret, s[:index]...)
	ret = append(ret, s[index+1:]...)
	return ret, removed
}

// Selects a random element from a slice and returns
// it and it's index
func RandomElementFromSlice[T any](s []T) (T, int) {
	rand.Seed(time.Now().Unix())
	randomChoice := rand.Intn(len(s))
	return s[randomChoice], randomChoice
}

func RemoveDuplicatesFromSlice[T comparable](s []T) []T {
	set := make(map[T]bool)
	list := []T{}
	for _, item := range s {
		if _, exists := set[item]; !exists {
			set[item] = true
			list = append(list, item)
		}
	}
	return list
}

func RandomElementFromSet[T comparable](s map[T]struct{}) T {
	rand.Seed(time.Now().Unix())
	randomChoice := rand.Intn(len(s))
	for password := range s {
		if randomChoice == 0 {
			return password
		}
		randomChoice--
	}
	panic("unable to choose element from set")
}

func MapKeysToSlice[T comparable, K any](m map[T]K) []T {
	s := make([]T, len(m))
	idx := 0
	for key := range m {
		s[idx] = key
		idx++
	}
	return s
}
