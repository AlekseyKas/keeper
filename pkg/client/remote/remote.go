package remote

import "github.com/AlekseyKas/keeper/pkg/models"

// Actions - interface of actions with server
type Actions interface {
	Login(username, password, public string) error
	Register(username, password, public string) error
	NewStream() error

	Push(payload, hashsum string) error
	Pull(checksum string) ([]byte, error)
	Versions() ([]models.Version, error)
	GetLogs() ([]models.Action, error)

	ListPGP() ([]models.PGP, error)
	AddPGP(publickey string) error
	ConfirmPGP(publickey string) error
	RevokePGP(publickey string) error

	Close() error
	Delete() error
}
