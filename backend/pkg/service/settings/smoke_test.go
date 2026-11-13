/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package settings

import "testing"

func TestSmokeDocServerVersion(t *testing.T) {
	ok := []string{"8.2.0", "8.3.1.10", "9.0.0"}
	bad := []string{"8.1.3", "7.5.0", "", "x.y"}
	for _, v := range ok {
		if err := validateDocServerVersion(v); err != nil {
			t.Errorf("%q should be supported: %v", v, err)
		}
	}
	for _, v := range bad {
		if err := validateDocServerVersion(v); err == nil {
			t.Errorf("%q should be rejected", v)
		}
	}
}
