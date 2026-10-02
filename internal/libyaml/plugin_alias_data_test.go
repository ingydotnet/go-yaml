// Copyright 2026 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package libyaml

import (
	"reflect"
	"testing"
)

type tracingAliasDataPlugin struct {
	actions *[]string
}

func (p tracingAliasDataPlugin) NewAliasDataContext() (AliasDataContext, error) {
	return &tracingAliasDataContext{actions: p.actions}, nil
}

type tracingAliasDataContext struct {
	actions *[]string
	anchors map[string]*Node
}

func (c *tracingAliasDataContext) add(action string) {
	*c.actions = append(*c.actions, action)
}

func (c *tracingAliasDataContext) BeginStream() error {
	c.add("begin-stream")
	return nil
}

func (c *tracingAliasDataContext) BeginDocument() error {
	c.add("begin-document")
	c.anchors = map[string]*Node{}
	return nil
}

func (c *tracingAliasDataContext) DefineAnchor(name string, node *Node) error {
	c.add("define:" + name)
	c.anchors[name] = node
	return nil
}

func (c *tracingAliasDataContext) ResolveAlias(name string) (*Node, bool, error) {
	c.add("resolve:" + name)
	node, found := c.anchors[name]
	return node, found, nil
}

func (c *tracingAliasDataContext) EndDocument() error {
	c.add("end-document")
	return nil
}

func (c *tracingAliasDataContext) EndStream() error {
	c.add("end-stream")
	return nil
}

func TestAliasDataLifecycle(t *testing.T) {
	var actions []string
	options, err := ApplyOptions(func(options *Options) error {
		options.AliasData = tracingAliasDataPlugin{actions: &actions}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	composer := NewComposer([]byte("value: &a 1\ncopy: *a\n"), options)
	defer composer.Destroy()
	if composer.Compose() == nil || composer.Compose() != nil {
		t.Fatal("unexpected composed documents")
	}
	want := []string{
		"begin-stream",
		"begin-document",
		"define:a",
		"resolve:a",
		"end-document",
		"end-stream",
	}
	if !reflect.DeepEqual(actions, want) {
		t.Fatalf("got %#v, want %#v", actions, want)
	}
}
