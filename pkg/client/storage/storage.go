package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/mgutz/ansi"

	"github.com/AlekseyKas/keeper/pkg/client/storage/secrets"
	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
)

var (
	ErrHashValid = errors.New("hashsum is identical") // storage hashsum not changeg
	once         sync.Once
	_storage     Storage
)

func init() {
	gob.Register(&secrets.AnyBinary{})
	gob.Register(&secrets.AnyText{})
	gob.Register(&secrets.CreditCard{})
	gob.Register(&secrets.OTP{})
	gob.Register(&secrets.UserPass{})
}

// Secret - record interface
type Secret interface {
	Scope() string
	Value() any
}

// Keeper - local vault struct
type Keeper struct {
	secrets []Value
	logger  logging.Logger
	rwMutex *sync.RWMutex
	hashsum string
}

// Value - value secret struct
type Value struct {
	Name        string
	Description string
	Record      Secret
}

// New() - initialize local storage
func New() Storage {
	once.Do(func() {
		k := new(Keeper)
		k.logger = zerolog.New().WithPrefix("storage")
		k.rwMutex = new(sync.RWMutex)
		k.secrets = []Value{}
		_storage = k
	})
	return _storage
}

// HashSum - get current hashsum of secrets
func (k *Keeper) HashSum() string {
	return k.hashsum
}

// Save - save vault to disk
func (k *Keeper) Save() ([]byte, error) {
	if len(k.secrets) == 0 {
		return []byte{}, nil
	}
	k.rwMutex.RLock()
	defer k.rwMutex.RUnlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(k.secrets)
	if err != nil {
		return []byte{}, err
	}
	hash := helpers.GenHash(buf.Bytes())
	if hash == k.HashSum() {
		k.logger.Info(nil, "hashsum identical. skipping...")
		return []byte{}, ErrHashValid
	}
	k.hashsum = hash
	return buf.Bytes(), nil
}

// ReEncrypt - reencrypt saved vault
func (k *Keeper) ReEncrypt() ([]byte, error) {
	if len(k.secrets) == 0 {
		return []byte{}, nil
	}
	k.rwMutex.RLock()
	defer k.rwMutex.RUnlock()
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(k.secrets)
	if err != nil {
		return []byte{}, err
	}
	hash := helpers.GenHash(buf.Bytes())
	k.hashsum = hash
	return buf.Bytes(), nil
}

// Load - load to memory vault
func (k *Keeper) Load(b []byte) error {
	hash := helpers.GenHash(b)
	if hash == k.HashSum() {
		k.logger.Info(nil, "hashsum identical. skipping...")
		return ErrHashValid
	}
	k.rwMutex.Lock()
	defer k.rwMutex.Unlock()
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	v := []Value{}
	err := dec.Decode(&v)
	if err != nil {
		if err == io.EOF {
			k.secrets = []Value{}
			return nil
		}
		return err
	}
	k.hashsum = hash
	k.secrets = v
	return nil
}

// InsertSecret - insert new secret
func (k *Keeper) InsertSecret(name, description string, secret Secret) Storage {
	list := k.ListSecrets()
	if len(list) != 0 {
		if _, ok := list[name]; ok {
			k.logger.Warnf("Create new secret failed. Such secret name %s already exists", nil, ansi.Color(name, "red+b"))
			return k
		}
	}
	k.rwMutex.Lock()
	defer k.rwMutex.Unlock()
	k.secrets = append(k.secrets,
		Value{
			Name:        name,
			Description: description,
			Record:      secret,
		})
	return k
}

// DeleteSecret - delete secret
func (k *Keeper) DeleteSecret(name string) Storage {
	k.rwMutex.Lock()
	defer k.rwMutex.Unlock()
	kk := make([]Value, len(k.secrets)-1)
	var found bool
	for _, v := range k.secrets {
		if strings.EqualFold(name, v.Name) {
			found = true
			continue
		}
		kk = append(kk, v)
	}
	if !found {
		k.logger.Warnf("Delete secret failed. No such secret name %s", nil, ansi.Color(name, "red+b"))
		return k
	}
	k.secrets = kk
	return k
}

// GetSecret - get secret value
func (k *Keeper) GetSecret(name string) Secret {
	k.rwMutex.RLock()
	defer k.rwMutex.RUnlock()
	for _, v := range k.secrets {
		if strings.EqualFold(name, v.Name) {
			return v.Record
		}
	}
	k.logger.Warnf("Get secret failed. No such secret name %s", nil, ansi.Color(name, "red+b"))
	return nil
}

// ListSecrets - get list of secrets
func (k *Keeper) ListSecrets() map[string]string {
	k.rwMutex.RLock()
	defer k.rwMutex.RUnlock()
	list := make(map[string]string, len(k.secrets))
	for _, v := range k.secrets {
		list[v.Name] = v.Description
	}
	delete(list, "")
	return list
}
