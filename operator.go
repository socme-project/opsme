package opsme

type Operator struct {
	OperatorID   string    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Machines     []Machine `json:"machines"`
}

func (o *Operator) GetMachineByID(machineID string) *Machine {
	for _, machine := range o.Machines {
		if machine.ID == machineID {
			return &machine
		}
	}
	return nil
}

func (o *Operator) GetMachineByName(machineName string) *Machine {
	for _, machine := range o.Machines {
		if machine.Name == machineName {
			return &machine
		}
	}
	return nil
}

func (o *Operator) AddMachine(machine Machine) {
	o.Machines = append(o.Machines, machine)
}

func (o *Operator) RemoveMachine(machineID string) {
	for i, machine := range o.Machines {
		if machine.ID == machineID {
			o.Machines = append(o.Machines[:i], o.Machines[i+1:]...)
			return
		}
	}
}
