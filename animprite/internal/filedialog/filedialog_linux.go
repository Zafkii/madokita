//go:build !windows

package filedialog

import (
	"fmt"
	"os/exec"
	"strings"
)

func openLinux(title, filter string) (string, error) {
	cmd := exec.Command("zenity", "--file-selection", "--title="+title,
		"--file-filter="+filter)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("zenity: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
