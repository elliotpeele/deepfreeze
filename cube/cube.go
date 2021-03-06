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

package cube

import (
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/elliotpeele/deepfreeze/fileinfo"
	"github.com/elliotpeele/deepfreeze/log"
	"github.com/elliotpeele/deepfreeze/molecule"
	"github.com/elliotpeele/deepfreeze/tarfile"
	"github.com/elliotpeele/deepfreeze/utils"
	"github.com/satori/go.uuid"
)

// Cubes are the actual files that get uploaded to Glacier. These
// contain entire or parts of backed up files.
type Cube struct {
	Id          string               `json:"id"`
	TrayId      string               `json:"tray_id"`
	CreatedAt   time.Time            `json:"created_at"`
	Hash        string               `json:"hash"`
	AWSLocation string               `json:"aws_location"`
	UploadedAt  time.Time            `json:"uploaded_at"`
	ParentId    string               `json:"parent_id"`
	ChildId     string               `json:"child_id"`
	Parent      *Cube                `json:"-"`
	Child       *Cube                `json:"-"`
	Molecules   []*molecule.Molecule `json:"-"`
	Size        int64                `json:"size"`

	backingfile *os.File
	backupdir   string
	tf          *tarfile.TarFile
	max_size    int64
	size        int64
}

// Create a new cube isntance.
func New(size int64, backupdir string) (*Cube, error) {
	id := uuid.NewV4().String()
	fobj, err := os.Create(path.Join(backupdir, id))
	if err != nil {
		return nil, err
	}
	return &Cube{
		Id:        id,
		CreatedAt: time.Now(),
		Parent:    nil,
		Child:     nil,
		Size:      0,

		backingfile: fobj,
		backupdir:   backupdir,
		tf:          tarfile.New(fobj),
		max_size:    size * 1024 * 1024, // Size in bytes
		size:        size,
	}, nil
}

// Create a cube instance from backing store file.
func Open(name string) (*Cube, error) {
	fobj, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	c := &Cube{
		backingfile: fobj,
		tf:          tarfile.Open(fobj),
	}
	if err := c.unpackHeader(); err != nil {
		return nil, err
	}
	return c, nil
}

// Write a molecule to the cube backing store.
func (c *Cube) WriteMolecule(m *molecule.Molecule) (n int, err error) {
	cur := c

	// Add molecule to cube.
	c.Molecules = append(c.Molecules, m)

	// Make sure there is enough space to store some of the file.
	orig_size := cur.tf.Size()
	if cur.max_size-orig_size < 0 {
		return 0, fmt.Errorf("not enough space left to write file")
	}

	// Write the molecule header.
	molHeader, err := m.Header()
	if err != nil {
		return 0, err
	}
	if _, err := cur.tf.WriteMetadata("molecule", molHeader); err != nil {
		return 0, err
	}

	// Write the file info for the original file so that it can be restored later.
	finfo, err := fileinfo.NewFileInfo(m.OrigInfo()).ToJSON()
	if err != nil {
		return 0, err
	}
	if _, err := cur.tf.WriteMetadata("finfo", finfo); err != nil {
		return 0, err
	}

	// Write the file info for the backing file.
	bfinfo, err := fileinfo.NewFileInfo(m.Info()).ToJSON()
	if err != nil {
		return 0, err
	}
	if _, err := cur.tf.WriteMetadata("bfinfo", bfinfo); err != nil {
		return 0, err
	}

	// Write the current file contents.
	written := int64(0)
	for m.Size() > 0 {
		size := cur.max_size - cur.tf.Size()
		if size > m.Size() {
			size = m.Size()
		}

		// Create a new atom
		a := m.NewAtom(cur.Id, size)

		// Write the atom metadata
		atomHeader, err := a.Header()
		if err != nil {
			return 0, err
		}
		if _, err := cur.tf.WriteMetadata("atom", atomHeader); err != nil {
			return 0, err
		}

		log.Debugf("attempting to write %d, info size is %d", size, m.Info().Size())
		lr := &io.LimitedReader{
			R: m,
			N: size,
		}
		info := &fileinfo.FileInfo{
			Name: a.Id,
			Size: size,
		}
		if _, err := cur.tf.WriteFile(info.FileInfo(), lr); err != nil {
			return 0, err
		}

		written += size

		if cur.IsFull() {
			log.Debug("moving to next cube")
			next, err := cur.Next()
			if err != nil {
				return int(written), err
			}
			if err := cur.Close(); err != nil {
				return 0, err
			}
			cur = next
			log.Debugf("cur: %v, next: %v", cur, next)
		}

		log.Debugf("written: %d", written)
	}

	return int(c.tf.Size() - orig_size), nil
}

