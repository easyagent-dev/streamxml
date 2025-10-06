package streamxml

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenText           TokenType = iota
	TokenOpenBracket              // <
	TokenCloseBracket             // >
	TokenSlash                    // /
	TokenElementName              // element name
	TokenAttributeName            // attribute name
	TokenEquals                   // =
	TokenAttributeValue           // attribute value
	TokenIncomplete               // incomplete token
)

type Token struct {
	Type     TokenType
	Start    int
	End      int
	Complete bool
}

type StreamXmlTokenizer struct {
	buffer          string
	position        int
	allowedElements map[string]bool
	consumed        int

	// State tracking
	inTag        bool
	tagStartPos  int
	tagBuffer    strings.Builder
	textBuffer   strings.Builder
	textStartPos int

	// Pending tokens from a tag being parsed
	pendingTokens []*Token
	pendingIndex  int

	// Track if incomplete token was already returned
	incompleteReturned bool
}

func NewStreamXmlTokenizer() *StreamXmlTokenizer {
	return &StreamXmlTokenizer{
		buffer:          "",
		position:        0,
		allowedElements: nil, // nil means all elements are allowed
		consumed:        0,
		pendingTokens:   make([]*Token, 0),
		pendingIndex:    0,
	}
}

// SetAllowedElements configures which XML elements should be treated as XML tokens.
// If nil, all elements are allowed (default behavior).
// If empty slice, no elements are allowed (all tags treated as text).
// If set with elements, only those elements will be tokenized as XML; others will be treated as text.
func (t *StreamXmlTokenizer) SetAllowedElements(elements []string) {
	if elements == nil {
		t.allowedElements = nil
		return
	}

	// Empty slice means no elements allowed
	t.allowedElements = make(map[string]bool)
	for _, elem := range elements {
		t.allowedElements[elem] = true
	}
}

// Append adds more data to the tokenizer
func (t *StreamXmlTokenizer) Append(data string) {
	t.buffer += data
	// Reset incomplete flag when new data arrives
	t.incompleteReturned = false
}

// GetBuffer returns the current buffer for value extraction
func (t *StreamXmlTokenizer) GetBuffer() string {
	return t.buffer
}

// NextToken returns the next token from the buffer.
// Returns nil if no complete token is available yet.
func (t *StreamXmlTokenizer) NextToken() *Token {
	// First return any pending tokens
	if t.pendingIndex < len(t.pendingTokens) {
		token := t.pendingTokens[t.pendingIndex]
		t.pendingIndex++

		// Clear pending tokens if all consumed
		if t.pendingIndex >= len(t.pendingTokens) {
			t.pendingTokens = t.pendingTokens[:0]
			t.pendingIndex = 0
		}

		return token
	}

	// Try to get next token
	for t.position < len(t.buffer) {
		if t.inTag {
			if t.tryCompleteTag() {
				// Tag complete, check if we have pending tokens
				if t.pendingIndex < len(t.pendingTokens) {
					token := t.pendingTokens[t.pendingIndex]
					t.pendingIndex++

					if t.pendingIndex >= len(t.pendingTokens) {
						t.pendingTokens = t.pendingTokens[:0]
						t.pendingIndex = 0
					}

					return token
				}
			} else {
				// Tag incomplete, return nil or incomplete token
				break
			}
		} else {
			token := t.processText()
			if token != nil {
				return token
			}
		}
	}

	// Return incomplete text if any
	if t.textBuffer.Len() > 0 && !t.inTag {
		token := &Token{
			Type:     TokenText,
			Start:    t.textStartPos,
			End:      t.position,
			Complete: true, // Text at end of buffer is complete
		}
		// Reset the text buffer to avoid returning the same token repeatedly
		t.textBuffer.Reset()
		return token
	}

	// Return incomplete tag if any
	if t.inTag && t.tagBuffer.Len() > 0 && !t.incompleteReturned {
		t.incompleteReturned = true
		return &Token{
			Type:     TokenIncomplete,
			Start:    t.tagStartPos,
			End:      t.position,
			Complete: false,
		}
	}

	return nil
}

