package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/mgutz/ansi"
	"golang.org/x/crypto/bcrypt"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
	"github.com/AlekseyKas/keeper/pkg/models"
	"github.com/AlekseyKas/keeper/pkg/server/config"
)

const (
	dbSchema = `CREATE TABLE IF NOT EXISTS public.users (
		id bigserial NOT NULL UNIQUE,
		username varchar NOT NULL UNIQUE,
		secret text NOT NULL
	);

	CREATE TABLE IF NOT EXISTS public.keeper (
		id bigserial NOT NULL UNIQUE,
		username varchar NOT NULL,
		checksum text NOT NULL,
		time  timestamp NOT NULL,
		content text,
		CONSTRAINT keeper_fk FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS public.openpgp (
		id bigserial NOT NULL UNIQUE,
		username varchar NOT NULL,
		publickey text NOT NULL,
		time timestamp NOT NULL,
		confirmed bool NULL DEFAULT false,
		revoked bool NULL DEFAULT false,
		CONSTRAINT publickeys_fk FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS public.logs (
		id bigserial NOT NULL UNIQUE,
		username varchar NOT NULL,
		ip varchar NOT NULL,
		action text NOT NULL,
		time timestamp NOT NULL,
		checksum text NOT NULL
	);

	CREATE UNIQUE INDEX IF NOT EXISTS users_names_idx ON public.users USING btree (username);
	CREATE UNIQUE INDEX IF NOT EXISTS keeper_content_idx ON public.keeper USING btree (username, checksum, content);
	CREATE UNIQUE INDEX IF NOT EXISTS openpgp_idx ON public.openpgp USING btree (username, publickey);`

	createUser    = `INSERT INTO public.users (username, secret) VALUES( $1, $2 );`
	deleteUser    = `DELETE FROM public.users WHERE username=$1;`
	deleteSecrets = `DELETE FROM public.keeper WHERE username=$1;`
	checkUser     = `SELECT username, secret FROM public.users WHERE username=$1;`
	checkPGP      = `SELECT publickey, confirmed FROM public.openpgp WHERE username=$1;`
	logAction     = `INSERT INTO public.logs (username, ip, action, time, checksum) VALUES( $1, $2, $3, $4, $5 );`
	logView       = `SELECT time, ip, action, checksum FROM public.logs WHERE username=$1;`
	pushData      = `INSERT INTO public.keeper (username, checksum, time, content) VALUES ( $1, $2, $3, $4);`
	pullData      = `SELECT content, time FROM public.keeper WHERE username=$1 AND checksum=$2`
	pullVersions  = `SELECT time, checksum FROM public.keeper WHERE username=$1`
	listPGP       = `SELECT time, publickey, confirmed FROM public.openpgp WHERE revoked=false AND username=$1;`
	addPGP        = `INSERT INTO public.openpgp (username, publickey, time, confirmed) VALUES ( $1, $2, $3, $4 );`
	revokePGP     = `UPDATE public.openpgp	SET revoked=true WHERE publickey=$1 AND username=$2;`
	confirmPGP    = `UPDATE public.openpgp	SET confirmed=true WHERE username=$1 AND publickey=$2;`
)

var (
	_storage Storage
	once     sync.Once
)

type psqlStorage struct {
	logger logging.Logger
	db     *pgxpool.Pool
}

func New() (Storage, error) {
	var err error
	s := new(psqlStorage)
	once.Do(func() {
		var db *pgxpool.Pool
		s.logger = zerolog.New().WithPrefix("storage")
		s.log().Debug(nil, "Connecting to Postgres database")
		db, err = pgxpool.Connect(context.Background(), config.New().DSN)
		if err != nil {
			s.log().Error(err, "error  in  database connection")
			return
		}
		s.log().Trace(nil, "Database storage connected")
		s.db = db
		err = s.create()
		if err != nil {
			s.log().Fatal(err, "db schema could not applied")
		}
		_storage = s
	})
	if err != nil {
		return nil, err
	}
	return _storage, nil
}

