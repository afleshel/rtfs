package krab

import (
	"bytes"
	"errors"
	"strings"

	"github.com/RTradeLtd/crypto"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/ipfs/go-ipfs-keystore"
	ci "github.com/libp2p/go-libp2p-crypto"
)

// Krab is used to manage an encrypted IPFS keystore
type Krab struct {
	em *crypto.EncryptManager
	ds *badger.Datastore
	keystore.Keystore
}

// Opts is used to configure a Krab keystore
type Opts struct {
	Passphrase string
	DSPath     string
}

// NewKrab is used to create a new krab ipfs keystore manager
func NewKrab(opts Opts) (*Krab, error) {
	ds, err := badger.NewDatastore(opts.DSPath, &badger.DefaultOptions)
	if err != nil {
		return nil, err
	}
	return &Krab{
		em: crypto.NewEncryptManager(opts.Passphrase),
		ds: ds,
	}, nil
}

// Has is used to check whether or not the given key name exists
func (km *Krab) Has(name string) (bool, error) {
	return km.ds.Has(ds.NewKey(name))
}

// Put is used to store a key in our keystore
func (km *Krab) Put(name string, privKey ci.PrivKey) error {
	if has, err := km.Has(name); err != nil {
		return err
	} else if has {
		return errors.New("key with name already exists")
	}
	pkBytes, err := privKey.Bytes()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(pkBytes)
	// encrypt the private key
	encryptedPK, err := km.em.Encrypt(reader)
	if err != nil {
		return err
	}
	return km.ds.Put(ds.NewKey(name), encryptedPK)
}

// Get is used to retrieve a key from our keystore
func (km *Krab) Get(name string) (ci.PrivKey, error) {
	encryptedPKBytes, err := km.ds.Get(ds.NewKey(name))
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(encryptedPKBytes)
	pkBytes, err := km.em.Decrypt(reader)
	if err != nil {
		return nil, err
	}
	return ci.UnmarshalPrivateKey(pkBytes)
}

// Delete is used to remove a key from our keystore
func (km *Krab) Delete(name string) error {
	return km.ds.Delete(ds.NewKey(name))
}

// List is used to list all key identifiers in our keystore
func (km *Krab) List() ([]string, error) {
	entries, err := km.ds.Query(query.Query{})
	if err != nil {
		return nil, err
	}
	keys, err := entries.Rest()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, v := range keys {
		ids = append(ids, strings.Split(v.Key, "/")[1])
	}
	return ids, nil
}
