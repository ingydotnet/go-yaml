// Copyright 2026 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

// Package toml provides the YAMLStar TOML parser for go-yaml.
package toml

import (
	"errors"

	tomlparser "github.com/yamlstar/yamlstar-plugin-parser-toml/parser"
	"go.yaml.in/yaml/v4"
)

// Plugin parses TOML 1.1 through the shared YAMLStar implementation.
type Plugin struct{}

var _ yaml.ParserPlugin = (*Plugin)(nil)

// Version is the linked TOML parser release version.
const Version = tomlparser.Version

// New creates a TOML parser plugin.
func New() *Plugin { return &Plugin{} }

// Register enables parser=toml in yaml.OptsYAML.
func Register() error {
	return yaml.RegisterPlugin(yaml.PluginRegistration{
		API: "parser", Name: "toml", Version: Version,
		Factory: func(cfg map[string]any) (any, error) {
			if len(cfg) != 0 {
				return nil, errors.New(
					"TOML parser configuration must be empty")
			}
			return New(), nil
		},
	})
}

// Parse implements yaml.ParserPlugin.
func (p *Plugin) Parse(input []byte) ([]yaml.PluginEvent, error) {
	source, err := tomlparser.Parse(input)
	if err != nil {
		return nil, err
	}
	events := make([]yaml.PluginEvent, len(source))
	for i, event := range source {
		events[i] = yaml.PluginEvent{
			Type: event.Type, Value: event.Value,
			Tag: event.Tag, Style: event.Style,
			HeadComment: event.HeadComment,
			LineComment: event.LineComment,
			FootComment: event.FootComment,
		}
	}
	return events, nil
}
