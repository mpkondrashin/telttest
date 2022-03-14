package hybridanalysis

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
	"golang.org/x/sync/errgroup"
)

var (
	//	ErrTooBigFile    = errors.New("too big file size")
	ErrResponseError = errors.New("response error")
)

type Client struct {
	APIKey    string
	userAgent string
}

func New(APIKey string) *Client {
	return &Client{
		APIKey:    APIKey,
		userAgent: "Falcon Sandbox",
	}
}

func (c *Client) SetUserAgent(userAgent string) *Client {
	c.userAgent = userAgent
	return c
}

func (c *Client) ListLatestSamples() (*ListLatest, error) {
	client := &http.Client{}
	url := "https://www.AAAhybrid-analysis.com/api/v2/feed/latest"
	//fmt.Printf("URL: %s\n", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Add("Api-Key", c.APIKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", c.userAgent)
	//req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http.Client.Do: %w", err)
	}
	defer resp.Body.Close()
	//fmt.Printf("Respond: %v", resp)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %w: %d", url, ErrResponseError, resp.StatusCode)
	}
	jsonData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w: %v", ErrResponseError, err)
	}
	//fmt.Printf("%v\n", string(jsonData))
	var data ListLatest
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w: %v\n%s", ErrResponseError, err, string(jsonData))
	}
	log.Printf("Count: %d\n", data.Count)
	log.Printf("Status: %s\n", data.Status)
	return &data, nil
}

func (c *Client) Report(jobID, reportType string) ([]byte, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://www.hybrid-analysis.com/api/v2/report/%s/report/%s", jobID, reportType)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Api-Key", c.APIKey)
	//req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", c.userAgent)
	//req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, err
	//fmt.Printf("%v\n", string(jsonData))
}

func (c *Client) DownloadGzipSample(id string) (io.ReadCloser, error) {
	client := &http.Client{}
	url := fmt.Sprintf("https://www.hybrid-analysis.com/api/v2/report/%s/sample", id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", url, err)
	}
	req.Header.Add("Api-Key", c.APIKey)
	req.Header.Add("Accept", "application/gzip")
	req.Header.Add("User-Agent", c.userAgent)
	//req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	//defer resp.Body.Close()
	return resp.Body, nil
}

func (c *Client) DownloadSample(id string) (io.Reader, io.Closer, error) {
	g, err := c.DownloadGzipSample(id)
	if err != nil {
		return nil, nil, err
	}
	r, err := gzip.NewReader(g)
	if err != nil {
		return nil, nil, err
	}
	return r, g, nil
}

/*
func (c *Client) __IterateReader(callback func(data *ListLatestData, r io.Reader) error,
	filter func(data *ListLatestData) bool) error {
	d, err := c.ListLatest()
	if err != nil {
		return err
	}
	for i := range d.Data {
		each := &d.Data[i]
		if filter != nil && !filter(each) {
			continue
		}
		//fmt.Printf("%s\n", each.JobID)
		u, toClose, err := c.DownloadSample(each.JobID)
		if err != nil {
			if errors.Is(err, gzip.ErrHeader) {
				log.Printf("Missing sample for %s", each.JobID)
				continue
			} else {
				return err
			}
		}
		err = callback(each, u)
		if err != nil {
			return err
		}
		toClose.Close()
	}
	return nil
}
*/

func (c *Client) IterateReader(callback func(data *ListLatestData, r io.Reader) error,
	filter func(data *ListLatestData) bool) error {
	d, err := c.ListLatestSamples()
	if err != nil {
		return err
	}
	eGroup := new(errgroup.Group)
	for i := range d.Data {
		each := &d.Data[i]
		if filter != nil && !filter(each) {
			continue
		}
		eGroup.Go(func() error {
			u, toClose, err := c.DownloadSample(each.JobID)
			if err != nil {
				if errors.Is(err, gzip.ErrHeader) {
					log.Printf("Missing sample for %s", each.JobID)
					return nil
				} else {
					return err
				}
			}
			defer toClose.Close()
			return callback(each, u)
		})
	}
	return eGroup.Wait()
}

func (c *Client) IterateFiles(
	callback func(data *ListLatestData, path string) error,
	filter func(data *ListLatestData) bool) error {
	dir, err := ioutil.TempDir("", "ha")
	if err != nil {
		return fmt.Errorf("ioutil.TempDir: %w", err)
	}
	//fmt.Printf("Temp folder: %s\n", dir)
	defer os.Remove(dir)
	return c.IterateReader(func(data *ListLatestData, r io.Reader) error {
		content, err := ioutil.ReadAll(r)
		if err != nil {
			return fmt.Errorf("IterateReader: %w", err)
		}
		name := ""
		sha256 := Sha256ForData(content)
		for _, each := range data.Processes {
			if each.Sha256 == sha256 {
				name = each.Name
				break
			}
		}
		if name == "" {
			name = GuessFileName(data)
		}
		if name == "" {
			kind, _ := filetype.Match(content)
			ext := "bin"
			if kind != filetype.Unknown {
				ext = kind.Extension
			}
			name = data.Sha1 + "." + ext
		}
		path := filepath.Join(dir, name)
		err = os.WriteFile(path, content, 0700)
		if err != nil {
			return err
		}
		return callback(data, path)
	}, filter)
}

