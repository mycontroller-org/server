package ziputils

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"os"

	"go.uber.org/zap"
)

// Zip func
func Zip(sourceDirectory, zipFilename string) error {
	// Get a Buffer to Write To
	zipFile, err := os.Create(zipFilename)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}
	defer zipFile.Close()

	// Create a new zip archive.
	writer := zip.NewWriter(zipFile)

	// Add some files to the archive.
	err = addFiles(writer, sourceDirectory, "")
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	return writer.Close()
}

func addFiles(writer *zip.Writer, basePath, baseInZip string) error {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			// Recurse
			newBase := fmt.Sprintf("%s/%s/", basePath, file.Name())
			var newBaseInZip string
			if baseInZip == "" {
				newBaseInZip = file.Name()
			} else {
				newBaseInZip = fmt.Sprintf("%s/%s/", baseInZip, file.Name())
			}
			zap.L().Debug("file names", zap.String("newbase", newBase), zap.String("newBaseInZip", newBaseInZip))
			err = addFiles(writer, newBase, newBaseInZip)
			if err != nil {
				return err
			}
		} else {
			filename := fmt.Sprintf("%s/%s", basePath, file.Name())
			dat, err := ioutil.ReadFile(filename)
			if err != nil {
				zap.L().Error("error", zap.Error(err))
				return err
			}

			filenameInZip := fmt.Sprintf("%s/%s", baseInZip, file.Name())
			f, err := writer.Create(filenameInZip)
			if err != nil {
				zap.L().Error("error", zap.Error(err))
				return err
			}
			_, err = f.Write(dat)
			if err != nil {
				zap.L().Error("error", zap.Error(err))
				return err
			}
		}
	}
	return nil
}
