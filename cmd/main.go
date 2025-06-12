package main

import (
	"fmt"
	"log"

	"github.com/socme-project/opsme"
)

func main() {
	operator := opsme.New()

	m1, err := operator.NewMachine(
		"hyrule",
		"app-systeme-ch13",
		"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBM7gEfaLJmXJwkDnS6BlTG2IVwAnr6GT3GSCXgEmjBOkRjk5se9sVPoRE2IT8XR/uXm2w/q8QJmdC2VOvIa1Jnw=",
		"challenge02.root-me.org",
		2222,
		opsme.Auth{AuthType: opsme.AuthTypePassword, Password: "app-systeme-ch13"},
	)
	if err != nil {
		log.Fatalf("Error creating machine hyrule: %v", err)
	}
	fmt.Println("Machine created:", m1.Name)

	m2, err := operator.NewMachine(
		"zeus",
		"zeus",
		"",
		"192.168.1.101",
		22,
		opsme.Auth{
			AuthType: opsme.AuthTypeSshKey,
			SshKey: "-----BEGIN OPENSSH PRIVATE KEY-----\n" + "b3BlbnNzaC1rZXktdjEAAAAABG5vbm"
		},
	)
	if err != nil {
		log.Fatalf("Error creating machine machineName2: %v", err)
	}
	fmt.Println("Machine created:", m2.Name)

	fmt.Println("\n--- Running 'id' command on all machines ---")
	results, errors := operator.Run("id")

	var successfulOutputs []opsme.Output
	var collectedErrors []error

	for i, result := range results {
		if errors[i] == nil {
			successfulOutputs = append(successfulOutputs, result)
		} else {
			collectedErrors = append(collectedErrors, fmt.Errorf("Error on machine %s: %v", operator.Machines[i].Name, errors[i]))
		}
	}

	fmt.Println("Command execution attempted on all machines.")

	// Print successful outputs first
	for _, result := range successfulOutputs {
		if result.MachineName != "" {
			fmt.Println("Machine name:", result.MachineName)
			fmt.Println("Output:", result.Output)
		}
	}

	// Then print errors
	for _, err := range collectedErrors {
		fmt.Println(err)
	}

	fmt.Println("\n--- Running 'uname -a' command on specific machines ---")

	var hyruleMachine, machine2FromOperator opsme.Machine
	for _, m := range operator.Machines {
		if m.Name == "hyrule" {
			hyruleMachine = m
		} else if m.Name == "machineName2" {
			machine2FromOperator = m
		}
	}

	if hyruleMachine.Name != "" {
		resultHyrule, errHyrule := hyruleMachine.Run("uname -a")
		if errHyrule != nil {
			fmt.Println("Error running command on hyrule:", errHyrule)
		} else {
			fmt.Println("Machine name:", resultHyrule.MachineName)
			fmt.Println("Output:", resultHyrule.Output)
		}
	} else {
		fmt.Println("Machine 'hyrule' not found in operator's list.")
	}

	if machine2FromOperator.Name != "" {
		resultM2, errM2 := machine2FromOperator.Run("uname -a")
		if errM2 != nil {
			fmt.Println("Error running command on machineName2:", errM2)
		} else {
			fmt.Println("Machine name:", resultM2.MachineName)
			fmt.Println("Output:", resultM2.Output)
		}
	} else {
		fmt.Println("Machine 'machineName2' not found in operator's list.")
	}
}
