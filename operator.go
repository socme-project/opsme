package opsme

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"sync"
	"time"
)

// Output represents the output of a command executed on a machine.
type Output struct {
	MachineName string
	Output      string
}

// Operator manages multiple machines and provides methods to run commands on them.
type Operator struct {
	Machines        []*Machine
	AddToKnownHosts bool
	KnownHostsPath  string
	Timeout         time.Duration
}

// New initializes a new Operator instance with the option to add to known_hosts and sets a timeout for operations.
func New(addToKnownHosts bool, timeout time.Duration) (Operator, error) {
	var defaultKnownHostsPath string
	currentUser, err := user.Current()
	if err != nil {
		return Operator{}, fmt.Errorf(
			"failed to determine user home directory for default known_hosts path: %w",
			err,
		)
	}
	defaultKnownHostsPath = filepath.Join(currentUser.HomeDir, ".ssh", "known_hosts")

	op := Operator{
		Machines:        []*Machine{},
		AddToKnownHosts: addToKnownHosts,
		KnownHostsPath:  defaultKnownHostsPath,
		Timeout:         timeout * time.Second,
	}
	return op, nil
}

// NewMachine creates a new Machine instance with the specified parameters.
func (op *Operator) NewMachine(
	machineName, username, host string,
	port int,
) (*Machine, error) {
	for _, m := range op.Machines {
		if m.Name == machineName {
			return nil, errors.New("machine already exists")
		}
	}

	if machineName == "" || username == "" || host == "" {
		return nil, errors.New("arguments cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return nil, errors.New("port must be between 1 and 65535")
	}

	machine := &Machine{
		Name:            machineName,
		Username:        username,
		Host:            host,
		Port:            port,
		auth:            auth{},
		KnownHostsPath:  op.KnownHostsPath,
		AddToKnownHosts: op.AddToKnownHosts,
		Timeout:         op.Timeout,
	}

	op.Machines = append(op.Machines, machine)
	return machine, nil
}

// Run executes a command on all machines managed by the Operator concurrently.
func (op Operator) Run(command string) (outputs []Output, errs []error) {
	wg := sync.WaitGroup{}
	outputs = make([]Output, len(op.Machines))
	errs = make([]error, len(op.Machines))

	for i, machine := range op.Machines {
		wg.Add(1)

		go func(index int, m *Machine) {
			defer wg.Done()
			outputs[index], errs[index] = m.Run(command)
		}(i, machine)

	}

	wg.Wait()

	return outputs, errs
}
