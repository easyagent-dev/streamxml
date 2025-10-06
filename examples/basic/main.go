// Copyright 2025 EasyAgent
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/easyagent-dev/streamxml"
)

func main() {
	fmt.Println("=== Basic StreamXML Example ===\n")

	// Create a new parser
	parser := streamxml.NewStreamXmlParser()

	// Simulate streaming data
	fmt.Println("Streaming data in chunks:")

	// Chunk 1
	chunk1 := "I must call tools to get more information.\n"
	fmt.Printf("Chunk 1: %q\n", chunk1)
	parser.Append(chunk1)

	// Chunk 2
	chunk2 := "<use-tool name=\"get_info\">\n"
	fmt.Printf("Chunk 2: %q\n", chunk2)
	parser.Append(chunk2)

	// Chunk 3
	chunk3 := "{\"name\":\"product\"}\n"
	fmt.Printf("Chunk 3: %q\n", chunk3)
	parser.Append(chunk3)

	// Chunk 4
	chunk4 := "</use-tool>"
	fmt.Printf("Chunk 4: %q\n", chunk4)
	parser.Append(chunk4)

	fmt.Println("\n--- Results ---")

	// Get all text (excluding XML tags)
	text, _ := parser.GetText()
	fmt.Printf("Text content: %q\n\n", text)

	// Get all XML nodes
	nodes, _ := parser.GetXmlNodes()
	fmt.Printf("Found %d XML node(s):\n", len(nodes))
	for i, node := range nodes {
		fmt.Printf("  Node %d:\n", i+1)
		fmt.Printf("    Name: %s\n", node.Name)
		fmt.Printf("    Partial: %v\n", node.Partial)
		fmt.Printf("    Attributes: %v\n", node.Attributes)
		fmt.Printf("    Content: %q\n", node.Content)
	}
}