func (s *psqlStorage) SignUp(username, password string, ip net.IP) error {
	now := time.Now()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	s.log().Tracef("Password hash is %v", nil, hex.EncodeToString(hashed))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = s.db.Exec(ctx, createUser, username, hex.EncodeToString(hashed))
	if err != nil {
		return err
	}
	s.log().Infof("user %s created", nil, ansi.Color(username, "green+b"))
	s.SaveLog(username, "signup", "", ip, now)
	return nil
}

func (s *psqlStorage) SignIn(user models.User, ip net.IP) error {
	now := time.Now()
	usr := new(string)
	pwd := new(string)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.db.QueryRow(ctx, checkUser, user.Username).Scan(usr, pwd)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.log().Trace(nil, "provided username not found")
			return errors.New("no such user or password")
		}
		s.log().Trace(nil, "error in user request query")
		return err
	}
	hashed, err := hex.DecodeString(*pwd)
	if err != nil {
		s.log().Trace(err, "hex decode password hash completed with error")
		return errors.New("no such user or password")
	}
	err = bcrypt.CompareHashAndPassword(hashed, []byte(user.Password))
	if err != nil {
		s.log().Trace(err, "error in password verification")
		s.log().Debugf("wrong creditentials provided [user \"%s\" with password \"%s\"]", nil, ansi.Color(user.Username, "red+b"), ansi.Color(user.Password, "red+b"))
		return errors.New("no such user or password")
	}
	s.SaveLog(user.Username, "signin", "", ip, now)
	return nil
}

func (s *psqlStorage) DeleteUser(username string, ip net.IP) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tag, err := s.db.Exec(ctx, deleteUser, username)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		s.log().Trace(nil, "no affected users in delete query")
		return errors.New("no such user")
	}
	s.log().Infof("user %s deleted", nil, ansi.Color(username, "green+b"))
	s.SaveLog(username, "delete user", "", ip, now)
	s.SaveLog(username, "delete secrets", "", ip, now)
	return nil
}

func (s *psqlStorage) Push(username, checksum string, data string, ip net.IP) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()
	_, err := s.db.Exec(ctx, pushData, username, checksum, now, data)
	if err != nil {
		s.log().Debug(err, "content save action failed")
		return errors.New("content push failed")
	}

	s.SaveLog(username, "push", checksum, ip, now)
	return nil
}

func (s *psqlStorage) Pull(username, checksum string, ip net.IP) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()
	rows, err := s.db.Query(ctx, pullData, username, checksum)
	if err != nil {
		s.log().Error(err, "get secret failed")
		return nil, err
	}
	var latest []byte
	var timestamp time.Time
	for rows.Next() {
		var content string
		var ts time.Time
		err := rows.Scan(&content, &ts)
		if err != nil {
			s.log().Error(err, "row scan failed")
			return nil, err
		}
		if ts.After(timestamp) {
			s.log().Tracef("latest timestamp %s with hash %s", nil, ts.Format(time.RFC822), checksum)
			s.log().Trace(nil, content)
			latest = []byte(content)
		}
	}
	if len(latest) == 0 {
		s.log().Error(nil, "no content found")
		return nil, errors.New("no content")
	}
	s.log().Trace(nil, "result ", string(latest))
	s.SaveLog(username, "pull", checksum, ip, now)
	return latest, nil
}

func (s *psqlStorage) Versions(username string, ip net.IP) ([]models.Version, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()
	rows, err := s.db.Query(ctx, pullVersions, username)
	if err != nil {
		s.log().Error(err, "get versions failed")
		return nil, err
	}
	defer rows.Close()
	versions := []models.Version{}
	for rows.Next() {
		var (
			checksum string
			date     time.Time
		)
		err := rows.Scan(&date, &checksum)
		defer rows.Close()
		if err != nil {
			s.log().Error(err, "parse sql row failed")
			return nil, err
		}
		v := models.Version{Date: date, Hash: checksum}
		s.log().Tracef("%+v", nil, v)
		versions = append(versions, v)
	}
	if len(versions) == 0 {
		s.log().Error(nil, "no content found")
		return nil, errors.New("no content")
	}
	s.SaveLog(username, "get versions", "", ip, now)
	return helpers.OnlyOne(versions), nil
}

