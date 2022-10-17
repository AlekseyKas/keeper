package apigrpc

import "sync"

type APIandGRPC interface {
	Start(wg *sync.WaitGroup) error
	Stop() error
}
