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

func TestBasicElementType_GetType(t *testing.T) {
	et := &BasicElementType{}
	require.Equal(t, "basic", et.GetType())
}

func TestBasicElementType_ValidateSchema(t *testing.T) {
	et := &BasicElementType{}

	// Schema is optional for basic — any value including nil is accepted.
	require.Nil(t, et.ValidateSchema(nil))
	s := ""
	require.Nil(t, et.ValidateSchema(&s))
	s = `{"type":"string"}`
	require.Nil(t, et.ValidateSchema(&s))
	s = "not json at all"
	require.Nil(t, et.ValidateSchema(&s))
}

func TestBasicElementType_ValidateProperties(t *testing.T) {
	et := &BasicElementType{}

	// Basic type accepts any properties — none are required.
	require.Nil(t, et.ValidateProperties(nil))
	require.Nil(t, et.ValidateProperties(map[string]string{}))
	require.Nil(t, et.ValidateProperties(map[string]string{
		"validationSchema": `{"type":"string"}`,
		"resourcePath":     "/accounts",
		"jsonPath":         "Data.amount",
	}))
	require.Nil(t, et.ValidateProperties(map[string]string{"unknownProp": "value"}))
}

