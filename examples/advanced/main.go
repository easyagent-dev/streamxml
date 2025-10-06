package main

import (
	"fmt"
	"github.com/easyagent-dev/streamxml"
)

func main() {
	fmt.Println("=== Advanced StreamXML Example ===\n")

	// Create a custom configuration
	config := streamxml.ParserConfig{
		MaxDepth:               50,
		MaxBufferSize:          5 * 1024 * 1024, // 5MB
		AllowedElements:        []string{"tool", "thinking"},
		BufferCleanupThreshold: 512,
	}

	// Create parser with custom configuration
	parser := streamxml.NewStreamXmlParserWithConfig(config)

	fmt.Println("Configuration:")
	fmt.Printf("  Max Depth: %d\n", config.MaxDepth)
	fmt.Printf("  Max Buffer Size: %d bytes\n", config.MaxBufferSize)
	fmt.Printf("  Allowed Elements: %v\n", config.AllowedElements)
	fmt.Println()

	// Example: Multiple XML nodes with allowed/disallowed elements
	fmt.Println("Streaming mixed content with allowed elements filter:")

	data := []string{
		"Text before\n",
		"<tool name=\"search\">",
		"query text",
		"</tool>\n",
		"<disallowed>This will be treated as text</disallowed>\n",
		"<thinking>",
		"Analyzing...",
		"</thinking>\n",
		"Text after",
	}

	for i, chunk := range data {
		fmt.Printf("Chunk %d: %q\n", i+1, chunk)
		if err := parser.Append(chunk); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Print current state after each chunk
		currentText, _ := parser.GetText()
		currentNodes, _ := parser.GetXmlNodes()
		fmt.Printf("  Current text: %q\n", currentText)
		fmt.Printf("  Current elements: %d\n", len(currentNodes))
		if len(currentNodes) > 0 {
			for j, node := range currentNodes {
				fmt.Printf("    Element %d: <%s> (partial=%v) content=%q\n",
					j+1, node.Name, node.Partial, node.Content)
			}
		}
		fmt.Println()
	}

	fmt.Println("\n--- Results ---")

	// Get text content
	text, _ := parser.GetText()
	fmt.Printf("Text content: %q\n\n", text)

	// Get XML nodes
	nodes, _ := parser.GetXmlNodes()
	fmt.Printf("Found %d XML node(s):\n", len(nodes))
	for i, node := range nodes {
		fmt.Printf("\n  Node %d:\n", i+1)
		fmt.Printf("    Name: %s\n", node.Name)
		fmt.Printf("    Partial: %v\n", node.Partial)
		fmt.Printf("    Content: %q\n", node.Content)
		if len(node.Attributes) > 0 {
			fmt.Printf("    Attributes:\n")
			for k, v := range node.Attributes {
				fmt.Printf("      %s = %q\n", k, v)
			}
		}
	}
}
