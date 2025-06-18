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
func main() {
 operator, err := opsme.New(
  true, // this indicates to add to known_hosts file
  3,    // this is the timeout for each operation in seconds
 )
 if err != nil {
  log.Fatalf("Failed to initialize operator: %v", err)
 }

 m1, err := operator.NewMachine(
  "machine1",
  "user",
  "192.168.1.2",
  22,
 )
 if err != nil {
  log.Fatalf("Failed to create machine1: %v", err)
 }

 err = m1.WithPasswordAuth("test1234")
 if err != nil {
  log.Fatalf("Failed to authenticate machine %s: %v\n", m1.Name, err)
 }

 result, err := m1.Run("pwd")
 if err != nil {
  log.Fatalf("Error 'pwd' on %s: %v\n", m1.Name, err)
 }

 fmt.Printf("Output from %s ('pwd'): %s\n", m1.Name, strings.TrimSpace(result.Output))

 m2, _ := operator.NewMachine(
  "machine2",
  "dilounix",
  "hyrule",
  22,
 )

 sshKey, _ := opsme.GetKeyFromFile("/home/dilounix/.ssh/id_ed25519")
 _ = m2.WithSSHKeyAuth(sshKey)

 result, _ = m2.Run("pwd")
 fmt.Printf("Output from %s ('pwd'): %s\n", m2.Name, strings.TrimSpace(result.Output))
```

You can also run commands concurrently on multiple machines:

```go
 fmt.Println("\nRunning 'id' on all machines...")
 results, errors := operator.Run("id")
 for i, result := range results {
  if errors[i] != nil {
   log.Printf("Error 'id' on %s: %v\n", operator.Machines[i].Name, errors[i])
   continue
  }
  fmt.Printf(
   "Output from %s ('id'): %s\n",
   result.MachineName,
   strings.TrimSpace(result.Output),
  )
 }
}
```
