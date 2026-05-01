package backfill

import (
	"errors"
	"os"
	"path/filepath"
)

// archivePaths returns every JSON file path the orchestrator should rename.
func archivePaths(autoDir, accountsPath, schedulesPath string) []string {
	var out []string
	if entries, err := os.ReadDir(autoDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
				out = append(out, filepath.Join(autoDir, e.Name()))
			}
		}
	}
	if _, err := os.Stat(accountsPath); err == nil {
		out = append(out, accountsPath)
	}
	if _, err := os.Stat(schedulesPath); err == nil {
		out = append(out, schedulesPath)
	}
	return out
}

// archive renames src to src + ".preP2.bak". If the .bak target already
// exists, it is replaced; this keeps the operation idempotent across retries.
func archive(src string) error {
	dst := src + ".preP2.bak"
	if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(src, dst)
}
