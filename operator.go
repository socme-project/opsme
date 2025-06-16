package opsme

import (
	"errors"
	"sync"
)

type Output struct {
	MachineName string
	Output      string
}

type Operator struct {
	Machines []Machine
}

func New() Operator {
	op := Operator{
		Machines: []Machine{},
	}
	return op
}

func (op *Operator) NewMachine(
	machineName, username, hostkey, host string,
	port int,
	auth Auth, // Now Auth is passed directly here
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
		Name:     machineName,
		Username: username,
		HostKey:  hostkey,
		Host:     host,
		Port:     port,
		Auth:     auth,
	}

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
