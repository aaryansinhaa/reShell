package shell

import (
	"os"
	"path/filepath"
	"reshell/pkg/aliases"
	"reshell/pkg/env"
	"strings"
	"testing"
)

func TestShellCompiler(t *testing.T) {
	// Set up temporary environment homes
	tempHome, err := os.MkdirTemp("", "reshell-shell-test-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Set active shell to bash for testing
	oldShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "/bin/bash")
	defer os.Setenv("SHELL", oldShell)

	// Create a mock .bashrc profile
	bashrcPath := filepath.Join(tempHome, ".bashrc")
	initialBashrc := "# Some existing bashrc content\necho hello\n"
	err = os.WriteFile(bashrcPath, []byte(initialBashrc), 0644)
	if err != nil {
		t.Fatalf("failed to write mock .bashrc: %v", err)
	}

	// 1. Add env var and alias
	err = env.AddOrUpdate("COMPILER_VAR", "comp_val", "testing shell compiler", true)
	if err != nil {
		t.Fatalf("env.AddOrUpdate failed: %v", err)
	}
	err = aliases.AddOrUpdate("compal", "echo 123", "compiler alias", "all", true)
	if err != nil {
		t.Fatalf("aliases.AddOrUpdate failed: %v", err)
	}

	// 2. Compile and apply configurations
	err = Apply()
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// 3. Verify compiled script contents
	compiledPath := filepath.Join(tempHome, ".config", "reshell", "shell", "reshell.sh")
	data, err := os.ReadFile(compiledPath)
	if err != nil {
		t.Fatalf("failed to read compiled script: %v", err)
	}
	scriptContent := string(data)

	if !strings.Contains(scriptContent, `export COMPILER_VAR="comp_val"`) {
		t.Errorf("compiled script missing env var export, got: %s", scriptContent)
	}
	if !strings.Contains(scriptContent, `alias compal="echo 123"`) {
		t.Errorf("compiled script missing alias definition, got: %s", scriptContent)
	}

	// 4. Verify .bashrc was updated with integration hooks
	bashrcData, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("failed to read .bashrc: %v", err)
	}
	bashrcContent := string(bashrcData)

	if !strings.Contains(bashrcContent, StartMarker) || !strings.Contains(bashrcContent, EndMarker) {
		t.Errorf(".bashrc missing reshell initialization block, got: %s", bashrcContent)
	}
	if !strings.Contains(bashrcContent, compiledPath) {
		t.Errorf(".bashrc missing source path to compiled script, got: %s", bashrcContent)
	}

	// 5. Clean / remove integrations
	err = Clean()
	if err != nil {
		t.Fatalf("Clean failed: %v", err)
	}

	cleanedData, err := os.ReadFile(bashrcPath)
	if err != nil {
		t.Fatalf("failed to read .bashrc after clean: %v", err)
	}
	cleanedContent := string(cleanedData)

	if strings.Contains(cleanedContent, StartMarker) || strings.Contains(cleanedContent, EndMarker) {
		t.Errorf(".bashrc should not contain reshell initialization block after Clean()")
	}
	if strings.Contains(cleanedContent, "reshell initialize") {
		t.Errorf(".bashrc has remaining hooks after Clean(): %s", cleanedContent)
	}
}
