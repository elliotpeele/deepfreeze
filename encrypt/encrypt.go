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

package encrypt

import (
	"crypto"
	"io"
	"os"
	"path"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

// TODO: Figure out how to add a passphrase to the secret key.

const key_name = "deepfreeze"
const key_desc = "This is the encryption key used by deepfreeze to encrypt/decrypt backup data"

const pubring_name = "pubring"
const secring_name = "secring"

type EncryptionManager struct {
	ringDir string
}

func New(ringDir string) (*EncryptionManager, error) {
	return &EncryptionManager{
		ringDir: ringDir,
	}, nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	if err == nil {
		return true
	}
	return false
}

func (em *EncryptionManager) GenKey() error {
	secRingPath := path.Join(em.ringDir, secring_name)
	pubRingPath := path.Join(em.ringDir, pubring_name)

	config := &packet.Config{
		DefaultHash: crypto.SHA256,
	}

	ent, err := openpgp.NewEntity(key_name, key_desc, "", config)
	if err != nil {
		return err
	}

	// If the keyrings already exist, don't regenerate.
	if exists(secRingPath) && exists(pubRingPath) {
		return nil
	}

	// Write out secret keyring
	f, err := os.Create(secRingPath)
	if err != nil {
		return err
	}
	if err := ent.SerializePrivate(f, nil); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	// Write out public keyring
	f, err = os.Create(pubRingPath)
	if err != nil {
		return err
	}
	if err := ent.Serialize(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func readKeyRingFromFile(path string) (openpgp.EntityList, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	entList, err := openpgp.ReadKeyRing(f)
	if err != nil {
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	return entList, nil
}

func getPubRing(prefix string) (openpgp.EntityList, error) {
	return readKeyRingFromFile(path.Join(prefix, pubring_name))
}

func getSecRing(prefix string) (openpgp.EntityList, error) {
	return readKeyRingFromFile(path.Join(prefix, secring_name))
}

func (em *EncryptionManager) Encrypt(w io.Writer) (io.WriteCloser, error) {
	pubRing, err := getPubRing(em.ringDir)
	if err != nil {
		return nil, err
	}
	return openpgp.Encrypt(w, pubRing, nil, nil, nil)
}

func (em *EncryptionManager) Decrypt(r io.Reader) (io.Reader, error) {
	secRing, err := getSecRing(em.ringDir)
	if err != nil {
		return nil, err
	}
	md, err := openpgp.ReadMessage(r, secRing, nil, nil)
	if err != nil {
		return nil, err
	}
	return md.UnverifiedBody, nil
}
