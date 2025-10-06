package streamxml

import (
	"strings"
)

type ASTNodeType int

const (
	ASTNodeText ASTNodeType = iota
	ASTNodeXml
)

type ASTNode struct {
	Type     ASTNodeType
	Text     string
	XmlNode  *XmlNode
	Position int
}

type XmlNode struct {
	Name       string
	Attributes map[string]string
	Content    string
	Partial    bool
	StartPos   int
	EndPos     int
}

type StreamXmlParser struct {
	tokenizer      *StreamXmlTokenizer
	astNodes       []ASTNode
	xmlStack       []*XmlNode
	textParts      []string
	currentContent strings.Builder
	depth          int

	// Tag reconstruction state
	collectingTag bool
	tagTokens     []*Token
	tagStartPos   int

	// Track current incomplete node being built
	currentPartialNode *XmlNode
	partialNodeIndex   int
}

func NewStreamXmlParser() *StreamXmlParser {
	return &StreamXmlParser{
		tokenizer:          NewStreamXmlTokenizer(),
		astNodes:           make([]ASTNode, 0),
		xmlStack:           make([]*XmlNode, 0),
		textParts:          make([]string, 0),
		depth:              0,
		collectingTag:      false,
		tagTokens:          make([]*Token, 0),
		currentPartialNode: nil,
		partialNodeIndex:   -1,
	}
}

// SetAllowedElements configures which XML elements should be treated as XML tokens.
// If nil, all elements are allowed (default behavior).
// If empty slice, no elements are allowed (all tags treated as text).
// If set with elements, only those elements will be tokenized as XML; others will be treated as text.
func (p *StreamXmlParser) SetAllowedElements(elements []string) {
	p.tokenizer.SetAllowedElements(elements)
}

// Append adds new data to the parser and processes new tokens incrementally
func (p *StreamXmlParser) Append(data string) {
	p.tokenizer.Append(data)
	p.processNewTokens()
}

// processNewTokens processes new tokens from the tokenizer incrementally
func (p *StreamXmlParser) processNewTokens() {
	for {
		token := p.tokenizer.NextToken()
		if token == nil {
			// No more tokens available
			break
		}

		p.processToken(token)
	}
}

// getValue extracts the value from buffer using token positions
func (p *StreamXmlParser) getValue(token *Token) string {
	buffer := p.tokenizer.GetBuffer()
	if token.Start >= 0 && token.End <= len(buffer) {
		return buffer[token.Start:token.End]
	}
	return ""
}

// processToken processes a single token and updates the AST incrementally
func (p *StreamXmlParser) processToken(token *Token) {
	switch token.Type {
	case TokenText:
		value := p.getValue(token)
		if p.depth > 0 {
			// We're inside an XML tag, accumulate as content
			p.currentContent.WriteString(value)
			// Update content in current open node
			if len(p.xmlStack) > 0 {
				p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
			}
		} else {
			// We're outside XML tags, add as text node
			p.astNodes = append(p.astNodes, ASTNode{
				Type:     ASTNodeText,
				Text:     value,
				Position: token.Start,
			})
			p.textParts = append(p.textParts, value)
		}

	case TokenOpenBracket:
		// Start collecting tag tokens
		p.collectingTag = true
		p.tagTokens = []*Token{token}
		p.tagStartPos = token.Start

	case TokenSlash, TokenElementName, TokenAttributeName, TokenEquals, TokenAttributeValue:
		// Continue collecting tag tokens
		if p.collectingTag {
			p.tagTokens = append(p.tagTokens, token)
		}

	case TokenCloseBracket:
		// Tag is complete
		if p.collectingTag {
			p.tagTokens = append(p.tagTokens, token)
			p.processCompleteTag()
			p.collectingTag = false
			p.tagTokens = nil
		}

	case TokenIncomplete:
		// Incomplete token - this means we have an incomplete tag
		if !token.Complete {
			if p.depth == 0 {
				value := p.getValue(token)
				tagName := extractPartialTagName(value)

				// Check if we already have a partial node being built
				if p.currentPartialNode != nil && p.partialNodeIndex >= 0 {
					// Update existing partial node
					if tagName != "" && tagName != p.currentPartialNode.Name {
						p.currentPartialNode.Name = tagName
					}
				} else {
					// Create new partial node - even if no tag name yet
					xmlNode := &XmlNode{
						Name:       tagName,
						Partial:    true,
						Content:    "",
						Attributes: make(map[string]string),
						StartPos:   token.Start,
					}

					// Add to AST as partial
					p.astNodes = append(p.astNodes, ASTNode{
						Type:     ASTNodeXml,
						XmlNode:  xmlNode,
						Position: token.Start,
					})

					// Track this as current partial node
					p.currentPartialNode = xmlNode
					p.partialNodeIndex = len(p.astNodes) - 1
				}
			} else {
				// Inside a tag - check if this is a closing tag fragment
				value := p.getValue(token)

				if isClosingTagFragment(value) {
					// This is a closing tag start (</...)
					// Need to remove any trailing '<' that was previously added to content
					currentContentStr := p.currentContent.String()
					if strings.HasSuffix(currentContentStr, "<") {
						// Remove the trailing '<'
						p.currentContent.Reset()
						p.currentContent.WriteString(strings.TrimSuffix(currentContentStr, "<"))
						// Update content in current open node
						if len(p.xmlStack) > 0 {
							p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
						}
					}
				} else {
					// Not a closing tag, add to content
					p.currentContent.WriteString(value)
					// Update content in current open node
					if len(p.xmlStack) > 0 {
						p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
					}
				}
			}
		}
	}
}

