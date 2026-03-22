package examples

import (
	"context"
	"fmt"
	"time"
)

// ExampleControlledClient demonstrates the ControlledClient wrapper.
// It shows how to safely execute commands in a concurrent context.
func ExampleControlledClient() error {
	client, err := NewControlledClient("localhost:7778")
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Authenticate
	ok, err := client.Authenticate("your_password")
	if err != nil || !ok {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Execute a command
	response, err := client.Execute("status")
	if err != nil {
		return fmt.Errorf("execute failed: %w", err)
	}
	fmt.Printf("Status: %s\n", response)

	return nil
}

// ExampleConnectionPool demonstrates the ConnectionPool for managing multiple connections.
// This is useful when you have many concurrent commands that don't share a single connection.
func ExampleConnectionPool() error {
	pool := NewConnectionPool("localhost:7778", "your_password", 5, time.Minute)
	defer pool.Close()

	ctx := context.Background()

	// Execute a command using a pooled connection
	err := pool.WithClient(ctx, func(client *ControlledClient) error {
		response, err := client.Execute("playerlist")
		if err != nil {
			return err
		}
		fmt.Printf("Players: %s\n", response)
		return nil
	})
	if err != nil {
		return fmt.Errorf("pool operation failed: %w", err)
	}

	return nil
}

// ExampleEventListener demonstrates streaming server events.
// This is useful for listening to server broadcasts like player logins, chat, etc.
func ExampleEventListener() error {
	listener, err := NewEventListener("localhost:7778", "your_password")
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	listener.Run(ctx)

	// Listen for events
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Listener stopped")
			return nil
		case event := <-listener.Events:
			fmt.Printf("Event: %s\n", event)
		}
	}
}
