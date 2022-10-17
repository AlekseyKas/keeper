package storage

// Storage - storage interface
type Storage interface {
	Save() ([]byte, error)
	ReEncrypt() ([]byte, error)
	Load(b []byte) error
	InsertSecret(name, description string, secret Secret) Storage
	DeleteSecret(name string) Storage
	GetSecret(name string) Secret
	ListSecrets() map[string]string
	HashSum() string
}
