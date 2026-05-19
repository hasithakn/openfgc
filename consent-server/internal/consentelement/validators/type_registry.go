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
	"fmt"
	"sync"
)

// TypeRegistry holds all registered element types
type TypeRegistry struct {
	mu      sync.RWMutex
	types   map[string]ElementType
}

var (
	// defaultRegistry is the global registry singleton
	defaultRegistry *TypeRegistry
)

// init registers all built-in element types at package init time
func init() {
	defaultRegistry = NewTypeRegistry()

	if err := defaultRegistry.Register(&BasicElementType{}); err != nil {
		panic(fmt.Sprintf("failed to register BasicElementType: %v", err))
	}
	if err := defaultRegistry.Register(&JSONElementType{}); err != nil {
		panic(fmt.Sprintf("failed to register JSONElementType: %v", err))
	}
	if err := defaultRegistry.Register(&XMLElementType{}); err != nil {
		panic(fmt.Sprintf("failed to register XMLElementType: %v", err))
	}
}

// NewTypeRegistry creates a new registry instance
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]ElementType),
	}
}

// Register adds an element type to the registry.
// Returns error if a type with the same key is already registered.
func (registry *TypeRegistry) Register(et ElementType) error {
	typeStr := et.GetType()

	registry.mu.Lock()
	defer registry.mu.Unlock()

	if _, exists := registry.types[typeStr]; exists {
		return fmt.Errorf("element type %q already registered", typeStr)
	}
	registry.types[typeStr] = et
	return nil
}

// Get retrieves an element type by its type string.
// Returns error if no type is registered for the given key.
func (registry *TypeRegistry) Get(typeStr string) (ElementType, error) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	et, exists := registry.types[typeStr]
	if !exists {
		return nil, fmt.Errorf("no element type registered: %q", typeStr)
	}
	return et, nil
}

// GetAllTypes returns a list of all registered element type keys.
func (registry *TypeRegistry) GetAllTypes() []string {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	types := make([]string, 0, len(registry.types))
	for typeStr := range registry.types {
		types = append(types, typeStr)
	}
	return types
}

// GetTypeRegistry returns the global registry singleton.
func GetTypeRegistry() *TypeRegistry {
	return defaultRegistry
}
