package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type FileResult struct {
	Path string
	Hash string
	Err  error
}

func main() {
	// ---- Timer start ----
	start := time.Now()

	// ---- CLI FLAGS ----
	pathFlag := flag.String("path", "./testdata", "Root directory to scan")
	workersFlag := flag.Int("workers", 4, "Number of worker goroutines")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Concurrent File Integrity Checker")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  filechecker --path <directory> --workers <num>")
		fmt.Println("")
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	jobs := make(chan string)
	results := make(chan FileResult)

	var wg sync.WaitGroup

	/// ---- WORKER POOL ----
	for i := 0; i < *workersFlag; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	// Walk files and send jobs
	go func() {
		filepath.Walk(*pathFlag, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				jobs <- path
			}
			return nil
		})
		close(jobs)
	}()

	// ---- CLOSE RESULTS AFTER WORKERS FINISH ----
	go func() {
		wg.Wait()
		close(results)
	}()

	// ---- COLLECT RESULTS ----
	var all []FileResult
	for r := range results {
		all = append(all, r)
	}

	// ---- SORT OUTPUT ----
	sort.Slice(all, func(i, j int) bool {
		return all[i].Path < all[j].Path
	})

	// ---- PRINT ----
	for _, r := range all {
		if r.Err != nil {
			fmt.Printf("ERR: %s (%v)\n", r.Path, r.Err)
			continue
		}
		fmt.Printf("%s  %s\n", r.Hash, r.Path)
	}

	// ---- Timer end ----
	fmt.Printf("\nCompleted in %v\n", time.Since(start))
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
