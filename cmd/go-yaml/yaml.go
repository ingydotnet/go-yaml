// Copyright 2025 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

// Package main provides YAML formatting utilities for the go-yaml tool.

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v4"
)

// ProcessYAML reads YAML from reader and outputs formatted YAML
func ProcessYAML(reader io.Reader, preserve, unmarshalMode, decodeMode,
	marshalMode, encodeMode bool, selection documentSelection,
	opts []yaml.Option,
) error {
	if unmarshalMode {
		return processYAMLUnmarshal(reader, preserve, marshalMode, selection)
	}
	if decodeMode {
		return processYAMLDecode(reader, preserve, encodeMode, selection,
			nil) // Decode API doesn't support options
	}
	// Default: use Load API with options
	return processYAMLLoad(reader, preserve, marshalMode, encodeMode,
		selection, opts)
}

// processYAMLLoad uses Loader.Load for YAML processing with options
func processYAMLLoad(reader io.Reader, preserve, marshal, encode bool,
	selection documentSelection, opts []yaml.Option,
) error {
	if preserve {
		loader, err := yaml.NewLoader(reader, opts...)
		if err != nil {
			return fmt.Errorf("failed to create loader: %w", err)
		}
		var documents []yaml.Node
		for {
			var node yaml.Node
			err := loader.Load(&node)
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to decode YAML: %w", err)
			}

			documents = append(documents, node)
		}
		return writeYAMLNodes(selectDocuments(documents, selection),
			marshal, encode, opts)
	}

	loader, err := yaml.NewLoader(reader, opts...)
	if err != nil {
		return fmt.Errorf("failed to create loader: %w", err)
	}
	var documents []any
	for {
		var data any
		err := loader.Load(&data)
		if err != nil {
			if err == io.EOF || err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode YAML: %w", err)
		}
		documents = append(documents, data)
	}
	return writeYAMLValues(selectDocuments(documents, selection),
		marshal, encode, opts)
}

