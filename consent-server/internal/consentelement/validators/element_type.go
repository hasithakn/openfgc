/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package validators

import "encoding/json"

// ValidationError represents a single validation error for a property
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ElementType defines behavior for a specific consent element type
type ElementType interface {
	// GetType returns the type string this element type manages (e.g., "basic", "json", "xml")
	GetType() string

	// ValidateSchema validates the element schema field for this type.
	// Returns a ValidationError if the schema is invalid or missing when required, nil otherwise.
	ValidateSchema(schema *string) *ValidationError

	// ValidateProperties checks type-specific property constraints.
	ValidateProperties(properties map[string]string) []ValidationError
}

// Helper function to validate JSON string
func isValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
