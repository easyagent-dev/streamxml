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

package streamxml

import (
	"testing"
)

// TestMultiRoundAppendTextOnly tests appending text in multiple rounds
func TestMultiRoundAppendTextOnly(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Append first part
	parser.Append("Hello ")
	text, _ := parser.GetText()
	if text != "Hello " {
		t.Errorf("Round 1: expected 'Hello ', got '%s'", text)
	}

	// Round 2: Append second part
	parser.Append("World")
	text, _ = parser.GetText()
	if text != "Hello World" {
		t.Errorf("Round 2: expected 'Hello World', got '%s'", text)
	}

	// Round 3: Append third part
	parser.Append("!\n")
	text, _ = parser.GetText()
	if text != "Hello World!\n" {
		t.Errorf("Round 3: expected 'Hello World!\\n', got '%s'", text)
	}

	// Round 4: Append fourth part
	parser.Append("How are you?")
	text, _ = parser.GetText()
	if text != "Hello World!\nHow are you?" {
		t.Errorf("Round 4: expected 'Hello World!\\nHow are you?', got '%s'", text)
	}
}

// TestMultiRoundAppendBreakInTagName tests breaking in the middle of tag name
func TestMultiRoundAppendBreakInTagName(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Break in middle of tag name
	parser.Append("Text before <too")
	text, _ := parser.GetText()
	nodes, _ := parser.GetXmlNodes()
	if text != "Text before " {
		t.Errorf("Round 1: expected 'Text before ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 1: expected 1 node, got %d", len(nodes))
	} else if !nodes[0].Partial {
		t.Errorf("Round 1: expected partial node")
	}

	// Round 2: Complete tag name and add attribute
	parser.Append("l name=\"te")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text before " {
		t.Errorf("Round 2: expected 'Text before ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 2: expected 1 node, got %d", len(nodes))
	} else if nodes[0].Name != "tool" {
		t.Errorf("Round 2: expected tag name 'tool', got '%s'", nodes[0].Name)
	}

	// Round 3: Complete attribute value
	parser.Append("st\">\nCon")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text before " {
		t.Errorf("Round 3: expected 'Text before ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 3: expected 1 node, got %d", len(nodes))
	}

	// Round 4: Add more content
	parser.Append("tent here")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text before " {
		t.Errorf("Round 4: expected 'Text before ', got '%s'", text)
	}

	// Round 5: Close tag
	parser.Append("\n</tool>")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text before " {
		t.Errorf("Round 5: expected 'Text before ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 5: expected 1 node, got %d", len(nodes))
	} else {
		if nodes[0].Partial {
			t.Errorf("Round 5: expected complete node")
		}
		if nodes[0].Content != "\nContent here\n" {
			t.Errorf("Round 5: expected content '\\nContent here\\n', got '%s'", nodes[0].Content)
		}
	}
}

// TestMultiRoundAppendBreakInAttribute tests breaking in attribute
func TestMultiRoundAppendBreakInAttribute(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Break in attribute name
	parser.Append("<tag at")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}

	// Round 2: Complete attribute name, break in value
	parser.Append("tr=\"val")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 2: expected 1 partial node")
	}

	// Round 3: Complete attribute value and tag
	parser.Append("ue\">content</tag>")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 3: expected 1 node, got %d", len(nodes))
	} else {
		if nodes[0].Partial {
			t.Errorf("Round 3: expected complete node")
		}
		if nodes[0].Attributes["attr"] != "value" {
			t.Errorf("Round 3: expected attr='value', got '%s'", nodes[0].Attributes["attr"])
		}
		if nodes[0].Content != "content" {
			t.Errorf("Round 3: expected content 'content', got '%s'", nodes[0].Content)
		}
	}
}

// TestMultiRoundAppendBreakInContent tests breaking in content
func TestMultiRoundAppendBreakInContent(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Open tag
	parser.Append("<data>")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}

	// Round 2: Add partial content
	parser.Append("First ")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 2: expected 1 node")
	} else if nodes[0].Content != "First " {
		t.Errorf("Round 2: expected content 'First ', got '%s'", nodes[0].Content)
	}

	// Round 3: Add more content
	parser.Append("part\nSecond ")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 3: expected 1 node")
	} else if nodes[0].Content != "First part\nSecond " {
		t.Errorf("Round 3: expected content 'First part\\nSecond ', got '%s'", nodes[0].Content)
	}

	// Round 4: Add final content
	parser.Append("part")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 4: expected 1 node")
	} else if nodes[0].Content != "First part\nSecond part" {
		t.Errorf("Round 4: expected content 'First part\\nSecond part', got '%s'", nodes[0].Content)
	}

	// Round 5: Close tag
	parser.Append("</data>")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 5: expected 1 node")
	} else {
		if nodes[0].Partial {
			t.Errorf("Round 5: expected complete node")
		}
		if nodes[0].Content != "First part\nSecond part" {
			t.Errorf("Round 5: expected content 'First part\\nSecond part', got '%s'", nodes[0].Content)
		}
	}
}

