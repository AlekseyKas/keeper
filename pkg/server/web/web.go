package web

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	mw "github.com/labstack/echo/v4/middleware"
	"github.com/mgutz/ansi"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
	"github.com/AlekseyKas/keeper/pkg/server/config"
	"github.com/AlekseyKas/keeper/pkg/server/storage"
	"github.com/AlekseyKas/keeper/pkg/server/tls"
	"github.com/AlekseyKas/keeper/pkg/server/web/auth"
	mymws "github.com/AlekseyKas/keeper/pkg/server/web/middlewares"
	"github.com/AlekseyKas/keeper/pkg/server/web/static"
	"github.com/AlekseyKas/keeper/pkg/server/web/websockets"
)

var (
	errRegFail   = echo.NewHTTPError(http.StatusBadRequest, "registration failed")
	errBadReq    = echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	errUnAuth    = echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	errIntSrvErr = echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	errStorage   = echo.NewHTTPError(http.StatusInternalServerError, "Storage unavailable")
	errEmpty     = echo.NewHTTPError(http.StatusNoContent, "no data availible")
)

// Server - web api part of application
type Server struct {
	echo *echo.Echo      // pointer of echo framework
	log  logging.Logger  // instance logger
	wg   *sync.WaitGroup // pointer of waitgroup
	db   storage.Storage
	sig  chan struct{}
	msg  map[string]chan models.Message
}

// New - web api intialization
//		bind - string with network binding address
//		logger - application logger
func New() *Server {
	tls.Prepare()
	s := new(Server)
	s.echo = echo.New()
	s.log = zerolog.New().WithPrefix("web")
	s.db, _ = storage.New()
	s.applyMiddlewares()
	s.msg = make(map[string]chan models.Message)
	static.ApplyStatic(s.echo)
	s.sig = make(chan struct{})
	s.createRouter()
	return s
}

func (s *Server) chanCleaner() {
	s.wg.Add(1)
	mux := websockets.GetMutex()
	chs := websockets.GetMsgChan().Cleanup()
	ticker := time.NewTicker(10 * time.Second)
	s.log.Trace(nil, "channel cleaner started")
	for {
		select {
		case <-s.sig:
			s.log.Trace(nil, "channel cleaner stopped")
			s.wg.Done()
			return
		case <-ticker.C:
			s.log.Trace(nil, "try to make cleanup")
			for _, vv := range chs {
				s.log.Trace(nil, "cleanup for ", vv.Vault)
				for k, v := range vv.Channels {
					exp, err := helpers.GetExpirationFromToken(k)
					if err != nil {
						s.log.Trace(nil, "force to cleanup token ", k)
						mux.Lock()
						close(v)
						delete(vv.Channels, k)
						mux.Unlock()
					}
					if time.Now().After(*exp) {
						s.log.Trace(nil, "standart cleanup token ", k)
						mux.Lock()
						close(v)
						delete(vv.Channels, k)
						mux.Unlock()
					}
				}
			}
		}
	}
}

// Start - func to start web-api
//  wg - application WaitGroup
func (s *Server) Start(wg *sync.WaitGroup) error {
	s.wg = wg
	s.wg.Add(1)
	go s.chanCleaner()
	err := s.echo.StartTLS(config.New().WebBind, "./ssl/server.crt", "./ssl/server.pem")
	if err != nil && !strings.Contains(err.Error(), "Server closed") {
		s.log.Error(err, "filed to start")
		return err
	}
	return nil
}

// Stop - gracefull shutdown of web api with shutdown timeout in 10 seconds
func (s *Server) Stop() error {
	close(s.sig)
	s.log.Info(nil, "Graceful shutdown in progress...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.echo.Shutdown(ctx)
	if err != nil {
		s.log.Error(err, "shutting down ends with error")
		return err
	}
	s.log.Info(nil, "Web server stopped")
	s.wg.Done()
	return nil
}

// applyMiddlewares - apply middleware set to echo framework instance
func (s *Server) applyMiddlewares() *Server {
	s.echo.Use(mw.Gzip())
	s.echo.Use(mw.Decompress())
	s.echo.Use(mw.Recover())
	s.echo.Use(mw.RequestLoggerWithConfig(mw.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogRemoteIP: true,
		LogHost:     true,
		LogValuesFunc: func(c echo.Context, v mw.RequestLoggerValues) error {
			s.log.
				WithFields(logging.Fields{
					"URI":      v.URI,
					"Status":   v.Status,
					"RemoteIP": v.RemoteIP,
					"Host":     v.Host,
				}).
				Info(nil, "request")
			return nil
		},
	}))
	s.echo.Use(mymws.JSONHeader)
	return s
}

