// Copyright 2025 SIXT SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tensorlake

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestValueOrValuesUnmarshalJSON(t *testing.T) {
	tests := []struct {
		value    string
		expected []int
	}{
		{
			value:    "1",
			expected: []int{1},
		},
		{
			value:    "[1, 2, 3]",
			expected: []int{1, 2, 3},
		},
	}

	for _, test := range tests {
		var v UnionValues[int]
		if err := json.Unmarshal([]byte(test.value), &v); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if !reflect.DeepEqual(v, UnionValues[int](test.expected)) {
			t.Fatalf("expected %v, got %v", test.expected, v)
		}
	}
}

func TestValueOrValuesMarshalJSON(t *testing.T) {
	type testType struct {
		Value UnionValues[int] `json:"value"`
	}

	tests := []struct {
		value    string
		expected testType
	}{
		{
			value: `{"value": 1}`,
			expected: testType{
				Value: UnionValues[int]{1},
			},
		},
		{
			value: `{"value": [1, 2, 3]}`,
			expected: testType{
				Value: UnionValues[int]{1, 2, 3},
			},
		},
	}

	for _, test := range tests {
		var v testType
		if err := json.Unmarshal([]byte(test.value), &v); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if !reflect.DeepEqual(v, test.expected) {
			t.Fatalf("expected %+v, got %+v", test.expected, v)
		}
	}
}
