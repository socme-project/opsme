package opsme

type SSHLogin struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Machine struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (m *Machine) sendCommand(command string) (string, error) {
	return "", nil
}
