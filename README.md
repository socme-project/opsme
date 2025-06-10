# OPSme

OPSme is a golang library that provides a simple way to manage multiple machines via SSH. It allows you to execute commands on different machines concurrently, making it easier to manage large infrastructures.

## Installation

To install OPSme, you can use the following command:

```bash
go get github.com/opsme/opsme
```

## Usage

To use OPSme, you need to import the package in your Go code:

```go
import "github.com/opsme/opsme"
```

You can then create a new instance of OPSme and use it to execute commands on multiple machines. Here is a simple example:

```go
package main

import (
  "fmt"
  "github.com/opsme/opsme"
)

func main() {
  operator := opsme.New()

  m1 := operator.NewMachine("machineName", "user", "192.168.1.2")
  m1.WithPasswordAuth("test1234")

  m2, err := operator.NewMachine("machineName2", "user", "192.168.1.3").WithSshKeyAuth("-----BEGIN OPENSSH PRIVATE KEY-----\n...")
  if err != nil {
   panic(err)
  }

 results, err := operator.Run("id")
 if err != nil {
  panic(err)
 }

 for _, result := range results {
  fmt.Println("Machine name:", result.MachineName)
  fmt.Println("Output:", result.Output)
 }

 result, err := m2.Run("uname -a")
 if err != nil {
  panic(err)
 }
 fmt.Println("Machine name:", result.MachineName)
 fmt.Println("Output:", result.Output)

 result, err := operator.Machines[0].Run("uname -a")
 if err != nil {
  panic(err)
 }
 fmt.Println("Machine name:", result.MachineName)
 fmt.Println("Output:", result.Output)
}
```
