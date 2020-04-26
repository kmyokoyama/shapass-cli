package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh/terminal"
)

const apiUrl = "https://shapass.com/api"

type Rule struct {
	Name      string
	Length    int
	Prefix    string
	Suffix    string
	Algorithm string
	Metadata  map[string]string
}

type ListResponse struct {
	Status string
	Rules  []Rule
	Code   int
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	Status string
	Token  string
}

type TokenRequest struct {
	Token string
}

func login(email string, password string) (string, error) {
	loginRequest := LoginRequest{Email: email, Password: password}
	reqBody := new(bytes.Buffer)
	json.NewEncoder(reqBody).Encode(loginRequest)

	resp, err := http.Post(apiUrl+"/login", "application/json", reqBody)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResponse LoginResponse
	json.NewDecoder(resp.Body).Decode(&loginResponse)

	return loginResponse.Token, nil
}

func list(token string) ([]Rule, error) {
	listRequest := TokenRequest{Token: token}
	reqBody := new(bytes.Buffer)
	json.NewEncoder(reqBody).Encode(listRequest)

	resp, err := http.Post(apiUrl+"/list", "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var listResponse ListResponse
	json.NewDecoder(resp.Body).Decode(&listResponse)

	return listResponse.Rules, nil
}

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

func fetchServiceFromAPI(email string, masterPasswd string) (*Rule, error) {
	token, err := login("your@email.com", "your-shapass-generated-password")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rules, err := list(token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, rule := range rules {
		if rule.Name == "service" {
			return &rule, nil
		}
	}

	return nil, fmt.Errorf("shapass: Service not found")
}

func main() {
	lengthOpt := flag.Int("len", 10, "Length of the password")
	suffixOpt := flag.String("suffix", "", "Suffix to generate the output password (default \"\")")
	showOpt := flag.Bool("show", false, "Should show output password? (default false)")
	clipOpt := flag.Bool("clip", true, "Should copy output password to system clipboard? (default true)")
	setupOpt := flag.Bool("setup", false, "Should fetch and setup? (default false)")

	flag.Parse()

	if *setupOpt {
		rule, err := fetchServiceFromAPI("yokoyama.km@gmail.com", "H8TAc/fO3M0MgD9D0vfk55AK4YQkeRUZ")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(rule)
	}

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
