package opsme

import "errors"

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

func (op *Operator) NewMachine(machineName, username, host string, port int) (Machine, error) {
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
		Host:     host,
		Port:     port,
	}
	op.Machines = append(op.Machines, machine)
	return machine, nil
}

func (op Operator) Run(command string) (outputs []Output, errs []error) {
	for _, machine := range op.Machines {
		output, err := machine.Run(command)
		outputs = append(outputs, output)
		errs = append(errs, err)
	}
	return outputs, errs
}
