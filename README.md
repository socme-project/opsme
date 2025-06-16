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
	operator := opsme.New(true) // true means HostKey will be automatically added if missing

	m1, err := operator.NewMachine(
		"machine1",
		"user",
		"192.168.1.2",
		22,
	)
	if err != nil {
		log.Printf("Failed to create machine1: %v", err)
	} else {
		fmt.Printf("Machine %s created. Setting password auth...\n", m1.Name)
		authErr := m1.WithPasswordAuth("test1234")
		if authErr != nil {
			log.Printf("Failed to authenticate machine %s: %v\n", m1.Name, authErr)
		} else {
			fmt.Printf("Machine %s authenticated. Running 'pwd'...\n", m1.Name)
			result, runErr := m1.Run("pwd")
			if runErr != nil {
				log.Printf("Error 'pwd' on %s: %v\n", m1.Name, runErr)
			} else {
				fmt.Printf("Output from %s ('pwd'): %s\n", m1.Name, strings.TrimSpace(result.Output))
			}
		}
	}

	privateKey := []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
	.
	.
	.
	.
	.
	.
	.
	-----END OPENSSH PRIVATE KEY-----`)

	m2, err := operator.NewMachine("machine2", "dilounix", "hyrule", 22)
	if err != nil {
		log.Printf("Failed to create machine2: %v", err)
	} else {
		fmt.Printf("Machine %s created. Setting SSH key auth...\n", m2.Name)
		authErr := m2.WithSshKeyAuth(privateKey)
		if authErr != nil {
			log.Printf("Failed to authenticate machine %s: %v\n", m2.Name, authErr)
		} else {
			fmt.Printf("Machine %s authenticated. Running 'pwd'...\n", m2.Name)
			result, runErr := m2.Run("pwd")
			if runErr != nil {
				log.Printf("Error 'pwd' on %s: %v\n", m2.Name, runErr)
			} else {
				fmt.Printf("Output from %s ('pwd'): %s\n", m2.Name, strings.TrimSpace(result.Output))
			}
		}
	}

	fmt.Println("\nRunning 'id' on all machines...")
	results, runErrors := operator.Run("id")
	for i, result := range results {
		if runErrors[i] != nil {
			fmt.Printf("Error on machine %s ('id'): %v\n", result.MachineName, runErrors[i])
		} else {
			fmt.Printf("Output from %s ('id'): %s\n", result.MachineName, strings.TrimSpace(result.Output))
		}
	}

	fmt.Println("\nRunning 'uname -a' on first machine...")
	if len(operator.Machines) > 0 {
		firstMachine := operator.Machines[0]
		result, runErr := firstMachine.Run("uname -a")
		if runErr != nil {
			log.Printf("Error 'uname -a' on %s: %v\n", firstMachine.Name, runErr)
		} else {
			fmt.Printf("Output from %s ('uname -a'): %s\n", firstMachine.Name, strings.TrimSpace(result.Output))
		}
	} else {
		fmt.Println("No machines available for 'uname -a'.")
	}

	fmt.Printf("\nTotal machines managed: %d\n", len(operator.Machines))
}

