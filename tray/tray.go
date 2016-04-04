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
	"github.com/satori/go.uuid"
)

type Tray struct {
	Id          string       `json:"tray_id"`
	CreatedAt   time.Time    `json:"created_at"`
	IsUploaded  bool         `json:"-"`
	Hash        string       `json:"-"`
	Full        bool         `json:"full"`
	Incremental bool         `json:"incremental"`
	Parent      *Tray        `json:"-"`
	UploadedAt  time.Time    `json:"-"`
	Cubes       []*cube.Cube `json:"-"`
	Size        int64        `json:"-"`
}

func New() (*Tray, error) {
	c, err := cube.New(1024)
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
		Cubes: []*cube.Cube{
			c,
		},
	}
	c.TrayId = t.Id
	if err := t.packHeader(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Tray) CurrentCube() *cube.Cube {
	return t.Cubes[len(t.Cubes)-1]
}

func (t *Tray) NextCube() (*cube.Cube, error) {
	if err := t.CurrentCube().Close(); err != nil {
		return nil, err
	}
	c, err := t.CurrentCube().Next()
	if err != nil {
		return nil, err
	}
	t.Cubes = append(t.Cubes, c)
	return c, nil
}

func (t *Tray) WriteMolecule(m *molecule.Molecule) (n int, err error) {
	log.Debugf("Backing up %s", m.Path)
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
	s := int(m.Size()) // FIXME: This might be a problem when handling large files.
	written := 0
	for written != s {
		n, err := t.CurrentCube().WriteMolecule(m)
		if err != nil {
			return written + n, err
		}
		if t.CurrentCube().IsFull() {
			if _, err := t.NextCube(); err != nil {
				return written + n, err
			}
		}
		written += n
	}

	return written, nil
}

func (t *Tray) Upload() error {
	return nil
}

// Write header to current cube.
func (t *Tray) packHeader() error {
	return nil
}

// Read header from current cube.
func (t *Tray) unpackHeader() error {
	return nil
}
