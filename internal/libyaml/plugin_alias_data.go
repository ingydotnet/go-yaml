// Copyright 2026 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package libyaml

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// AliasDataPlugin creates operation-local anchor policy contexts.
// Implementations must support concurrent calls.
type AliasDataPlugin interface {
	NewAliasDataContext() (AliasDataContext, error)
}

// AliasDataContext controls anchor storage and alias resolution for one YAML
// stream. Methods are called in stream order by the composer.
type AliasDataContext interface {
	BeginStream() error
	BeginDocument() error
	DefineAnchor(name string, node *Node) error
	ResolveAlias(name string) (*Node, bool, error)
	EndDocument() error
	EndStream() error
}

// AliasDataEnv configures an environment snapshot.
type AliasDataEnv struct {
	All     bool
	Names   []string
	Pattern string
}

// AliasDataConfig configures the built-in Alias-Data implementation.
type AliasDataConfig struct {
	Data   map[string]any
	Nodes  map[string]*Node
	File   string
	Env    *AliasDataEnv
	Stream bool
}

// BuiltinAliasDataPlugin is the configurable default Alias-Data policy.
type BuiltinAliasDataPlugin struct {
	inline map[string]*Node
	file   map[string]*Node
	env    map[string]*Node
	stream bool
}

var aliasDataName = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// NewAliasDataPlugin snapshots configured sources and returns the built-in
// Alias-Data policy.
func NewAliasDataPlugin(config AliasDataConfig) (*BuiltinAliasDataPlugin, error) {
	inline := make(map[string]*Node, len(config.Data)+len(config.Nodes))
	for name, value := range config.Data {
		if err := validateAliasDataName(name); err != nil {
			return nil, err
		}
		node := new(Node)
		if err := node.Encode(value); err != nil {
			return nil, fmt.Errorf("alias-data: encode %q: %w", name, err)
		}
		inline[name] = node
	}
	for name, node := range config.Nodes {
		if err := validateAliasDataName(name); err != nil {
			return nil, err
		}
		if node == nil {
			return nil, fmt.Errorf("alias-data: node %q must not be nil", name)
		}
		inline[name] = node
	}

	file := map[string]*Node{}
	if config.File != "" {
		data, err := os.ReadFile(config.File)
		if err != nil {
			return nil, fmt.Errorf("alias-data: read file %q: %w", config.File, err)
		}
		file, err = parseAliasDataFile(data)
		if err != nil {
			return nil, fmt.Errorf("alias-data: file %q: %w", config.File, err)
		}
	}

	env, err := snapshotAliasDataEnv(config.Env)
	if err != nil {
		return nil, err
	}
	return &BuiltinAliasDataPlugin{
		inline: cloneAliasDataNodes(inline),
		file:   file,
		env:    env,
		stream: config.Stream,
	}, nil
}

// NewAliasDataPluginFromYAML constructs the default implementation from a
// named-plugin configuration.
func NewAliasDataPluginFromYAML(cfg map[string]any) (*BuiltinAliasDataPlugin, error) {
	config := AliasDataConfig{}
	if len(cfg) == 0 {
		config.Stream = true
	}
	for key, value := range cfg {
		switch key {
		case "data":
			data, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("alias-data: data must be a mapping, got %T", value)
			}
			config.Data = data
		case "file":
			path, ok := value.(string)
			if !ok || path == "" {
				return nil, fmt.Errorf("alias-data: file must be a non-empty string")
			}
			config.File = path
		case "env":
			env, err := aliasDataEnvFromYAML(value)
			if err != nil {
				return nil, err
			}
			config.Env = env
		case "stream":
			stream, ok := value.(bool)
			if !ok {
				return nil, fmt.Errorf("alias-data: stream must be a boolean")
			}
			config.Stream = stream
		default:
			return nil, fmt.Errorf("alias-data: unknown key %q", key)
		}
	}
	return NewAliasDataPlugin(config)
}

func aliasDataEnvFromYAML(value any) (*AliasDataEnv, error) {
	switch value := value.(type) {
	case bool:
		if !value {
			return nil, nil
		}
		return &AliasDataEnv{All: true}, nil
	case string:
		if value == "" {
			return nil, fmt.Errorf("alias-data: env pattern must not be empty")
		}
		return &AliasDataEnv{Pattern: value}, nil
	case []any:
		names := make([]string, len(value))
		for i, item := range value {
			name, ok := item.(string)
			if !ok || name == "" {
				return nil, fmt.Errorf("alias-data: env names must be non-empty strings")
			}
			names[i] = name
		}
		return &AliasDataEnv{Names: names}, nil
	case []string:
		return &AliasDataEnv{Names: append([]string(nil), value...)}, nil
	default:
		return nil, fmt.Errorf(
			"alias-data: env must be a boolean, string, or sequence, got %T", value)
	}
}

