package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

type Lexer struct {
	tokens []string
}

func NewLexer(input string) *Lexer {
	var tokens []string
	for _, field := range strings.FieldsFunc(input, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	}) {
		if field != "" {
			tokens = append(tokens, strings.ToLower(field))
		}
	}

	return &Lexer{tokens: tokens}
}

func (lexer *Lexer) SingleThreadIndexer() map[string]int {
	m := make(map[string]int)

	for _, v := range lexer.tokens {
		currentValue, ok := m[v]

		if ok {
			m[v] = currentValue + 1
		} else {
			m[v] = 1
		}
	}

	return m
}

func (lexer *Lexer) MultiThreadIndexer(numChunks int) map[string]int {
	chunkSize := (len(lexer.tokens) + numChunks - 1) / numChunks
	results := make(chan map[string]int, chunkSize)
	var wg sync.WaitGroup

	for i := 0; i < len(lexer.tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(lexer.tokens) {
			end = len(lexer.tokens)
		}
		chunk := lexer.tokens[i:end]

		wg.Add(1)
		go func(chunk []string) {
			defer wg.Done()
			chunkResult := make(map[string]int)
			for _, token := range chunk {
				chunkResult[token]++
			}
			results <- chunkResult
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	finalResult := make(map[string]int)
	for result := range results {
		for token, count := range result {
			finalResult[token] += count
		}
	}

	return finalResult
}

func main() {
	data, err := os.ReadFile("test/alice.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fileContent := string(data)
	lexer := NewLexer(fileContent)

	start := time.Now()
	lexer.SingleThreadIndexer()
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("SingleThreadIndexer time taken: %v\n", duration)

	start2 := time.Now()
	lexer.MultiThreadIndexer(5)
	end2 := time.Now()
	duration2 := end2.Sub(start2)
	fmt.Printf("MultiThreadIndexer time taken with 5 threads: %v\n", duration2)

	start3 := time.Now()
	lexer.MultiThreadIndexer(10)
	end3 := time.Now()
	duration3 := end3.Sub(start3)
	fmt.Printf("MultiThreadIndexer time taken with 10 threads: %v\n", duration3)
}
