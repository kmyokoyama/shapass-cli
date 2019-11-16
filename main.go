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

func encode(secret string) string {
	ss := sha256.Sum256([]byte(secret))

	return base64.StdEncoding.EncodeToString(ss[:])
}

func main() {
	lengthOpt := flag.Int("len", 10, "Length of the password")

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

	encoded := encode(string(append([]byte(service), masterPasswd...)))

	fmt.Println(len(encoded))

	passTruncated := encoded[:*lengthOpt]

	fmt.Println(passTruncated)

	clipboard.WriteAll(string(passTruncated))
}