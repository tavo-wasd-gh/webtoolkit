package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/tavo-wasd-gh/webtoolkit/session"
)

const (
	numSessions     = 100_000 // Make sure to not exceed this in validation loop
	sessionLifetime = 10 * time.Minute
)

func printMem(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("[%s] HeapAlloc: %.2f MB, TotalAlloc: %.2f MB, NumGC: %d\n",
		label,
		float64(m.HeapAlloc)/(1024*1024),
		float64(m.TotalAlloc)/(1024*1024),
		m.NumGC,
	)
}

func main() {
	// Start cleanup goroutine (optional)
	session.StartCleanup(1 * time.Minute)

	fmt.Printf("Creating %d sessions...\n", numSessions)
	printMem("BEFORE")

	start := time.Now()

	sessionTokens := make([]string, 0, numSessions)
	csrfTokens := make([]string, 0, numSessions)

	for i := 0; i < numSessions; i++ {
		st, ct, err := session.New(sessionLifetime, fmt.Sprintf("data-%d", i))
		if err != nil {
			panic(err)
		}
		sessionTokens = append(sessionTokens, st)
		csrfTokens = append(csrfTokens, ct)
	}

	fmt.Printf("Created %d sessions in %s\n", numSessions, time.Since(start))
	printMem("AFTER CREATE")

	// Validate sessions multiple times to test speed
	validateCount := 1000
	start = time.Now()
	for i := 0; i < validateCount; i++ {
		idx := i % numSessions
		st := sessionTokens[idx]
		ct := csrfTokens[idx]
		newSt, newCt, _, err := session.Validate(st, ct)
		if err != nil {
			fmt.Printf("validation error at iteration %d: %v\n", i, err)
		} else {
			// update tokens to use rotated tokens for next iteration
			sessionTokens[idx] = newSt
			csrfTokens[idx] = newCt
		}
	}
	fmt.Printf("Validated %d sessions in %s\n", validateCount, time.Since(start))
	printMem("AFTER VALIDATE")

	// Wait for cleanup to run
	fmt.Println("Sleeping 2s to allow cleanup goroutine to run...")
	time.Sleep(2 * time.Second)
	printMem("FINAL")
}
