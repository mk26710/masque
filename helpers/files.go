package helpers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
)

const (
	MAP_FILE_NAME string = "map.masque.json"
)

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func GetSha256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	hash := hasher.Sum([]byte{})
	hashStr := fmt.Sprintf("%x", hash)

	return hashStr, nil
}

func IsMasqued(dirpath string) bool {
	fp := path.Join(dirpath, MAP_FILE_NAME)
	_, err := os.Stat(fp)

	return err == nil
}
