package web

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
	"github.com/mgutz/ansi"

	"github.com/AlekseyKas/keeper/pkg/client/config"
	"github.com/AlekseyKas/keeper/pkg/client/remote"
	"github.com/AlekseyKas/keeper/pkg/client/storage"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
)

var (
	once    sync.Once
	_client *WebClient
)

// WebClient - go-resty and websocket client struct
type WebClient struct {
	client *resty.Client
	logger logging.Logger
	jar    *cookiejar.Jar
	wsSig  chan struct{}
}

// New() -initialize http client
func New() remote.Actions {
	client := new(WebClient)
	client.wsSig = make(chan struct{})
	once.Do(func() {
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatal(err, "cookie initialization failed")
		}
		client.jar = jar
		client.client = resty.New()
		client.client.SetCookieJar(client.jar)
		client.client.SetBaseURL(config.GetRunning().RemoteHTTP)
		client.client.SetTLSClientConfig(&tls.Config{
			Rand:               rand.Reader,
			InsecureSkipVerify: true,
		})
		client.logger = zerolog.New().WithPrefix("resty-client")
		_client = client
		client.log().Info(nil, "go-resty client initialized")
	})
	return _client
}

// log - usefull shot for call log
func (c *WebClient) log() logging.Logger {
	return c.logger
}

// Login - login action
//   Statuses:
//   200 - login successfull
//	 208 - PGP key not confirmed yet, but other users notified
//   400 - wrong request
//   401 - username or password not found
//   403 - PGP key not confirmed
//   500 - internal server error
func (c *WebClient) Login(username, password, public string) error {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(
			models.User{
				Username:  username,
				Password:  password,
				PublicKey: public,
			}).
		Post("api/v1/signin")
	if err != nil {
		c.log().Error(err, "client login failed")
		return err
	}
	if response.StatusCode() == http.StatusForbidden {
		err := c.AddPGP(public)
		if err != nil {
			c.log().Error(err, "public key add failed")
			return err
		}
		c.log().Error(nil, "pgp key not confirmed. status ", http.StatusForbidden)
		c.log().Info(nil, "please try again later")
		return nil
	}
	if response.StatusCode() == http.StatusAlreadyReported {
		c.log().Error(nil, "public key not confirmed yet. status", http.StatusAlreadyReported)
		c.log().Info(nil, "please try again later")
		return nil
	}
	if response.StatusCode() == http.StatusInternalServerError {
		c.log().Error(nil, "sign in failed. status", http.StatusInternalServerError)
		return errors.New("internal server error")
	}
	if response.StatusCode() == http.StatusBadRequest {
		c.log().Error(nil, "invalid sign in form. status", http.StatusBadRequest)
		return errors.New("bad request")
	}
	if response.StatusCode() == http.StatusUnauthorized {
		c.log().Error(nil, "wrong username or password. status", http.StatusUnauthorized)
		return errors.New("unauthorized")
	}
	return nil
}

// Register - signup action
//   Statuses:
//   201 - success
//   500 - internal server error
func (c *WebClient) Register(username, password, public string) error {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(
			models.User{
				Username:  username,
				Password:  password,
				PublicKey: public,
			}).
		Post("api/v1/signup")
	if err != nil {
		c.log().Error(err, "client registration failed")
		return err
	}
	if response.StatusCode() != http.StatusCreated {
		c.log().Error(nil, "sign up failed")
		return errors.New("registration failed")
	}
	return nil
}

// Delete - delete action
//   Statuses:
//   202 - deleted
//   500 - internal server error
func (c *WebClient) Delete() error {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Delete("api/v1/keeper/remove")
	if err != nil {
		c.log().Error(err, "client deletion failed")
		return err
	}
	if response.StatusCode() != http.StatusAccepted {
		c.log().Error(nil, "user deleteion failed")
		return errors.New("user deletion failed")
	}
	return nil
}

// Push - delete action
//   Statuses:
//   200 - saved
//   500 - internal server error
func (c *WebClient) Push(payload, hashsum string) error {
	data := models.Content{Payload: payload, Hash: hashsum}
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(data).
		Post("api/v1/keeper/push")
	if err != nil {
		c.log().Error(err, "save action failed")
		return err
	}
	if response.StatusCode() != http.StatusOK {
		c.log().Error(nil, "save action failed")
		return errors.New("save action failed")
	}
	return nil
}

// GetLogs - get action logs
//   Statuses:
//   200 - success
//   202 - no logs
//   500 - internal server error
func (c *WebClient) GetLogs() ([]models.Action, error) {
	actions := []models.Action{}
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get("api/v1/keeper/logs")
	if err != nil {
		c.log().Error(err, "get logs action failed")
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		if response.StatusCode() == http.StatusNoContent {
			c.log().Debug(nil, "no logs found")
			return []models.Action{}, nil
		}
		c.log().Error(nil, "get logs action failed")
		return nil, errors.New("get logs action failed")
	}
	err = json.Unmarshal(response.Body(), &actions)
	if err != nil {
		c.log().Error(nil, "response parse failed")
		return nil, err
	}
	return actions, nil
}

// Pull - get specific vrecion of vault
//   Statuses:
//   200 - success
//   202 - not found
//   500 - internal server error
func (c *WebClient) Pull(checksum string) ([]byte, error) {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetQueryParams(map[string]string{"checksum": checksum}).
		Get("api/v1/keeper/pull")
	if err != nil {
		c.log().Error(err, "get secret action failed")
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		if response.StatusCode() == http.StatusNoContent {
			c.log().Debug(nil, "no secrets found")
			return nil, nil
		}
		c.log().Error(nil, "get secrets failed failed")
		return nil, errors.New("get secrets failed failed")
	}
	return response.Body(), nil
}

