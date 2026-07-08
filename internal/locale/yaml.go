package locale

import (
	"fmt"

	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// ParseYAML parses YAML locale data.
func ParseYAML(tag language.Tag, data []byte) (Catalog, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Catalog{}, fmt.Errorf("parse yaml: %w", err)
	}
	if len(root.Content) == 0 {
		return Catalog{Tag: tag, Entries: make(map[string]Entry)}, nil
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return Catalog{}, fmt.Errorf("line %d: top level must be a mapping", doc.Line)
	}
	n, err := yamlNode(doc)
	if err != nil {
		return Catalog{}, err
	}
	return catalogFrom(tag, n)
}

func yamlNode(y *yaml.Node) (node, error) {
	switch y.Kind {
	case yaml.ScalarNode:
		if y.Tag != "!!str" {
			return node{}, fmt.Errorf("line %d: value must be a string", y.Line)
		}
		return node{line: y.Line, str: y.Value}, nil
	case yaml.MappingNode:
		n := node{line: y.Line, mapping: true}
		for i := 0; i < len(y.Content); i += 2 {
			keyNode, valNode := y.Content[i], y.Content[i+1]
			if keyNode.Kind != yaml.ScalarNode || keyNode.Tag != "!!str" {
				return node{}, fmt.Errorf("line %d: key must be a string", keyNode.Line)
			}
			v, err := yamlNode(valNode)
			if err != nil {
				return node{}, err
			}
			n.children = append(n.children, child{key: keyNode.Value, line: keyNode.Line, node: v})
		}
		return n, nil
	case yaml.DocumentNode, yaml.SequenceNode, yaml.AliasNode:
		return node{}, fmt.Errorf("line %d: value must be a string or a mapping", y.Line)
	default:
		return node{}, fmt.Errorf("line %d: unsupported node", y.Line)
	}
}
