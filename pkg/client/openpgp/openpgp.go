package openpgp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	pgpcrypto "github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/caarlos0/env"
	"github.com/denisbrodbeck/machineid"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
)

// passphrase - struct for pgp passhrase from env
type passpharse struct {
	PP string `env:"KEEPER_PGP_PASSPHRASE"`
}

var (
	log      logging.Logger                 // package logger
	errPGP   = errors.New("openpgp failed") // pgp error
	once     sync.Once                      // initialization with singletone pattern
	_keyring *OpenPGP                       // current keyring
)

// OpenPGP - openpgp endpoint struct
type OpenPGP struct {
	pubkey     string             // current public key of machine in armored format
	passphrase []byte             // current pgpg passphrase
	Public     *pgpcrypto.KeyRing // public keyring
	Private    *pgpcrypto.KeyRing // private keyring
}

type OPENPGP interface {
	GetPublicKey() string
	AddPublicKey(armored []byte) error
	AddPrivateKey(armored []byte) error
	GeneratePair() error
	ReloadPublicKeys(keys []string) error
	ReadFolder(name string) error
	EncryptWithKeys(data []byte) ([]byte, error)
	DecryptWithKeys(data []byte) ([]byte, error)
}

// New() - pgp initialization
func New() (OPENPGP, error) {
	var err error
	once.Do(func() {
		p, err := emptyKeyring()
		if err != nil {
			return
		}
		_keyring = p
		name, err := machineid.ID()
		if err != nil {
			log.Fatal(err, "get machineID failed")
		}
		if !helpers.FileExists(fmt.Sprintf("./openpgp/%s.gpg", name)) || !helpers.FileExists(fmt.Sprintf("./openpgp/%s.pub", name)) {
			err := os.MkdirAll("./openpgp", 0777)
			log.Warn(err, "openpgp folder creation failed")
			err = p.GeneratePair()
			if err != nil {
				log.Fatal(err, "encryption subsystem failed")
			}
		} else {
			err := p.ReadFolder(name)
			if err != nil {
				log.Fatal(err, "encryption subsystem failed")
			}
		}
		folder, err := os.ReadDir("./openpgp/")
		if err != nil {
			log.Error(err, "read openpgp folder failed")
		}
		for _, file := range folder {
			if file.IsDir() {
				continue
			}
			if strings.Contains(file.Name(), name) {
				continue
			}
			f, err := os.Open(fmt.Sprintf("./openpgp/%s", file.Name()))
			if err != nil {
				log.Warn(err, "public key open failed")
			}
			buf, err := io.ReadAll(f)
			if err != nil {
				log.Warn(err, "public key read failed")
			}
			err = p.AddPublicKey(buf)
			log.Warn(err, "public key add failed")
		}
	})
	if err != nil {
		return nil, errPGP
	}
	return _keyring, nil
}

func emptyKeyring() (*OpenPGP, error) {
	log = zerolog.New().WithPrefix("crypto")
	pp := new(passpharse)
	err := env.Parse(pp)
	if err != nil {
		log.Error(err, "read environment failed")
		return nil, err
	}
	if _, ok := os.LookupEnv("KEEPER_PGP_PASSPHRASE"); pp.PP == "" && !ok {
		p, err := helpers.ReadSecret("Enter pgp passphrase:")
		if err != nil {
			log.Warn(err, "read passphrase failed")
		}
		pp.PP = p
	}
	p := new(OpenPGP)
	p.passphrase = []byte(pp.PP)
	p.Private, err = pgpcrypto.NewKeyRing(nil)
	if err != nil {
		log.Debug(err, "empty private keyring creation failed")
		return nil, err
	}
	p.Public, err = pgpcrypto.NewKeyRing(nil)
	if err != nil {
		log.Debug(err, "empty public keyring creation failed")
		return nil, err
	}
	return p, nil
}

// GetPublicKey - retrun runnig public key
func (p *OpenPGP) GetPublicKey() string {
	return p.pubkey
}

// AddPublicKey - add new public key to keyring
func (p *OpenPGP) AddPublicKey(armored []byte) error {
	key, err := pgpcrypto.NewKeyFromArmored(string(armored))
	if err != nil {
		log.Debug(err, "key parse failed")
		return errPGP
	}
	err = p.Public.AddKey(key)
	if err != nil {
		log.Debug(err, "key add failed")
		return errPGP
	}
	log.Debug(nil, "key added")
	return nil
}

