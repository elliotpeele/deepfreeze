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

package fileinfo

import (
	"bytes"
	"encoding/json"
	"os"
	"time"
)

type FileInfo struct {
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
	Mode    os.FileMode `json:"mode"`
	ModTime time.Time   `json:"mod_time"`
	IsDir   bool        `json:"is_dir"`
	Sys     interface{} `json:"sys"`
}

type finfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
	sys     interface{}
}

func NewFileInfo(fi os.FileInfo) *FileInfo {
	return &FileInfo{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
		Sys:     fi.Sys(),
	}
}

func ParseFileInfo(buf []byte) (*FileInfo, error) {
	fi := &FileInfo{}
	if err := json.Unmarshal(buf, fi); err != nil {
		return nil, err
	}
	return fi, nil
}

func (fi *FileInfo) ToJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(fi); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (fi *FileInfo) FileInfo() os.FileInfo {
	return &finfo{
		name:    fi.Name,
		size:    fi.Size,
		mode:    fi.Mode,
		modTime: fi.ModTime,
		isDir:   fi.IsDir,
		sys:     fi.Sys,
	}
}

func (i *finfo) Name() string {
	return i.name
}

func (i *finfo) Size() int64 {
	return i.size
}

func (i *finfo) Mode() os.FileMode {
	return i.mode
}

func (i *finfo) ModTime() time.Time {
	return i.modTime
}

func (i *finfo) IsDir() bool {
	return i.isDir
}

func (i *finfo) Sys() interface{} {
	return i.sys
}
