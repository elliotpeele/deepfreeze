/*
 * Copyright (c) Elliot Peele <elliot@bentlogic.net>
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
 */

package molecule

import "testing"

func TestMoleculeCompress(t *testing.T) {
	m, err := New("../testdata/foo",
		"84e99c21df3d69d6bcb82420dc1c5ab9e877aa19ca516fa2644cd2f1e6c35840")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Open(); err != nil {
		t.Fatal(err)
	}

	if err := m.Compress(); err != nil {
		t.Fatal(err)
	}
}
