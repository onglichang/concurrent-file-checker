package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type FileResult struct {
	Path string
	Hash string
	Err  error
}

func main() {
	root := "./testdata"

	jobs := make(chan string)
	results := make(chan FileResult)

	var wg sync.WaitGroup

	// Start workers
	workerCount := 4
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Walk files and send jobs
	go func() {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				fmt.Println("Hashing:", path)
				jobs <- path
			}
			return nil
		})
		close(jobs)
	}()

	// Close results channel after workers finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Consume results
	for res := range results {
		if res.Err != nil {
			fmt.Printf("ERR: %s (%v)\n", res.Path, res.Err)
			continue
		}
		fmt.Printf("%s  %s\n", res.Hash, res.Path)
	}
}

func worker(jobs <-chan string, results chan<- FileResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range jobs {
		hash, err := hashFile(path)
		results <- FileResult{Path: path, Hash: hash, Err: err}
	}
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
