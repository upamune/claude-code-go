package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/upamune/claude-code-go"
)

func main() {
	// Check if Claude CLI is available
	if !claude.IsClaudeAvailable() {
		log.Fatal("Claude CLI is not available. Please install it first.")
	}

	// Simple query with streaming
	prompt := "Write a simple hello world program in Go"
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	ctx := context.Background()

	// Create options (optional)
	opts := &claude.Options{
		Model: "claude-3-5-sonnet-20241022", // Use latest model
	}

	fmt.Println("Querying Claude with streaming...")
	fmt.Println("Prompt:", prompt)
	fmt.Println("---")

	// Execute query with streaming
	stream, err := claude.QueryStream(ctx, prompt, opts)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}
	defer stream.Close()

	// Process messages from channel
	for msgOrErr := range stream.Messages {
		if msgOrErr.Err != nil {
			log.Fatalf("Stream error: %v", msgOrErr.Err)
		}

		msg := msgOrErr.Message
		switch m := msg.(type) {
		case *claude.UserMessage:
			fmt.Println("[USER]", m.Message)
		case *claude.AssistantMessage:
			fmt.Print(m.Message)
		case *claude.ResultMessage:
			fmt.Println("\n---")
			fmt.Println("[RESULT]")
			fmt.Println("Status:", m.Status)
			if m.Error != nil {
				fmt.Println("Error:", *m.Error)
			}
		case *claude.SystemMessage:
			// System messages are typically verbose, you might want to skip them
			// fmt.Println("[SYSTEM]", m.Message)
		case *claude.PermissionRequestMessage:
			fmt.Printf("\n[PERMISSION] %s (tool: %s)\n", m.Message, m.Tool)
			// In a real application, you would handle permission requests
		}
	}

	fmt.Println("\nStreaming completed!")
}
