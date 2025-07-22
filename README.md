# OPSme

OPSme est une bibliothèque Go qui offre un moyen simple de gérer
plusieurs machines via SSH. Elle vous permet d'exécuter des commandes
sur différentes machines simultanément, facilitant ainsi la gestion de
grandes infrastructures.

## Installation

Pour installer OPSme, vous pouvez utiliser la commande suivante :

```bash
go get github.com/socme-project/opsme
```

## Utilisation

Pour utiliser OPSme, vous devez importer le package dans votre code Go :

```go
import "github.com/socme-project/opsme"
```

Vous pouvez ensuite créer une nouvelle instance d'OPSme et l'utiliser
pour exécuter des commandes sur plusieurs machines. Voici un exemple simple :

```go
func main() {
 operator, err := opsme.New(
  true, // ceci indique d'ajouter au fichier known_hosts
  3,    // ceci est le délai d'attente pour chaque opération en secondes
 )
 if err != nil {
  log.Fatalf("Échec de l'initialisation de l'opérateur : %v", err)
 }

 m1, err := operator.NewMachine(
  "machine1",
  "user",
  "192.168.1.2",
  22,
 )
 if err != nil {
  log.Fatalf("Échec de la création de machine1 : %v", err)
 }

 err = m1.WithPasswordAuth("test1234")
 if err != nil {
  log.Fatalf("Échec de l'authentification de la machine %s : %v\n", m1.Name, err)
 }

 result, err := m1.Run("pwd")
 if err != nil {
  log.Fatalf("Erreur 'pwd' sur %s : %v\n", m1.Name, err)
 }

 fmt.Printf("Sortie de %s ('pwd') : %s\n", m1.Name, strings.TrimSpace(result.Output))

 m2, _ := operator.NewMachine(
  "machine2",
  "dilounix",
  "hyrule",
  22,
 )

 sshKey, _ := opsme.GetKeyFromFile("/home/dilounix/.ssh/id_ed25519")
 _ = m2.WithSSHKeyAuth(sshKey)

 result, _ = m2.Run("pwd")
 fmt.Printf("Sortie de %s ('pwd') : %s\n", m2.Name, strings.TrimSpace(result.Output))
}
```

Vous pouvez également exécuter des commandes concurremment sur plusieurs machines :

```go
fmt.Println("\nExécution de 'id' sur toutes les machines...")
results, errors := operator.Run("id")
for i, result := range results {
 if errors[i] != nil {
  log.Printf("Erreur 'id' sur %s : %v\n", operator.Machines[i].Name, errors[i])
  continue
 }
 fmt.Printf(
  "Sortie de %s ('id') : %s\n",
  result.MachineName,
  strings.TrimSpace(result.Output),
 )
}
```
