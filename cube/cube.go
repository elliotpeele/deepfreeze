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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/elliotpeele/deepfreeze/atom"
	"github.com/elliotpeele/deepfreeze/molecule"
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
	tarw        *tar.Writer
	tarr        *tar.Reader
	remaining   int64
}

func New(size int64) (*Cube, error) {
	id := uuid.NewV4().String()
	fobj, err := ioutil.TempFile("", id)
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
		tarw:        tar.NewWriter(fobj),
		tarr:        nil,
		remaining:   size * 1024 * 1024, // Size in bytes
	}, nil
}

func Open(name string) (*Cube, error) {
	fobj, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	c := &Cube{
		backingfile: fobj,
		tarw:        nil,
		tarr:        tar.NewReader(fobj),
	}
	if err := c.unpackHeader(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Cube) Read(p []byte) (n int, err error) {
	if c.tarr == nil {
		return 0, fmt.Errorf("can not read from write only cube")
	}
	return 0, nil
}

func (c *Cube) Write(p []byte) (n int, err error) {
	if c.tarw == nil {
		return 0, fmt.Errorf("can not write to read only cube")
	}
	return 0, nil
}

func (c *Cube) WriteMolecule(m *molecule.Molecule) (n int, err error) {
	if c.tarw == nil {
		return 0, fmt.Errorf("can not write to read only cube")
	}

	// Info is the file info object for the original file.
	size, err := c.getTarHeaderSize(m.Info())
	if err != nil {
		return 0, err
	}
	// Can't fit header plus some data in this cube, go to next cube.
	if size >= c.remaining {
		return 0, nil
	}

	// Create header from file info.
	header, err := tar.FileInfoHeader(m.Info(), "")
	if err != nil {
		return 0, err
	}
	// Write tar header
	if err := c.tarw.WriteHeader(header); err != nil {
		return 0, err
	}

	return 0, nil
}

func (c *Cube) Seek(offset int64, whence int) (int64, error) {
	return c.backingfile.Seek(offset, whence)
}

func (c *Cube) Close() error {
	if c.tarw != nil {
		if err := c.tarw.Close(); err != nil {
			return err
		}
	}
	return c.backingfile.Close()
}

func (c *Cube) Next() (*Cube, error) {
	if c.Child == nil {
		c2, err := New(c.Size)
		if err != nil {
			return nil, err
		}
		c.Child = c2
		c.Child.Parent = c
	}
	return c.Child, nil
}

func (c *Cube) IsFull() bool {
	return c.remaining == 0
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
	binary.Write(c, binary.BigEndian, buf.Len())
	c.Write(buf.Bytes())
	return nil
}

func (c *Cube) unpackHeader() error {
	var size int
	if err := binary.Read(c, binary.BigEndian, &size); err != nil {
		return err
	}
	buf := make([]byte, size)
	n, err := c.Read(buf)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("Read the wrong size, found %d instead of %d", n, size)
	}
	if err := json.Unmarshal(buf, c); err != nil {
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
