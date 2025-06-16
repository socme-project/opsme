package opsme

import (
	"errors"
	"fmt"
	"os/user"
	"path/filepath"
	"sync"
)

type Output struct {
	MachineName string
	Output      string
}

type Operator struct {
	Machines        []Machine
	AddToKnownHosts bool
	KnownHostsPath  string
}

func New(addToKnownHosts bool) Operator {
	var defaultKnownHostsPath string
	currentUser, _ := user.Current()
	defaultKnownHostsPath = filepath.Join(currentUser.HomeDir, ".ssh", "known_hosts")

	op := Operator{
		Machines:        []Machine{},
		AddToKnownHosts: addToKnownHosts,
		KnownHostsPath:  defaultKnownHostsPath,
	}
	return op
}

func (op *Operator) NewMachine(
	machineName, username, host string,
	port int,
	auth Auth,
) (Machine, error) {
	for _, m := range op.Machines {
		if m.Name == machineName {
			return Machine{}, errors.New("machine already exists")
		}
	}

	if machineName == "" || username == "" || host == "" {
		return Machine{}, errors.New("arguments cannot be empty")
	}
	if port <= 0 || port > 65535 {
		return Machine{}, errors.New("port must be between 1 and 65535")
	}

	machine := Machine{
		Name:            machineName,
		Username:        username,
		Host:            host,
		Port:            port,
		Auth:            auth,
		KnownHostsPath:  op.KnownHostsPath,
		AddToKnownHosts: op.AddToKnownHosts,
	}

	client, err := machine.newSSHClient()
	if err != nil {
		return Machine{}, fmt.Errorf(
			"initial connection and authentication failed for machine '%s': %w",
			machineName,
			err,
		)
	}

	defer func() {
		_ = client.Close()
	}()

	op.Machines = append(op.Machines, machine)
	return machine, nil
}

func (op Operator) Run(command string) (outputs []Output, errs []error) {
	numMachines := len(op.Machines)
	outputs = make([]Output, numMachines)
	errs = make([]error, numMachines)

	var wg sync.WaitGroup

	for i, machine := range op.Machines {
		wg.Add(1)

		mCopy := machine

		go func(index int, m Machine) {
			defer wg.Done()

			output, err := m.Run(command)
			outputs[index] = output
			errs[index] = err
		}(i, mCopy)
	}

	wg.Wait()

	return outputs, errs
}
