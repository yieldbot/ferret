/*
 * Ferret
 * Copyright (c) 2016 Yieldbot, Inc.
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

// Package config provides configuration functionality
package config

import (
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
)

// Config represents a configuration
type Config struct {
	File      string
	Search    Search     `yaml:"search"`
	Listen    Listen     `yaml:"listen"`
	Providers []Provider `yaml:"providers"`
}

// Search represents the structure of the config search field
type Search struct {
	GotoCmd    string        `yaml:"gotoCmd"`
	TimeoutStr string        `yaml:"timeout"`
	Timeout    time.Duration `yaml:"-"`
}

// Listen represents the structure of the config listen field
type Listen struct {
	Address   string `yaml:"address"`
	Path      string `yaml:"path"`
	Providers string `yaml:"providers"`
}

// Provider represents the structure of the config provider field
type Provider struct {
	Provider string `yaml:"provider"`
	Name     string `yaml:"name"`
	Title    string `yaml:"title"`
	Priority int64  `yaml:"priority"`
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
	Key      string `yaml:"key"`
	Repo     string `yaml:"repo"`
	Query    string `yaml:"query"`
	Rewrite  string `yaml:"rewrite"`
}

// Load loads the configuration from the given file
func (config *Config) Load() error {

	// Check and load the job file
	if config.File == "" {
		return errors.New("missing config file")
	}
	confData, err := ioutil.ReadFile(config.File)
	if err != nil {
		return errors.New("failed to load config file due to " + err.Error())
	}

	// Parse template
	t, err := template.New("config").Funcs(template.FuncMap{
		"env": tmplFuncEnv,
	}).Parse(string(confData))

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, nil); err != nil {
		return errors.New("failed to parse config file due to " + err.Error())
	}
	confData = buf.Bytes()

	// YAML
	if err := yaml.Unmarshal(confData, &config); err != nil {
		return errors.New("failed to unmarshal config file due to " + err.Error())
	}

	return nil
}

// tmplFuncEnv returns an environment variable
func tmplFuncEnv(args ...interface{}) string {
	if len(args) > 0 {
		if s, ok := args[0].(string); ok && s != "" {
			return os.Getenv(s)
		}
	}

	return ""
}
