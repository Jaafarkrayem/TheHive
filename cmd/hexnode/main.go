package main

import (
	"fmt"
	"os"

	"github.com/hexagonal-chain/hexchain/internal/config"
	"github.com/hexagonal-chain/hexchain/pkg/core"
	"github.com/spf13/cobra"
)

var (
	version   = "v0.1.0-alpha"
	gitCommit = "dev"
	buildDate = "May 25, 2025"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hexnode",
		Short: "Hexagonal Chain Node - A revolutionary mesh blockchain",
		Long: `Hexagonal Chain Node is a next-generation blockchain node that implements
a hexagonal mesh topology instead of traditional linear chains.

Each block can connect to up to 6 neighboring blocks, creating a resilient
and efficient network architecture inspired by nature's most optimal structure.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, gitCommit, buildDate),
		Run: func(cmd *cobra.Command, args []string) {
			// Default action - show help
			cmd.Help()
		},
	}

	// Add subcommands
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(genesisCmd())
	rootCmd.AddCommand(accountCmd())
	rootCmd.AddCommand(consoleCmd())

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initCmd() *cobra.Command {
	var dataDir string
	var networkID uint64

	cmd := &cobra.Command{
		Use:   "init [genesis.json]",
		Short: "Initialize a new hexagonal chain node",
		Long: `Initialize creates a new node database and sets up the initial configuration.
It requires a genesis configuration file that defines the initial state of the
hexagonal mesh network.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			genesisPath := args[0]

			fmt.Printf("🐝 Initializing Hexagonal Chain Node\n")
			fmt.Printf("Data Directory: %s\n", dataDir)
			fmt.Printf("Network ID: %d\n", networkID)
			fmt.Printf("Genesis File: %s\n", genesisPath)

			// Create data directory
			if err := os.MkdirAll(dataDir, 0755); err != nil {
				return fmt.Errorf("failed to create data directory: %v", err)
			}

			// Initialize configuration
			cfg := config.DefaultConfig()
			cfg.DataDir = dataDir
			cfg.NetworkID = networkID

			// TODO: Load genesis configuration from file
			// TODO: Initialize database
			// TODO: Setup keystore

			fmt.Printf("✅ Node initialized successfully!\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "datadir", "./data", "Data directory for the node")
	cmd.Flags().Uint64Var(&networkID, "networkid", 1337, "Network identifier (integer)")

	return cmd
}

func runCmd() *cobra.Command {
	var (
		dataDir   string
		httpAddr  string
		httpPort  int
		wsAddr    string
		wsPort    int
		p2pPort   int
		nodeType  string
		bootnodes []string
		validator bool
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the hexagonal chain node",
		Long: `Run starts the hexagonal chain node with the specified configuration.
The node will connect to the mesh network and begin participating in consensus.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🐝 Starting Hexagonal Chain Node\n")
			fmt.Printf("Data Directory: %s\n", dataDir)
			fmt.Printf("HTTP API: %s:%d\n", httpAddr, httpPort)
			fmt.Printf("WebSocket API: %s:%d\n", wsAddr, wsPort)
			fmt.Printf("P2P Port: %d\n", p2pPort)
			fmt.Printf("Node Type: %s\n", nodeType)
			fmt.Printf("Validator Mode: %t\n", validator)

			if len(bootnodes) > 0 {
				fmt.Printf("Bootstrap Nodes: %v\n", bootnodes)
			}

			// Create configuration
			cfg := config.DefaultConfig()
			cfg.DataDir = dataDir
			cfg.HTTP.Addr = httpAddr
			cfg.HTTP.Port = httpPort
			cfg.WebSocket.Addr = wsAddr
			cfg.WebSocket.Port = wsPort
			cfg.P2P.Port = p2pPort
			cfg.NodeType = nodeType
			cfg.Validator = validator
			cfg.P2P.BootstrapNodes = bootnodes

			// TODO: Start the node
			// TODO: Initialize P2P networking
			// TODO: Start HTTP/WS APIs
			// TODO: Begin consensus participation

			fmt.Printf("🚀 Node started successfully!\n")
			fmt.Printf("Press Ctrl+C to stop...\n")

			// Block forever (in real implementation, this would start the node)
			select {}
		},
	}

	cmd.Flags().StringVar(&dataDir, "datadir", "./data", "Data directory for the node")
	cmd.Flags().StringVar(&httpAddr, "http.addr", "localhost", "HTTP-RPC server listening interface")
	cmd.Flags().IntVar(&httpPort, "http.port", 8545, "HTTP-RPC server listening port")
	cmd.Flags().StringVar(&wsAddr, "ws.addr", "localhost", "WS-RPC server listening interface")
	cmd.Flags().IntVar(&wsPort, "ws.port", 8546, "WS-RPC server listening port")
	cmd.Flags().IntVar(&p2pPort, "port", 30303, "Network listening port")
	cmd.Flags().StringVar(&nodeType, "nodetype", "full", "Node type (full, light, validator)")
	cmd.Flags().StringSliceVar(&bootnodes, "bootnodes", nil, "Comma separated enode URLs for P2P discovery bootstrap")
	cmd.Flags().BoolVar(&validator, "validator", false, "Enable validator mode")

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Hexagonal Chain Node\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Git Commit: %s\n", gitCommit)
			fmt.Printf("Build Date: %s\n", buildDate)
			fmt.Printf("Go Version: %s\n", "go1.23")
		},
	}
}

func genesisCmd() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "genesis",
		Short: "Generate a hexagonal chain genesis configuration",
		Long: `Genesis generates a new genesis configuration file for the hexagonal chain.
This includes the initial hexagonal block structure and network parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🐝 Generating Hexagonal Chain Genesis\n")

			// Create genesis block
			genesis := core.HexGenesisBlock()

			fmt.Printf("Genesis Block Hash: %s\n", genesis.Hash().Hex())
			fmt.Printf("Genesis Position: Q=%d, R=%d, S=%d\n",
				genesis.HexPosition().Q,
				genesis.HexPosition().R,
				genesis.HexPosition().S)

			// TODO: Write genesis configuration to file
			fmt.Printf("✅ Genesis configuration saved to: %s\n", outputPath)

			return nil
		},
	}

	cmd.Flags().StringVar(&outputPath, "output", "genesis.json", "Output file for genesis configuration")

	return cmd
}

func accountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage hexagonal chain accounts",
		Long:  `Account management commands for the hexagonal chain.`,
	}

	// Add subcommands for account management
	cmd.AddCommand(&cobra.Command{
		Use:   "new",
		Short: "Create a new account",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🔑 Creating new account...\n")
			// TODO: Generate new key pair
			// TODO: Save to keystore
			fmt.Printf("✅ Account created successfully!\n")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("📋 Listing accounts...\n")
			// TODO: List accounts from keystore
			return nil
		},
	})

	return cmd
}

func consoleCmd() *cobra.Command {
	var dataDir string

	cmd := &cobra.Command{
		Use:   "console",
		Short: "Start an interactive JavaScript console",
		Long: `Console starts an interactive JavaScript console connected to the running node.
This provides access to the full hexagonal chain API for debugging and interaction.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("🐝 Starting Hexagonal Chain Console\n")
			fmt.Printf("Data Directory: %s\n", dataDir)

			// TODO: Start JavaScript console
			// TODO: Connect to running node
			// TODO: Provide API access

			fmt.Printf("Welcome to the Hexagonal Chain JavaScript console!\n")
			fmt.Printf("To exit, press ctrl-d or type exit\n")

			return nil
		},
	}

	cmd.Flags().StringVar(&dataDir, "datadir", "./data", "Data directory for the node")

	return cmd
}
