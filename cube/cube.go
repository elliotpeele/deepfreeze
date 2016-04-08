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
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/elliotpeele/deepfreeze/atom"
	"github.com/elliotpeele/deepfreeze/fileinfo"
	"github.com/elliotpeele/deepfreeze/log"
	"github.com/elliotpeele/deepfreeze/molecule"
	"github.com/elliotpeele/deepfreeze/tarfile"
	"github.com/satori/go.uuid"
)

// Cubes are the actual files that get uploaded to Glacier.
type Cube struct {
	Id          string               `json:"id"`
	TrayId      string               `json:"tray_id"`
	CreatedAt   time.Time            `json:"created_at"`
	Hash        string               `json:"hash"`
	AWSLocation string               `json:"aws_location"`
	UploadedAt  time.Time            `json:"uploaded_at"`
	Parent      *Cube                `json:"-"`
	Child       *Cube                `json:"-"`
	Molecules   []*molecule.Molecule `json:"-"`
	Atoms       []*atom.Atom         `json:"-"`
	Size        int64                `json:"size"`

	backingfile *os.File
	backupdir   string
	tf          *tarfile.TarFile
	max_size    int64
}

func New(size int64, backupdir string) (*Cube, error) {
	id := uuid.NewV4().String()
	fobj, err := ioutil.TempFile(backupdir, id)
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
	}, nil
}

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

func (c *Cube) WriteMolecule(m *molecule.Molecule) (n int, err error) {
	// Make sure there is enough space to store some of the file.
	orig_size := c.tf.Size()
	if c.max_size-orig_size < 0 {
		return 0, fmt.Errorf("not enough space left to write file")
	}

	// Write the molecule header.
	molHeader, err := m.Header()
	if err != nil {
		return 0, err
	}
	if _, err := c.tf.WriteMetadata("molecule", molHeader); err != nil {
		return 0, err
	}

	// Write the file info for the original file so that it can be restored later.
	finfo, err := fileinfo.NewFileInfo(m.OrigInfo()).ToJSON()
	if err != nil {
		return 0, err
	}
	if _, err := c.tf.WriteMetadata("finfo", finfo); err != nil {
		return 0, err
	}

	// Write the current file contents.
	size := c.max_size - c.tf.Size()
	if size > m.Size() {
		size = m.Size()
	}
	log.Debugf("attempting to write %d, info size is %d", size, m.Info().Size())
	lr := &io.LimitedReader{
		R: m,
		N: size,
	}
	if _, err := c.tf.WriteFile(m.Info(), lr); err != nil {
		return 0, err
	}

	return int(c.tf.Size() - orig_size), nil
}

func (c *Cube) Close() error {
	return c.tf.Close()
}

func (c *Cube) Next() (*Cube, error) {
	if c.Child == nil {
		c2, err := New(c.Size, c.backupdir)
		if err != nil {
			return nil, err
		}
		c.Child = c2
		c.Child.Parent = c
	}
	return c.Child, nil
}

func (c *Cube) IsFull() bool {
	return c.tf.Size() >= c.max_size
}

func (c *Cube) Freeze() error {
	if err := c.packHeader(); err != nil {
		return err
	}
	return nil
}

func (c *Cube) packHeader() error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(c); err != nil {
		return err
	}
	_, err := c.tf.WriteMetadata("cube", buf.Bytes())
	if err != nil {
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

// Get the size that a tar header would take up given a set of file information.
func (c *Cube) getTarHeaderSize(info os.FileInfo) (int64, error) {
	// Pack the file info into a header buffer to see if it fits in the cube.
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return 0, err
	}
	if err := tw.WriteHeader(header); err != nil {
		return 0, err
	}
	if err := tw.Flush(); err != nil {
		return 0, err
	}
	return int64(buf.Len()), nil
}
