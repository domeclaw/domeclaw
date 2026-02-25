// DomeClaw - Ultra-lightweight personal AI agent
// License: MIT

package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sipeed/domeclaw/pkg/config"
)

//go:generate cp -r ../../workspace .
//go:embed workspace
var embeddedFiles embed.FS

func onboard() {
	configPath := getConfigPath()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s\n", configPath)
		fmt.Println("Use existing config. Run 'domeclaw onboard' again with 'y' to overwrite.")
		fmt.Println("")
		// Load existing config to get workspace path
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Printf("Warning: Could not load config: %v\n", err)
			cfg = config.DefaultConfig()
		}
		workspace := cfg.WorkspacePath()
		createWorkspaceTemplates(workspace)
	} else {
		// Read config.example.json from the project's config folder
		exampleConfigPath := "config/config.example.json"
		exampleData, err := os.ReadFile(exampleConfigPath)
		if err != nil {
			fmt.Printf("Warning: Could not read %s: %v\n", exampleConfigPath, err)
			fmt.Println("Using default config instead...")
			cfg := config.DefaultConfig()
			if err := config.SaveConfig(configPath, cfg); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created config at %s\n", configPath)
			workspace := cfg.WorkspacePath()
			createWorkspaceTemplates(workspace)
		} else {
			// Copy config.example.json to config.json
			if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
				fmt.Printf("Error creating config directory: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(configPath, exampleData, 0o600); err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Copied config.example.json to %s\n", configPath)

			// Load config to get workspace path
			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				fmt.Printf("Warning: Could not load config: %v\n", err)
				cfg = config.DefaultConfig()
			}
			workspace := cfg.WorkspacePath()
			createWorkspaceTemplates(workspace)
		}
	}

	fmt.Printf("%s domeclaw is ready!\n", logo)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add your API key to", configPath)
	fmt.Println("")
	fmt.Println("     Recommended:")
	fmt.Println("     - Qwen: https://qwen.ai/")
	fmt.Println("     - OpenRouter: https://openrouter.ai/keys (access 100+ models)")
	fmt.Println("     - Ollama:     https://ollama.com (local, free)")
	fmt.Println("")
	fmt.Println("     See README.md for 17+ supported providers.")
	fmt.Println("")
	fmt.Println("  2. Chat: domeclaw agent -m \"Hello!\"")
}

func copyEmbeddedToTarget(targetDir string) error {
	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("Failed to create target directory: %w", err)
	}

	// Walk through all files in embed.FS
	err := fs.WalkDir(embeddedFiles, "workspace", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Read embedded file
		data, err := embeddedFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("Failed to read embedded file %s: %w", path, err)
		}

		new_path, err := filepath.Rel("workspace", path)
		if err != nil {
			return fmt.Errorf("Failed to get relative path for %s: %v\n", path, err)
		}

		// Build target file path
		targetPath := filepath.Join(targetDir, new_path)

		// Ensure target file's directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return fmt.Errorf("Failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}

		// Write file
		if err := os.WriteFile(targetPath, data, 0o644); err != nil {
			return fmt.Errorf("Failed to write file %s: %w", targetPath, err)
		}

		return nil
	})

	return err
}

func createWorkspaceTemplates(workspace string) {
	err := copyEmbeddedToTarget(workspace)
	if err != nil {
		fmt.Printf("Error copying workspace templates: %v\n", err)
	}
}