// TestMultiRoundAppendBreakInClosingTag tests breaking in closing tag
func TestMultiRoundAppendBreakInClosingTag(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Open tag with content
	parser.Append("<element>data")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}
	if nodes[0].Content != "data" {
		t.Errorf("Round 1: expected content 'data', got '%s'", nodes[0].Content)
	}

	// Round 2: Start closing tag - content includes the '<'
	parser.Append("<")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 2: expected 1 partial node")
	}
	if nodes[0].Content != "data<" {
		t.Errorf("Round 2: expected content 'data<', got '%s'", nodes[0].Content)
	}

	// Round 3: Add slash - becomes closing tag indicator
	parser.Append("/")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 3: expected 1 partial node")
	}

	// Round 4: Add partial tag name
	parser.Append("ele")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 4: expected 1 partial node")
	}

	// Round 5: Complete closing tag
	parser.Append("ment>")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 5: expected 1 node")
	} else if nodes[0].Partial {
		t.Errorf("Round 5: expected complete node")
	}
	if nodes[0].Content != "data" {
		t.Errorf("Round 5: expected content 'data', got '%s'", nodes[0].Content)
	}
}

// TestMultiRoundAppendMultipleNodes tests multiple nodes with breaks
func TestMultiRoundAppendMultipleNodes(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Text and partial first tag
	parser.Append("Text 1\n<tag1>con")
	text, _ := parser.GetText()
	nodes, _ := parser.GetXmlNodes()
	if text != "Text 1\n" {
		t.Errorf("Round 1: expected 'Text 1\\n', got '%s'", text)
	}
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}

	// Round 2: Complete first tag, add text
	parser.Append("tent1</tag1>\nText ")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text 1\n\nText " {
		t.Errorf("Round 2: expected 'Text 1\\n\\nText ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 2: expected 1 node, got %d", len(nodes))
	} else if nodes[0].Partial {
		t.Errorf("Round 2: expected complete node")
	}

	// Round 3: More text and start second tag
	parser.Append("2\n<tag2")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text 1\n\nText 2\n" {
		t.Errorf("Round 3: expected 'Text 1\\n\\nText 2\\n', got '%s'", text)
	}
	if len(nodes) != 2 {
		t.Errorf("Round 3: expected 2 nodes, got %d", len(nodes))
	} else if nodes[1].Partial != true {
		t.Errorf("Round 3: expected second node to be partial")
	}

	// Round 4: Complete second tag
	parser.Append(">content2</tag2>")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Text 1\n\nText 2\n" {
		t.Errorf("Round 4: expected 'Text 1\\n\\nText 2\\n', got '%s'", text)
	}
	if len(nodes) != 2 {
		t.Errorf("Round 4: expected 2 nodes, got %d", len(nodes))
	} else {
		if nodes[0].Partial || nodes[1].Partial {
			t.Errorf("Round 4: expected both nodes to be complete")
		}
		if nodes[0].Content != "content1" {
			t.Errorf("Round 4: expected first content 'content1', got '%s'", nodes[0].Content)
		}
		if nodes[1].Content != "content2" {
			t.Errorf("Round 4: expected second content 'content2', got '%s'", nodes[1].Content)
		}
	}
}

// TestMultiRoundAppendBreakEveryChar tests breaking at every single character
func TestMultiRoundAppendBreakEveryChar(t *testing.T) {
	parser := NewStreamXmlParser()
	fullInput := "Before <tool name=\"test\">content here</tool> After"

	// Append one character at a time and verify state
	for i := 0; i < len(fullInput); i++ {
		parser.Append(string(fullInput[i]))
		println(fullInput[0:i])
		// Get current state
		nodes, _ := parser.GetXmlNodes()

		// After tag starts (position 7), should have at least one node
		if i >= 7 && i < 45 {
			if len(nodes) == 0 {
				t.Errorf("Round %d: expected at least one node", i+1)
			} else {
				// At position 44 (after closing '>'), node should be complete
				if i == 44 && nodes[0].Partial {
					t.Errorf("Round %d: expected complete node", i+1)
				}
			}
		}
	}

	// Final verification
	text, _ := parser.GetText()
	nodes, _ := parser.GetXmlNodes()
	expectedFinalText := "Before  After"
	if text != expectedFinalText {
		t.Errorf("Final: expected '%s', got '%s'", expectedFinalText, text)
	}
	if len(nodes) != 1 {
		t.Errorf("Final: expected 1 node, got %d", len(nodes))
	} else {
		if nodes[0].Partial {
			t.Errorf("Final: expected complete node")
		}
		if nodes[0].Name != "tool" {
			t.Errorf("Final: expected name 'tool', got '%s'", nodes[0].Name)
		}
		if nodes[0].Content != "content here" {
			t.Errorf("Final: expected content 'content here', got '%s'", nodes[0].Content)
		}
		if nodes[0].Attributes["name"] != "test" {
			t.Errorf("Final: expected attribute name='test', got '%s'", nodes[0].Attributes["name"])
		}
	}
}