// processCompleteTag processes a complete tag (reconstructed from tokens)
func (p *StreamXmlParser) processCompleteTag() {
	if len(p.tagTokens) < 3 {
		// Invalid tag (need at least <, name, >)
		return
	}

	// Determine tag type
	isClosing := false
	isSelfClosing := false
	elementName := ""
	attributes := make(map[string]string)

	i := 1 // Skip opening <

	// Check for closing tag
	if i < len(p.tagTokens) && p.tagTokens[i].Type == TokenSlash {
		isClosing = true
		i++
	}

	// Get element name
	if i < len(p.tagTokens) && p.tagTokens[i].Type == TokenElementName {
		elementName = p.getValue(p.tagTokens[i])
		i++
	}

	// Parse attributes
	for i < len(p.tagTokens)-1 { // -1 to exclude closing >
		if p.tagTokens[i].Type == TokenSlash {
			isSelfClosing = true
			i++
			continue
		}

		if p.tagTokens[i].Type == TokenAttributeName {
			attrName := p.getValue(p.tagTokens[i])
			i++

			// Expect =
			if i < len(p.tagTokens) && p.tagTokens[i].Type == TokenEquals {
				i++

				// Expect value
				if i < len(p.tagTokens) && p.tagTokens[i].Type == TokenAttributeValue {
					attributes[attrName] = p.getValue(p.tagTokens[i])
					i++
				}
			}
		} else {
			i++
		}
	}

	// Process based on tag type
	if isClosing {
		// Closing tag
		if p.depth > 0 {
			p.depth--
		}

		if p.depth == 0 && len(p.xmlStack) > 0 {
			// Closing top-level tag
			xmlNode := p.xmlStack[len(p.xmlStack)-1]
			p.xmlStack = p.xmlStack[:len(p.xmlStack)-1]

			xmlNode.Content = p.currentContent.String()
			xmlNode.EndPos = p.tagStartPos
			xmlNode.Partial = false

			// Update existing node if it was partial, or add new one
			if p.currentPartialNode == xmlNode && p.partialNodeIndex >= 0 {
				// Already in AST, just mark as complete
				p.currentPartialNode = nil
				p.partialNodeIndex = -1
			} else {
				// Add to AST
				p.astNodes = append(p.astNodes, ASTNode{
					Type:     ASTNodeXml,
					XmlNode:  xmlNode,
					Position: xmlNode.StartPos,
				})
			}

			// Reset content builder
			p.currentContent.Reset()
		} else if p.depth > 0 {
			// Nested closing tag - add to content as raw text
			p.currentContent.WriteString(p.reconstructTag())
			// Update content in current open node
			if len(p.xmlStack) > 0 {
				p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
			}
		}
	} else if isSelfClosing {
		// Self-closing tag
		if p.depth == 0 {
			// Top-level self-closing tag
			if p.currentPartialNode != nil && p.partialNodeIndex >= 0 {
				// Update existing partial node
				p.currentPartialNode.Name = elementName
				p.currentPartialNode.Attributes = attributes
				p.currentPartialNode.Partial = false
				p.currentPartialNode.EndPos = p.tagStartPos
				p.currentPartialNode = nil
				p.partialNodeIndex = -1
			} else {
				xmlNode := &XmlNode{
					Name:       elementName,
					Attributes: attributes,
					Partial:    false,
					Content:    "",
					StartPos:   p.tagStartPos,
					EndPos:     p.tagStartPos,
				}

				p.astNodes = append(p.astNodes, ASTNode{
					Type:     ASTNodeXml,
					XmlNode:  xmlNode,
					Position: p.tagStartPos,
				})
			}
		} else {
			// Nested self-closing tag - add to content as raw text
			p.currentContent.WriteString(p.reconstructTag())
			// Update content in current open node
			if len(p.xmlStack) > 0 {
				p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
			}
		}
	} else {
		// Opening tag
		if p.depth == 0 {
			// Check if we have a partial node to update
			if p.currentPartialNode != nil && p.partialNodeIndex >= 0 {
				// Update existing partial node with complete info
				p.currentPartialNode.Name = elementName
				p.currentPartialNode.Attributes = attributes

				// Push to stack if not already there
				if len(p.xmlStack) == 0 || p.xmlStack[len(p.xmlStack)-1] != p.currentPartialNode {
					p.xmlStack = append(p.xmlStack, p.currentPartialNode)
					p.currentContent.Reset()
					p.depth++
				}
			} else {
				// Top-level tag - create new XML node
				xmlNode := &XmlNode{
					Name:       elementName,
					Attributes: attributes,
					Partial:    true,
					StartPos:   p.tagStartPos,
				}

				// Add to AST immediately
				p.astNodes = append(p.astNodes, ASTNode{
					Type:     ASTNodeXml,
					XmlNode:  xmlNode,
					Position: p.tagStartPos,
				})

				// Track as current partial node
				p.currentPartialNode = xmlNode
				p.partialNodeIndex = len(p.astNodes) - 1

				// Push to stack for tracking
				p.xmlStack = append(p.xmlStack, xmlNode)
				p.currentContent.Reset()
				p.depth++
			}
		} else {
			// Nested tag - add to content as raw text
			p.currentContent.WriteString(p.reconstructTag())
			// Update content in current open node
			if len(p.xmlStack) > 0 {
				p.xmlStack[len(p.xmlStack)-1].Content = p.currentContent.String()
			}
			p.depth++
		}
	}
}

