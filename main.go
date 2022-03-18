package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mpkondrashin/telttest/pkg/demomw"
)

const (
	title          = "TunnelEffect Long-term testing utility"
	label          = "telttest"
	configFileName = label + ".yaml"
)

func main() {
	config := NewConfig()
	err := config.ParseAll(configFileName)
	if err != nil {
		panic(fmt.Errorf("%s: %v", configFileName, err))
	}
	configLogging(config)
	log.Printf("%s Started", title)

	for {
		paths := demomw.Generate(config.SourceDir)
		for _, path := range paths {
			go checkSample(path, config)
		}
		log.Printf("Cycle finished")
		time.Sleep(1 * time.Minute)
	}
}

/*
func processSamples(samplesChannel chan string, config *Config) {
	for sample := range samplesChannel {
		checkSample(sample, config)
	}
	samplesChannel <- ""
}
*/

func checkSample(path string, config *Config) {
	sha1, err := FileSHA1(path)
	if err != nil {
		log.Printf("FileSHA1: %v", err)
		return
	}
	stopTime := time.Now().Add(config.Timeout)
	for i := 0; time.Now().Before(stopTime); i++ {
		//log.Printf("%d: %s", i, path)
		time.Sleep(10 * time.Second)
		//time.Sleep(1 * time.Minute)
		ok := checkConsistency(path, sha1, config)
		if ok {
			return
		}
	}
	log.Printf("Timeout for %s", path)
}

func checkConsistency(path string, sha1 string, config *Config) bool {
	s, err := inDir(path)
	if err != nil {
		log.Print("inDir ", err)
		return true
	}
	t, err := inTargetDir(path, config)
	if err != nil {
		log.Print("inTargetDir ", err)
		return true
	}
	q, err := inQuarantineDir(sha1, config)
	if err != nil {
		log.Print("inQuarantineDir ", err)
		return true
	}
	if s && !t && !q {
		// Still only in source
		return false
	}
	if s && t && !q {
		// Copied to target
		log.Printf("Copied: %s", path)
		return true
	}
	if !s && !t && q {
		// Quarantined
		log.Printf("Quarantined: %s", path)
		return true
	}
	log.Printf("Consistency check error: source = %v, target = %v, quarantine = %v, for %s", s, t, q, path)
	return true
}

func exist(path string) (bool, error) {
	_, err := os.Stat(path)
	//return err == nil, nil
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("%s: %w", path, err)
	}
}

func inTargetDir(path string, config *Config) (bool, error) {
	fileName := filepath.Base(path)
	p := filepath.Join(config.TargetDir, fileName)
	return exist(p)
}

func inQuarantineDir(sha1 string, config *Config) (bool, error) {
	p := filepath.Join(config.QuarantineDir, sha1+".zip")
	return exist(p)
}

func inDir(path string) (bool, error) {
	return exist(path)
}

/*
func HASamples(conf *Config) *hybridanalysis.Samples {
	ha := hybridanalysis.New(conf.HybridAnalysisAPIKey)
	samples := hybridanalysis.NewSamples(ha).SetThreatLevelThreshold(conf.ThreatLevelThreshold)
	if len(conf.SkipList) > 0 {
		for _, each := range strings.Split(conf.SkipList, ",") {
			samples.SetSkip(each)
		}
	}
	if len(conf.IncludeList) > 0 {
		for _, each := range strings.Split(conf.IncludeList, ",") {
			//	fmt.Printf("\"%s\" INCLUDE: %s\n", IncludeList, each)
			samples.SetInclude(each)
		}
	}
	if len(conf.ExtList) > 0 {
		for _, each := range strings.Split(conf.ExtList, ",") {
			//	fmt.Printf("\"%s\" INCLUDE: %s\n", IncludeList, each)
			samples.SetExtension(each)
		}
	}
	return samples
}
*/
func FileSHA1(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	hash := sha1.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func configLogging(conf *Config) (closeLog func() error, err error) {
	f, err := os.OpenFile(conf.Log,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	log.SetOutput(f)
	return f.Close, nil
}