// AddPrivateKey - add private key to keyring
func (p *OpenPGP) AddPrivateKey(armored []byte) error {
	key, err := pgpcrypto.NewKeyFromArmored(string(armored))
	if err != nil {
		log.Debug(err, "key parse failed")
		return errPGP
	}
	ok, err := key.IsLocked()
	if err != nil {
		log.Debug(err, "key locked state check failed")
		return errPGP
	}
	if ok {
		unlocked, err := key.Unlock(p.passphrase)
		if err != nil {
			log.Debug(err, "key unlock failed")
			return errPGP
		}
		err = p.Private.AddKey(unlocked)
		if err != nil {
			log.Debug(err, "key add failed")
			return errPGP
		}
		log.Debug(nil, "locked key added")
		return nil
	}
	err = p.Private.AddKey(key)
	if err != nil {
		log.Debug(err, "key add failed")
		return errPGP
	}
	log.Debug(nil, "key added")
	return nil
}

// GeneratePair - generate key pair for current PC. name is machine-id
func (p *OpenPGP) GeneratePair() error {
	name, err := machineid.ID()
	if err != nil {
		log.Fatal(err, "get machineID failed")
	}
	var email string
	fmt.Print("Ente email address:")
	_, err = fmt.Scanln(&email)
	if err != nil || email == "" {
		log.Warn(err, "email reading failed. using default")
		email = fmt.Sprintf("%s@localhost", name)
	}
	key, err := pgpcrypto.GenerateKey(name, email, "x25519", 0)
	if err != nil {
		log.Debug(err, "openpgp key pair generation failed")
		return errPGP
	}
	defer key.ClearPrivateParams()
	if len(p.passphrase) == 0 {
		log.Warn(nil, "passphrase is blank. this is not secure")
	}
	locked, err := key.Lock([]byte(p.passphrase))
	if err != nil {
		log.Debug(err, "unable to lock openpgp key")
		return errPGP
	}
	if !helpers.FileExists("./openpgp") {
		err := os.MkdirAll("./openpgp", 0777)
		if err != nil {
			log.Debug(err, "openpgp folder creation failed")
			return errPGP
		}
	}
	keyfilename := fmt.Sprintf("./openpgp/%s.gpg", name)
	pubfilename := fmt.Sprintf("./openpgp/%s.pub", name)
	keyFile, err := os.OpenFile(keyfilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Debug(err, "openpgp private key file creation failed")
		return errPGP
	}
	defer keyFile.Close()
	pubFile, err := os.OpenFile(pubfilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0777)
	if err != nil {
		log.Debug(err, "openpgp public key file creation failed")
		return errPGP
	}
	defer pubFile.Close()
	public, err := key.GetArmoredPublicKey()
	if err != nil {
		log.Debug(err, "get openpgp public key failed")
		return errPGP
	}
	p.pubkey = public
	log.Trace(nil, "Public key:\n", public)
	fmt.Fprint(pubFile, public)
	private, err := locked.Armor()
	if err != nil {
		log.Debug(err, "get openpgp locked private key failed")
		return errPGP
	}
	fmt.Fprint(keyFile, private)
	log.Trace(nil, "Private key:\n", public)
	err = p.AddPublicKey([]byte(public))
	if err != nil {
		log.Debug(err, "add public key to keyring failed")
		return errPGP
	}
	p.AddPrivateKey([]byte(private))
	if err != nil {
		log.Debug(err, "add private key to keyring failed")
		return errPGP
	}
	return nil
}

// ReloadPublicKeys - reload public key ring with keys hashsums
func (p *OpenPGP) ReloadPublicKeys(keys []string) error {
	pub, err := pgpcrypto.NewKeyRing(nil)
	if err != nil {
		log.Error(err, "recreate public Keyring failed")
		return errPGP
	}
	p.Public.ClearPrivateParams()
	p.Public = pub
	for _, armor := range keys {
		err := p.AddPublicKey([]byte(armor))
		if err != nil {
			log.Error(err, "reloading public Keys failed")
			return errPGP
		}
	}
	return nil
}

