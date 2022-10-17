package models

import (
	"net"
	"time"
)

// User - vault client-server model for signup and registartion
type User struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	PublicKey string `json:"publickey"`
}

// Action - struct defines user action in client-server operations
type Action struct {
	Action   string    `json:"action"`   //
	Checksum string    `json:"checksum"` //
	IP       net.IP    `json:"ip"`       //
	Date     time.Time `json:"time"`     //
}

// Content - define content wich will be saved or loaded
type Content struct {
	Payload string `json:"payload"`
	Hash    string `json:"hashsum"`
}

// Version - struct with vault version description
type Version struct {
	Date time.Time `json:"time"`
	Hash string    `json:"hashsum"`
}

// PGP - struct to operate with list of pgp public keys
type PGP struct {
	Date      time.Time `json:"time,omitempty"`
	Publickey string    `json:"publickey"`
	Confirmed bool      `json:"confirmed,omitempty"`
}

// Message - struct to notify logged in user
type Message struct {
	Text    string `json:"text"`
	Content string `json:"content"`
}
