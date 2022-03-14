package hybridanalysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Samples struct {
	ha *Client
	//	tagetFolder          string
	skipList             []string
	includeList          []string
	extensions           []string
	threatLevelThreshold int
}

func NewSamples(ha *Client) *Samples {
	return &Samples{
		ha: ha,
		//	tagetFolder:          tagetFolder,
		threatLevelThreshold: 2,
	}
}

func (s *Samples) SetSkip(keyword string) *Samples {
	s.skipList = append(s.skipList, keyword)
	return s
}

func (s *Samples) SetInclude(keyword string) *Samples {
	s.includeList = append(s.includeList, keyword)
	return s
}

func (s *Samples) SetExtension(keyword string) *Samples {
	s.extensions = append(s.extensions, keyword)
	return s
}

func (s *Samples) SetThreatLevelThreshold(threatLevelThreshold int) *Samples {
	s.threatLevelThreshold = threatLevelThreshold
	return s
}

func (s *Samples) MatchExtension(fileName string) bool {
	if len(s.extensions) == 0 {
		return true
	}
	e := strings.ToLower(filepath.Ext(fileName))
	if len(e) < 2 {
		return false
	}
	if e[0] == '.' {
		e = e[1:]
	}
	for _, ext := range s.extensions {
		if e == ext {
			return true
		}
	}
	return false
}

func (s *Samples) Download(targetFolder string, pathChan chan string) error {
	err := s.ha.IterateFiles(
		func(data *ListLatestData, path string) error {
			if !s.MatchExtension(path) {
				log.Printf("Skip unwanted type: %s", filepath.Base(path))
				return nil
			}
			folderName := data.Sha1
			folderPath := filepath.Join(targetFolder, folderName)
			err := os.Mkdir(folderPath, 0700)
			if err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
			sampleFolderName := "sample"
			sampleFolderPath := filepath.Join(folderPath, sampleFolderName)
			err = os.Mkdir(sampleFolderPath, 0700)
			if err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}

			fileName := filepath.Base(path)
			newPath := filepath.Join(sampleFolderPath, fileName)
			err = os.Rename(path, newPath)
			if err != nil {
				return err
			}
			log.Printf("New sample: %s", fileName)

			repName := "hybridanalysis.json"
			repPath := filepath.Join(folderPath, repName)
			s, _ := json.MarshalIndent(data, "", "\t")
			err = os.WriteFile(repPath, s, 0700)
			if err != nil {
				return err
			}
			sha256, err := FileSHA256(newPath)
			if err != nil {
				return err
			}
			sha256FileName := "sha256.txt"
			sha256FilePath := filepath.Join(folderPath, sha256FileName)
			err = os.WriteFile(sha256FilePath, []byte(sha256), 0700)
			if err != nil {
				return err
			}
			if pathChan != nil {
				pathChan <- newPath
			}
			return nil
		},
		func(data *ListLatestData) bool {
			//repName := fmt.Sprintf("%s.txt", data.JobID)
			repPath := filepath.Join(targetFolder, data.Sha1)
			_, err := os.Stat(repPath)
			if err == nil {
				log.Printf("%s: already have it", repPath)
				return false
			}
			if data.ThreatLevel < s.threatLevelThreshold {
				log.Printf("%d: skip low threat level", data.ThreatLevel)
				return false
			}
			include := true
			for _, keyword := range s.includeList {
				fmt.Printf("ZZZ: Check %s\n", keyword)
				include = false
				if strings.Contains(data.EnvironmentDescription, keyword) {
					log.Printf("%s: include as \"%s\" found", data.EnvironmentDescription, keyword)
					include = true
					break
				}
			}
			if !include {
				return false
			}
			for _, keyword := range s.skipList {
				if strings.Contains(data.EnvironmentDescription, keyword) {
					//fmt.Println("Skip Linux")
					log.Printf("%s: skip as \"%s\" found", data.EnvironmentDescription, keyword)
					return false
				}
			}
			return true
		})
	//close(pathChan)
	return err
}

func FileSHA256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil)), nil
}