func (t *StreamXmlTokenizer) processText() *Token {
	for t.position < len(t.buffer) {
		ch := t.buffer[t.position]

		if ch == '<' {
			// Found start of potential XML tag
			var token *Token
			if t.textBuffer.Len() > 0 {
				// Return accumulated text as complete token
				token = &Token{
					Type:     TokenText,
					Start:    t.textStartPos,
					End:      t.position,
					Complete: true,
				}
				t.textBuffer.Reset()
			}

			// Switch to tag mode
			t.inTag = true
			t.tagStartPos = t.position
			t.tagBuffer.Reset()

			if token != nil {
				return token
			}

			// Continue to try completing the tag
			break
		} else {
			// Accumulate text
			if t.textBuffer.Len() == 0 {
				t.textStartPos = t.position
			}
			t.textBuffer.WriteByte(ch)
			t.position++
		}
	}

	return nil
}

func (t *StreamXmlTokenizer) tryCompleteTag() bool {
	// Try to parse the tag to completion
	// Look for the closing >
	for t.position < len(t.buffer) {
		ch := t.buffer[t.position]
		t.tagBuffer.WriteByte(ch)
		t.position++

		if ch == '>' {
			// Tag is complete, parse it
			tagContent := t.tagBuffer.String()
			t.parseAndEmitTag(tagContent)

			t.inTag = false
			t.tagBuffer.Reset()
			t.consumed = t.position
			return true
		}
	}

	// Tag is incomplete
	return false
}

func (t *StreamXmlTokenizer) parseAndEmitTag(tagContent string) {
	// Tag format: <name attr="value"> or </name> or <name/>
	if len(tagContent) < 2 {
		// Invalid tag, treat as text
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenText,
			Start:    t.tagStartPos,
			End:      t.tagStartPos + len(tagContent),
			Complete: true,
		})
		return
	}

	// Remove < and >
	inner := tagContent[1 : len(tagContent)-1]

	// Determine if closing tag
	isClosing := len(inner) > 0 && inner[0] == '/'
	if isClosing {
		inner = inner[1:]
	}

	// Determine if self-closing
	isSelfClosing := len(inner) > 0 && inner[len(inner)-1] == '/'
	if isSelfClosing {
		inner = inner[:len(inner)-1]
	}

	// Extract element name
	inner = strings.TrimSpace(inner)
	elementName := ""
	restOfTag := ""

	// Find where element name ends (space or end of string)
	spaceIdx := -1
	for i, ch := range inner {
		if unicode.IsSpace(ch) {
			spaceIdx = i
			break
		}
	}

	if spaceIdx >= 0 {
		elementName = inner[:spaceIdx]
		restOfTag = strings.TrimSpace(inner[spaceIdx+1:])
	} else {
		elementName = inner
	}

	// Check if element is allowed
	if t.allowedElements != nil && !t.allowedElements[elementName] {
		// Not in allowed list, treat entire tag as text
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenText,
			Start:    t.tagStartPos,
			End:      t.tagStartPos + len(tagContent),
			Complete: true,
		})
		return
	}

	// Element is allowed, emit detailed tokens
	currentPos := t.tagStartPos

	// Emit <
	t.pendingTokens = append(t.pendingTokens, &Token{
		Type:     TokenOpenBracket,
		Start:    currentPos,
		End:      currentPos + 1,
		Complete: true,
	})
	currentPos++

	// Emit / for closing tag
	if isClosing {
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenSlash,
			Start:    currentPos,
			End:      currentPos + 1,
			Complete: true,
		})
		currentPos++
	}

	// Emit element name
	t.pendingTokens = append(t.pendingTokens, &Token{
		Type:     TokenElementName,
		Start:    currentPos,
		End:      currentPos + len(elementName),
		Complete: true,
	})
	currentPos += len(elementName)

	// Parse and emit attributes if present
	if restOfTag != "" {
		// Calculate position offset
		// We need to account for spaces between element name and attributes
		tagContentStart := t.tagStartPos + 1 // Skip <
		if isClosing {
			tagContentStart++ // Skip /
		}
		attrStartInTag := strings.Index(tagContent[tagContentStart-t.tagStartPos:], restOfTag)
		if attrStartInTag >= 0 {
			currentPos = tagContentStart + attrStartInTag
			t.parseAndEmitAttributes(restOfTag, currentPos)
		}
	}

	// Emit / for self-closing tag
	if isSelfClosing {
		slashPos := t.tagStartPos + len(tagContent) - 2 // Before >
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenSlash,
			Start:    slashPos,
			End:      slashPos + 1,
			Complete: true,
		})
	}

	// Emit >
	closeBracketPos := t.tagStartPos + len(tagContent) - 1
	t.pendingTokens = append(t.pendingTokens, &Token{
		Type:     TokenCloseBracket,
		Start:    closeBracketPos,
		End:      closeBracketPos + 1,
		Complete: true,
	})
}

