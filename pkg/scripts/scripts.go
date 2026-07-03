package scripts

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"reshell/pkg/config"
	"reshell/pkg/git"
	"strings"
	"time"
)

// Script represents a stored script with its category and metadata.
type Script struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Path        string   `json:"path"`
	Parameters  []string `json:"parameters"`
	Description string   `json:"description"`
}

// CreateOrUpdate writes the script body under ~/.config/reshell/scripts/<category>/<name>.sh.
func CreateOrUpdate(category, name, code string) error {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		return err
	}

	catDir := filepath.Join(scriptsDir, category)
	if err := os.MkdirAll(catDir, 0755); err != nil {
		return err
	}

	path := filepath.Join(catDir, name+".sh")
	if err := os.WriteFile(path, []byte(code), 0755); err != nil {
		return err
	}

	_ = git.CommitWorkspace(fmt.Sprintf("Update script: %s/%s", category, name))
	return nil
}

// Get reads the content of a script.
func Get(category, name string) (string, error) {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(scriptsDir, category, name+".sh")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Remove deletes the script file and cleans up empty category folder.
func Remove(category, name string) error {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		return err
	}

	path := filepath.Join(scriptsDir, category, name+".sh")
	if err := os.Remove(path); err != nil {
		return err
	}

	catDir := filepath.Join(scriptsDir, category)
	files, err := os.ReadDir(catDir)
	if err == nil && len(files) == 0 {
		_ = os.Remove(catDir)
	}

	_ = git.CommitWorkspace(fmt.Sprintf("Remove script: %s/%s", category, name))
	return nil
}

// List scans scripts folder and groups them.
func List() ([]Script, error) {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var list []Script
	err = filepath.Walk(scriptsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".sh" {
			rel, _ := filepath.Rel(scriptsDir, path)
			dir, file := filepath.Split(rel)

			category := strings.TrimSuffix(filepath.ToSlash(dir), "/")
			if category == "" {
				category = "general"
			}

			name := strings.TrimSuffix(file, ".sh")
			codeBytes, _ := os.ReadFile(path)
			code := string(codeBytes)

			desc := ""
			// Simple extraction of description
			lines := strings.Split(code, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") && !strings.Contains(line, "@param") {
					desc = strings.TrimPrefix(line, "# ")
					break
				}
			}

			params := ParseParameters(code)

			list = append(list, Script{
				Name:        name,
				Category:    category,
				Path:        path,
				Parameters:  params,
				Description: desc,
			})
		}
		return nil
	})

	return list, err
}

// ParseParameters parses parameters from script code:
// 1. Detects positional parameters like $1, $2, $3...
// 2. Detects comment tags: # @param <Name>
func ParseParameters(code string) []string {
	// Look for # @param <Name>
	reParam := regexp.MustCompile(`(?m)^\s*#\s*@param\s+([a-zA-Z0-9_-]+)`)
	matches := reParam.FindAllStringSubmatch(code, -1)
	var params []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			p := match[1]
			if !seen[p] {
				seen[p] = true
				params = append(params, p)
			}
		}
	}

	// Also detect positional arguments ($1, $2) up to $9
	for i := 1; i <= 9; i++ {
		pat := fmt.Sprintf(`\$%d`, i)
		matched, _ := regexp.MatchString(pat, code)
		if matched {
			argName := fmt.Sprintf("Arg%d", i)
			if !seen[argName] {
				seen[argName] = true
				params = append(params, argName)
			}
		}
	}

	return params
}

// Execute runs the script using bash and saves output to a log file.
func Execute(category, name string, args []string) (string, string, int, error) {
	scriptsDir, err := config.GetScriptsDir()
	if err != nil {
		return "", "", -1, err
	}

	scriptPath := filepath.Join(scriptsDir, category, name+".sh")

	cmd := exec.Command("bash", append([]string{scriptPath}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	errRun := cmd.Run()
	exitCode := 0
	if errRun != nil {
		if exitError, ok := errRun.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = -1
		}
	}

	// Write execution logs
	configDir, err := config.GetConfigDir()
	if err == nil {
		now := time.Now().Format("20060102_150405")
		logFilename := fmt.Sprintf("%s_%s.log", name, now)
		logPath := filepath.Join(configDir, "logs", "scripts", logFilename)

		var logBuf bytes.Buffer
		logBuf.WriteString(fmt.Sprintf("Script: %s/%s\n", category, name))
		logBuf.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		logBuf.WriteString(fmt.Sprintf("Arguments: %v\n", args))
		logBuf.WriteString(fmt.Sprintf("Exit Code: %d\n", exitCode))
		if errRun != nil {
			logBuf.WriteString(fmt.Sprintf("Execution Error: %s\n", errRun.Error()))
		}
		logBuf.WriteString("\n--- STDOUT ---\n")
		logBuf.Write(stdout.Bytes())
		logBuf.WriteString("\n--- STDERR ---\n")
		logBuf.Write(stderr.Bytes())

		_ = os.WriteFile(logPath, logBuf.Bytes(), 0644)
	}

	return stdout.String(), stderr.String(), exitCode, errRun
}

// GetLogs returns execution log contents from log files.
func GetLogs() ([]string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	logsDir := filepath.Join(configDir, "logs", "scripts")
	files, err := os.ReadDir(logsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var logs []string
	for i := len(files) - 1; i >= 0; i-- { // return newest first
		f := files[i]
		if !f.IsDir() {
			logs = append(logs, filepath.Join(logsDir, f.Name()))
		}
	}
	return logs, nil
}
