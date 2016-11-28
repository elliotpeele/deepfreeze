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

package indexer

import (
	"crypto/sha512"

	"github.com/elliotpeele/deepfreeze/log"
)

// File system indexer.
type Indexer struct {
	excludes []string
	root     string
}

// Create a new indexer instance.
func New(root string, excludes []string) *Indexer {
	return &Indexer{
		root:     root,
		excludes: excludes,
	}
}

// Index filesystem with content hashes.
func (idx *Indexer) Index() (map[string][sha512.Size]byte, error) {
	log.Infof("indexing directory tree")
	return HashAll(idx.root)
}
