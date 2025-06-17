package opsme

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type authType int

const (
	authTypePassword authType = iota
	authTypeSSHKey
)

type auth struct {
	authType authType
	password string
	sshKey   []byte
}

type Machine struct {
	Name              string
	Username          string
	Host              string
	Port              int
	auth              auth
	KnownHostsPath    string
	AddToKnownHosts   bool
	Timeout           time.Duration
	knownHostsChecker ssh.HostKeyCallback // Stored on Machine after initialization
}

func (m *Machine) WithPasswordAuth(password string) error {
	m.auth = auth{
		authType: authTypePassword,
		password: password,
	}
	client, err := m.newSSHClient()
	if err != nil {
		return err
	}

	err = client.Close()
	return err
}

func (m *Machine) WithSSHKeyAuth(sshKey []byte) error {
	m.auth = auth{
		authType: authTypeSSHKey,
		sshKey:   sshKey,
	}

	client, err := m.newSSHClient()
	if err != nil {
		return fmt.Errorf("failed to authenticate machine '%s' with SSH key: %w", m.Name, err)
	}
	err = client.Close()
	return err
}

// hostKeyCallbackMethod handles host key verification.
// If AddToKnownHosts is true and the key is not in known_hosts, it attempts to add it.
func (m *Machine) hostKeyCallbackMethod(
	hostname string,
	remote net.Addr,
	key ssh.PublicKey,
) error {
	err := m.knownHostsChecker(hostname, remote, key)

	// TODO: Check the error
	if err == nil {
		return nil
	}

	if m.AddToKnownHosts {
		f, err := os.OpenFile(m.KnownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf(
				"machine '%s': failed to open known_hosts file '%s' for appending: %w",
				m.Name,
				m.KnownHostsPath,
				err,
			)
		}
		defer func() {
			_ = f.Close()
		}()

		// Format the host entry and write it.
		hostEntry := knownhosts.Line([]string{hostname}, key)
		if _, err := f.WriteString(hostEntry + "\n"); err != nil {
			return fmt.Errorf(
				"machine '%s': failed to write host key to known_hosts file '%s': %w",
				m.Name,
				m.KnownHostsPath,
				err,
			)
		}
		// Key successfully added, so the connection can now proceed.
		return nil
	}

	return fmt.Errorf(
		"machine '%s': host key validation failed for '%s' in known_hosts '%s'. Auto-adding is disabled. Original error: %w",
		m.Name,
		hostname,
		m.KnownHostsPath,
		err,
	)
}

// keyboardInteractiveMethod handles responding to keyboard-interactive prompts (e.g., for passwords).
func (m *Machine) keyboardInteractiveMethod(
	user, instruction string,
	questions []string,
	echoprompts []bool,
) ([]string, error) {
	answers := make([]string, len(questions))
	for i, q := range questions {
		if (q == "Password:" || q == "password:" || q == fmt.Sprintf("%s@%s's password:", user, m.Host)) &&
			!echoprompts[i] {
			answers[i] = m.auth.password
		} else {
			return nil, fmt.Errorf("unsupported keyboard-interactive question: %s", q)
		}
	}
	return answers, nil
}

func (m *Machine) newSSHClient() (*ssh.Client, error) {
	if m.KnownHostsPath == "" {
		return nil, fmt.Errorf(
			"machine '%s': KnownHostsPath is empty. It must be set by the Operator",
			m.Name,
		)
	}

	// Initialize knownHostsChecker directly on the Machine instance
	var err error
	m.knownHostsChecker, err = knownhosts.New(m.KnownHostsPath)
	if err != nil {
		return nil, fmt.Errorf(
			"machine '%s': failed to initialize known_hosts checker for path '%s': %w",
			m.Name,
			m.KnownHostsPath,
			err,
		)
	}

	config := &ssh.ClientConfig{
		User:            m.Username,
		HostKeyCallback: m.hostKeyCallbackMethod, // Refer to the method directly
		Timeout:         m.Timeout,
	}

	config.Auth = []ssh.AuthMethod{}
	switch m.auth.authType {
	case authTypePassword:
		config.Auth = append(
			config.Auth,
			ssh.Password(m.auth.password),
		)
		config.Auth = append(
			config.Auth,
			ssh.KeyboardInteractive(m.keyboardInteractiveMethod), // Refer to the method directly
		)
	case authTypeSSHKey:
		signer, err := ssh.ParsePrivateKey(m.auth.sshKey)
		if err != nil {
			return nil, fmt.Errorf(
				"machine '%s': failed to parse SSH key: %w",
				m.Name,
				err,
			)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	default:
		return nil, fmt.Errorf("machine '%s': authentication type not set or unsupported", m.Name)
	}

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

func (m *Machine) Run(command string) (Output, error) {
	if m.auth.authType == 0 && m.auth.password == "" &&
		len(m.auth.sshKey) == 0 {
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
