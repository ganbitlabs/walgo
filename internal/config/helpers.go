package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// UpdateWalgoYAMLProjectID sets walrus.projectID in walgo.yaml.
// It edits the YAML node tree in place, so comments and key order are kept.
func UpdateWalgoYAMLProjectID(sitePath, objectID string) error {
	path := filepath.Join(sitePath, "walgo.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse walgo.yaml: %w", err)
	}

	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return fmt.Errorf("failed to parse walgo.yaml: top level is not a mapping")
	}

	walrus := mappingChild(root, "walrus")
	setScalar(walrus, "projectID", objectID)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return fmt.Errorf("failed to marshal walgo.yaml: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("failed to marshal walgo.yaml: %w", err)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write walgo.yaml: %w", err)
	}
	return nil
}

// mappingChild returns the mapping value stored under key, creating (or
// replacing a non-mapping value with) an empty mapping when needed.
func mappingChild(m *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			v := m.Content[i+1]
			if v.Kind != yaml.MappingNode {
				*v = yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", HeadComment: v.HeadComment, LineComment: v.LineComment}
			}
			return v
		}
	}
	child := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, child)
	return child
}

// setScalar sets key to a string value in mapping m, appending the key if absent.
func setScalar(m *yaml.Node, key, value string) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			v := m.Content[i+1]
			v.Kind, v.Tag, v.Value, v.Style = yaml.ScalarNode, "!!str", value, 0
			v.Content = nil
			return
		}
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value})
}
