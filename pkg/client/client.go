package client

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/AlekseyKas/keeper/pkg/client/cli"
	"github.com/AlekseyKas/keeper/pkg/client/config"
	"github.com/AlekseyKas/keeper/pkg/client/openpgp"
	"github.com/AlekseyKas/keeper/pkg/client/storage"
	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
)

// KeeperClient - keeper client struct
type KeeperClient struct {
	wg      *sync.WaitGroup
	log     logging.Logger
	config  *config.Config
	storage storage.Storage
	crypto  openpgp.OPENPGP
	cli     *cli.CLI
}

// New - returns struct of KeeperServer
func New() *KeeperClient {
	kc := new(KeeperClient)
	kc.log = zerolog.New().WithPrefix("client")
	kc.wg = new(sync.WaitGroup)
	kc.config = config.New()
	c, err := openpgp.New()
	if err != nil {
		kc.log.Fatal(err, "encryption subsystem failed")
	}
	kc.crypto = c
	if kc.config.Mode != "standalone" {
		if len(kc.config.Username) == 0 {
			kc.log.Debug(nil, "username not set")
			var u string
			fmt.Print("Enter username: ")
			_, err := fmt.Scanln(&u)
			if err != nil {
				kc.log.Fatal(err, "username read failed")
			}
			kc.config.Username = u
		}
		if len(kc.config.Password) == 0 {
			kc.log.Debug(nil, "password not set")
			pp, err := helpers.ReadSecret("Enter password: ")
			if err != nil {
				kc.log.Fatal(err, "password read failed")
			}
			fmt.Println()
			kc.config.Password = pp
		}
	}
	kc.storage = storage.New()
	storage, err := os.OpenFile(kc.config.Storage, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		kc.log.Fatal(err, "storage file can not be open")
	}
	buf, err := io.ReadAll(storage)
	if err != nil {
		kc.log.Fatal(err, "storage can not be read")
	}
	if len(buf) != 0 {
		msg, err := kc.crypto.DecryptWithKeys(buf)
		if err != nil {
			kc.log.Fatal(err, "storage can not be decrypted")
		}
		err = kc.storage.Load(msg)
		if err != nil {
			kc.log.Fatal(err, "storage can not be loaded")
		}
	}
	kc.cli = cli.New(kc.WG())
	return kc
}

// Start - start cleint
func (kc *KeeperClient) Start() {
	helpers.SaveTermState()
	kc.WG().Add(1)
	go kc.cli.Start()
	kc.WG().Wait()
	kc.log.Info(nil, "Keeper stopped")
	helpers.RestoreTermState()
	os.Exit(0)
}

// WG - return application waitgroup
func (ks *KeeperClient) WG() *sync.WaitGroup {
	return ks.wg
}
