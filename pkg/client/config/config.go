package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync"

	"github.com/caarlos0/env"
	"github.com/creasty/defaults"

	"github.com/AlekseyKas/keeper/pkg/helpers"
	"github.com/AlekseyKas/keeper/pkg/logging"
	"github.com/AlekseyKas/keeper/pkg/logging/zerolog"
)

var (
	_running *Config
	once     sync.Once
)

//Config - struct for handling configuration
type Config struct {
	Mode       string         `json:"mode" default:"standalone"`
	Username   string         `env:"KEEPER_REMOTE_USERNAME" json:"username,omitempty" default:"-"`
	Password   string         `env:"KEEPER_REMOTE_PASSWORD" json:"password,omitempty" default:"-"`
	RemoteHTTP string         `env:"KEEPER_REMOTE_URL" json:"remote-http,omitempty" default:"-"`
	RemoteGRPC string         `env:"KEEPER_REMOTE_GRPC" json:"remote-grpc,omitempty" default:"-"`
	LogLevel   uint8          `json:"LogLevel,omitempty" default:"-"`
	Storage    string         `json:"storage" default:"secrets.db"`
	file       string         `json:"-" default:"config.json"`
	logger     logging.Logger `default:"-"`
}

//NewConfig - config initialization
func New() *Config {
	once.Do(func() {
		c := new(Config)
		c.SetDefaults()
		c.SetByFlags()
		c.SetByFile()
		c.SetByEnv()
		if c.RemoteHTTP != "" || c.RemoteGRPC != "" {
			c.Mode = "client-server"
		}
		c.log().Trace(nil, fmt.Sprintf("%+v\n", c))
		_running = c
	})

	return _running
}

// log - returns default application Logger
func (c *Config) log() logging.Logger {
	return c.logger.WithPrefix("configuration")
}

//SetDefaults - set configuration values to default state
func (c *Config) SetDefaults() *Config {
	c.logger = zerolog.New().WithPrefix("configuration")
	c.logger.SetLevel(logging.InfoLevel)
	err := defaults.Set(c)
	if err != nil {
		c.log().Fatal(err)
	}
	c.LogLevel = uint8(logging.InfoLevel)
	c.file = "config.json"
	c.Mode = "standalone"
	return c
}

//SetByEnv - set configuration values from evironment
func (c *Config) SetByEnv() *Config {
	cc := Config{}
	err := env.Parse(&cc)
	if err != nil {
		c.log().Warnf("read environment failed with error: %v", err)
		return c
	}
	if cc.Username != "" {
		c.Username = cc.Username
	}
	if cc.Password != "" {
		c.Password = cc.Password
	}
	if cc.RemoteHTTP != "" {
		c.RemoteHTTP = cc.RemoteHTTP
	}
	if cc.RemoteGRPC != "" {
		c.RemoteGRPC = cc.RemoteGRPC
	}
	return c
}

//Configuring flags
var (
	remoteHTTPFlag = flag.String("remote-http", "", "Set up remote URL.\nExample: -remote-http=\"https://localhost.ltd:8443\"")
	// remoteGRPCFlag = flag.String("remote-grpc", "", "Set up remote URL.\nExample: -remote-grpc=\"localhost.ltd:3200\"")
	usernameFlag = flag.String("username", "", "Set up username for authorization on remote.\nExample: -username=\"username\"")
	passwordFlag = flag.String("password", "", "Set up password for authorization on remote.\nExample: -password=\"password\"")
	loglevelFlag = flag.String("loglevel", "info", "Set up logging level.\nExample: -loglevel=info\nAvailible level are trace, debug, info, warn, error, fatal, panic.")
	storageFlag  = flag.String("storage", "", "Set up storage path.\nExample: -storage=\"storage.db\"")
	configFlag   = flag.String("config", "", "Set up configuration file path.\n Example: -config=config.json")
)

//SetByFlags - set configuration values from cli flags
//	-remote-http http address of server
//	-username vault name flag
//	-password vault password
//  -loglevel setup logging level
//  -config path to config file
func (c *Config) SetByFlags() *Config {
	flag.Parse()
	if flag.Parsed() {
		if remoteHTTPFlag != nil && *remoteHTTPFlag != "" {
			c.RemoteHTTP = *remoteHTTPFlag
		}
		// if remoteGRPCFlag != nil && *remoteGRPCFlag != "" {
		// 	c.RemoteGRPC = *remoteGRPCFlag
		// }
		if usernameFlag != nil && *usernameFlag != "" {
			c.Username = *usernameFlag
		}
		if passwordFlag != nil && *passwordFlag != "" {
			c.Password = *passwordFlag
		}
		if configFlag != nil && *configFlag != "" {
			c.file = *configFlag
		}
		if storageFlag != nil && *storageFlag != "" {
			c.Storage = *storageFlag
		}
		if loglevelFlag != nil {
			if helpers.IsFlagPassed("loglevel") {
				switch *loglevelFlag {
				case "trace":
					c.LogLevel = uint8(logging.TraceLevel)
				case "debug":
					c.LogLevel = uint8(logging.DebugLevel)
				case "info":
					c.LogLevel = uint8(logging.InfoLevel)
				case "warn":
					c.LogLevel = uint8(logging.WarnLevel)
				case "error":
					c.LogLevel = uint8(logging.ErrorLevel)
				case "fatal":
					c.LogLevel = uint8(logging.FatalLevel)
				case "panic":
					c.LogLevel = uint8(logging.PanicLevel)
				default:
					c.LogLevel = uint8(logging.InfoLevel)
				}
				c.logger.SetLevel(logging.Level(c.LogLevel))
			}
		}
	}
	return c
}

//SetByFile - set configuration values from configuration file
func (c *Config) SetByFile() *Config {
	cfg, err := os.ReadFile(c.file)
	if err != nil {
		c.log().Warn(err, "configuration file can not be read")
		return c
	}
	cc := new(Config)
	err = json.Unmarshal(cfg, cc)
	if err != nil {
		c.log().Warn(err, "configuration can not be parsed")
		return c
	}
	cc.file = c.file
	return cc
}

// Level - get current logging level
func (c *Config) Level() logging.Level {
	return logging.Level(c.LogLevel)
}

// GetRunning - get pinter of running configuration
func GetRunning() *Config {
	return _running
}