// createRouter - define http route logic
func (s *Server) createRouter() *Server {

	s.echo.GET("/api/v1/ping", s.ping)                    // db liveness probe
	s.echo.POST("/api/v1/signup", s.registration)         // signup web handler
	s.echo.POST("/api/v1/signin", s.signin)               // login web handler
	s.echo.GET("/api/swagger/*", echoSwagger.WrapHandler) // swagger web page

	restricted := s.echo.Group("/api/v1/keeper") // router group with restrected access
	jwtconfig := mw.JWTConfig{
		Claims:                  &auth.JWTClaims{},
		TokenLookup:             "cookie:token",
		SigningKey:              auth.Key(),
		ErrorHandler:            auth.JWTErrorHandler,
		ErrorHandlerWithContext: auth.JWTErrorHandlerWithContext,
	}
	restricted.Use(mw.JWTWithConfig(jwtconfig))
	restricted.DELETE("/remove", s.remove)        // user deletion handler
	restricted.POST("/push", s.save)              // api save handler
	restricted.GET("/pgp/list", s.listpgp)        // api pgp list public keys
	restricted.POST("/pgp/add", s.addpgp)         // api pgp add public keys
	restricted.POST("/pgp/confirm", s.confirmpgp) // api pgp add public keys
	restricted.POST("/pgp/revoke", s.revokepgp)   // api pgp revoke public keys
	restricted.GET("/pull", s.get)                // api get secret content hadler
	restricted.GET("/pull/versions", s.versions)  // api get versions of secrets
	restricted.GET("/ws", websockets.New)         // api websocket
	restricted.GET("/logs", s.logs)               // api get logs

	return s
}

