package storage

import (
	"net"
	"time"

	"github.com/AlekseyKas/keeper/pkg/models"
)

// Storage
type Storage interface {
	SignUp(username, password string, ip net.IP) error
	SignIn(user models.User, ip net.IP) error
	DeleteUser(username string, ip net.IP) error

	Push(username, checksum string, data string, ip net.IP) error
	Pull(username, checksum string, ip net.IP) ([]byte, error)
	Versions(username string, ip net.IP) ([]models.Version, error)
	SaveLog(username, action, checksum string, ip net.IP, date time.Time) error
	GetLog(username string, ip net.IP) ([]models.Action, error)
	ListPGP(username string, ip net.IP) ([]models.PGP, error)
	AddPGP(username, publickey string, confirm bool, ip net.IP) error
	RevokePGP(username, publickey string, ip net.IP) error
	ConfirmPGP(username, publickey string, ip net.IP) error

	Close() error
	Ping() error
}
