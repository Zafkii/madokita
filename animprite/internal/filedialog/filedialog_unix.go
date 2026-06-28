//go:build !windows

package filedialog

import (
	"os/exec"
	"strings"
)

func openFile(title, filter string) (string, error) {
	cmd := exec.Command("zenity", "--file-selection", "--title="+title,
		"--file-filter="+filter)
	out, err := cmd.Output()
	if err != nil {
		// Try kdialog fallback
		cmd2 := exec.Command("kdialog", "--getopenfilename", ".", filter)
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return "", err
		}
		return strings.TrimSpace(string(out2)), nil
	}
	return strings.TrimSpace(string(out)), nil
}

func saveFile(title, filter string) (string, error) {
	cmd := exec.Command("zenity", "--file-selection", "--save", "--title="+title,
		"--file-filter="+filter)
	out, err := cmd.Output()
	if err != nil {
		cmd2 := exec.Command("kdialog", "--getsavefilename", ".", filter)
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return "", err
		}
		return strings.TrimSpace(string(out2)), nil
	}
	return strings.TrimSpace(string(out)), nil
}
