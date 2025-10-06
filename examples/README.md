## StreamXML Examples

This directory contains example programs demonstrating various features of the StreamXML library.

### Basic Example

Location: `examples/basic/main.go`

Demonstrates:
- Creating a basic parser
- Streaming data in chunks
- Extracting text and XML nodes
- Accessing node attributes and content

Run:
```bash
cd examples/basic
go run main.go
```

### Advanced Example

Location: `examples/advanced/main.go`

Demonstrates:
- Custom parser configuration
- Setting maximum depth limits
- Filtering allowed XML elements
- Error handling
- Buffer size limits

Run:
```bash
cd examples/advanced
go run main.go
```

### Running All Examples

From the repository root:
```bash
# Basic example
go run ./examples/basic

# Advanced example
go run ./examples/advanced
```

### Key Concepts

#### Streaming Parsing
The parser maintains state across multiple `Append()` calls, allowing you to process data as it arrives:

```go
parser := streamxml.NewStreamXmlParser()
parser.Append("Hello ")
parser.Append("<tag>")
parser.Append("content")
parser.Append("</tag>")
```

#### Partial Nodes
When XML tags are incomplete, the parser creates partial nodes that are updated as more data arrives:

```go
parser.Append("<too")  // Partial node created
parser.Append("l>")    // Node completed with name "tool"
```

#### Configuration
Customize parser behavior with `ParserConfig`:

```go
config := streamxml.ParserConfig{
    MaxDepth:               50,
    MaxBufferSize:          5 * 1024 * 1024,
    AllowedElements:        []string{"tool", "thinking"},
    BufferCleanupThreshold: 512,
}
parser := streamxml.NewStreamXmlParserWithConfig(config)
