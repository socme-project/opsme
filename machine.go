package opsme

import (
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHLogin struct {
	IP       string `json:"ip"`
	Username string `json:"username"`
}

type Machine struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SSHLogin SSHLogin
}

func (m *Machine) connect(signer ssh.Signer) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: m.SSHLogin.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: 5 * time.Second,
		HostKeyCallback: ssh.NewCallback(
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				log.Printf(
					"INFO: SSH Host Key Callback for %s. Fingerprint: %s",
					hostname,
					ssh.FingerprintSHA256(key),
				)
				return nil
			},
		),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", m.SSHLogin.IP), config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH to %s (%s): %w", m.Name, m.SSHLogin.IP, err)
	}
	return client, nil
}

func (m *Machine) sendCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session for %s: %w", m.Name, err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(
				output,
			), fmt.Errorf(
				"failed to execute command '%s' on %s: %w",
				command,
				m.Name,
				err,
			)
	}

	return string(output), nil
}

