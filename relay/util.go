package relay

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gen2brain/beeep"
)

func notify(message string) error {
	if err := beeep.Notify(path.Base(os.Args[0]), message, ""); err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open src file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create dst file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat src file: %w", err)
	}
	if err := os.Chtimes(dstFile.Name(), srcFileInfo.ModTime(), srcFileInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to chtimes: %w", err)
	}
	return nil
}
