package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mpkondrashin/telttest/pkg/hybridanalysis"
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
	samples := HASamples(config)
	samplesChannel := make(chan string)
	err = samples.Download(config.TargetDir, samplesChannel)
	if err != nil {
		log.Printf("samples.Download: %v", err)
	}
	go processSamples(samplesChannel, config)
	<-samplesChannel
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
		time.Sleep(1 * time.Second)
		//time.Sleep(1 * time.Minute)
		ok := checkConsistency(path, config)
		if ok {
			break
		}
	}
	log.Printf("Timeout for %s", path)
}

func checkConsistency(path string, config *Config) bool {
	t, err := inTargetDir(path, config)
	if err != nil {
		log.Printf("For file %s: %v", path, err)
		return true
	}
	q, err := inQuarantineDir(path, config)
	if err != nil {
		log.Printf("For file %s: %v", path, err)
		return true
	}
	if !t && !q {
		return false
	}
	if t != q {
		return true
	}
	log.Printf("consistency check error: q = %v, t = %v", q, t)
	return true
}

func inTargetDir(path string, config *Config) (bool, error) {
	fileName := filepath.Base(path)
	p := filepath.Join(config.TargetDir, fileName)
	_, err := os.Stat(p)
	return err != os.ErrNotExist, err

}

func inQuarantineDir(path string, config *Config) (bool, error) {
	sha1, err := FileSHA1(path)
	if err != nil {
		return false, err
	}
	p := filepath.Join(config.QuarantineDir, sha1+".zip")
	_, err = os.Stat(p)
	return err != os.ErrNotExist, err
}

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