// registration - handler for /signup
func (s *Server) registration(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	defer c.Request().Body.Close()
	if c.Request().Body == nil {
		return errBadReq
	}
	u := new(models.User)
	if err := c.Bind(u); err != nil {
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	s.log.Debugf("recieved registration request for user %v", nil, ansi.Color(u.Username, "green+b"))
	err := s.db.SignUp(u.Username, u.Password, addr)
	if err != nil {
		s.log.Warnf("registration failed with error", err, "")
		return errRegFail
	}
	err = auth.Token(u.Username, auth.Key(), c)
	if err != nil {
		s.log.Error(err, "authorization token creation error")
		return errIntSrvErr
	}
	s.log.Trace(nil, "auth token generated")
	err = s.db.AddPGP(u.Username, u.PublicKey, true, addr)
	if err != nil {
		s.log.Error(err, "public key saving error")
		return errIntSrvErr
	}
	s.log.Trace(nil, "pgp public key added")
	return c.String(http.StatusCreated, "{\"status\": \"OK\"}")
}

// signin - handler for /login
func (s *Server) signin(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	defer c.Request().Body.Close()
	if c.Request().Body == nil {
		return errBadReq
	}
	user := new(models.User)
	if err := c.Bind(user); err != nil {
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	s.log.Debugf("recieved login request for user %v", nil, ansi.Color(user.Username, "green+b"))
	err := s.db.SignIn(*user, addr)
	if err != nil {
		s.log.Warn(err)
		return errUnAuth
	}
	err = auth.Token(user.Username, auth.Key(), c)
	if err != nil {
		s.log.Error(err, "authorization token creation error")
		return errIntSrvErr
	}
	keys, err := s.db.ListPGP(user.Username, addr)
	if err != nil {
		s.log.Error(err, "user can not be validated")
		return errIntSrvErr
	}
	if len(keys) == 0 {
		s.log.Error(err, "public key not registered")
		return errIntSrvErr
	}
	var found bool
	hash := helpers.GenHash([]byte(user.PublicKey))
	for _, key := range keys {
		if helpers.CompareHash(hash, []byte(key.Publickey)) {
			found = true
			if !key.Confirmed {
				return c.String(http.StatusAlreadyReported, "{\"status\": \"Forbidden\"}")
			}
		}
	}
	if !found {
		return c.String(http.StatusForbidden, "{\"status\": \"Forbidden\"}")
	}
	return c.String(http.StatusOK, "{\"status\": \"OK\"}")
}

// signin - handler for /remove
func (s *Server) remove(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	defer c.Request().Body.Close()
	if c.Request().Body == nil {
		return errBadReq
	}
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	err = s.db.DeleteUser(name, addr)
	if err != nil {
		s.log.Error(err, "delete user's records failed")
		return errIntSrvErr
	}
	return c.String(http.StatusAccepted, "{\"status\": \"Removed\"}")
}

// ping - handle for /ping
func (s *Server) ping(c echo.Context) error {
	err := s.db.Ping()
	if err != nil {
		return errStorage
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(http.StatusOK, "{\"status\": \"OK\"}")
}

// save - handle for /api/v1/keeper/save
func (s *Server) save(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	defer c.Request().Body.Close()
	if c.Request().Body == nil {
		s.log.Debug(nil, "empty request")
		return errBadReq
	}
	body := new(models.Content)
	if err := c.Bind(body); err != nil {
		s.log.Debug(err, "body parse failed")
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	err = s.db.Push(name, body.Hash, body.Payload, addr)
	if err != nil {
		s.log.Debug(err, "save payload failed")
		return errBadReq
	}
	websockets.GetMsgChan().Notify(name, token.Value, models.Message{Text: "new version recieved", Content: body.Hash})
	return c.String(http.StatusOK, "{\"status\": \"OK\"}")
}

// save - handle for /api/v1/keeper/save
func (s *Server) logs(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	actions, err := s.db.GetLog(name, addr)
	if err != nil {
		s.log.Debug(err, "get logs request failed")
		return errBadReq
	}
	if len(actions) == 0 {
		return errEmpty
	}
	return c.JSON(http.StatusOK, actions)
}

// get - handle for /api/v1/keeper/pull
func (s *Server) get(c echo.Context) error {
	type req struct {
		Checksum string `query:"checksum"`
	}
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	r := new(req)
	s.log.Trace(nil, c.Request().RequestURI)
	err = c.Bind(r)
	if err != nil {
		s.log.Debug(err, "failed to parse url params")
		return errIntSrvErr
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	secret, err := s.db.Pull(name, r.Checksum, addr)
	if err != nil {
		s.log.Debug(err, "pull request failed")
		return errBadReq
	}
	if len(secret) == 0 {
		return errEmpty
	}
	return c.String(http.StatusOK, string(secret))
}

// get - handle for /api/v1/keeper/pull
func (s *Server) versions(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	s.log.Trace(nil, c.Request().RequestURI)
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	versions, err := s.db.Versions(name, addr)
	if err != nil {
		s.log.Debug(err, "get versions failed")
		return errBadReq
	}
	if len(versions) == 0 {
		return errEmpty
	}
	return c.JSON(http.StatusOK, versions)
}

// get - handle for /api/v1/keeper/pull
func (s *Server) listpgp(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	list, err := s.db.ListPGP(name, addr)
	if err != nil {
		s.log.Debug(err, "get versions failed")
		return errBadReq
	}
	if len(list) == 0 {
		return errEmpty
	}
	return c.JSON(http.StatusOK, list)
}

// get - handle for /api/v1/keeper/pull
func (s *Server) addpgp(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	key := new(models.PGP)
	err = c.Bind(key)
	if err != nil {
		s.log.Debug(err, "parse body failed")
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	err = s.db.AddPGP(name, key.Publickey, false, addr)
	if err != nil {
		s.log.Debug(err, "add pgp public key failed failed")
		return errBadReq
	}
	websockets.GetMsgChan().Notify(name, token.Value, models.Message{Text: "new client with unknown pgp key", Content: key.Publickey})
	return c.String(http.StatusCreated, "{\"status\": \"OK\"}")
}

// get - handle for /api/v1/keeper/pull
func (s *Server) revokepgp(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	key := new(models.PGP)
	err = c.Bind(key)
	if err != nil {
		s.log.Debug(err, "parse body failed")
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	err = s.db.RevokePGP(name, key.Publickey, addr)
	if err != nil {
		s.log.Debug(err, "pgp public key revoke failed")
		return errBadReq
	}
	return c.String(http.StatusGone, "{\"status\": \"GONE\"}")
}

// get - handle for /api/v1/keeper/pull
func (s *Server) confirmpgp(c echo.Context) error {
	token, err := c.Request().Cookie("token")
	if err != nil {
		s.log.Error(err, "get cookie token failed")
	}
	name, err := helpers.GetNameFromToken(token.String())
	if err != nil {
		s.log.Debug(err, "get username from cookie failed")
		return errIntSrvErr
	}
	key := new(models.PGP)
	err = c.Bind(key)
	if err != nil {
		s.log.Debug(err, "parse body failed")
		return errBadReq
	}
	ip := strings.Split(c.Request().RemoteAddr, ":")
	addr := net.ParseIP(ip[0])
	err = s.db.ConfirmPGP(name, key.Publickey, addr)
	if err != nil {
		s.log.Debug(err, "pgp public key confirm failed")
		return errBadReq
	}
	return c.String(http.StatusOK, "{\"status\": \"OK\"}")
}
