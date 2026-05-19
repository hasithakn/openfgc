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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidJSON(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid JSON object", `{"key":"value"}`, true},
		{"Valid JSON array", `["item1","item2"]`, true},
		{"Valid JSON string", `"string"`, true},
		{"Valid JSON number", `123`, true},
		{"Valid JSON boolean", `true`, true},
		{"Valid JSON null", `null`, true},
		{"Valid complex JSON", `{"type":"object","properties":{"name":{"type":"string"}}}`, true},
		{"Valid nested JSON", `{"outer":{"inner":{"deep":"value"}}}`, true},
		{"Invalid JSON - missing quotes", `{key:value}`, false},
		{"Invalid JSON - trailing comma", `{"key":"value",}`, false},
		{"Invalid JSON - missing closing brace", `{"key":"value"`, false},
		{"Invalid JSON - single quotes", `{'key':'value'}`, false},
		{"Invalid JSON - plain text", `not json at all`, false},
		{"Empty string", ``, false},
		{"Whitespace only", `   `, false},
		{"Valid JSON with whitespace", `  {"key": "value"}  `, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, isValidJSON(tc.input))
		})
	}
}
