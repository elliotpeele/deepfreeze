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

package freezer

import (
	"fmt"

	"github.com/elliotpeele/deepfreeze/indexer"
	"github.com/elliotpeele/deepfreeze/molecule"
	"github.com/elliotpeele/deepfreeze/tray"
)

type Freezer struct {
	tray      *tray.Tray
	indexer   *indexer.Indexer
	backupdir string
}

func New(root string, backupdir string, excludes []string) (*Freezer, error) {
	t, err := tray.New()
	if err != nil {
		return nil, err
	}
	return &Freezer{
		tray:      t,
		indexer:   indexer.New(root, excludes),
		backupdir: backupdir,
	}, nil
}

func (f *Freezer) Freeze() error {
	files, err := f.indexer.Index()
	if err != nil {
		return err
	}
	mols := make(map[string]*molecule.Molecule)
	for path, hash := range files {
		mol, err := molecule.New(path, fmt.Sprintf("%x", hash))
		if err != nil {
			return err
		}
		mols[path] = mol
	}

	t, err := tray.New()
	if err != nil {
		return err
	}

	for _, mol := range mols {
		if _, err := t.WriteMolecule(mol); err != nil {
			return err
		}
	}

	if err := t.CurrentCube().Close(); err != nil {
		return err
	}

	return nil
}
