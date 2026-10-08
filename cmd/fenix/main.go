package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/kernel"
	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit()
	case "execute":
		handleExecute()
	case "status":
		handleStatus()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Fénix AI SDLC - Command Line Interface")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  fenix <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init     Initialize the Fénix Kernel")
	fmt.Println("  execute  Execute a task")
	fmt.Println("  status   Show kernel status")
	fmt.Println("  help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  fenix init")
	fmt.Println("  fenix execute")
	fmt.Println("  fenix status")
}

func handleInit() {
	fmt.Println("Initializing Fénix Kernel...")

	k := kernel.NewKernel()

	ctx := context.Background()
	if err := k.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize kernel: %v", err)
	}

	fmt.Println("✓ Kernel initialized successfully")
	fmt.Printf("✓ Current state: %s\n", k.State())

	// Register state transition callback
	k.OnTransition(func(from, to contracts.KernelState) {
		fmt.Printf("  State transition: %s → %s\n", from, to)
	})

	// Shutdown
	if err := k.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown kernel: %v", err)
	}

	fmt.Println("✓ Kernel shutdown complete")
}

func handleExecute() {
	fmt.Println("Executing sample task...")

	k := kernel.NewKernel()

	ctx := context.Background()
	if err := k.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize kernel: %v", err)
	}
	defer k.Shutdown(ctx)

	// Create a sample execution package
	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID:            "pkg-001",
			Version:       "v1",
			SchemaVersion: "1.0",
			KernelVersion: "0.1.0",
			TraceID:       "trace-123",
			CorrelationID: "corr-456",
		},
		Task: contracts.TaskDescriptor{
			ID:   "task-001",
			Type: "GenerateCode",
		},
		Executor: contracts.ExecutorDescriptor{
			ID:           "executor-001",
			Type:         "Agent",
			Version:      "1.0",
			Capabilities: []string{"code_generation", "architecture"},
		},
		Snapshot: contracts.ExecutionSnapshot{
			AstNodes:         []string{"node-1", "node-2"},
			MemoryReferences: []string{"ref-1", "ref-2"},
		},
		Policies: contracts.ExecutionPolicies{
			MaxTokens:           100000,
			MaxFiles:            100,
			MaxAstDepth:         50,
			CompressionStrategy: "gzip",
			RetryPolicy: contracts.RetryConfig{
				MaxRetries:      3,
				BackoffStrategy: "exponential",
			},
		},
		Resources: contracts.ResourceAllocation{
			Provider:        "claude-3.5-sonnet",
			EstimatedTokens: 50000,
			ReservedTokens:  60000,
			EstimatedCost:   0.15,
			Priority:        1,
			Deadline:        time.Now().Add(30 * time.Minute),
		},
		Contracts: contracts.Contracts{
			ExpectedFormat:  "json",
			JsonSchema:      "{}",
			ValidationRules: []string{"required_fields", "type_check"},
		},
	}

	// Execute the package
	handle, err := k.Execute(ctx, pkg)
	if err != nil {
		log.Fatalf("Failed to execute: %v", err)
	}

	fmt.Println("✓ Task executed successfully")
	fmt.Printf("  Package ID: %s\n", handle.PackageID)
	fmt.Printf("  Version: %s\n", handle.CurrentVersion)
	fmt.Printf("  Attempt: %d\n", handle.Attempt)
	fmt.Printf("  Status: %s\n", handle.Status)
	fmt.Printf("  Phase: %s\n", handle.CurrentPhase)
	fmt.Printf("  Duration: %v\n", handle.Duration)
}

func handleStatus() {
	fmt.Println("Fénix Kernel Status")

	k := kernel.NewKernel()

	ctx := context.Background()
	if err := k.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize kernel: %v", err)
	}
	defer k.Shutdown(ctx)

	fmt.Printf("  State: %s\n", k.State())
	fmt.Printf("  Active Handles: %d\n", len(k.ListHandles()))
}
