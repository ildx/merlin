package scripts

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ildx/merlin/internal/models"
)

// ScriptResult represents the outcome of a script execution
type ScriptResult struct {
	Script   string
	ExitCode int
	Duration time.Duration
	Success  bool
	Output   string
	Error    error
}

// ScriptRunner handles script execution
type ScriptRunner struct {
	ToolRoot    string
	Environment map[string]string
	DryRun      bool
	Verbose     bool
	Output      io.Writer
}

// NewScriptRunner creates a new script runner
func NewScriptRunner(toolRoot string, env map[string]string, dryRun, verbose bool, output io.Writer) *ScriptRunner {
	if output == nil {
		output = os.Stdout
	}

	return &ScriptRunner{
		ToolRoot:    toolRoot,
		Environment: env,
		DryRun:      dryRun,
		Verbose:     verbose,
		Output:      output,
	}
}

// RunScripts executes all scripts from a tool's configuration
func (r *ScriptRunner) RunScripts(config *models.ToolMerlinConfig) ([]*ScriptResult, error) {
	if !config.HasScripts() {
		return nil, nil
	}

	// Determine script directory
	scriptDir := filepath.Join(r.ToolRoot, config.Scripts.Directory)
	if config.Scripts.Directory == "" {
		scriptDir = filepath.Join(r.ToolRoot, "scripts")
	}

	// Check if script directory exists
	if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("script directory does not exist: %s", scriptDir)
	}

	var results []*ScriptResult

	for _, scriptName := range config.Scripts.Scripts {
		scriptPath := filepath.Join(scriptDir, scriptName)
		
		result := r.RunScript(scriptPath)
		results = append(results, result)

		// Stop on error unless we're being lenient
		if !result.Success && result.Error != nil {
			break
		}
	}

	return results, nil
}

// RunScript executes a single script
func (r *ScriptRunner) RunScript(scriptPath string) *ScriptResult {
	result := &ScriptResult{
		Script:  filepath.Base(scriptPath),
		Success: false,
	}

	// Check if script exists
	info, err := os.Stat(scriptPath)
	if err != nil {
		result.Error = fmt.Errorf("script not found: %w", err)
		return result
	}

	// Check if script is executable
	if info.Mode()&0111 == 0 {
		result.Error = fmt.Errorf("script is not executable (run: chmod +x %s)", scriptPath)
		return result
	}

	// Dry run mode
	if r.DryRun {
		fmt.Fprintf(r.Output, "  [DRY RUN] Would execute: %s\n", result.Script)
		result.Success = true
		return result
	}

	// Execute script
	startTime := time.Now()

	cmd := exec.Command(scriptPath)
	cmd.Dir = filepath.Dir(scriptPath)

	// Set up environment
	cmd.Env = os.Environ()
	for key, value := range r.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Errorf("failed to create stdout pipe: %w", err)
		return result
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Errorf("failed to create stderr pipe: %w", err)
		return result
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		result.Error = fmt.Errorf("failed to start script: %w", err)
		return result
	}

	// Stream output
	var outputLines []string
	done := make(chan bool)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			outputLines = append(outputLines, line)
			if r.Verbose {
				fmt.Fprintf(r.Output, "    %s\n", line)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			outputLines = append(outputLines, line)
			if r.Verbose {
				fmt.Fprintf(r.Output, "    %s\n", line)
			}
		}
		done <- true
	}()

	// Wait for command to complete
	<-done
	err = cmd.Wait()

	result.Duration = time.Since(startTime)
	result.Output = strings.Join(outputLines, "\n")

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Error = fmt.Errorf("script failed with exit code %d", result.ExitCode)
		return result
	}

	result.ExitCode = 0
	result.Success = true
	return result
}

// RunScriptByName finds and runs a script by name
func (r *ScriptRunner) RunScriptByName(scriptDir, scriptName string) *ScriptResult {
	scriptPath := filepath.Join(scriptDir, scriptName)
	return r.RunScript(scriptPath)
}

// GetDefaultEnvironment returns default environment variables for scripts
func GetDefaultEnvironment(toolRoot, toolName string, homeDir, configDir string) map[string]string {
	return map[string]string{
		"MERLIN_TOOL":       toolName,
		"MERLIN_TOOL_ROOT":  toolRoot,
		"MERLIN_HOME":       homeDir,
		"MERLIN_CONFIG_DIR": configDir,
		"HOME":              homeDir,
	}
}

// FormatScriptResult formats a script result for display
func FormatScriptResult(result *ScriptResult, verbose bool) string {
	var sb strings.Builder

	if result.Success {
		sb.WriteString(fmt.Sprintf("  ✓ %s", result.Script))
		if verbose {
			sb.WriteString(fmt.Sprintf(" (%.2fs)", result.Duration.Seconds()))
		}
	} else {
		sb.WriteString(fmt.Sprintf("  ✗ %s", result.Script))
		if result.Error != nil {
			sb.WriteString(fmt.Sprintf(" - %s", result.Error.Error()))
		}
	}

	return sb.String()
}

// ValidateScripts checks if all scripts exist and are executable
func ValidateScripts(toolRoot string, config *models.ToolMerlinConfig) []error {
	if !config.HasScripts() {
		return nil
	}

	var errors []error

	scriptDir := filepath.Join(toolRoot, config.Scripts.Directory)
	if config.Scripts.Directory == "" {
		scriptDir = filepath.Join(toolRoot, "scripts")
	}

	// Check if script directory exists
	if _, err := os.Stat(scriptDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Errorf("script directory does not exist: %s", scriptDir))
		return errors
	}

	for _, scriptName := range config.Scripts.Scripts {
		scriptPath := filepath.Join(scriptDir, scriptName)
		
		info, err := os.Stat(scriptPath)
		if err != nil {
			errors = append(errors, fmt.Errorf("script not found: %s", scriptName))
			continue
		}

		// Check if executable
		if info.Mode()&0111 == 0 {
			errors = append(errors, fmt.Errorf("script not executable: %s", scriptName))
		}
	}

	return errors
}

