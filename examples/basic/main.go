// Package main demonstrates basic usage of the Claude Code Go SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	claude "github.com/upamune/claude-code-go"
)

func main() {
	// Get prompt from command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: basic <prompt>")
		fmt.Println("Example: basic \"Help me write a function to calculate fibonacci numbers\"")
		os.Exit(1)
	}

	prompt := os.Args[1]

	// Create context
	ctx := context.Background()

	// Configure options
	opts := &claude.Options{
		Model: "claude-3-5-sonnet-20241022",
		// Uncomment to use specific tools
		// AllowedTools: []string{"Read", "Write", "Bash"},
	}

	// Execute query
	fmt.Printf("Sending prompt: %s\n", prompt)
	fmt.Println("---")

	result, err := claude.Query(ctx, prompt, opts)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Display result
	if result.IsError {
		fmt.Printf("Error: %s\n", result.Result)
	} else {
		fmt.Printf("Result: %s\n", result.Result)
	}

	fmt.Printf("\nDuration: %dms\n", result.DurationMS)
	fmt.Printf("Cost: $%.4f\n", result.TotalCostUSD)
	fmt.Printf("Turns: %d\n", result.NumTurns)
	fmt.Printf("Session ID: %s\n", result.SessionID)

	// Example of streaming output
	fmt.Println("\n--- Streaming Example ---")
	err = claude.QueryStream(ctx, "Count from 1 to 5", opts, func(msg claude.Message) error {
		switch m := msg.(type) {
		case *claude.SystemMessage:
			fmt.Printf("[SYSTEM] %s\n", m.Subtype)
		case *claude.AssistantMessage:
			fmt.Printf("[ASSISTANT] Processing...\n")
		case *claude.ResultMessage:
			fmt.Printf("[RESULT] %s\n", m.Result)
		}
		return nil
	})
	if err != nil {
		log.Printf("Streaming error: %v", err)
	}
}
