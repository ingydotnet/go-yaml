// Copyright 2026 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package aliasdata_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
	"go.yaml.in/yaml/v4/plugin/aliasdata"
)

func TestDataAndMerge(t *testing.T) {
	plugin, err := aliasdata.New(aliasdata.Data(map[string]any{
		"defaults": map[string]any{"color": "blue", "size": 3},
	}))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	err = yaml.Load([]byte("<<: *defaults\nsize: 5\n"), &got,
		yaml.WithPlugin(plugin))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"color": "blue", "size": 5}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestNodesPreserveTags(t *testing.T) {
	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "123"}
	plugin, err := aliasdata.New(aliasdata.Nodes(map[string]*yaml.Node{
		"number": node,
	}))
	if err != nil {
		t.Fatal(err)
	}
	node.Value = "changed"
	var got map[string]any
	if err := yaml.Load([]byte("value: *number\n"), &got,
		yaml.WithPlugin(plugin)); err != nil {
		t.Fatal(err)
	}
	if got["value"] != "123" {
		t.Fatalf("got %#v", got)
	}
}

func TestEnvironmentPatternMayMatchNothing(t *testing.T) {
	plugin, err := aliasdata.New(
		aliasdata.EnvPattern("ALIAS_DATA_NOT_SET_FOR_TEST_*"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := yaml.Load([]byte("value: true\n"), &got,
		yaml.WithPlugin(plugin)); err != nil {
		t.Fatal(err)
	}
}

func TestFileEnvironmentAndPrecedence(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "data.yaml")
	if err := os.WriteFile(path, []byte(
		"file_only: {source: file}\nshared: file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ALIAS_ENV", "environment")
	plugin, err := aliasdata.New(
		aliasdata.Data(map[string]any{"shared": "inline"}),
		aliasdata.File(path),
		aliasdata.EnvNames("ALIAS_ENV"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("changed: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	input := strings.Join([]string{
		"inline: *shared",
		"file: *file_only",
		"env: *ALIAS_ENV",
		"local: &shared local",
		"winner: *shared",
		"",
	}, "\n")
	if err := yaml.Load([]byte(input), &got,
		yaml.WithPlugin(plugin)); err != nil {
		t.Fatal(err)
	}
	if got["inline"] != "inline" || got["env"] != "environment" ||
		got["winner"] != "local" {
		t.Fatalf("got %#v", got)
	}
	if !reflect.DeepEqual(got["file"], map[string]any{"source": "file"}) {
		t.Fatalf("got file %#v", got["file"])
	}
}

func TestStreamScope(t *testing.T) {
	input := []byte("--- &saved {x: 1}\n---\ncopy: *saved\n")
	var without []map[string]any
	if err := yaml.Load(input, &without, yaml.WithAllDocuments()); err == nil {
		t.Fatal("expected the default policy to reject a cross-document alias")
	}
	plugin, err := aliasdata.New(aliasdata.Stream())
	if err != nil {
		t.Fatal(err)
	}
	var got []map[string]any
	if err := yaml.Load(input, &got, yaml.WithAllDocuments(),
		yaml.WithPlugin(plugin)); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[1]["copy"] == nil {
		t.Fatalf("got %#v", got)
	}
}

func TestOptsYAMLEmptyMeansStream(t *testing.T) {
	option, err := yaml.OptsYAML("plugin: {alias-data: true}\n")
	if err != nil {
		t.Fatal(err)
	}
	input := []byte("--- &saved {x: 1}\n---\ncopy: *saved\n")
	var got []map[string]any
	if err := yaml.Load(input, &got, yaml.WithAllDocuments(), option); err != nil {
		t.Fatal(err)
	}
}

func TestEnvironmentPatternAndSnapshot(t *testing.T) {
	t.Setenv("ALIAS_ONE", "one")
	plugin, err := aliasdata.New(aliasdata.EnvPattern("ALIAS_*"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("ALIAS_ONE", "changed")
	var got map[string]string
	if err := yaml.Load([]byte("value: *ALIAS_ONE\n"), &got,
		yaml.WithPlugin(plugin)); err != nil {
		t.Fatal(err)
	}
	if got["value"] != "one" {
		t.Fatalf("got %#v", got)
	}
}

func TestConfigurationErrors(t *testing.T) {
	tests := []struct {
		name    string
		options []aliasdata.Option
	}{
		{"invalid data key", []aliasdata.Option{
			aliasdata.Data(map[string]any{"bad.key": 1}),
		}},
		{"missing environment", []aliasdata.Option{
			aliasdata.EnvNames("ALIAS_DATA_NOT_SET_FOR_TEST"),
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := aliasdata.New(test.options...); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
