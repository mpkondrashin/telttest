package main

import (
	"encoding/json"
	"errors"
	"flag"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const ENV_PREFIX = "TELTTEST"

type Config struct {
	SourceDir     string `yaml:"SourceDir"`
	TargetDir     string `yaml:"TargetDir"`
	QuarantineDir string `yaml:"QuarantineDir"`
	Log           string `yaml:"Log"`
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

func (c *Config) ParseEnvWithPrefix(prefix string) error {
	p := func(s string) string {
		return strings.Join([]string{prefix, s}, "_")
	}
	v, ok := os.LookupEnv(p("SOURCE_DIR"))
	if ok {
		c.SourceDir = v
	}
	v, ok = os.LookupEnv(p("TARGET_DIR"))
	if ok {
		c.TargetDir = v
	}
	v, ok = os.LookupEnv(p("QUARANTINE_DIR"))
	if ok {
		c.QuarantineDir = v
	}
	v, ok = os.LookupEnv(p("LOG"))
	if ok {
		c.Log = v
	}
	return nil
}

func (c *Config) ParseEnv() error {
	return c.ParseEnvWithPrefix(ENV_PREFIX)
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
	return nil
}
