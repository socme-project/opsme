package opsme

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
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
	Name           string
	Username       string
	Host           string
	Port           int
	HostKey        string
	Auth           Auth
	KnownHostsPath string
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

func (m Machine) Run(command string) (Output, error) {
	var hostKeyCallback ssh.HostKeyCallback
	var err error

	if m.HostKey != "" {
		var parsedKey ssh.PublicKey
		parsedKey, err = ssh.ParsePublicKey([]byte(m.HostKey))
		if err != nil {
			return Output{}, fmt.Errorf(
				"machine '%s': invalid explicit HostKey provided: %w",
				m.Name,
				err,
			)
		}
		hostKeyCallback = ssh.FixedHostKey(parsedKey)
	} else {
		lookupPath := m.KnownHostsPath
		if lookupPath == "" {
			currentUser, userErr := user.Current()
			if userErr != nil {
				return Output{}, fmt.Errorf("machine '%s': failed to get current user home directory for default known_hosts path: %w", m.Name, userErr)
			}
			lookupPath = filepath.Join(currentUser.HomeDir, ".ssh", "known_hosts")
		}

		if _, statErr := os.Stat(lookupPath); os.IsNotExist(statErr) {
			return Output{}, fmt.Errorf("machine '%s': no explicit host key provided and known_hosts file (%s) not found. Host key enforcement requires a known key or a valid known_hosts file.", m.Name, lookupPath)
		}

		hostKeyCallback, err = knownhosts.New(lookupPath)
		if err != nil {
			return Output{}, fmt.Errorf("machine '%s': failed to load known_hosts file %s: %w", m.Name, lookupPath, err)
		}
	}

	config := &ssh.ClientConfig{
		User:            m.Username,
		HostKeyCallback: hostKeyCallback,
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
							!echoprompts[i] {
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
