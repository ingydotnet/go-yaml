package toml_test

import (
	"reflect"
	"testing"

	"go.yaml.in/yaml/v4"
	"go.yaml.in/yaml/v4/plugin/parser/toml"
)

func TestPlugin(t *testing.T) {
	var got map[string]any
	input := []byte("title = \"TOML\"\nanswer = 42\nvalues = [true, 2.5]\n")
	err := yaml.Load(input, &got, yaml.WithPlugin(toml.New()))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"title": "TOML", "answer": 42,
		"values": []any{true, 2.5},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestComments(t *testing.T) {
	var node yaml.Node
	err := yaml.Load([]byte("# file\nanswer = 42 # value\n"), &node,
		yaml.WithPlugin(toml.New()))
	if err != nil {
		t.Fatal(err)
	}
	root := node.Content[0]
	if root.HeadComment != "# file" {
		t.Fatalf("head comment is %q", root.HeadComment)
	}
	if root.Content[1].LineComment != "# value" {
		t.Fatalf("line comment is %q", root.Content[1].LineComment)
	}
}
