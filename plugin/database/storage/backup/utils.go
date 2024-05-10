package backup

import (
	"path"
	"path/filepath"

	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

// copies files from a location to another location
func CopyFiles(logger *zap.Logger, sourceDir, dstDir string, overwrite bool) error {
	err := utils.CreateDir(dstDir)
	if err != nil {
		return err
	}

	files, err := utils.ListFiles(sourceDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		// have to copy recursive directories too
		// get relative path and append with target directory
		relativePath, err := filepath.Rel(sourceDir, file.FullPath)
		if err != nil {
			logger.Error("error on getting relative path", zap.String("basePath", dstDir), zap.String("targetPath", file.FullPath))
		}
		relativePathBasePath := path.Dir(relativePath)
		finalDstPath := path.Join(dstDir, relativePathBasePath)
		destPath := path.Join(finalDstPath, file.Name)
		err = utils.CopyFileForce(file.FullPath, destPath, overwrite)
		if err != nil {
			return err
		}
	}

	return nil
}