func (s *psqlStorage) SaveLog(username, action, checksum string, ip net.IP, date time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()
	_, err := s.db.Exec(ctx, logAction, username, ip.String(), action, now, checksum)
	if err != nil {
		s.log().Debug(err, "content save action failed")
		return errors.New("content push failed")
	}
	return nil
}

func (s *psqlStorage) GetLog(username string, ip net.IP) ([]models.Action, error) {
	now := time.Now()
	actions := []models.Action{}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	rows, err := s.db.Query(ctx, logView, username)
	if err != nil {
		s.log().Error(err, "get log failed")
	}
	defer rows.Close()
	for rows.Next() {
		var (
			action, checksum, ip string
			date                 time.Time
		)
		err := rows.Scan(&date, &ip, &action, &checksum)
		defer rows.Close()
		if err != nil {
			s.log().Error(err, "parse sql row failed")
			return nil, err
		}
		addr := net.ParseIP(ip)
		act := models.Action{Action: action, Checksum: checksum, IP: addr, Date: date}
		s.log().Tracef("%+v", nil, act)
		actions = append(actions, act)
	}
	s.SaveLog(username, "get log", "", ip, now)
	return actions, nil
}

func (s *psqlStorage) ListPGP(username string, ip net.IP) ([]models.PGP, error) {
	now := time.Now()
	keys := []models.PGP{}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	rows, err := s.db.Query(ctx, listPGP, username)
	if err != nil {
		s.log().Error(err, "list pgp request failed")
	}
	defer rows.Close()
	for rows.Next() {
		var (
			publickey string
			confirmed bool
			date      time.Time
		)
		err := rows.Scan(&date, &publickey, &confirmed)
		defer rows.Close()
		if err != nil {
			s.log().Error(err, "parse sql row failed")
			return nil, err
		}
		key := models.PGP{Date: date, Publickey: publickey, Confirmed: confirmed}
		s.log().Tracef("%+v", nil, key)
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		s.log().Error(nil, "vault keys not found")
		return nil, errors.New("no content")
	}
	s.log().Trace(nil, "found ", keys)
	s.SaveLog(username, "list pgp", "", ip, now)
	return keys, nil
}

func (s *psqlStorage) AddPGP(username, publickey string, confirm bool, ip net.IP) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	now := time.Now()
	_, err := s.db.Exec(ctx, addPGP, username, publickey, now, confirm)
	if err != nil {
		s.log().Debug(err, "content save action failed")
		return errors.New("content push failed")
	}

	s.SaveLog(username, "add pgp public key", "", ip, now)
	return nil
}

func (s *psqlStorage) RevokePGP(username, publickey string, ip net.IP) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tag, err := s.db.Exec(ctx, revokePGP, publickey, username)
	if err != nil {
		s.log().Debug(err, "revoke failed")
		return errors.New("revoke failed")
	}
	s.log().Tracef("Affected rows are: %d", nil, tag.RowsAffected())
	if tag.RowsAffected() == 0 {
		s.log().Error(nil, "combination of vault and piublic key not found")
		return errors.New("no such vault or public key")
	}
	s.SaveLog(username, "revoke pgp public key", "", ip, now)
	return nil
}

func (s *psqlStorage) ConfirmPGP(username, publickey string, ip net.IP) error {
	now := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	tag, err := s.db.Exec(ctx, confirmPGP, username, publickey)
	if err != nil {
		s.log().Debug(err, "confirm failed")
		return errors.New("confirm failed")
	}
	s.log().Tracef("Affected rows are: %d", nil, tag.RowsAffected())
	if tag.RowsAffected() == 0 {
		s.log().Error(nil, "combination of vault and piublic key not found")
		return errors.New("no such vault or public key")
	}
	s.SaveLog(username, "confirm pgp public key", "", ip, now)
	return nil
}

func (s *psqlStorage) Close() error {
	s.db.Close()
	s.log().Trace(nil, "database connection closed")
	return nil
}

func (s *psqlStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.db.Ping(ctx)
	return err
}

func (s *psqlStorage) create() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := s.db.Exec(ctx, dbSchema)
	if err != nil {
		return err
	}
	s.log().Debug(nil, "database schema applied successfully")
	return nil
}

func (s *psqlStorage) log() logging.Logger {
	return s.logger
}
