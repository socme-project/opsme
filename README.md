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

 "github.com/socme-project/opsme"
)

func main() {
 operator := opsme.New()

 m1, err := operator.NewMachine(
  "machineName",
  "user",
  "", // HostKey is automatically taken from the known_hosts file, however, you can specify it manually if needed
  "192.168.1.2",
  22,
  opsme.Auth{AuthType: opsme.AuthTypePassword, Password: "test1234"},
 )
 if err != nil {
  fmt.Println("Error creating machine:", err)
 } else {
  fmt.Println("Machine created successfully:", m1.Name)
 }

 result, err := m1.Run("pwd")
 if err != nil {
  fmt.Println("Error running command on machine:", err)
 } else {
  fmt.Println("Output:", result.Output)
 }

 m2, err := operator.NewMachine(
  "siuu",
  "dilounix",
  "", // HostKey is automatically taken from the known_hosts file, however, you can specify it manually if needed
  "192.168.1.3",
  22,
  opsme.Auth{
   AuthType: opsme.AuthTypeSshKey,
   SshKey: `-----BEGIN OPENSSH PRIVATE KEY-----
   .
   .
   .
   .
   .
   .
   -----END OPENSSH PRIVATE KEY-----
   `,
  },
 )
 if err != nil {
  fmt.Println("Error creating machine:", err)
 } else {
  fmt.Println("Machine created successfully:", m2.Name)
 }

 result, err = m2.Run("pwd")
 if err != nil {
  fmt.Println("Error running command on machine:", err)
 } else {
  fmt.Println("Output:", result.Output)
 }

 results, errors := operator.Run("id")
 if len(errors) > 0 {
  for i, individualErr := range errors {
   if individualErr != nil {
    fmt.Printf("Error on machine %s: %v\n", operator.Machines[i].Name, individualErr)
   } else {
    fmt.Println("Machine name:", operator.Machines[i].Name)
    fmt.Println("Output:", results[i].Output)
   }
  }
 } else {
  for _, result := range results {
   fmt.Println("Machine name:", result.MachineName)
   fmt.Println("Output:", result.Output)
  }
 }

 result, err = operator.Machines[0].Run("uname -a")
 if err != nil {
  fmt.Println("Error running command on machine:", err)
 } else {
  fmt.Println("Output:", result.Output)
 }
}
```

You can also change the authentication method for a machine after it has been created:

```go
m1.WithPasswordAuth("newpassword")
```
