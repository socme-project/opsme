package opsme

import (
	"encoding/base64"
	"fmt" // Import the log package
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

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
	HostKey  string
	Auth     Auth
}

func (m *Machine) WithPasswordAuth(password string) {
	m.Auth = Auth{
		AuthType: AuthTypePassword,
		Password: password,
	}
}

func (m *Machine) WithSshKeyAuth(sshKey string) {
	m.Auth = Auth{
		AuthType: AuthTypeSshKey,
		SshKey:   sshKey,
	}
}

func keyString(k ssh.PublicKey) string {
	return k.Type() + " " + base64.StdEncoding.EncodeToString(
		k.Marshal(),
	)
}

func trustedHostKeyCallback(trustedKey string) ssh.HostKeyCallback {
	if trustedKey == "" {
		return func(_ string, _ net.Addr, k ssh.PublicKey) error {
			return nil
		}
	}

	return func(_ string, _ net.Addr, k ssh.PublicKey) error {
		ks := keyString(k)
		if trustedKey != ks {
			return fmt.Errorf("SSH-key verification: expected %q but got %q", trustedKey, ks)
		}
		return nil
	}
}

func (m Machine) Run(command string) (Output, error) {
	config := &ssh.ClientConfig{
		User:            m.Username,
		HostKeyCallback: trustedHostKeyCallback(m.HostKey),
		Timeout:         10 * time.Second,
	}

	authMethods := []ssh.AuthMethod{}

	switch m.Auth.AuthType {
	case AuthTypePassword:
		authMethods = append(authMethods, ssh.Password(m.Auth.Password))
		authMethods = append(
			authMethods,
			ssh.KeyboardInteractive(
				func(user, instruction string, questions []string, echoprompts []bool) ([]string, error) {
					answers := make([]string, len(questions))
					for i, q := range questions {
						if (q == "Password:" || q == "password:" || q == fmt.Sprintf("%s@%s's password:", user, m.Host)) &&
							echoprompts[i] == false {
							answers[i] = m.Auth.Password
						} else {
							return nil, fmt.Errorf("unsupported keyboard-interactive question: %s", q)
						}
					}
					return answers, nil
				},
			),
		)

	case AuthTypeSshKey:
		signer, err := ssh.ParsePrivateKey([]byte(m.Auth.SshKey))
		if err != nil {
			return Output{
					MachineName: m.Name,
				}, fmt.Errorf(
					"machine '%s': failed to parse SSH key: %w",
					m.Name,
					err,
				)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	default:
		return Output{
			MachineName: m.Name,
		}, fmt.Errorf("machine '%s': authentication type not set or unsupported", m.Name)
	}

	config.Auth = authMethods

	addr := m.Host + ":" + strconv.Itoa(m.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return Output{
				MachineName: m.Name,
			}, fmt.Errorf(
				"machine '%s': failed to connect to %s: %w",
				m.Name,
				addr,
				err,
			)
	}

	defer func() {
		_ = client.Close()
	}()

	session, err := client.NewSession()
	if err != nil {
		return Output{
				MachineName: m.Name,
			}, fmt.Errorf(
				"machine '%s': failed to create SSH session: %w",
				m.Name,
				err,
			)
	}

	defer func() {
		_ = session.Close()
	}()

	outputBytes, err := session.CombinedOutput(command)
	if err != nil {
		return Output{
				MachineName: m.Name,
			}, fmt.Errorf(
				"machine '%s': command '%s' failed: %w. Output: %s",
				m.Name,
				command,
				err,
				string(outputBytes),
			)
	}

	return Output{
		MachineName: m.Name,
		Output:      string(outputBytes),
	}, nil
}
