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

// High level package around reading and writing tar archives.
// A tarfile can either be written to or read from, not both.
package tarfile

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	readError  = fmt.Errorf("can not read from write only file")
	writeError = fmt.Errorf("can not write to read only file")
)

// Interface for storing archive metadata.
type MetadataStore interface {
	WriteMetadata(name string, data []byte) (n int, err error)
	ReadMetadata() (md *metadataRecord, err error)
}

// Interface for reading tarfiles.
type FileReader interface {
	ReadFile(w io.Writer) (info os.FileInfo, err error)
}

// Interface for writing tarfiles.
type FileWriter interface {
	WriteFile(info os.FileInfo, r io.Reader) (n int, err error)
}

// Interface for reading and writing tarfiles.
type FileReadWriter interface {
	FileReader
	FileWriter
}

// Private structure for storing record metadata.
type metadataRecord struct {
	Name string
	Data []byte
}

// High level structure for handling tarfiles.
type TarFile struct {
	w    *tar.Writer
	r    *tar.Reader
	size int64
}

// Create a new tar file for writing.
func New(w io.Writer) *TarFile {
	return &TarFile{
		w:    tar.NewWriter(w),
		r:    nil,
		size: 0,
	}
}

// Open a tar file for reading
func Open(r io.Reader) *TarFile {
	return &TarFile{
		w:    nil,
		r:    tar.NewReader(r),
		size: 0,
	}
}

// Close tarfile, flushes content to underlying writer.
func (tf *TarFile) Close() error {
	if tf.w == nil {
		return writeError
	}
	return tf.w.Close()
}

// Finish writin current file data.
func (tf *TarFile) Flush() error {
	if tf.w == nil {
		return writeError
	}
	return tf.w.Flush()
}

// Get the current size of the underlying writer.
func (tf *TarFile) Size() int64 {
	return tf.size
}

// Read content of next archive file into specified writer.
func (tf *TarFile) ReadFile(w io.Writer) (info os.FileInfo, err error) {
	if tf.r == nil {
		return nil, readError
	}

	header, err := tf.r.Next()
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(w, tf.r); err != nil {
		return nil, err
	}

	return header.FileInfo(), nil
}

// Write content of reader to tarfile.
func (tf *TarFile) WriteFile(info os.FileInfo, r io.Reader) (n int, err error) {
	if tf.w == nil {
		return 0, writeError
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return 0, err
	}

	if err := tf.w.WriteHeader(header); err != nil {
		return 0, err
	}

	written, err := io.Copy(tf.w, r)
	if err != nil {
		return 0, err
	}

	tf.size += written

	return int(written), nil
}

// Read metadata from the tarfile.
func (tf *TarFile) ReadMetadata() (md *metadataRecord, err error) {
	if tf.r == nil {
		return nil, readError
	}

	header, err := tf.r.Next()
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, tf.r); err != nil {
		return nil, err
	}

	md = &metadataRecord{
		Name: header.Name,
		Data: buf.Bytes(),
	}

	return md, nil
}

// Write metadata to the tarfile.
func (tf *TarFile) WriteMetadata(name string, data []byte) (n int, err error) {
	if tf.w == nil {
		return 0, writeError
	}

	header := &tar.Header{
		Name: name,
		Size: int64(len(data)),
	}

	if err := tf.w.WriteHeader(header); err != nil {
		return 0, err
	}

	n, err = tf.w.Write(data)
	if err != nil {
		return 0, err
	}
	tf.size += int64(n)
	return n, nil
}