// reconstructTag reconstructs the full tag string from collected tokens
func (p *StreamXmlParser) reconstructTag() string {
	var result strings.Builder

	for _, token := range p.tagTokens {
		value := p.getValue(token)
		switch token.Type {
		case TokenOpenBracket:
			result.WriteString("<")
		case TokenCloseBracket:
			result.WriteString(">")
		case TokenSlash:
			result.WriteString("/")
		case TokenElementName:
			result.WriteString(value)
		case TokenAttributeName:
			result.WriteString(" ")
			result.WriteString(value)
		case TokenEquals:
			result.WriteString("=")
		case TokenAttributeValue:
			result.WriteString("\"")
			result.WriteString(value)
			result.WriteString("\"")
		}
	}

	return result.String()
}

// GetText returns all accumulated text (excluding XML tags)
func (p *StreamXmlParser) GetText() (string, error) {
	var result strings.Builder

	for _, node := range p.astNodes {
		if node.Type == ASTNodeText {
			result.WriteString(node.Text)
		}
	}

	return result.String(), nil
}

// GetXmlNode returns the first XML node (complete or partial)
func (p *StreamXmlParser) GetXmlNode() (*XmlNode, error) {
	for _, node := range p.astNodes {
		if node.Type == ASTNodeXml && node.XmlNode != nil {
			return node.XmlNode, nil
		}
	}
	return nil, nil
}

// GetXmlNodes returns all XML nodes (complete and partial)
func (p *StreamXmlParser) GetXmlNodes() ([]*XmlNode, error) {
	nodes := make([]*XmlNode, 0)

	for _, node := range p.astNodes {
		if node.Type == ASTNodeXml && node.XmlNode != nil {
			nodes = append(nodes, node.XmlNode)
		}
	}

	return nodes, nil
}

// GetAST returns the complete AST
func (p *StreamXmlParser) GetAST() []ASTNode {
	return p.astNodes
}

// extractPartialTagName tries to extract tag name from incomplete tag
func extractPartialTagName(tagValue string) string {
	if len(tagValue) < 2 {
		return ""
	}

	// Remove leading <
	content := strings.TrimPrefix(tagValue, "<")
	content = strings.TrimSpace(content)

	// Extract first word as tag name
	parts := strings.Fields(content)
	if len(parts) > 0 {
		return parts[0]
	}

	return content
}

// isClosingTagFragment checks if an incomplete token value looks like a closing tag fragment
func isClosingTagFragment(value string) bool {
	if len(value) == 0 {
		return false
	}

	// Only filter out if it's clearly a closing tag: "</", "</t", "</ta", etc.
	// A single "<" might just be content, so don't filter it
	if len(value) >= 2 && value[0] == '<' && value[1] == '/' {
		return true
	}

	return false
}
