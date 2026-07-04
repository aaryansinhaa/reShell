package workflows

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reshell/pkg/config"
	"strings"
	"time"
)

// StepStatus represents the execution state of a single workflow step.
type StepStatus struct {
	Index    int
	Command  string
	Stdout   string
	Stderr   string
	Error    error
	Finished bool
}

// AddOrUpdate creates or updates a workflow.
func AddOrUpdate(name, description string, steps []config.WorkflowStep) error {
	cfg, err := config.LoadWorkflows()
	if err != nil {
		return err
	}

	found := false
	for i, wf := range cfg.Workflows {
		if wf.Name == name {
			cfg.Workflows[i].Description = description
			cfg.Workflows[i].Steps = steps
			found = true
			break
		}
	}

	if !found {
		cfg.Workflows = append(cfg.Workflows, config.Workflow{
			Name:        name,
			Description: description,
			Steps:       steps,
		})
	}

	return config.SaveWorkflows(cfg)
}

// Remove deletes a workflow.
func Remove(name string) error {
	cfg, err := config.LoadWorkflows()
	if err != nil {
		return err
	}

	newWfs := make([]config.Workflow, 0, len(cfg.Workflows))
	found := false
	for _, wf := range cfg.Workflows {
		if wf.Name == name {
			found = true
			continue
		}
		newWfs = append(newWfs, wf)
	}

	if !found {
		return errors.New("workflow not found")
	}

	cfg.Workflows = newWfs
	return config.SaveWorkflows(cfg)
}

// Get finds a workflow by name.
func Get(name string) (*config.Workflow, error) {
	cfg, err := config.LoadWorkflows()
	if err != nil {
		return nil, err
	}

	for _, wf := range cfg.Workflows {
		if wf.Name == name {
			return &wf, nil
		}
	}

	return nil, fmt.Errorf("workflow '%s' not found", name)
}

// Run executes a workflow step-by-step, sending updates down statusChan.
// It is non-blocking when run in a goroutine.
func Run(wf *config.Workflow, statusChan chan<- StepStatus) {
	defer close(statusChan)

	var logBuf bytes.Buffer
	logBuf.WriteString(fmt.Sprintf("Workflow Execution: %s\n", wf.Name))
	logBuf.WriteString(fmt.Sprintf("Start Time: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	for i, step := range wf.Steps {
		statusChan <- StepStatus{
			Index:    i,
			Command:  step.Command,
			Finished: false,
		}

		// Set up execution command
		shellName := "bash"
		if shellEnv := os.Getenv("SHELL"); shellEnv != "" {
			if strings.Contains(shellEnv, "zsh") {
				shellName = "zsh"
			} else if strings.Contains(shellEnv, "fish") {
				shellName = "fish"
			}
		}
		cmd := exec.Command(shellName, "-c", step.Command)
		if step.Dir != "" {
			expandedDir := os.ExpandEnv(step.Dir)
			if len(expandedDir) > 0 && expandedDir[0] == '~' {
				home, _ := os.UserHomeDir()
				expandedDir = filepath.Join(home, expandedDir[1:])
			}
			cmd.Dir = expandedDir
			logBuf.WriteString(fmt.Sprintf("[%d] Execute: %s (in %s)\n", i, step.Command, expandedDir))
		} else {
			logBuf.WriteString(fmt.Sprintf("[%d] Execute: %s\n", i, step.Command))
		}

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		logBuf.WriteString(fmt.Sprintf("[%d] Status: ", i))
		if err != nil {
			logBuf.WriteString(fmt.Sprintf("FAILED (%s)\n", err.Error()))
		} else {
			logBuf.WriteString("SUCCESS\n")
		}
		logBuf.WriteString(fmt.Sprintf("[%d] STDOUT:\n%s\n", i, stdout.String()))
		logBuf.WriteString(fmt.Sprintf("[%d] STDERR:\n%s\n", i, stderr.String()))
		logBuf.WriteString("----------------------------------------\n")

		statusChan <- StepStatus{
			Index:    i,
			Command:  step.Command,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Error:    err,
			Finished: true,
		}

		// Halt execution of subsequent steps on failure
		if err != nil {
			break
		}
	}

	// Write log file
	configDir, err := config.GetConfigDir()
	if err == nil {
		now := time.Now().Format("20060102_150405")
		logFilename := fmt.Sprintf("wf_%s_%s.log", wf.Name, now)
		logPath := filepath.Join(configDir, "logs", "workflows", logFilename)
		_ = os.WriteFile(logPath, logBuf.Bytes(), 0644)
	}
}

// GetLogs returns execution log files for workflows.
func GetLogs() ([]string, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	logsDir := filepath.Join(configDir, "logs", "workflows")
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
