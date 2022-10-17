package secrets

const scopeUserPass = "user-password"

// UserPass - type for UserPass secret
type UserPass struct {
	Username string
	Password string
}

// NewUserPass - create USERPASS secret
func NewUserPass(username, password string) *UserPass {
	up := UserPass{
		Username: username,
		Password: password,
	}
	return &up
}

// Scope - secret scope
func (s *UserPass) Scope() string {
	return scopeUserPass
}

// Value - secret value
func (s *UserPass) Value() interface{} {
	return s
}
