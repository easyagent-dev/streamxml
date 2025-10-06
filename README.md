# StreamXML Parser

A Go library for parsing streaming mixed text and XML content, particularly useful for processing LLM (Large Language Model) stream outputs that contain both text and XML tags.

## Features

- **Stateful Tokenizer**: Supports multi-round `Append()` operations with state preservation
- **Partial Token Support**: Returns partial tokens for incomplete XML tags
- **AST Generation**: Builds an Abstract Syntax Tree for mixed text/XML content
- **Incremental Updates**: AST is automatically updated as more data is appended
- **Multiple XML Fragments**: Supports parsing multiple XML nodes in a single stream
- **Incomplete XML Handling**: Gracefully handles incomplete XML syntax with partial nodes

## Installation

```bash
go get github.com/easymvp/streamxml
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/easymvp/streamxml"
)

func main() {
    parser := streamxml.NewStreamXmlParser()
    
    // Stream data in chunks
    parser.Append("I must call tools to get more information.\n")
    parser.Append("<use-tool name=\"get_info\">\n")
    parser.Append("{\"name\":\"product\"}\n")
    parser.Append("</use-tool>")
    
    // Get all text (excluding XML tags)
    text, _ := parser.GetText()
    fmt.Println("Text:", text)
    
    // Get all XML nodes
    nodes, _ := parser.GetXmlNodes()
    for _, node := range nodes {
        fmt.Printf("XML Node: name=%s, partial=%v, content=%s\n", 
            node.Name, node.Partial, node.Content)
    }
}
```

### Handling Partial/Incomplete XML

```go
parser := streamxml.NewStreamXmlParser()

// First append - incomplete XML tag
parser.Append("<use-tool name=\"get")

nodes, _ := parser.GetXmlNodes()
if len(nodes) > 0 {
    fmt.Printf("Partial: %v\n", nodes[0].Partial) // true
}

// Complete the tag
parser.Append("_info\">\ncontent\n</use-tool>")

nodes, _ = parser.GetXmlNodes()
if len(nodes) > 0 {
    fmt.Printf("Partial: %v\n", nodes[0].Partial) // false
    fmt.Printf("Content: %s\n", nodes[0].Content) // "content\n"
}
```

### Multiple XML Fragments

```go
parser := streamxml.NewStreamXmlParser()

parser.Append("Text before.\n")
parser.Append("<tool name=\"search\">query</tool>\n")
parser.Append("Text between.\n")
parser.Append("<tool name=\"read\">file.txt</tool>")

nodes, _ := parser.GetXmlNodes()
fmt.Printf("Found %d XML nodes\n", len(nodes)) // 2
```

## API Reference

### StreamXmlParser

#### `NewStreamXmlParser() *StreamXmlParser`
Creates a new parser instance.

#### `Append(data string)`
Appends new data to the parser. The parser maintains state across multiple `Append()` calls and automatically updates the AST.

#### `GetText() (string, error)`
Returns all accumulated text content, excluding XML tags.

#### `GetXmlNode() (*XmlNode, error)`
Returns the first XML node (complete or partial).

#### `GetXmlNodes() ([]*XmlNode, error)`
Returns all XML nodes found in the stream (both complete and partial).

#### `GetAST() []ASTNode`
Returns the complete Abstract Syntax Tree.

### XmlNode

```go
type XmlNode struct {
    Name       string            // Tag name
    Attributes map[string]string // Tag attributes
    Content    string            // Inner content
    Partial    bool              // Whether node is incomplete
    StartPos   int               // Start position in stream
    EndPos     int               // End position in stream
}
```

### ASTNode

```go
type ASTNode struct {
    Type     ASTNodeType // ASTNodeText or ASTNodeXml
    Text     string      // Text content (if Type is ASTNodeText)
    XmlNode  *XmlNode    // XML node (if Type is ASTNodeXml)
    Position int         // Position in stream
}
```

### StreamXmlTokenizer

The tokenizer is used internally by the parser but can also be used standalone:

#### `NewStreamXmlTokenizer() *StreamXmlTokenizer`
Creates a new tokenizer instance.

#### `Append(data string)`
Appends new data to the tokenizer and processes it into tokens.

#### `GetTokens() []Token`
Returns all tokens including partial/incomplete ones.

## Implementation Details

### Stateful Multi-Round Append

The tokenizer maintains internal state across multiple `Append()` calls:
- `buffer`: Accumulates all received data
- `position`: Current parsing position
- `inXmlTag`: Whether currently inside an XML tag
- `currentTagBuf`: Buffer for incomplete XML tag
- `textBuffer`: Buffer for accumulating text

This allows the parser to handle streaming data where XML tags may be split across multiple chunks.

### AST Construction

The parser builds an AST that reflects the structure of mixed text/XML content:
1. Text nodes for plain text content
2. XML nodes for complete or partial XML elements
3. Automatic updates when new data arrives

### Partial Token Handling

When an XML tag is incomplete (e.g., `<use-tool name="get`), the tokenizer:
1. Keeps the incomplete tag in its buffer
2. Returns it as a partial token with `Complete: false`
3. Continues parsing when more data is appended
4. Updates the token to complete when the closing `>` is received

## Example Output Format

For LLM stream output like:
```
I must call tools to get more information.
<use-tool name="get_info">
{"name":"product"}
</use-tool>
```

The parser generates:
- **Text nodes**: "I must call tools to get more information.\n"
- **XML nodes**: 
  - Name: "use-tool"
  - Attributes: {"name": "get_info"}
  - Content: "\n{\"name\":\"product\"}\n"
  - Partial: false

## License

MIT License - see LICENSE file for details.
