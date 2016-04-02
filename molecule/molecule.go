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
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/elliotpeele/deepfreeze/atom"
	"github.com/elliotpeele/deepfreeze/log"
	"github.com/satori/go.uuid"
	"github.com/ulikunitz/xz"
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

	finfo os.FileInfo
	fobj  *os.File
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
	return m.fobj.Read(p)
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
		return m.EncryptedSize
	case m.CompressedSize != 0:
		return m.CompressedSize
	case m.OriginalSize != 0:
		return m.OriginalSize
	default:
		return -1
	}
}

func (m *Molecule) Info() os.FileInfo {
	return m.finfo
}

func (m *Molecule) Header() (string, error) {

	return "", nil
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
	w := xz.NewWriter(tmpf)
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
	// Close underlying file object.
	m.fobj.Close()
	// Replace with tmp file.
	m.fobj = tmpf
	return nil
}
