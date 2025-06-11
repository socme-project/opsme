package opsme

import (
	"errors"
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
		Auth:     auth, // Auth is assigned directly from the parameter
	}

	op.Machines = append(op.Machines, machine)
	return machine, nil
}

func (op Operator) Run(command string) (outputs []Output, errs []error) {
	outputs = make([]Output, 0, len(op.Machines))
	errs = make([]error, 0, len(op.Machines))

	for _, machine := range op.Machines {
		output, err := machine.Run(command)
		outputs = append(outputs, output)
		errs = append(errs, err)
	}
	return outputs, errs
}