func GuessFileName(data *ListLatestData) string {
	for _, each := range data.Processes {
		switch each.Name {
		case "rundll32.exe":
			// "C:\\steam_api64.dll",#1
			start := strings.Index(each.CommandLine, "\"")
			end := strings.LastIndex(each.CommandLine, "\"")
			if start == -1 || end == -1 {
				return ""
			}
			path := each.CommandLine[start+1 : end]
			backSlashPosition := strings.LastIndex(path, "\\")
			if backSlashPosition == -1 {
				return ""
			}
			return path[backSlashPosition+1:]
		case "iexplore.exe":
			// "command_line": "C:\\5302eb21e43123811ca5935e079c1e516c24ed7ea21113dd266.html"
			start := strings.Index(each.CommandLine, "\\")
			if start == -1 {
				return data.Sha1 + ".html"
			}
			return each.CommandLine[start+1:]
		case "msiexec.exe":
			// /i \"C:\\Endpoint20Agent-x86-1.92.0.msi\"
			extIndex := strings.Index(strings.ToLower(each.CommandLine), ".msi")
			if extIndex == -1 {
				continue
			}
			backslashIndex := strings.LastIndex(each.CommandLine[:extIndex], "\\")
			if backslashIndex == -1 {
				continue
			}
			return each.CommandLine[backslashIndex+1 : extIndex+4]
		case "javaw.exe":
			// /i \"C:\\Endpoint20Agent-x86-1.92.0.jar\"
			extIndex := strings.Index(strings.ToLower(each.CommandLine), ".jar")
			if extIndex == -1 {
				continue
			}
			backslashIndex := strings.LastIndex(each.CommandLine[:extIndex], "\\")
			if backslashIndex == -1 {
				continue
			}
			return each.CommandLine[backslashIndex+1 : extIndex+4]
		case "WScript.exe":
			//  "command_line": "\"C:\\JeppesenInvoice#450051235.vbs\"",
			extIndex := strings.LastIndex(each.CommandLine, ".")
			fmt.Println(extIndex)
			if extIndex == -1 {
				continue
			}
			lastIndex := strings.Index(each.CommandLine[extIndex:], "\\")
			if lastIndex == -1 {
				lastIndex = strings.Index(each.CommandLine[extIndex:], "\"")
			}
			if lastIndex == -1 {
				lastIndex = len(each.CommandLine[extIndex:])
			}
			backslashIndex := strings.LastIndex(each.CommandLine[:extIndex+lastIndex], "\\")
			if backslashIndex == -1 {
				backslashIndex = 0
			}
			return each.CommandLine[backslashIndex+1 : extIndex+lastIndex]
		}
	}
	return ""
}

/*
func Unbackslash(s string) string {
	var sb strings.Builder
	backslash := false
	for _, r := range s {
		if backslash {
			switch r {
			case 'a':
				sb.WriteRune('\a')
			case 'b':
				sb.WriteRune('\b')
			case '\\':
				sb.WriteRune('\\')
			case 't':
				sb.WriteRune('\t')
			case 'n':
				sb.WriteRune('\n')
			case 'f':
				sb.WriteRune('\f')
			case 'r':
				sb.WriteRune('\r')
			case 'v':
				sb.WriteRune('\v')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			}
			backslash = false
		} else {
			if r == '\\' {
				backslash = true
			} else {
				sb.WriteRune(r)
			}
		}
	}
	return sb.String()
}
*/
type ListLatest struct {
	Count  int              `json:"count"`
	Status string           `json:"status"`
	Data   []ListLatestData `json:"data"`
}

type ListLatestData struct {
	JobID             string   `json:"job_id"`
	Md5               string   `json:"md5"`
	Sha1              string   `json:"sha1"`
	Sha256            string   `json:"sha256"`
	Interesting       bool     `json:"interesting"`
	AnalysisStartTime string   `json:"analysis_start_time"`
	ThreatScore       int      `json:"threat_score"`
	ThreatLevel       int      `json:"threat_level"`
	ThreatLevelHuman  string   `json:"threat_level_human"`
	Unknown           bool     `json:"unknown"`
	Domains           []string `json:"domains"`
	Hosts             []string `json:"hosts"`
	HostsGeolocation  []struct {
		IP        string `json:"ip"`
		Latitude  string `json:"latitude"`
		Longitude string `json:"longitude"`
		Country   string `json:"country"`
	} `json:"hosts_geolocation"`
	EnvironmentID          int    `json:"environment_id"`
	EnvironmentDescription string `json:"environment_description"`
	SharedAnalysis         bool   `json:"shared_analysis"`
	Reliable               bool   `json:"reliable"`
	ReportURL              string `json:"report_url"`
	Processes              []struct {
		UID            string `json:"uid"`
		Name           string `json:"name"`
		NormalizedPath string `json:"normalized_path"`
		CommandLine    string `json:"command_line"`
		Sha256         string `json:"sha256"`
		Parentuid      string `json:"parentuid,omitempty"`
	} `json:"processes"`
	ExtractedFiles []struct {
		Name                    string   `json:"name"`
		FileSize                int      `json:"file_size"`
		Sha1                    string   `json:"sha1"`
		Sha256                  string   `json:"sha256"`
		Md5                     string   `json:"md5"`
		TypeTags                []string `json:"type_tags,omitempty"`
		Description             string   `json:"description"`
		RuntimeProcess          string   `json:"runtime_process"`
		ThreatLevel             int      `json:"threat_level"`
		ThreatLevelReadable     string   `json:"threat_level_readable"`
		AvMatched               int      `json:"av_matched,omitempty"`
		AvTotal                 int      `json:"av_total,omitempty"`
		FileAvailableToDownload bool     `json:"file_available_to_download"`
		FilePath                string   `json:"file_path,omitempty"`
	} `json:"extracted_files"`
	Ssdeep string `json:"ssdeep"`
}

func Sha256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func Sha256ForData(data []byte) string {
	h := sha256.New()
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
