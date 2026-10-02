// Copyright 2026 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

// Package aliasdata configures external values and stream-scoped anchors.
package aliasdata

import (
	"go.yaml.in/yaml/v4"
	"go.yaml.in/yaml/v4/internal/libyaml"
)

// Option configures the built-in Alias-Data implementation.
type Option func(*libyaml.AliasDataConfig)

// New creates a configured Alias-Data plugin and snapshots file and
// environment sources.
func New(options ...Option) (yaml.AliasDataPlugin, error) {
	config := libyaml.AliasDataConfig{}
	for _, option := range options {
		option(&config)
	}
	return libyaml.NewAliasDataPlugin(config)
}

// Data supplies ordinary Go values as alias targets.
func Data(values map[string]any) Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Data = values
	}
}

// Nodes supplies exact YAML nodes as alias targets.
func Nodes(values map[string]*yaml.Node) Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Nodes = values
	}
}

// File loads alias targets from one YAML mapping document.
func File(path string) Option {
	return func(config *libyaml.AliasDataConfig) {
		config.File = path
	}
}

// EnvAll selects every environment variable.
func EnvAll() Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Env = &libyaml.AliasDataEnv{All: true}
	}
}

// EnvNames selects exact environment variable names.
func EnvNames(names ...string) Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Env = &libyaml.AliasDataEnv{
			Names: append([]string(nil), names...),
		}
	}
}

// EnvPattern selects environment names using * as the only wildcard.
func EnvPattern(pattern string) Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Env = &libyaml.AliasDataEnv{Pattern: pattern}
	}
}

// Stream retains anchors for aliases in later documents.
func Stream() Option {
	return func(config *libyaml.AliasDataConfig) {
		config.Stream = true
	}
}
