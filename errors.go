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

import "errors"

// Error definitions for the parser
var (
	// ErrMaxDepthExceeded is returned when XML nesting exceeds the maximum allowed depth
	ErrMaxDepthExceeded = errors.New("maximum XML nesting depth exceeded")

	// ErrMaxBufferSizeExceeded is returned when the internal buffer exceeds the maximum allowed size
	ErrMaxBufferSizeExceeded = errors.New("maximum buffer size exceeded")

	// ErrInvalidConfiguration is returned when parser configuration is invalid
	ErrInvalidConfiguration = errors.New("invalid parser configuration")
)
