package functions

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/config"
	"reshell/pkg/git"
	"strings"
)

// CreateOrUpdate writes the function script body to functions/ folder.
func CreateOrUpdate(name, code string) error {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		return err
	}

	// Always default to .sh for portability, unless it's fish specific
	filename := fmt.Sprintf("%s.sh", name)
	if strings.HasPrefix(strings.TrimSpace(code), "#!/usr/bin/env fish") || strings.HasPrefix(strings.TrimSpace(code), "#!/bin/fish") {
		filename = fmt.Sprintf("%s.fish", name)
	}

	path := filepath.Join(funcDir, filename)
	if err := os.WriteFile(path, []byte(code), 0755); err != nil {
		return err
	}

	_ = git.CommitWorkspace("Update custom function: " + name)
	return nil
}

// Get reads the function file code contents.
func Get(name string) (string, string, error) {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		return "", "", err
	}

	// Try .sh then .fish
	pathSh := filepath.Join(funcDir, name+".sh")
	data, err := os.ReadFile(pathSh)
	if err == nil {
		return string(data), ".sh", nil
	}

	pathFish := filepath.Join(funcDir, name+".fish")
	data, err = os.ReadFile(pathFish)
	if err == nil {
		return string(data), ".fish", nil
	}

	return "", "", os.ErrNotExist
}

// Remove deletes the function script file.
func Remove(name string) error {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		return err
	}

	pathSh := filepath.Join(funcDir, name+".sh")
	errSh := os.Remove(pathSh)

	pathFish := filepath.Join(funcDir, name+".fish")
	errFish := os.Remove(pathFish)

	if errSh != nil && errFish != nil {
		return fmt.Errorf("function '%s' not found", name)
	}

	_ = git.CommitWorkspace("Remove custom function: " + name)
	return nil
}

// Validate executes a dry-run check on the file to check shell syntax (e.g. bash -n).
func Validate(name string) (string, error) {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		return "", err
	}

	var path string
	var shellCmd string

	pathSh := filepath.Join(funcDir, name+".sh")
	if _, err := os.Stat(pathSh); err == nil {
		path = pathSh
		shellCmd = "bash"
	} else {
		pathFish := filepath.Join(funcDir, name+".fish")
		if _, err := os.Stat(pathFish); err == nil {
			path = pathFish
			shellCmd = "fish"
		} else {
			return "", fmt.Errorf("function script not found")
		}
	}

	cmd := exec.Command(shellCmd, "-n", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	errRun := cmd.Run()
	if errRun != nil {
		return stderr.String(), errRun
	}

	return "", nil
}

// List returns names of all registered custom functions.
func List() ([]string, error) {
	funcDir, err := config.GetFunctionsDir()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(funcDir)
	if err != nil {
		return nil, err
	}

	var funcs []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		ext := filepath.Ext(f.Name())
		if ext == ".sh" || ext == ".fish" {
			funcs = append(funcs, strings.TrimSuffix(f.Name(), ext))
		}
	}

	return funcs, nil
}
