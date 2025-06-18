package opsme

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// authType represents the type of authentication used for SSH connections.
type authType int

// Constants for different authentication types.
const (
	authTypePassword authType = iota
	authTypeSSHKey
)

// auth holds the authentication details for a Machine.
type auth struct {
	authType authType
	password string
	sshKey   []byte
}

// Machine holds the details of a remote machine that can be accessed via SSH.
type Machine struct {
	Name              string
	Username          string
	Host              string
	Port              int
	auth              auth
	KnownHostsPath    string
	AddToKnownHosts   bool
	Timeout           time.Duration
	knownHostsChecker ssh.HostKeyCallback
}

// WithPasswordAuth sets the authentication method to password-based for the Machine and takes a password as input.
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

// WithSSHKeyAuth sets the authentication method to SSH key-based for the Machine and takes the private key as input.
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

// newSSHClient creates a new SSH client for the Machine.
func (m *Machine) newSSHClient() (*ssh.Client, error) {
	if m.KnownHostsPath == "" {
		return nil, fmt.Errorf(
			"machine '%s': KnownHostsPath is empty. It must be set by the Operator",
			m.Name,
		)
	}

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
		HostKeyCallback: m.hostKeyCallbackMethod,
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
			ssh.KeyboardInteractive(m.keyboardInteractiveMethod),
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

// Run executes a command on the Machine via SSH and returns the output.
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

// hostKeyCallbackMethod handles host key verification.
// If AddToKnownHosts is true and the key is not in known_hosts, it attempts to add it.
func (m *Machine) hostKeyCallbackMethod(
	hostname string,
	remote net.Addr,
	key ssh.PublicKey,
) error {
	err := m.knownHostsChecker(hostname, remote, key)
	if err == nil {
		return nil
	}

	var keyError *knownhosts.KeyError
	if !errors.As(err, &keyError) {
		return fmt.Errorf(
			"machine '%s': failed to check known_hosts file '%s': %w",
			m.Name,
			m.KnownHostsPath,
			err,
		)
	}

	if len(keyError.Want) > 0 {
		return fmt.Errorf(
			"machine '%s': @@@@@ WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED! @@@@@\n"+
				"IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY (man-in-the-middle attack)!\n"+
				"The key for host '%s' in '%s' has changed. The new key's fingerprint is %s",
			m.Name,
			hostname,
			m.KnownHostsPath,
			ssh.FingerprintSHA256(key),
		)
	}

	if !m.AddToKnownHosts {
		return fmt.Errorf(
			"machine '%s': host key for '%s' is not trusted. Auto-adding is disabled",
			m.Name,
			hostname,
		)
	}

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

	hostEntry := knownhosts.Line([]string{hostname}, key)
	if _, err := f.WriteString(hostEntry + "\n"); err != nil {
		return fmt.Errorf(
			"machine '%s': failed to write host key to known_hosts file '%s': %w",
			m.Name,
			m.KnownHostsPath,
			err,
		)
	}

	return nil
}

// keyboardInteractiveMethod handles responding to keyboard-interactive prompts for authentication.
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