func parseAliasDataFile(data []byte) (result map[string]*Node, err error) {
	defer handleErr(&err)
	composer := NewComposer(data, nil)
	defer composer.Destroy()
	document := composer.Compose()
	if document == nil || len(document.Content) != 1 {
		return nil, fmt.Errorf("must contain exactly one document")
	}
	if composer.Compose() != nil {
		return nil, fmt.Errorf("must contain exactly one document")
	}
	NewResolver(nil).Resolve(document)
	root := document.Content[0]
	if root.Kind != MappingNode {
		return nil, fmt.Errorf("document root must be a mapping")
	}
	result = make(map[string]*Node, len(root.Content)/2)
	for i := 0; i < len(root.Content); i += 2 {
		key := root.Content[i]
		if key.Kind != ScalarNode || key.ShortTag() != strTag {
			return nil, fmt.Errorf("mapping keys must be strings")
		}
		if err := validateAliasDataName(key.Value); err != nil {
			return nil, err
		}
		if _, found := result[key.Value]; found {
			return nil, fmt.Errorf("duplicate key %q", key.Value)
		}
		result[key.Value] = root.Content[i+1]
	}
	return result, nil
}

func snapshotAliasDataEnv(config *AliasDataEnv) (map[string]*Node, error) {
	result := map[string]*Node{}
	if config == nil {
		return result, nil
	}
	values := map[string]string{}
	for _, entry := range os.Environ() {
		name, value, found := strings.Cut(entry, "=")
		if found {
			values[name] = value
		}
	}
	selected := map[string]string{}
	switch {
	case config.All:
		selected = values
	case len(config.Names) != 0:
		for _, name := range config.Names {
			if err := validateAliasDataName(name); err != nil {
				return nil, err
			}
			value, found := values[name]
			if !found {
				return nil, fmt.Errorf("alias-data: environment variable %q is not set", name)
			}
			selected[name] = value
		}
	case config.Pattern != "":
		if !strings.Contains(config.Pattern, "*") {
			if err := validateAliasDataName(config.Pattern); err != nil {
				return nil, err
			}
			value, found := values[config.Pattern]
			if !found {
				return nil, fmt.Errorf(
					"alias-data: environment variable %q is not set",
					config.Pattern)
			}
			selected[config.Pattern] = value
		} else {
			for name, value := range values {
				if starMatch(config.Pattern, name) {
					selected[name] = value
				}
			}
		}
	}
	for name, value := range selected {
		if err := validateAliasDataName(name); err != nil {
			return nil, err
		}
		node := new(Node)
		node.SetString(value)
		result[name] = node
	}
	return result, nil
}

func validateAliasDataName(name string) error {
	if !aliasDataName.MatchString(name) {
		return fmt.Errorf(
			"alias-data: key %q must match [A-Za-z0-9_-]+", name)
	}
	return nil
}

func starMatch(pattern, value string) bool {
	parts := strings.Split(pattern, "*")
	position := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 && !strings.HasPrefix(value, part) {
			return false
		}
		index := strings.Index(value[position:], part)
		if index < 0 {
			return false
		}
		position += index + len(part)
	}
	last := parts[len(parts)-1]
	return last == "" || strings.HasSuffix(value, last)
}

// NewAliasDataContext implements AliasDataPlugin.
func (p *BuiltinAliasDataPlugin) NewAliasDataContext() (AliasDataContext, error) {
	inline := cloneAliasDataNodes(p.inline)
	for name, node := range inline {
		NewResolver(nil).Resolve(node)
		inline[name] = node
	}
	return &aliasDataContext{
		inline: inline,
		file:   cloneAliasDataNodes(p.file),
		env:    cloneAliasDataNodes(p.env),
		stream: p.stream,
	}, nil
}

type aliasDataContext struct {
	current map[string]*Node
	inline  map[string]*Node
	file    map[string]*Node
	env     map[string]*Node
	prior   map[string]*Node
	stream  bool
}

func newDefaultAliasDataContext() AliasDataContext {
	return &aliasDataContext{}
}

func (c *aliasDataContext) BeginStream() error {
	c.prior = map[string]*Node{}
	return nil
}

func (c *aliasDataContext) BeginDocument() error {
	c.current = map[string]*Node{}
	return nil
}

func (c *aliasDataContext) DefineAnchor(name string, node *Node) error {
	c.current[name] = node
	return nil
}

func (c *aliasDataContext) ResolveAlias(name string) (*Node, bool, error) {
	for _, data := range []map[string]*Node{
		c.current, c.inline, c.file, c.env, c.prior,
	} {
		if node, found := data[name]; found {
			return node, true, nil
		}
	}
	return nil, false, nil
}

func (c *aliasDataContext) EndDocument() error {
	if c.stream {
		for name, node := range c.current {
			c.prior[name] = node
		}
	}
	return nil
}

func (c *aliasDataContext) EndStream() error { return nil }

func cloneAliasDataNodes(source map[string]*Node) map[string]*Node {
	result := make(map[string]*Node, len(source))
	memo := map[*Node]*Node{}
	var clone func(*Node) *Node
	clone = func(node *Node) *Node {
		if node == nil {
			return nil
		}
		if copy, found := memo[node]; found {
			return copy
		}
		copy := *node
		copy.Content = nil
		copy.Alias = nil
		memo[node] = &copy
		for _, child := range node.Content {
			copy.Content = append(copy.Content, clone(child))
		}
		copy.Alias = clone(node.Alias)
		return &copy
	}
	for name, node := range source {
		result[name] = clone(node)
	}
	return result
}
