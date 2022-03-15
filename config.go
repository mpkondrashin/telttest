package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	HybridAnalysisAPIKey string `yaml:"HybridAnalysisAPIKey"`
	ThreatLevelThreshold int    `yaml:"ThreatLevelThreshold"`
	SkipList             string `yaml:"SkipList"`
	IncludeList          string `yaml:"IncludeList"`
	ExtList              string `yaml:"ExtList"`
	SourceDir            string `yaml:"SourceDir"`
	TargetDir            string `yaml:"TargetDir"`
	QuarantineDir        string `yaml:"QuarantineDir"`
	Log                  string `yaml:"Log"`
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) ParseAll(filePath string) error {
	err := c.LoadConfig(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	c.ParseArgs()
	err = c.ParseEnv()
	if err != nil {
		return err
	}
	return c.Validate()
}

func (c *Config) ParseArgs() {
	flag.StringVar(&c.HybridAnalysisAPIKey, "hakey",
		c.HybridAnalysisAPIKey, "Hybrid Analysis API key")
	flag.StringVar(&c.TargetDir, "output",
		c.TargetDir, "Target folder")
	flag.IntVar(&c.ThreatLevelThreshold, "level",
		c.ThreatLevelThreshold, "Threat level threshold")
	flag.StringVar(&c.SkipList, "skip",
		c.SkipList, "Coma separated list of platform keywords to skip")
	flag.StringVar(&c.IncludeList, "include",
		c.IncludeList, "Coma separated list of platform keywords to include")
	flag.StringVar(&c.ExtList, "ext",
		c.ExtList, "Coma separated list of extesions of files to include")
	flag.StringVar(&c.Log, "log",
		c.Log, "Log file")
	flag.Parse()
}

func (c *Config) LoadConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return c.ParseConfig(data)
}

func (c *Config) ParseConfig(data []byte) error {
	err := yaml.UnmarshalStrict(data, c)
	if err == nil {
		return nil
	}
	err = json.Unmarshal(data, c)
	if err == nil {
		return nil
	}
	return err
}

func (c *Config) ParseEnv() error {
	v, ok := os.LookupEnv("TELTTEST_HYBRID_ANALYSIS_API_KEY")
	if ok {
		c.HybridAnalysisAPIKey = v
	}
	mstlt := "TELTTEST_THREAT_LEVEL_THRESHOLD"
	v, ok = os.LookupEnv(mstlt)
	if ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s=%s: %w", mstlt, v, err)
		}
		c.ThreatLevelThreshold = i
	}
	v, ok = os.LookupEnv("TELTTEST_SKIP_LIST")
	if ok {
		c.SkipList = v
	}
	v, ok = os.LookupEnv("TELTTEST_INCLUDE_LIST")
	if ok {
		c.IncludeList = v
	}
	v, ok = os.LookupEnv("TELTTEST_EXT_LIST")
	if ok {
		c.ExtList = v
	}
	v, ok = os.LookupEnv("TELTTEST_SOURCE_DIR")
	if ok {
		c.SourceDir = v
	}
	v, ok = os.LookupEnv("TELTTEST_TARGET_DIR")
	if ok {
		c.TargetDir = v
	}
	v, ok = os.LookupEnv("TELTTEST_QUARANTINE_DIR")
	if ok {
		c.QuarantineDir = v
	}
	v, ok = os.LookupEnv("TELTTEST_LOG")
	if ok {
		c.Log = v
	}
	return nil
}

func (c *Config) Validate() error {
	if c.TargetDir == "" {
		return errors.New("no target folder provided")
	}
	if c.SourceDir == "" {
		return errors.New("no source folder provided")
	}
	if c.QuarantineDir == "" {
		return errors.New("no quarantine folder provided")
	}
	if c.HybridAnalysisAPIKey == "" {
		return errors.New("no Hybrid Analysis API key provided")
	}
	if c.ThreatLevelThreshold < 0 || c.ThreatLevelThreshold > 2 {
		return fmt.Errorf("wrong threat level threshold value (%d)", c.ThreatLevelThreshold)
	}
	return nil
}
