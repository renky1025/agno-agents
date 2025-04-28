package main

import (
	"flag"
	"log"
	"os"

	"go-mcp-k8s/internal/config"
	"go-mcp-k8s/internal/k8s"
	"go-mcp-k8s/internal/tools"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Parse command line arguments
	kubeconfigPath := flag.String("kubeconfig", "", "Path to Kubernetes configuration file (uses default config if not specified)")
	enableCreate := flag.Bool("enable-create", false, "Enable resource creation operations")
	enableUpdate := flag.Bool("enable-update", false, "Enable resource update operations")
	enableDelete := flag.Bool("enable-delete", false, "Enable resource deletion operations")
	enableScale := flag.Bool("enable-scale", true, "Enable scale operations")
	transport := flag.String("transport", "stdio", "Transport to use (stdio, sse)")

	flag.Parse()

	// Create configuration
	cfg := config.NewConfig(*kubeconfigPath)
	if err := cfg.Validate(); err != nil {
		//fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
		os.Exit(1)
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(cfg.KubeconfigPath)
	if err != nil {
		//fmt.Fprintf(os.Stderr, "Failed to create Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"Kubernetes MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// Add basic tools
	//fmt.Println("Registering basic tools...")
	mcpServer.AddTool(tools.CreateGetAPIResourcesTool(), tools.HandleGetAPIResources(client))
	mcpServer.AddTool(tools.CreateGetResourceTool(), tools.HandleGetResource(client))
	mcpServer.AddTool(tools.CreateListResourcesTool(), tools.HandleListResources(client))
	mcpServer.AddTool(tools.CreateGetNamespacesTool(), tools.HandleGetNamespaces(client))
	mcpServer.AddTool(tools.CreateGetPodsTool(), tools.HandleGetPods(client))
	mcpServer.AddTool(tools.CreateGetDeploymentsTool(), tools.HandleGetDeployments(client))
	mcpServer.AddTool(tools.CreateGetServicesTool(), tools.HandleGetServices(client))
	mcpServer.AddTool(tools.CreateGetConfigMapsTool(), tools.HandleGetConfigMaps(client))
	mcpServer.AddTool(tools.CreateDescribePodTool(), tools.HandleDescribePod(client))
	mcpServer.AddTool(tools.CreateQueryLogsByPodNameTool(), tools.HandleQueryLogsByPodName(client))

	// Add write operation tools (if enabled)
	if *enableCreate {
		//fmt.Println("Registering resource creation tool...")
		mcpServer.AddTool(tools.CreateCreateResourceTool(), tools.HandleCreateResource(client))
	}

	if *enableUpdate {
		//fmt.Println("Registering resource update tool...")
		mcpServer.AddTool(tools.CreateUpdateResourceTool(), tools.HandleUpdateResource(client))
	}

	if *enableDelete {
		//fmt.Println("Registering resource deletion tool...")
		mcpServer.AddTool(tools.CreateDeleteResourceTool(), tools.HandleDeleteResource(client))
	}
	if *enableScale {
		//fmt.Println("Registering scale deployment tool...")
		mcpServer.AddTool(tools.CreateScaleDeploymentTool(), tools.HandleScaleDeployment(client))
	}

	// Output functionality status
	// fmt.Println("\nStarting Kubernetes MCP Server...")
	// fmt.Printf("Read operations: Enabled\n")
	// fmt.Printf("Create operations: %v\n", *enableCreate)
	// fmt.Printf("Update operations: %v\n", *enableUpdate)
	// fmt.Printf("Delete operations: %v\n", *enableDelete)
	// fmt.Printf("Scale operations: %v\n", *enableScale)

	// Start stdio server
	//fmt.Println("\nServer started, waiting for MCP client connections...\n")
	// if err := server.ServeStdio(s); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
	// 	os.Exit(1)
	// }
	if *transport == "sse" {
		sseServer := server.NewSSEServer(mcpServer, server.WithBaseURL("http://localhost:8080"))
		log.Printf("SSE server listening on :8080")
		if err := sseServer.Start(":8080"); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}
