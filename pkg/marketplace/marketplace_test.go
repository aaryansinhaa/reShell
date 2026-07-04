package marketplace

import (
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/config"
	"testing"
)

func runGit(t *testing.T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		t.Fatalf("git command failed: %v", err)
	}
}

func createMockRepository(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "reshell-mock-repo-*")
	if err != nil {
		t.Fatalf("failed to create temp mock repo: %v", err)
	}

	runGit(t, tempDir, "init")

	manifestContent := `[package]
name = "test-pack"
description = "A mock package for testing"

[[variables]]
name = "MOCK_VAR"
value = "mock_val"
description = "a mock variable"
enabled = true

[[aliases]]
name = "mockal"
value = "echo mock"
description = "a mock alias"
shell = "all"
enabled = true
`
	err = os.WriteFile(filepath.Join(tempDir, "reshell.toml"), []byte(manifestContent), 0644)
	if err != nil {
		t.Fatalf("failed to write reshell.toml: %v", err)
	}

	// Create functions directory
	err = os.MkdirAll(filepath.Join(tempDir, "functions"), 0755)
	if err != nil {
		t.Fatalf("failed to create functions dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "functions", "mockfunc.sh"), []byte("#!/bin/bash\necho func"), 0755)
	if err != nil {
		t.Fatalf("failed to write mockfunc.sh: %v", err)
	}

	// Create scripts directory
	err = os.MkdirAll(filepath.Join(tempDir, "scripts"), 0755)
	if err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "scripts", "mockscript.sh"), []byte("#!/bin/bash\necho script"), 0755)
	if err != nil {
		t.Fatalf("failed to write mockscript.sh: %v", err)
	}

	runGit(t, tempDir, "config", "user.name", "test")
	runGit(t, tempDir, "config", "user.email", "test@test.com")
	runGit(t, tempDir, "add", "-A")
	runGit(t, tempDir, "commit", "-m", "initial commit")

	return tempDir
}

func TestMarketplaceFetchAndMerge(t *testing.T) {
	// Set up temporary environment homes
	tempHome, err := os.MkdirTemp("", "reshell-test-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create mock remote repo
	mockRepoPath := createMockRepository(t)
	defer os.RemoveAll(mockRepoPath)

	// 1. Fetch Manifest
	manifest, tempDir, err := FetchManifest(mockRepoPath)
	if err != nil {
		t.Fatalf("FetchManifest failed: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if manifest.Package.Name != "test-pack" {
		t.Errorf("expected package name 'test-pack', got %q", manifest.Package.Name)
	}
	if len(manifest.Variables) != 1 || manifest.Variables[0].Name != "MOCK_VAR" {
		t.Errorf("unexpected variables in manifest: %+v", manifest.Variables)
	}

	// 2. Merge Manifest
	err = MergeManifest(manifest, tempDir)
	if err != nil {
		t.Fatalf("MergeManifest failed: %v", err)
	}

	// Verify environment variables were written
	envCfg, err := config.LoadEnv()
	if err != nil {
		t.Fatalf("failed to load environment variables: %v", err)
	}
	if len(envCfg.Variables) != 1 || envCfg.Variables[0].Name != "MOCK_VAR" || envCfg.Variables[0].Value != "mock_val" {
		t.Errorf("merged variables mismatch: %+v", envCfg.Variables)
	}

	// Verify aliases were written
	aliasCfg, err := config.LoadAliases()
	if err != nil {
		t.Fatalf("failed to load aliases: %v", err)
	}
	if len(aliasCfg.Aliases) != 1 || aliasCfg.Aliases[0].Name != "mockal" || aliasCfg.Aliases[0].Value != "echo mock" {
		t.Errorf("merged aliases mismatch: %+v", aliasCfg.Aliases)
	}

	// Verify functions were copied
	funcDir, _ := config.GetFunctionsDir()
	if _, err := os.Stat(filepath.Join(funcDir, "mockfunc.sh")); os.IsNotExist(err) {
		t.Error("expected custom function mockfunc.sh to be merged")
	}

	// Verify scripts were copied
	scriptsDir, _ := config.GetScriptsDir()
	if _, err := os.Stat(filepath.Join(scriptsDir, "marketplace", "mockscript.sh")); os.IsNotExist(err) {
		t.Error("expected custom script marketplace/mockscript.sh to be merged")
	}
}
