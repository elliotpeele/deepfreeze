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

package tray

import (
	"time"

	"github.com/elliotpeele/deepfreeze/cube"
	"github.com/elliotpeele/deepfreeze/log"
	"github.com/elliotpeele/deepfreeze/molecule"
	"github.com/elliotpeele/deepfreeze/utils"
	"github.com/satori/go.uuid"
)

// Top level structure for any single backup. Contains reference to
// previous tray.
type Tray struct {
	Id          string       `json:"tray_id"`
	CreatedAt   time.Time    `json:"created_at"`
	IsUploaded  bool         `json:"-"`
	Hash        string       `json:"-"`
	Full        bool         `json:"full"`
	Incremental bool         `json:"incremental"`
	Parent      *Tray        `json:"-"`
	UploadedAt  time.Time    `json:"-"`
	Size        int64        `json:"size"`
	Cubes       []*cube_data `json:"cubes"`
	rootCube    *cube.Cube
	curCube     *cube.Cube
	backupdir   string
}

// Structure for storing cube metadata.
type cube_data struct {
	Id    string       `json:"cube_id"`
	Hash  string       `json:"hash"`
	Files []*file_data `json:"files"`
}

// Structure for storing file metaadata.
type file_data struct {
	Id   string `json:"file_id"`
	Hash string `json:"hash"`
	Path string `json:"path"`
}

// Create a new tray instance.
func New(backupdir string) (*Tray, error) {
	c, err := cube.New(1024, backupdir)
	if err != nil {
		return nil, err
	}
	t := &Tray{
		Id:          uuid.NewV4().String(),
		IsUploaded:  false,
		Full:        true,
		Incremental: false,
		Parent:      nil,
		Size:        0,
		rootCube:    c,
		backupdir:   backupdir,
	}
	c.TrayId = t.Id
	return t, nil
}

// Get the current cube from the tray.
func (t *Tray) CurrentCube() *cube.Cube {
	cur := t.rootCube
	for cur.Child != nil {
		cur = cur.Child
	}
	return cur
}

// Write a molecule to the tray.
func (t *Tray) WriteMolecule(m *molecule.Molecule) (n int, err error) {
	log.Infof("Backing up %s", m.Path)
	// Open the backend file, hopefully it still exists.
	if err := m.Open(); err != nil {
		return 0, err
	}
	// Compress molecule content.
	if err := m.Compress(); err != nil {
		return 0, err
	}
	// Encrypt molecule content.
	if err := m.Encrypt(); err != nil {
		return 0, err
	}

	// Pack molecule into cubes.
	return t.CurrentCube().WriteMolecule(m)
}

// Upload a frozen tray.
func (t *Tray) Upload() error {
	return nil
}

// Write header to current cube.
func (t *Tray) Header() ([]byte, error) {
	log.Debug("packing tray header")
	cube := t.rootCube
	for cube != nil {
		log.Debugf("packing cube %s", cube.Id)
		c := &cube_data{
			Id:   cube.Id,
			Hash: cube.Hash,
		}
		for _, mol := range cube.Molecules {
			f := &file_data{
				Id:   mol.Id,
				Hash: mol.Hash,
				Path: mol.Path,
			}
			log.Debugf("packing file %s", f.Id)
			c.Files = append(c.Files, f)
		}
		t.Cubes = append(t.Cubes, c)
		cube = cube.Child
	}

	return utils.ToJSON(t)
}

// Read header from current cube.
func (t *Tray) unpackHeader() error {
	return nil
}
