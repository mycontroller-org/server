package util

import (
	"fmt"
	"io/ioutil"
	"os"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
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
func CreateDir(dir string) {
	if !IsDirExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}
}

// WriteFile func
func WriteFile(dir, filename string, data []byte) error {
	CreateDir(dir)
	return ioutil.WriteFile(fmt.Sprintf("%s/%s", dir, filename), data, os.ModePerm)
}

// AppendFile func
func AppendFile(dir, filename string, data []byte) error {
	CreateDir(dir)
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
	CreateDir(dir)
	return ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, filename))
}

// ListFiles func
func ListFiles(dir string) ([]ml.File, error) {
	CreateDir(dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	items := make([]ml.File, 0)
	for _, file := range files {
		if !file.IsDir() {
			f := ml.File{
				Name:         file.Name(),
				Size:         file.Size(),
				ModifiedTime: file.ModTime(),
			}
			items = append(items, f)
		}
	}
	return items, nil
}
