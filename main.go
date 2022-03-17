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
	//samples := HASamples(config)
	//samplesChannel := make(chan string)
	//err = samples.Download(config.SourceDir, samplesChannel)
	//if err != nil {
	//		log.Printf("samples.Download: %v", err)
	//	}
	//	go processSamples(samplesChannel, config)
	//	<-samplesChannel

	paths := demomw.Generate(config.SourceDir)
	for _, path := range paths {
		checkSample(path, config)
	}
	log.Printf("Cycle finished")
}

func processSamples(samplesChannel chan string, config *Config) {
	for sample := range samplesChannel {
		checkSample(sample, config)
	}
	samplesChannel <- ""
}

func checkSample(path string, config *Config) {
	for i := 0; i < 20; i++ {
		log.Printf("%d: %s", i, path)
		time.Sleep(1 * time.Second)
		//time.Sleep(1 * time.Minute)
		ok := checkConsistency(path, config)
		if ok {
			return
		}
	}
	log.Printf("Timeout for %s", path)
}

func checkConsistency(path string, config *Config) bool {
	s, err := inDir(path)
	if err != nil {
		log.Print(err)
		return true
	}
	t, err := inTargetDir(path, config)
	if err != nil {
		log.Print(err)
		return true
	}
	q, err := inQuarantineDir(path, config)
	if err != nil {
		log.Print(err)
		return true
	}
	if s && !t && !q {
		// Still only in source
		return false
	}
	if s && t && !q {
		// Copyed to target
		return true
	}
	if !s && !t && q {
		// Quarantined
		return true
	}
	log.Printf("Consistency check error: source = %v, target = %v, quarantine = %v, for %s", s, t, q, path)
	return true
}

func exist(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	default:
		return false, fmt.Errorf("%p: %w", path, err)
	}
}

func inTargetDir(path string, config *Config) (bool, error) {
	fileName := filepath.Base(path)
	p := filepath.Join(config.TargetDir, fileName)
	return exist(p)
}

func inQuarantineDir(path string, config *Config) (bool, error) {
	sha1, err := FileSHA1(path)
	if err != nil {
		return false, err
	}
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
