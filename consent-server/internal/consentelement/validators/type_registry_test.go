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

// TestNewTypeRegistry tests creating a new registry
func TestNewTypeRegistry(t *testing.T) {
	registry := NewTypeRegistry()
	require.NotNil(t, registry)
	require.NotNil(t, registry.types)
	require.Equal(t, 0, len(registry.types))
}

// TestRegister tests registering handlers
func TestRegister(t *testing.T) {
	testCases := []struct {
		name          string
		handlers      []ElementType
		expectError   bool
		errorContains string
	}{
		{
			name:        "Register single handler",
			handlers:    []ElementType{&BasicElementType{}},
			expectError: false,
		},
		{
			name: "Register multiple different handlers",
			handlers: []ElementType{
				&BasicElementType{},
				&JSONElementType{},
				&XMLElementType{},
			},
			expectError: false,
		},
		{
			name: "Register duplicate handler",
			handlers: []ElementType{
				&BasicElementType{},
				&BasicElementType{},
			},
			expectError:   true,
			errorContains: "already registered",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewTypeRegistry()
			var err error

			for _, handler := range tc.handlers {
				err = registry.Register(handler)
				if err != nil {
					break
				}
			}

			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.handlers), len(registry.types))
			}
		})
	}
}

// TestGet tests retrieving handlers from registry
func TestGet(t *testing.T) {
	testCases := []struct {
		name          string
		setupHandlers []ElementType
		getType       string
		expectError   bool
		expectedType  string
	}{
		{
			name:          "Get existing handler",
			setupHandlers: []ElementType{&BasicElementType{}},
			getType:       "basic",
			expectError:   false,
			expectedType:  "basic",
		},
		{
			name:          "Get non-existent handler",
			setupHandlers: []ElementType{&BasicElementType{}},
			getType:       "non-existent-type",
			expectError:   true,
		},
		{
			name: "Get from multiple handlers",
			setupHandlers: []ElementType{
				&BasicElementType{},
				&JSONElementType{},
				&XMLElementType{},
			},
			getType:      "json",
			expectError:  false,
			expectedType: "json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewTypeRegistry()
			for _, handler := range tc.setupHandlers {
				_ = registry.Register(handler)
			}

			handler, err := registry.Get(tc.getType)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, handler)
				require.Contains(t, err.Error(), "no element type registered")
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
				require.Equal(t, tc.expectedType, handler.GetType())
			}
		})
	}
}

// TestGetAllTypes tests retrieving all registered types
func TestGetAllTypes(t *testing.T) {
	testCases := []struct {
		name          string
		setupHandlers []ElementType
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "Empty registry",
			setupHandlers: []ElementType{},
			expectedCount: 0,
			expectedTypes: []string{},
		},
		{
			name:          "Single handler",
			setupHandlers: []ElementType{&BasicElementType{}},
			expectedCount: 1,
			expectedTypes: []string{"basic"},
		},
		{
			name: "Multiple handlers",
			setupHandlers: []ElementType{
				&BasicElementType{},
				&JSONElementType{},
				&XMLElementType{},
			},
			expectedCount: 3,
			expectedTypes: []string{"basic", "json", "xml"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewTypeRegistry()
			for _, handler := range tc.setupHandlers {
				_ = registry.Register(handler)
			}

			types := registry.GetAllTypes()
			require.Equal(t, tc.expectedCount, len(types))

			for _, expectedType := range tc.expectedTypes {
				require.Contains(t, types, expectedType)
			}
		})
	}
}


// TestGetTypeRegistry tests the global registry getter
func TestGetTypeRegistry(t *testing.T) {
	registry := GetTypeRegistry()
	require.NotNil(t, registry)
	
	// Should be the same instance on multiple calls
	registry2 := GetTypeRegistry()
	require.Equal(t, registry, registry2)
	
	// Should have handlers from init()
	types := registry.GetAllTypes()
	require.Equal(t, 3, len(types))
}
