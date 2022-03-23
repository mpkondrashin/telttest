package demomw

import (
	"bytes"
	"crypto/sha1"
	"embed"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:embed embed/*
var embedded embed.FS

func Generate(targetDir string) (paths []string) {
	rand.Seed(time.Now().UnixNano())

	// ReadDir reads and returns the entire named directory.
	dir, err := embedded.ReadDir("embed")
	if err != nil {
		log.Printf("embedded.ReadDir: %v", err)
		return
	}
	for _, dirEntry := range dir { //}, networkScanCode} {
		filePath := "embed/" + dirEntry.Name()
		code, err := embedded.ReadFile(filePath)
		if err != nil {
			log.Printf("embedded.ReadFile: %v", err)
			continue
		}
		label := strings.TrimSuffix(dirEntry.Name(), filepath.Ext(dirEntry.Name()))
		path, err := build(label, code, "windows", "amd64", targetDir)
		if err != nil {
			log.Printf("Generate: %v", err)
		}
		paths = append(paths, path)
	}
	return
}

func exe(goos string) string {
	if goos == "windows" {
		return ".exe"
	}
	return ""
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func fileSHA1(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractLabel(code string) string {
	r := regexp.MustCompile(`^//([0-9A-Za-z]+)//`)
	s := r.FindStringSubmatch(code)
	if len(s) != 2 {
		return "unknown"
	}
	return s[1]
}

func build(label string, code []byte, goos, goarch, targetDir string) (string, error) {
	tempDir, err := os.MkdirTemp("", "demomw_*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempDir)

	codePath := filepath.Join(tempDir, "main.go")
	err = os.WriteFile(codePath, []byte(code), 0644)
	if err != nil {
		return "", err
	}

	output := filepath.Join(tempDir, label+exe(goos))
	ldflags := fmt.Sprintf("-X 'main.Ballast=%s'", randString(16))
	command := []string{
		"build", "-o", output, "-ldflags", ldflags, codePath,
	}
	//log.Println("go", strings.Join(command, " "))
	cmd := exec.Command("go", command...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS="+goos)
	cmd.Env = append(cmd.Env, "GOARCH="+goarch)
	var o bytes.Buffer
	cmd.Stdout = &o
	var e bytes.Buffer
	cmd.Stderr = &o
	err = cmd.Run()
	if err != nil {
		if o.Len() > 0 {
			err = fmt.Errorf("%s%w", o.String(), err)
		}
		if e.Len() > 0 {
			err = fmt.Errorf("%s%w", e.String(), err)
		}
		return "", err
	}
	sha1, err := fileSHA1(output)
	if err != nil {
		return "", err
	}
	newName := label + "_" + sha1 + exe(goos)
	targetPath := filepath.Join(targetDir, newName)
	err = os.Rename(output, targetPath)
	if err != nil {
		return "", fmt.Errorf("os.Rename to %s: %w", targetPath, err)
	}
	return targetPath, nil
}
