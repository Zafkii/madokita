package filedialog

import "fmt"

func OpenFile(title, filter string) (string, error) {
	return openFile(title, filter)
}

func SaveFile(title, filter string) (string, error) {
	return saveFile(title, filter)
}

var errNoSelection = fmt.Errorf("no file selected")
