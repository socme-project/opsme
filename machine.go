package opsme

import (
	"fmt"
	"net"
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
	AuthTypeSSHKey
)

type Auth struct {
	AuthType AuthType
	Password string
	SSHKey   []byte
}

type Machine struct {
	Name            string
	Username        string
	Host            string
	Port            int
	auth            Auth
	KnownHostsPath  string
	AddToKnownHosts bool
}

func (m *Machine) WithPasswordAuth(password string) error {
	m.auth = Auth{
		AuthType: AuthTypePassword,
		Password: password,
	}
	client, err := m.newSSHClient()
	if err != nil {
		return fmt.Errorf("failed to authenticate machine '%s' with password: %w", m.Name, err)
	}

	func() {
		_ = client.Close()
	}()
	return nil
}

func (m *Machine) WithSSHKeyAuth(sshKey []byte) error {
	m.auth = Auth{
		AuthType: AuthTypeSSHKey,
		SSHKey:   sshKey,
	}
	// Verify authentication immediately upon setting
	client, err := m.newSSHClient() // Create a client to test auth, then it's closed
	if err != nil {
		return fmt.Errorf("failed to authenticate machine '%s' with SSH key: %w", m.Name, err)
	}
	func() {
		_ = client.Close()
	}()
	return nil
}

func (m Machine) newSSHClient() (*ssh.Client, error) {
	var hostKeyCallback ssh.HostKeyCallback

	lookupPath := m.KnownHostsPath
	if lookupPath == "" {
		currentUser, userErr := user.Current()
		if userErr != nil {
			return nil, fmt.Errorf(
				"machine '%s': failed to get user home directory for known_hosts: %w",
				m.Name,
				userErr,
			)
		}
		lookupPath = filepath.Join(currentUser.HomeDir, ".ssh", "known_hosts")
	}

	knownHostsChecker, _ := knownhosts.New(lookupPath)

	hostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		err := knownHostsChecker(hostname, remote, key)

		if _, ok := err.(*knownhosts.KeyError); ok {
			if m.AddToKnownHosts {
				f, openErr := os.OpenFile(lookupPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
				if openErr != nil {
					return fmt.Errorf(
						"machine '%s': failed to open known_hosts file '%s' for appending: %w",
						m.Name,
						lookupPath,
						openErr,
					)
				}

				defer func() {
					_ = f.Close()
				}()

				hostEntry := knownhosts.Line([]string{hostname}, key)
				if _, writeErr := f.WriteString(hostEntry + "\n"); writeErr != nil {
					return fmt.Errorf(
						"machine '%s': failed to write host key to known_hosts file '%s': %w",
						m.Name,
						lookupPath,
						writeErr,
					)
				}
				return nil
			} else {
				return fmt.Errorf("machine '%s': host key for '%s' not found in known_hosts '%s' and auto-adding is disabled. Original error: %w", m.Name, hostname, lookupPath, err)
			}
		}
		return err
	}

	config := &ssh.ClientConfig{
		User:            m.Username,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	authMethods := []ssh.AuthMethod{}
	switch m.auth.AuthType {
	case AuthTypePassword:
		authMethods = append(authMethods, ssh.Password(m.auth.Password))
		authMethods = append(
			authMethods,
			ssh.KeyboardInteractive(
				func(user, instruction string, questions []string, echoprompts []bool) ([]string, error) {
					answers := make([]string, len(questions))
					for i, q := range questions {
						if (q == "Password:" || q == "password:" || q == fmt.Sprintf("%s@%s's password:", user, m.Host)) &&
							!echoprompts[i] {
							answers[i] = m.auth.Password
						} else {
							return nil, fmt.Errorf("unsupported keyboard-interactive question: %s", q)
						}
					}
					return answers, nil
				},
			),
		)
	case AuthTypeSSHKey:
		signer, err := ssh.ParsePrivateKey(m.auth.SSHKey)
		if err != nil {
			return nil, fmt.Errorf(
				"machine '%s': failed to parse SSH key: %w",
				m.Name,
				err,
			)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("machine '%s': authentication type not set or unsupported", m.Name)
	}

	config.Auth = authMethods

	addr := m.Host + ":" + strconv.Itoa(m.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf(
			"machine '%s': failed to connect to %s: %w",
			m.Name,
			addr,
			err,
		)
	}
	return client, nil
}

func (m Machine) Run(command string) (Output, error) {
	if m.auth.AuthType == 0 && m.auth.Password == "" && len(m.auth.SSHKey) == 0 {
		return Output{
				MachineName: m.Name,
			}, fmt.Errorf(
				"machine '%s': authentication not set",
				m.Name,
			)
	}

	client, err := m.newSSHClient()
	if err != nil {
		return Output{
			MachineName: m.Name,
		}, err
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
