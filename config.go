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

// ParserConfig holds configuration options for the StreamXmlParser
type ParserConfig struct {
	// MaxDepth limits the maximum nesting depth of XML elements (default: 100)
	MaxDepth int

	// MaxBufferSize limits the maximum size of the internal buffer in bytes (default: 10MB)
	MaxBufferSize int

	// AllowedElements specifies which XML elements should be parsed as XML.
	// If nil, all elements are allowed (default behavior).
	// If empty slice, no elements are allowed (all tags treated as text).
	AllowedElements []string

	// BufferCleanupThreshold determines when to cleanup consumed buffer data in bytes (default: 1KB)
	BufferCleanupThreshold int
}

// DefaultConfig returns the default parser configuration
func DefaultConfig() ParserConfig {
	return ParserConfig{
		MaxDepth:               100,
		MaxBufferSize:          10 * 1024 * 1024, // 10MB
		AllowedElements:        nil,              // Allow all elements
		BufferCleanupThreshold: 1024,             // 1KB
	}
}

// Validate checks if the configuration is valid
func (c ParserConfig) Validate() error {
	if c.MaxDepth < 1 {
		return ErrInvalidConfiguration
	}
	if c.MaxBufferSize < 1024 {
		return ErrInvalidConfiguration
	}
	if c.BufferCleanupThreshold < 0 {
		return ErrInvalidConfiguration
	}
	return nil
}
