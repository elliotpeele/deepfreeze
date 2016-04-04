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

package tarfile

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/elliotpeele/deepfreeze/fileinfo"
)

func TestWriteTarFile(t *testing.T) {
	dest, err := ioutil.TempFile("", "testsuite")
	if err != nil {
		t.Fatal(err)
	}

	src, err := os.Open("../testdata/foo")
	if err != nil {
		t.Fatal(err)
	}

	info, err := src.Stat()
	if err != nil {
		t.Fatal(err)
	}

	tf := New(dest)
	n, err := tf.WriteFile(info, src)
	if err != nil {
		t.Fatal(err)
	}

	if err := tf.Close(); err != nil {
		t.Fatal(err)
	}

	if n != 11358 {
		t.Fatalf("wrote unexpected ammount: %d", n)
	}
}

func TestWriteMetadataTarFile(t *testing.T) {
	dest, err := ioutil.TempFile("", "testsuite")
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat("../testdata/foo")
	if err != nil {
		t.Fatal(err)
	}

	finfo, err := fileinfo.NewFileInfo(info).ToJSON()
	if err != nil {
		t.Fatal(err)
	}

	tf := New(dest)
	n, err := tf.WriteMetadata("test", finfo)
	if err != nil {
		t.Fatal(err)
	}

	if err := tf.Close(); err != nil {
		t.Fatal(err)
	}

	if n != 450 {
		t.Fatalf("wrote unexpected ammount: %d", n)
	}
}
