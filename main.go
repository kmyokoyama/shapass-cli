package main

import (
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh/terminal"
	"os"
)

func makeSecret(parts ...[]byte) (secret []byte) {
	for _, part := range parts {
		secret = append(secret, part...)
	}

	return secret
}

func encode(secret []byte) string {
	ss := sha256.Sum256(secret)

	return base64.StdEncoding.EncodeToString(ss[:])
}

func main() {
	lengthOpt := flag.Int("len", 10, "Length of the password")
	suffixOpt := flag.String("suffix", "", "Suffix to generate the output password (default \"\")")
	showOpt := flag.Bool("show", false, "Should show output password? (default false)")
	clipOpt := flag.Bool("clip", true, "Should copy output password to system clipboard?")

	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Error: Provide exactly 1 service as last argument")
		os.Exit(1)
	}

	if *lengthOpt < 1 || *lengthOpt > 44 {
		fmt.Printf("Error: Invalid output length [%d]. It must be a number from 1 to 44.\n", *lengthOpt)
		os.Exit(1)
	}

	fmt.Println("Enter master password: ")
	masterPasswd, err := terminal.ReadPassword(int(os.Stdin.Fd()))

	if err != nil {
		fmt.Println(err)
	}

	service := flag.Args()[0]

	secret := makeSecret([]byte(service), masterPasswd)

	encoded := encode(secret)

	outputPasswd := encoded[:*lengthOpt] + (*suffixOpt)

	if *showOpt {
		fmt.Println(outputPasswd)
	}

	if *clipOpt {
		clipboard.WriteAll(string(outputPasswd))
	}
}