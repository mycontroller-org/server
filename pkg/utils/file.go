package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"go.uber.org/zap"
)

// IsFileExists checks the file availability
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirExists checks the directory availability
func IsDirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CreateDir func
func CreateDir(dir string) error {
	if !IsDirExists(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			zap.L().Error("failed to create a directory", zap.String("dir", dir))
			return err
		}
	}
	return nil
}

// RemoveDir func
func RemoveDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		zap.L().Error("failed to remove a directory", zap.String("dir", dir))
		return err
	}
	return nil
}

// RemoveFileOrEmptyDir func
func RemoveFileOrEmptyDir(file string) error {
	err := os.Remove(file)
	if err != nil {
		zap.L().Error("failed to remove a file/dir", zap.String("file", file))
		return err
	}
	return nil
}

// CopyFile from a location to another location
func CopyFile(src, dst string) error {
	bufferSize := 1024

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if IsFileExists(dst) {
		return fmt.Errorf("destination file exists: %s", dst)
	}

	// create target dir location
	dir, _ := filepath.Split(dst)
	err = CreateDir(dir)
	if err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, bufferSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}

// WriteFile func
func WriteFile(dir, filename string, data []byte) error {
	err := CreateDir(dir)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fmt.Sprintf("%s/%s", dir, filename), data, os.ModePerm)
}

// AppendFile func
func AppendFile(dir, filename string, data []byte) error {
	err := CreateDir(dir)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// ReadFile func
func ReadFile(dir, filename string) ([]byte, error) {
	err := CreateDir(dir)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, filename))
}

// ListFiles func
func ListFiles(dir string) ([]types.File, error) {
	err := CreateDir(dir)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	items := make([]types.File, 0)
	for _, file := range files {
		if !file.IsDir() {
			f := types.File{
				Name:         file.Name(),
				Size:         file.Size(),
				ModifiedTime: file.ModTime(),
				IsDir:        false,
				FullPath:     path.Join(dir, file.Name()),
			}
			items = append(items, f)
		}
	}
	return items, nil
}

// ListDirs func
func ListDirs(dir string) ([]types.File, error) {
	err := CreateDir(dir)
	if err != nil {
		return nil, err
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	items := make([]types.File, 0)
	for _, file := range files {
		if file.IsDir() {
			f := types.File{
				Name:         file.Name(),
				Size:         file.Size(),
				ModifiedTime: file.ModTime(),
				IsDir:        true,
			}
			items = append(items, f)
		}
	}
	return items, nil
}
