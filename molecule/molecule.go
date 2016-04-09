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
	"github.com/satori/go.uuid"
)

type Molecule struct {
	Id             string       `json:"id"`
	Path           string       `json:"path"`
	Hash           string       `json:"hash"`
	Atoms          []*atom.Atom `json:"-"`
	CreatedAt      time.Time    `json:"created_at"`
	OriginalSize   int64        `json:"size"`
	CompressedSize int64        `json:"compressed_size"`
	EncryptedSize  int64        `json:"encrypted_size"`

	finfo           os.FileInfo
	compressed_info os.FileInfo
	encrypted_info  os.FileInfo
	fobj            *os.File
	read_size       int64
}

func New(path string, hash string) (*Molecule, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &Molecule{
		Id:           uuid.NewV4().String(),
		Path:         path,
		Hash:         hash,
		CreatedAt:    time.Now(),
		OriginalSize: info.Size(),
		finfo:        info,
	}, nil
}

func (m *Molecule) Open() error {
	log.Debugf("opening %s", m.Path)
	f, err := os.Open(m.Path)
	if err != nil {
		return err
	}
	m.fobj = f
	return nil
}

func (m *Molecule) Read(p []byte) (n int, err error) {
	n, err = m.fobj.Read(p)
	m.read_size += int64(n)
	return
}

func (m *Molecule) Seek(offset int64, whence int) (int64, error) {
	return m.fobj.Seek(offset, whence)
}

func (m *Molecule) Close() error {
	return m.fobj.Close()
}

func (m *Molecule) Size() int64 {
	switch {
	case m.EncryptedSize != 0:
		return m.EncryptedSize - m.read_size
	case m.CompressedSize != 0:
		return m.CompressedSize - m.read_size
	case m.OriginalSize != 0:
		return m.OriginalSize - m.read_size
	default:
		return -1
	}
}

func (m *Molecule) Info() os.FileInfo {
	switch {
	case m.encrypted_info != nil:
		log.Debug("using encrypted info")
		return m.encrypted_info
	case m.compressed_info != nil:
		log.Debug("using compressed info")
		return m.compressed_info
	default:
		log.Debug("using orignal info")
		return m.finfo
	}
}

func (m *Molecule) OrigInfo() os.FileInfo {
	return m.finfo
}

func (m *Molecule) Header() ([]byte, error) {
	return []byte{}, nil
}

func (m *Molecule) NewAtom(cubeId string, size int64) *atom.Atom {
	a := atom.New(m.Id, cubeId, size)
	m.Atoms = append(m.Atoms, a)
	return a
}

func (m *Molecule) Encrypt() error {
	log.Debugf("encrypting %s", m.Path)
	log.Warn("encrytion not supported")
	return nil
}

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
	m.CompressedSize = info.Size()
	m.compressed_info = info
	// Close underlying file object.
	m.fobj.Close()
	// Replace with tmp file.
	m.fobj = tmpf
	return nil
}
