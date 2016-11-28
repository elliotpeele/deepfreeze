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

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/elliotpeele/deepfreeze/atom"
	"github.com/elliotpeele/deepfreeze/log"
	"github.com/elliotpeele/deepfreeze/utils"
	"github.com/satori/go.uuid"
)

// Container for storing backed up files. Handles splitting content
// between cubes.
type Molecule struct {
	Id           string       `json:"id"`
	Path         string       `json:"path"`
	Hash         string       `json:"hash"`
	Atoms        []*atom.Atom `json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	OriginalSize int64        `json:"size"`

	cur_size        int64
	orig_info       os.FileInfo
	cur_info        os.FileInfo
	fobj            *os.File
	read_size       int64
	delete_on_close bool
}

// Create a new molecule.
func New(path string, hash string) (*Molecule, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &Molecule{
		Id:              uuid.NewV4().String(),
		Path:            path,
		Hash:            hash,
		CreatedAt:       time.Now(),
		OriginalSize:    info.Size(),
		orig_info:       info,
		cur_info:        info,
		delete_on_close: false,
	}, nil
}

// Open the file to be backed up.
func (m *Molecule) Open() error {
	log.Debugf("opening %s", m.Path)
	f, err := os.Open(m.Path)
	if err != nil {
		return err
	}
	m.fobj = f
	return nil
}

// Read the file being backed up.
func (m *Molecule) Read(p []byte) (n int, err error) {
	n, err = m.fobj.Read(p)
	m.read_size += int64(n)
	return
}

// Seek the backup file.
func (m *Molecule) Seek(offset int64, whence int) (int64, error) {
	return m.fobj.Seek(offset, whence)
}

// Close the backup file.
func (m *Molecule) Close() error {
	if err := m.fobj.Close(); err != nil {
		return err
	}
	if m.delete_on_close {
		log.Debugf("deleting %s", m.fobj.Name())
		if err := os.Remove(m.fobj.Name()); err != nil {
			return err
		}
	}
	return nil
}

// Get the remaining size to be read.
func (m *Molecule) Size() int64 {
	return m.cur_size - m.read_size
}

// Get the current file info.
func (m *Molecule) Info() os.FileInfo {
	return m.cur_info
}

// Get the original file info.
func (m *Molecule) OrigInfo() os.FileInfo {
	return m.orig_info
}

// Serialize the molecule header.
func (m *Molecule) Header() ([]byte, error) {
	return utils.ToJSON(m)
}

// Create a new atom instance.
func (m *Molecule) NewAtom(cubeId string, size int64) *atom.Atom {
	a := atom.New(m.Id, cubeId, size)
	m.Atoms = append(m.Atoms, a)
	return a
}

// Encrypt the file contents.
func (m *Molecule) Encrypt() error {
	log.Debugf("encrypting %s", m.Path)
	log.Warn("encrytion not supported")
	return nil
}

// Compress file contents.
func (m *Molecule) Compress() error {
	log.Debugf("compressing %s", m.Path)
	// Get a tmp file to compress into.
	tmpf, err := ioutil.TempFile("", "deepfreeze")
	if err != nil {
		return err
	}
	// Compress orignal file.
	w, err := gzip.NewWriterLevel(tmpf, gzip.BestSpeed)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, m.fobj); err != nil {
		return err
	}
	// Flush the compressor once complete.
	if err := w.Close(); err != nil {
		return err
	}
	// Rewind tmp file.
	if _, err := tmpf.Seek(0, 0); err != nil {
		return err
	}
	// Check and store size.
	info, err := tmpf.Stat()
	if err != nil {
		return err
	}
	m.cur_size = info.Size()
	m.cur_info = info
	// Close underlying file object.
	m.fobj.Close()
	// Replace with tmp file.
	m.fobj = tmpf
	// Mark tmp file for deletion.
	m.delete_on_close = true
	return nil
}
