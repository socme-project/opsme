package opsme

type AuthType int

const (
	AuthTypePassword AuthType = iota
	AuthTypeSshKey
)

type Auth struct {
	AuthType AuthType
	Password string
	SshKey   string
}

type Machine struct {
	Name     string
	Username string
	Host     string
	Port     int
	Auth     Auth
}

func (m *Machine) WithPasswordAuth(password string) *Machine {
	m.Auth = Auth{
		AuthType: AuthTypePassword,
		Password: password,
	}
	return m
}

func (m *Machine) WithSshKeyAuth(sshKey string) *Machine {
	m.Auth = Auth{
		AuthType: AuthTypeSshKey,
		SshKey:   sshKey,
	}
	return m
}

//func (m Machine) Run(command string) (Output, error) {
	// Run the command without taking care of auth since auth will work like this : NetMachine.Run(command).WithAuth(m.Auth)


