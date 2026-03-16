// Package fayna provides operational work execution views for the Ryta OS monorepo.
//
// fayna owns the job lifecycle: templates (design-time), execution (runtime),
// activities (cost capture), and outcomes (Layer 7 quality assessment).
//
// From Filipino faena (Spanish origin) — labor, work, the decisive performance.
package fayna

import (
	"io/fs"
	"os"
	"path/filepath"
)

// CopyStaticAssets copies fayna's JavaScript assets to the target directory.
// Consumer apps call this at startup: fayna.CopyStaticAssets(jsDir)
func CopyStaticAssets(targetDir string) error {
	faynaDest := filepath.Join(targetDir, "fayna")
	if err := os.MkdirAll(faynaDest, 0755); err != nil {
		return err
	}
	// JS assets will be added as view modules are built
	return nil
}

// CopyStyles copies fayna's CSS assets to the target directory.
// Consumer apps call this at startup: fayna.CopyStyles(cssDir)
func CopyStyles(targetDir string) error {
	faynaDest := filepath.Join(targetDir, "fayna")
	if err := os.MkdirAll(faynaDest, 0755); err != nil {
		return err
	}
	// CSS assets will be added as view modules are built
	return nil
}

// copyFS copies all files from a source fs.FS to a target directory.
func copyFS(src fs.FS, targetDir string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(targetDir, path)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})
}
