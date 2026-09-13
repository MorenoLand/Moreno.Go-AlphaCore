package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultWork = "bin"

type Workspace struct{ root string }

func Open(value string) (Workspace, error) {
	if strings.TrimSpace(value) == "" {
		value = DefaultWork
	}
	root, err := filepath.Abs(value)
	if err != nil {
		return Workspace{}, fmt.Errorf("resolve work directory: %w", err)
	}
	if info, statErr := os.Stat(root); statErr == nil && !info.IsDir() {
		return Workspace{}, fmt.Errorf("work path is not a directory: %s", root)
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return Workspace{}, fmt.Errorf("inspect work directory: %w", statErr)
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return Workspace{}, fmt.Errorf("create work directory: %w", err)
	}
	return Workspace{root: root}, nil
}

func (w Workspace) Root() string { return w.root }

func (w Workspace) Database(name string) string { return filepath.Join(w.root, name+".sqlite3") }
