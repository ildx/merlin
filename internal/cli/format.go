package cli

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes (basic; avoid external deps for portability)
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

// Error prints a formatted error message to stderr with a red prefix.
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s✗ Error:%s %s\n", colorRed, colorReset, msg)
}

// Warning prints a yellow warning message to stderr.
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s⚠ Warning:%s %s\n", colorYellow, colorReset, msg)
}

// Info prints an informational message to stdout with a subtle prefix.
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "%sℹ%s %s\n", colorBlue, colorReset, msg)
}

// Success prints a green success checkmark.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stdout, "✓ %s\n", msg)
}

// Dim returns a dimmed (gray) version of a string for inline usage.
func Dim(s string) string { return colorGray + s + colorReset }

// BulletList formats a slice of strings as an indented bullet list.
func BulletList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	for _, it := range items {
		b.WriteString("  • ")
		b.WriteString(it)
		b.WriteString("\n")
	}
	return b.String()
}
