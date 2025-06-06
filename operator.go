package opsme

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"slices"

	"golang.org/x/crypto/ssh"
)

type CommandOutput struct {
	ClientName string
	MachineID  string
	Message    string
	Out        string
	Error      bool
}

type Operator struct {
	OperatorID   string     `json:"operator_id"`
	OperatorName string     `json:"operator_name"`
	Machines     []*Machine `json:"machines"`
	sshSigner    ssh.Signer
}

func NewOperator(operatorID, operatorName string) (*Operator, error) {
	op := &Operator{
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Machines:     []*Machine{},
	}

	signer, err := op.loadSSHKey()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize operator SSH signer: %w", err)
	}
	op.sshSigner = signer
	return op, nil
}

func (o *Operator) loadSSHKey() (ssh.Signer, error) {
	privateKeyB64 := os.Getenv("OPSME_SSH_KEY_BASE64")
	if privateKeyB64 == "" {
		return nil, fmt.Errorf("OPSME_SSH_KEY_BASE64 environment variable is not set")
	}

	privateKey, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("error decoding SSH key from base64: %w", err)
	}
	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing SSH private key: %w", err)
	}
	return key, nil
}

func (o *Operator) GetMachineByID(machineID string) *Machine {
	for _, machine := range o.Machines {
		if machine.ID == machineID {
			return machine
		}
	}
	return nil
}

func (o *Operator) GetMachineByName(machineName string) *Machine {
	for _, machine := range o.Machines {
		if machine.Name == machineName {
			return machine
		}
	}
	return nil
}

func (o *Operator) AddMachine(machine *Machine) {
	o.Machines = append(o.Machines, machine)
}

func (o *Operator) RemoveMachine(machineID string) {
	for i, machine := range o.Machines {
		if machine.ID == machineID {
			o.Machines = slices.Delete(o.Machines, i, i+1)
			return
		}
	}
}

func (o *Operator) LoadMachinesFromPocketBase() error {
	artemisRecords, err := GetArtemisMachines()
	if err != nil {
		return fmt.Errorf("failed to load machines from PocketBase: %w", err)
	}

	for _, record := range artemisRecords {
		o.AddMachine(&Machine{
			ID:   record.ID,
			Name: record.Hostname,
			SSHLogin: SSHLogin{
				IP:       record.IP,
				Username: "root",
			},
		})
	}
	return nil
}

func (o *Operator) ExecuteCommandOnMachines(command string, machineIDs ...string) []CommandOutput {
	var results []CommandOutput

	if o.sshSigner == nil {
		results = append(results, CommandOutput{
			Message: "SSH signer not loaded. Operator not initialized correctly.",
			Error:   true,
		})
		return results
	}

	machinesToTarget := []*Machine{}
	if len(machineIDs) == 0 {
		machinesToTarget = o.Machines
	} else {
		for _, id := range machineIDs {
			if m := o.GetMachineByID(id); m != nil {
				machinesToTarget = append(machinesToTarget, m)
			} else {
				results = append(results, CommandOutput{
					MachineID: id,
					Message:   fmt.Sprintf("Machine with ID '%s' not found in operator's fleet.", id),
					Error:     true,
				})
			}
		}
	}

	if len(machinesToTarget) == 0 {
		results = append(results, CommandOutput{
			Message: "No machines to execute command on.",
			Error:   true,
		})
		return results
	}

	for _, machine := range machinesToTarget {
		conn, err := machine.connect(o.sshSigner)
		if err != nil {
			results = append(results, CommandOutput{
				ClientName: machine.Name,
				MachineID:  machine.ID,
				Message:    fmt.Sprintf("SSH connection error: %s", err),
				Error:      true,
			})
			continue
		}
		defer func(c *ssh.Client, name string) {
			if err := c.Close(); err != nil {
				log.Printf("WARNING: Failed to close SSH connection for %s: %v", name, err)
			}
		}(conn, machine.Name)

		output, err := machine.sendCommand(conn, command)
		if err != nil {
			results = append(results, CommandOutput{
				ClientName: machine.Name,
				MachineID:  machine.ID,
				Message:    fmt.Sprintf("Command execution error: %s", err),
				Out:        output,
				Error:      true,
			})
		} else {
			results = append(results, CommandOutput{
				ClientName: machine.Name,
				MachineID:  machine.ID,
				Message:    "Command executed successfully.",
				Out:        output,
				Error:      false,
			})
		}
	}
	return results
}

func (o *Operator) GetMachineStatus(machineIDs ...string) []CommandOutput {
	return o.ExecuteCommandOnMachines("systemctl status", machineIDs...)
}

func (o *Operator) Update(machineIDs ...string) []CommandOutput {
	updateCommand := "cd /etc/nixos && git stash && git pull && nix flake update && sudo nixos-rebuild switch --upgrade --flake /etc/nixos#artemis && git stash pop"
	return o.ExecuteCommandOnMachines(updateCommand, machineIDs...)
}

func (o *Operator) Reboot(machineIDs ...string) []CommandOutput {
	return o.ExecuteCommandOnMachines("sudo reboot", machineIDs...)
}

func (o *Operator) FetchMe(machineIDs ...string) []CommandOutput {
	return o.ExecuteCommandOnMachines("fastfetch", machineIDs...)
}
