package filedialog

import "fmt"

func OpenFile(title, filter string) (string, error) {
	return openFile(title, filter)
}

var errNoSelection = fmt.Errorf("no file selected")