func (t *StreamXmlTokenizer) parseAndEmitAttributes(attrStr string, startPos int) {
	i := 0
	currentPos := startPos

	for i < len(attrStr) {
		// Skip whitespace
		for i < len(attrStr) && unicode.IsSpace(rune(attrStr[i])) {
			i++
			currentPos++
		}

		if i >= len(attrStr) {
			break
		}

		// Find attribute name
		nameStart := i
		for i < len(attrStr) && attrStr[i] != '=' && !unicode.IsSpace(rune(attrStr[i])) {
			i++
		}

		if i >= len(attrStr) {
			break
		}

		nameLen := i - nameStart

		// Emit attribute name
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenAttributeName,
			Start:    currentPos,
			End:      currentPos + nameLen,
			Complete: true,
		})
		currentPos += nameLen

		// Skip whitespace to =
		for i < len(attrStr) && unicode.IsSpace(rune(attrStr[i])) {
			i++
			currentPos++
		}

		if i >= len(attrStr) || attrStr[i] != '=' {
			break
		}

		// Emit =
		t.pendingTokens = append(t.pendingTokens, &Token{
			Type:     TokenEquals,
			Start:    currentPos,
			End:      currentPos + 1,
			Complete: true,
		})
		i++
		currentPos++

		// Skip whitespace after =
		for i < len(attrStr) && unicode.IsSpace(rune(attrStr[i])) {
			i++
			currentPos++
		}

		if i >= len(attrStr) {
			break
		}

		// Parse value
		var valueLen int
		if attrStr[i] == '"' || attrStr[i] == '\'' {
			quote := attrStr[i]
			i++
			currentPos++ // Skip opening quote
			valueStart := i
			for i < len(attrStr) && attrStr[i] != quote {
				i++
			}
			valueLen = i - valueStart

			// Emit attribute value (without quotes)
			t.pendingTokens = append(t.pendingTokens, &Token{
				Type:     TokenAttributeValue,
				Start:    currentPos,
				End:      currentPos + valueLen,
				Complete: true,
			})
			currentPos += valueLen

			if i < len(attrStr) {
				i++ // Skip closing quote
				currentPos++
			}
		} else {
			// Value without quotes
			valueStart := i
			for i < len(attrStr) && !unicode.IsSpace(rune(attrStr[i])) {
				i++
			}
			valueLen = i - valueStart

			// Emit attribute value
			t.pendingTokens = append(t.pendingTokens, &Token{
				Type:     TokenAttributeValue,
				Start:    currentPos,
				End:      currentPos + valueLen,
				Complete: true,
			})
			currentPos += valueLen
		}
	}
}