func writeYAMLNodes(documents []yaml.Node, marshal, encode bool,
	opts []yaml.Option,
) error {
	if !marshal && !encode {
		dumper, err := yaml.NewDumper(os.Stdout, opts...)
		if err != nil {
			return fmt.Errorf("failed to create dumper: %w", err)
		}
		defer dumper.Close()
		for i := range documents {
			if err := dumper.Dump(wrapDocument(&documents[i])); err != nil {
				return fmt.Errorf("failed to dump YAML: %w", err)
			}
		}
		return nil
	}
	for i := range documents {
		if i > 0 {
			fmt.Println("---")
		}
		node := wrapDocument(&documents[i])
		if marshal {
			output, err := yaml.Marshal(node)
			if err != nil {
				return fmt.Errorf("failed to marshal YAML: %w", err)
			}
			fmt.Print(string(output))
			continue
		}
		enc := yaml.NewEncoder(os.Stdout)
		if err := enc.Encode(node); err != nil {
			enc.Close()
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
		if err := enc.Close(); err != nil {
			return fmt.Errorf("failed to close encoder: %w", err)
		}
	}
	return nil
}

func writeYAMLValues(documents []any, marshal, encode bool,
	opts []yaml.Option,
) error {
	if !marshal && !encode {
		dumper, err := yaml.NewDumper(os.Stdout, opts...)
		if err != nil {
			return fmt.Errorf("failed to create dumper: %w", err)
		}
		defer dumper.Close()
		for _, data := range documents {
			if err := dumper.Dump(data); err != nil {
				return fmt.Errorf("failed to dump YAML: %w", err)
			}
		}
		return nil
	}
	for i, data := range documents {
		if i > 0 {
			fmt.Println("---")
		}
		if marshal {
			output, err := yaml.Marshal(data)
			if err != nil {
				return fmt.Errorf("failed to marshal YAML: %w", err)
			}
			fmt.Print(string(output))
			continue
		}
		enc := yaml.NewEncoder(os.Stdout)
		if err := enc.Encode(data); err != nil {
			enc.Close()
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
		if err := enc.Close(); err != nil {
			return fmt.Errorf("failed to close encoder: %w", err)
		}
	}
	return nil
}

// processYAMLDecode uses deprecated Decoder.Decode for YAML processing (no options support)
func processYAMLDecode(reader io.Reader, preserve, encode bool,
	selection documentSelection, opts []yaml.Option,
) error {
	decoder := yaml.NewDecoder(reader)
	if preserve {
		var documents []yaml.Node
		for {
			var node yaml.Node
			err := decoder.Decode(&node)
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to decode YAML: %w", err)
			}
			documents = append(documents, node)
		}
		return writeYAMLNodes(selectDocuments(documents, selection),
			!encode, encode, opts)
	}

	var documents []any
	for {
		var data any
		err := decoder.Decode(&data)
		if err != nil {
			if err == io.EOF || err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to decode YAML: %w", err)
		}
		documents = append(documents, data)
	}
	return writeYAMLValues(selectDocuments(documents, selection),
		!encode, encode, opts)
}

// processYAMLUnmarshal uses yaml.Unmarshal for YAML processing
func processYAMLUnmarshal(reader io.Reader, preserve, marshal bool,
	selection documentSelection,
) error {
	// Read all input from reader
	input, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Split input into documents
	var documents [][]byte
	for _, document := range bytes.Split(input, []byte("---")) {
		if len(bytes.TrimSpace(document)) > 0 {
			documents = append(documents, document)
		}
	}
	documents = selectDocuments(documents, selection)
	firstDoc := true

	for _, doc := range documents {
		// Add document separator for all documents except the first
		if !firstDoc {
			fmt.Println("---")
		}
		firstDoc = false

		if preserve {
			// Preserve comments and styles by using yaml.Node
			var node yaml.Node
			if err := yaml.Load(doc, &node); err != nil {
				return fmt.Errorf("failed to load YAML: %w", err)
			}

			// If the node is not a DocumentNode, wrap it in one
			var outNode *yaml.Node
			if node.Kind == yaml.DocumentNode {
				outNode = &node
			} else {
				outNode = &yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{&node},
				}
			}

			if marshal {
				// Use Dump for output
				output, err := yaml.Dump(outNode)
				if err != nil {
					return fmt.Errorf("failed to dump YAML: %w", err)
				}
				fmt.Print(string(output))
			} else {
				// Use Dumper for output
				dumper, err := yaml.NewDumper(os.Stdout)
				if err != nil {
					return fmt.Errorf("failed to create dumper: %w", err)
				}
				if err := dumper.Dump(outNode); err != nil {
					dumper.Close()
					return fmt.Errorf("failed to dump YAML: %w", err)
				}
				if err := dumper.Close(); err != nil {
					return fmt.Errorf("failed to close dumper: %w", err)
				}
			}
		} else {
			// For unmarshal mode with -y (not -Y), always use `any` to avoid preserving comments
			var data any
			if err := yaml.Load(doc, &data); err != nil {
				return fmt.Errorf("failed to load YAML: %w", err)
			}

			if marshal {
				// Use Dump for output
				output, err := yaml.Dump(data)
				if err != nil {
					return fmt.Errorf("failed to dump YAML: %w", err)
				}
				fmt.Print(string(output))
			} else {
				// Use Dumper for output
				dumper, err := yaml.NewDumper(os.Stdout)
				if err != nil {
					return fmt.Errorf("failed to create dumper: %w", err)
				}
				if err := dumper.Dump(data); err != nil {
					dumper.Close()
					return fmt.Errorf("failed to dump YAML: %w", err)
				}
				if err := dumper.Close(); err != nil {
					return fmt.Errorf("failed to close dumper: %w", err)
				}
			}
		}
	}

	return nil
}
