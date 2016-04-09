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

package atom

import (
	"time"

	"github.com/elliotpeele/deepfreeze/utils"
	"github.com/satori/go.uuid"
)

type Atom struct {
	Id         string    `json:"id"`
	MoleculeId string    `json:"molecule_id"`
	CubeId     string    `json:"cube_id"`
	PartId     int64     `json:"part_id"`
	Hash       string    `json:"hash"`
	CreatedAt  time.Time `json:"created_at"`
	Child      *Atom     `json:"-"`
	Delete     bool      `json:"delete"`
	Size       int64     `json:"size"`
}

func New(moleculeId string, cubeId string, size int64) *Atom {
	return &Atom{
		Id:         uuid.NewV4().String(),
		MoleculeId: moleculeId,
		CubeId:     cubeId,
		PartId:     0,
		CreatedAt:  time.Now(),
		Child:      nil,
		Delete:     false,
		Size:       size,
	}
}

func (a *Atom) Header() ([]byte, error) {
	return utils.ToJSON(a)
}
