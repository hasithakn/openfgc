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

func TestXMLElementType_GetType(t *testing.T) {
	et := &XMLElementType{}
	require.Equal(t, "xml", et.GetType())
}

func TestXMLElementType_ValidateSchema(t *testing.T) {
	et := &XMLElementType{}

	// nil schema — required, must fail
	require.NotNil(t, et.ValidateSchema(nil))

	// empty string — must fail
	empty := ""
	require.NotNil(t, et.ValidateSchema(&empty))

	// non-empty value is accepted
	s := "<root/>"
	require.Nil(t, et.ValidateSchema(&s))

	s = "any non-empty string"
	require.Nil(t, et.ValidateSchema(&s))
}

func TestXMLElementType_ValidateProperties(t *testing.T) {
	et := &XMLElementType{}

	require.Nil(t, et.ValidateProperties(nil))
	require.Nil(t, et.ValidateProperties(map[string]string{}))
	require.Nil(t, et.ValidateProperties(map[string]string{"anyKey": "anyValue"}))
}

