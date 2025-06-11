package main

import (
	"fmt"

	"github.com/socme-project/opsme"
)

func main() {
	ops, err := opsme.NewOperator()
	if err != nil {
		panic(err)
	}

	m1, err := ops.NewMachine("machine1", "root", "192.168.1.2")
	if err != nil {
		panic(err)
	}

	m1.WithPasswordAuth("test1234")

	results, err := ops.Run("id")
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Println("Machine:", result.Machine.Name)
		fmt.Println("Output:", result.Output)
	}

	m2, err := ops.NewMachine("machine2", "user", "192.168.1.3").
		WithSshKeyAuth("-----BEGIN OPENSSH PRIVATE KEY-----\n...")
	if err != nil {
		panic(err)
	}

	result, err := m2.Run("uname -a")
	if err != nil {
		panic(err)
	}
	fmt.Println("Machine:", result.Machine.Name)
	fmt.Println("Output:", result.Output)

	results, err = ops.Run("uname -a")
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		fmt.Println("Machine:", result.Machine.Name)
		fmt.Println("Output:", result.Output)
	}
}
