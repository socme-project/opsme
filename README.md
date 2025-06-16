# OPSme

OPSme is a golang library that provides a simple way to manage multiple machines via SSH. It allows you to execute commands on different machines concurrently, making it easier to manage large infrastructures.

## Installation

To install OPSme, you can use the following command:

```bash
go get github.com/socme-project/opsme
```

## Usage

To use OPSme, you need to import the package in your Go code:

```go
import "github.com/socme-project/opsme"
```

You can then create a new instance of OPSme and use it to execute commands on multiple machines. Here is a simple example:

```go

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/socme-project/opsme"
)

func main() {
	operator := opsme.New(
		true,
	) // true here means that the HostKey will be automatically added to the known_hosts file

	m1, err := operator.NewMachine(
		"machineName",
		"user",
		"192.168.1.2",
		22,
		opsme.Auth{AuthType: opsme.AuthTypePassword, Password: "test1234"},
	)
	if err != nil {
		log.Printf("%v", err)
	} else {
		fmt.Println("The following machine was created successfully:", m1.Name)
		result, runErr := m1.Run("pwd")
		fmt.Println("Running 'pwd' on machine:", m1.Name)
		if runErr != nil {
			log.Printf("Error running command on %v: %v", m1.Name, runErr)
		} else {
			fmt.Println("Output from 'machineName' :", strings.TrimSpace(result.Output))
		}
	}

	privateKey := `-----BEGIN OPENSSH PRIVATE KEY-----
	.
	.
	.
	.
	.
	.
	.
	-----END OPENSSH PRIVATE KEY-----`

	m2, err := operator.NewMachine(
		"machineName2",
		"dilounix",
		"hyrule",
		22,
		opsme.Auth{
			AuthType: opsme.AuthTypeSshKey,
			SshKey:   privateKey,
		},
	)

	if err != nil {
		log.Printf("%v", err)
	} else {
		fmt.Println("The following machine was created successfully:", m2.Name)
		result, runErr := m2.Run("pwd")
		fmt.Println("Running 'pwd' on machine:", m2.Name)
		if runErr != nil {
			log.Printf("Error running command on %v: %v", m2.Name, runErr)
		} else {
			fmt.Println("Output from 'machineName2':", strings.TrimSpace(result.Output))
		}
	}

	results, runErrors := operator.Run("id")

	for i := range results {
		machineName := results[i].MachineName
		if runErrors[i] != nil {
			fmt.Printf("Error on machine %s: %v\n", machineName, runErrors[i])
		} else {
			fmt.Printf("Machine name: %s\n", machineName)
			fmt.Printf("Output: %s\n", strings.TrimSpace(results[i].Output))
		}
	}

	if len(operator.Machines) > 0 {
		firstMachine := operator.Machines[0]
		result, runErr := firstMachine.Run("uname -a")
		fmt.Printf("Running 'uname -a' on machine: %s\n", firstMachine.Name)
		if runErr != nil {
			log.Printf("Error running command on %s: %v", firstMachine.Name, runErr)
		} else {
			fmt.Printf("Machine name: %s\n", result.MachineName)
			fmt.Printf("Output: %s\n", strings.TrimSpace(result.Output))
		}
	} else {
		fmt.Println("No machines were successfully added to the operator to run 'uname -a' on.")
	}

	numMachines := len(operator.Machines)
	fmt.Printf("Total number of machines: %d\n", numMachines)
}
```

You can also change the authentication method for a machine after it has been created:

```go
m1.WithPasswordAuth("newpassword")
```