// Versions - get list of vault versions
//   Statuses:
//   200 - success
//   202 - not found
//   500 - internal server error
func (c *WebClient) Versions() ([]models.Version, error) {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get("api/v1/keeper/pull/versions")
	if err != nil {
		c.log().Error(err, "get versions failed")
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		if response.StatusCode() == http.StatusNoContent {
			c.log().Debug(nil, "no versions found")
			return nil, nil
		}
		c.log().Error(nil, "get versions failed")
		return nil, errors.New("get versions failed")
	}
	versions := []models.Version{}
	err = json.Unmarshal(response.Body(), &versions)
	if err != nil {
		c.log().Error(nil, "parse versions failed")
		return nil, errors.New("get versions failed")
	}
	return versions, nil
}

// ListPGP - get list of vault pgp public keys
//   Statuses:
//   200 - success
//   202 - not found
//   500 - internal server error
func (c *WebClient) ListPGP() ([]models.PGP, error) {
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		Get("api/v1/keeper/pgp/list")
	if err != nil {
		c.log().Error(err, "get keys failed")
		return nil, err
	}
	if response.StatusCode() != http.StatusOK {
		if response.StatusCode() == http.StatusNoContent {
			c.log().Debug(nil, "no keys found")
			return nil, nil
		}
		c.log().Error(nil, "get keys failed")
		return nil, errors.New("get keys failed")
	}
	keys := []models.PGP{}
	err = json.Unmarshal(response.Body(), &keys)
	if err != nil {
		c.log().Error(nil, "parse keys failed")
		return nil, errors.New("get keys failed")
	}
	return keys, nil
}

// AddPGP - add public key to vault pgp public keys
//   Statuses:
//   202 - success
//   500 - internal server error
func (c *WebClient) AddPGP(publickey string) error {
	body := models.PGP{Publickey: publickey}
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("api/v1/keeper/pgp/add")
	if err != nil {
		c.log().Error(err, "add key failed")
		return err
	}
	if response.StatusCode() != http.StatusCreated {
		c.log().Error(nil, "add key failed")
		return errors.New("add key failed")
	}
	return nil
}

// ConfirmPGP - configrm  vault pgp public key
//   Statuses:
//   200 - success
//   500 - internal server error
func (c *WebClient) ConfirmPGP(publickey string) error {
	body := models.PGP{Publickey: publickey}
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("api/v1/keeper/pgp/confirm")
	if err != nil {
		c.log().Error(err, "key confirmation failed")
		return err
	}
	if response.StatusCode() != http.StatusOK {
		c.log().Error(nil, "key confirmation failed")
		return errors.New("key confirmation failed")
	}
	return nil
}

// RevokePGP - revoke  vault pgp public key
//   Statuses:
//   410 - success
//   500 - internal server error
func (c *WebClient) RevokePGP(publickey string) error {
	body := models.PGP{Publickey: publickey}
	response, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post("api/v1/keeper/pgp/revoke")
	if err != nil {
		c.log().Error(err, "revoke key failed")
		return err
	}
	if response.StatusCode() != http.StatusGone {
		c.log().Error(nil, "revoke key failed")
		return errors.New("revoke key failed")
	}
	return nil
}

// Close - close http connection
func (c *WebClient) Close() error {
	c.log().Trace(nil, "closing websocket connection")
	close(c.wsSig)
	return nil
}

// NewStream - initialize and start websocket connection
func (c *WebClient) NewStream() error {
	r := config.New().RemoteHTTP
	rr := strings.Split(r, "/")
	remote := rr[len(rr)-1]
	c.log().Trace(nil, remote)
	u := url.URL{Scheme: "wss", Host: remote, Path: "/api/v1/keeper/ws"}
	dialer := websocket.DefaultDialer
	dialer.Jar = c.jar
	dialer.EnableCompression = true
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true, Rand: rand.Reader}
	conn, _, err := dialer.Dial(u.String(), http.Header{"Content-Type": []string{"application/json"}})
	if err != nil {
		c.log().Error(err, "websocket connection failed")
		return err
	}

	go func() {
		defer conn.Close()
		for {
			select {
			case <-c.wsSig:
				return
			default:
				_, message, wsErr := conn.ReadMessage()
				if wsErr != nil {
					c.log().Error(wsErr, "read from websocket error")
					return
				}
				msg := string(message)
				if len(message) == 0 {
					continue
				}
				switch msg {
				case "pong":
					c.log().Trace(nil, msg)
				default:
					m := models.Message{}
					err := json.Unmarshal(message, &m)
					if err != nil {
						c.log().Error(err, "message can not be parsed")
					}
					if strings.Contains(msg, "new client with unknown pgp key") {
						fmt.Println(ansi.Color("------------------------------------------------------------------------", "reb+b"))
						c.log().Warn(nil, "Unknow PGP Public key registered. Please chek and confirm or revoke key")
						c.log().Warn(nil, "Registered key is\n", m.Content)
						fmt.Println(ansi.Color("------------------------------------------------------------------------", "reb+b"))
					}
					if strings.Contains(msg, "new version recieved") && storage.New().HashSum() != m.Content {
						fmt.Println(ansi.Color("------------------------------------------------------------------------", "reb+b"))
						c.log().Warn(nil, "New version on keeper storage saved to server. Please sync local with remote")
						c.log().Warn(nil, "Newly registered checksum is ", m.Content)
						fmt.Println(ansi.Color("------------------------------------------------------------------------", "reb+b"))
					}
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.wsSig:
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				c.log().Error(err, "write close error")
				return err
			}
			err = conn.Close()
			if err != nil {
				c.log().Error(err, "websocket connection close error")
				return err
			}
			return nil
		case <-ticker.C:
			err := conn.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				log.Println("write:", err)
				return err
			}
		}

	}
}
