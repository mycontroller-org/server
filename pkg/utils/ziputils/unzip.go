package ziputils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip func
func Unzip(zipFilename string, destinationDirectory string) error {
	reader, err := zip.OpenReader(zipFilename)
	if err != nil {
		return err
	}

	defer reader.Close()

	for _, file := range reader.File {
		fPath := filepath.Join(destinationDirectory, file.Name)

		if !strings.HasPrefix(fPath, filepath.Clean(destinationDirectory)+string(os.PathSeparator)) {
			return fmt.Errorf("%s is an illegal filepath", fPath)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}

		rc, err := file.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		if err != nil {
			return err
		}

		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