// ReadFolder - read PC keys from openpgp folder
func (p *OpenPGP) ReadFolder(name string) error {
	pubfilename := fmt.Sprintf("./openpgp/%s.pub", name)
	privfilename := fmt.Sprintf("./openpgp/%s.gpg", name)
	pub, err := os.Open(pubfilename)
	if err != nil {
		log.Debug(err, "public key file open failed")
		return errPGP
	}
	defer pub.Close()
	priv, err := os.Open(privfilename)
	if err != nil {
		log.Debug(err, "private key file open failed")
		return errPGP
	}
	defer priv.Close()
	pubkey, err := io.ReadAll(pub)
	if err != nil {
		log.Debug(err, "public key read failed")
		return errPGP
	}
	p.pubkey = string(pubkey)
	privkey, err := io.ReadAll(priv)
	if err != nil {
		log.Debug(err, "public key read failed")
		return errPGP
	}
	err = p.AddPublicKey(pubkey)
	if err != nil {
		return errPGP
	}
	err = p.AddPrivateKey(privkey)
	if err != nil {
		return errPGP
	}
	return nil
}

// EncryptWithKeys - encrypt data with public keyring
func (p *OpenPGP) EncryptWithKeys(data []byte) ([]byte, error) {
	if len(data) == 0 {
		log.Warn(errors.New("encryptable data is missing"), "nothing to encrypt")
		return nil, errPGP
	}
	if !p.Public.CanEncrypt() {
		log.Debug(nil, "pgp keyring can not be used for encryption")
		return nil, errPGP
	}
	msg := pgpcrypto.NewPlainMessage(data)
	encryptedMsg, err := p.Public.EncryptWithCompression(msg, p.Private)
	if err != nil {
		log.Debug(err, "data encryption failed")
		return nil, errPGP
	}
	armor, err := encryptedMsg.GetArmored()
	if err != nil {
		log.Debug(err, "data encryption failed")
		return nil, errPGP
	}
	return []byte(armor), nil
}

// DecryptWithKey - decrypt data with PC's private key
func (p *OpenPGP) DecryptWithKeys(data []byte) ([]byte, error) {
	if len(data) == 0 {
		log.Warn(errors.New("decryptable data is missing"), "nothing to encrypt")
		return nil, errPGP
	}
	encryptedMsg, err := pgpcrypto.NewPGPMessageFromArmored(string(data))
	if err != nil {
		log.Debug(err, "read armorred message failed")
		return nil, errPGP
	}
	clear, err := p.Private.Decrypt(encryptedMsg, p.Public, time.Now().Unix())
	if err != nil {
		log.Debug(err, "data decryption failed")
		return nil, errPGP
	}
	return clear.Data, nil
}

// func (p *OpenPGP) NewPGPStream(filename string) (*PGPFile, error) {
// 	pgp := new(PGPFile)
// 	pgp.filename = filename
// 	pgp.keyring = p
// 	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
// 	if err != nil {
// 		return nil, err
// 	}
// 	filebuf := bufio.NewReader(file)
// 	var buf bytes.Buffer
// 	part := make([]byte, 1024)
// 	for {
// 		count, err := filebuf.Read(part)
// 		if err != nil {
// 			if err != io.EOF {
// 				return nil, err
// 			}
// 			break
// 		}

// 		fmt.Println("!!!!!!!!!!", part)
// 		buf.Write(part[:count])
// 	}
// 	data, err := p.DecryptWithKey(buf.Bytes())
// 	if err != nil {
// 		return nil, err
// 	}
// 	pgp.data = data
// 	pgp.buf = bytes.NewBuffer(pgp.data)
// 	fmt.Println("@@@@@@@@@@@@@@@@@@", pgp.buf.Bytes())
// 	fmt.Println("@@@@@@@@@@@@@@@@@@", data)
// 	return pgp, nil
// }

// func (p *PGPFile) Close() error {
// 	encrypted, err := p.keyring.EncryptWithKeys(p.buf.Bytes())
// 	if err != nil {
// 		return err
// 	}
// 	file, err := os.OpenFile(p.filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	_, err = file.Write(encrypted.Data)
// 	if err != nil {
// 		return err
// 	}
// 	file.Close()
// 	return nil
// }

// func (p *PGPFile) Reader() io.Reader {
// 	return p.buf
// }

// func (p *PGPFile) Writer() io.Writer {
// 	return p.buf
// }

// func (p *PGPFile) ReadWriter() io.ReadWriter {
// 	x := bufio.NewReadWriter(bufio.NewReader(p.buf), bufio.NewWriter(p.buf))
// 	return x
// }