// TestMultiRoundAppendComplexBreaks tests breaking at various complex positions
func TestMultiRoundAppendComplexBreaks(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Break in opening bracket
	parser.Append("Start <")
	text, _ := parser.GetText()
	if text != "Start " {
		t.Errorf("Round 1: expected 'Start ', got '%s'", text)
	}

	// Round 2: Break after tag name before space
	parser.Append("cmd")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || nodes[0].Name != "cmd" {
		t.Errorf("Round 2: expected node with name 'cmd'")
	}

	// Round 3: Break in space before attribute
	parser.Append(" ")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 3: expected 1 node")
	}

	// Round 4: Break in attribute name
	parser.Append("ty")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 4: expected 1 node")
	}

	// Round 5: Break after equals sign
	parser.Append("pe=")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 5: expected 1 node")
	}

	// Round 6: Break in opening quote
	parser.Append("\"")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 6: expected 1 node")
	}

	// Round 7: Break in attribute value
	parser.Append("exec")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 7: expected 1 node")
	}

	// Round 8: Break after closing quote
	parser.Append("\"")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 8: expected 1 node")
	}

	// Round 9: Break before closing bracket
	parser.Append(">")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || nodes[0].Partial != true {
		t.Errorf("Round 9: expected 1 partial node")
	}

	// Round 10: Add content broken in middle
	parser.Append("some co")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || nodes[0].Content != "some co" {
		t.Errorf("Round 10: expected content 'some co', got '%s'", nodes[0].Content)
	}

	// Round 11: Complete content
	parser.Append("mmand")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || nodes[0].Content != "some command" {
		t.Errorf("Round 11: expected content 'some command', got '%s'", nodes[0].Content)
	}

	// Round 12: Start closing tag
	parser.Append("</")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 || nodes[0].Partial != true {
		t.Errorf("Round 12: expected 1 partial node")
	}

	// Round 13: Complete closing tag
	parser.Append("cmd>")
	text, _ = parser.GetText()
	nodes, _ = parser.GetXmlNodes()
	if text != "Start " {
		t.Errorf("Round 13: expected 'Start ', got '%s'", text)
	}
	if len(nodes) != 1 {
		t.Errorf("Round 13: expected 1 node")
	} else {
		if nodes[0].Partial {
			t.Errorf("Round 13: expected complete node")
		}
		if nodes[0].Attributes["type"] != "exec" {
			t.Errorf("Round 13: expected type='exec', got '%s'", nodes[0].Attributes["type"])
		}
		if nodes[0].Content != "some command" {
			t.Errorf("Round 13: expected content 'some command', got '%s'", nodes[0].Content)
		}
	}
}

// TestMultiRoundAppendNestedTags tests breaking with nested tags
func TestMultiRoundAppendNestedTags(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Start outer tag
	parser.Append("<outer>")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}

	// Round 2: Add text and start inner tag
	parser.Append("text<inn")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 2: expected 1 node, got %d", len(nodes))
	}

	// Round 3: Complete inner tag opening
	parser.Append("er>inner content")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 3: expected 1 node")
	}

	// Round 4: Close inner tag
	parser.Append("</inner>more")
	nodes, _ = parser.GetXmlNodes()
	// Should still have outer tag open
	if len(nodes) != 1 || nodes[0].Partial != true {
		t.Errorf("Round 4: expected 1 partial outer node")
	}

	// Round 5: Close outer tag
	parser.Append(" text</outer>")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 5: expected 1 node")
	} else if nodes[0].Partial {
		t.Errorf("Round 5: expected complete node")
	}
}

// TestMultiRoundAppendEmptyContent tests empty content scenarios
func TestMultiRoundAppendEmptyContent(t *testing.T) {
	parser := NewStreamXmlParser()

	// Round 1: Open and immediately close
	parser.Append("<tag>")
	nodes, _ := parser.GetXmlNodes()
	if len(nodes) != 1 || !nodes[0].Partial {
		t.Errorf("Round 1: expected 1 partial node")
	}

	// Round 2: Close tag with no content
	parser.Append("</tag>")
	nodes, _ = parser.GetXmlNodes()
	if len(nodes) != 1 {
		t.Errorf("Round 2: expected 1 node")
	} else {
		if nodes[0].Partial {
			t.Errorf("Round 2: expected complete node")
		}
		if nodes[0].Content != "" {
			t.Errorf("Round 2: expected empty content, got '%s'", nodes[0].Content)
		}
	}
}