// Close and finalize the cube.
func (c *Cube) Close() error {
	// Copy data to cube structure.
	if c.Parent != nil {
		c.ParentId = c.Parent.Id
	}
	if c.Child != nil {
		c.ChildId = c.ChildId
	}
	c.Size = c.tf.Size()

	// Close the tarfile abstraction.
	if err := c.tf.Close(); err != nil {
		return err
	}

	// Rewind backing file so that it can be hashed.
	if _, err := c.backingfile.Seek(0, 0); err != nil {
		return err
	}
	// Hash cube.
	h := sha512.New()
	if _, err := io.Copy(h, c.backingfile); err != nil {
		return err
	}
	c.Hash = fmt.Sprintf("%x", h.Sum(nil))

	// Create tmp file for writing cube header.
	tmpf, err := ioutil.TempFile(c.backupdir, "deepfreeze")
	if err != nil {
		return err
	}

	// Serialize cube.
	data, err := utils.ToJSON(c)
	if err != nil {
		return err
	}

	// Write out cube header to tmp file.
	tf := tarfile.New(tmpf)
	if _, err := tf.WriteMetadata("cube", data); err != nil {
		return err
	}
	if err := tf.Flush(); err != nil {
		return err
	}

	// Rewind the backingfile so that it can be copied.
	if _, err := c.backingfile.Seek(0, 0); err != nil {
		return err
	}

	// Copy the backingfile into the tmp file.
	if _, err := io.Copy(tmpf, c.backingfile); err != nil {
		return err
	}

	// Close the tmp file.
	if err := tmpf.Close(); err != nil {
		return err
	}

	// Close the backingfile.
	if err := c.backingfile.Close(); err != nil {
		return err
	}

	// Rename tmp file to backing file.
	if err := os.Rename(tmpf.Name(), c.backingfile.Name()); err != nil {
		return err
	}

	return nil
}

// Create and return the next cube instance when this one is full.
func (c *Cube) Next() (*Cube, error) {
	if c.Child == nil {
		c2, err := New(c.size, c.backupdir)
		if err != nil {
			return nil, err
		}
		c.Child = c2
		c.Child.Parent = c
		c.Child.TrayId = c.TrayId
	}
	return c.Child, nil
}

// Check if the cube has reached the max size.
func (c *Cube) IsFull() bool {
	return c.tf.Size() >= c.max_size
}

// Freeze the current cube.
func (c *Cube) Freeze() error {
	if err := c.packHeader(); err != nil {
		return err
	}
	return nil
}

func (c *Cube) packHeader() error {
	b, err := utils.ToJSON(c)
	if err != nil {
		return err
	}
	if _, err := c.tf.WriteMetadata("cube", b); err != nil {
		return err
	}
	return nil
}

func (c *Cube) unpackHeader() error {
	md, err := c.tf.ReadMetadata()
	if md.Name != "cube" {
		return fmt.Errorf("expected cube metadata, found %s", md.Name)
	}
	if err != nil {
		return err
	}
	if err := json.Unmarshal(md.Data, c); err != nil {
		return err
	}
	return nil
}
