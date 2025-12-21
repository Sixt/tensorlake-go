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
	"fmt"
)

// UnionValues is a union of values of type T.
// It can be a single value or an array of values.
type UnionValues[T any] []T

// UnmarshalJSON unmarshals a JSON array or a single value into a UnionValues.
func (v *UnionValues[T]) UnmarshalJSON(b []byte) error {
	// Try a single value
	var single T
	if err := json.Unmarshal(b, &single); err == nil {
		*v = []T{single}
		return nil
	}

	// Try an array of values
	var arr []T
	if err := json.Unmarshal(b, &arr); err == nil {
		*v = arr
		return nil
	}

	return fmt.Errorf("value must be a single value or an array of values: %s", string(b))
}

// MarshalJSON marshals a UnionValues into a JSON array.
func (v UnionValues[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal([]T(v))
}
