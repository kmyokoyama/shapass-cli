package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	"golang.org/x/crypto/ssh/terminal"
)

const apiUrl = "https://shapass.com/api"

type ShapassCLIConfig struct {
	Email string
}

type PasswordConfig struct {
	Service        string
	MasterPassword string
	Prefix         string
	Suffix         string
	Length         int
}

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

	if loginResponse.Status != "OK" {
		return "", errors.New("Error: login failed (wrong email or password)")
	}

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

func fetchRulesFromAPI(email string, masterPasswd string) ([]Rule, error) {
	token, err := login(email, masterPasswd)
	if err != nil {
		return nil, err
	}

	rules, err := list(token)
	if err != nil {
		return nil, err
	}

	return rules, nil
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

func getPassword() (string, error) {
	masterPassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", err
	}

	return string(masterPassword), nil
}

func generatePassword(pc PasswordConfig) string {
	secret := makeSecret([]byte(pc.Service), []byte(pc.MasterPassword))
	encoded := encode(secret)

	return (pc.Prefix + encoded[:pc.Length] + pc.Suffix)
}

func storeConfig(configFile string, config ShapassCLIConfig) {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return
	}

	ioutil.WriteFile(configFile, configJSON, 0600)
}

func configExists(configFile string) bool {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func loadConfig(configFile string) (ShapassCLIConfig, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return ShapassCLIConfig{}, err
	}

	configEncoded, err := ioutil.ReadAll(file)
	if err != nil {
		return ShapassCLIConfig{}, err
	}

	var config ShapassCLIConfig
	json.Unmarshal(configEncoded, &config)

	return config, nil
}

func main() {
	lengthOpt := flag.Int("length", 32, "Length of the password")
	prefixOpt := flag.String("prefix", "", "Prefix to generate the output password (default \"\")")
	suffixOpt := flag.String("suffix", "", "Suffix to generate the output password (default \"\")")
	showOpt := flag.Bool("display", false, "Should show output password? (default false)")
	clipOpt := flag.Bool("copy", true, "Should copy output password to system clipboard?")
	apiOpt := flag.Bool("api", false, "Should fetch configurations from shapass.com? (default false)")

	flag.Parse()

	outputLength := *lengthOpt
	prefix := *prefixOpt
	suffix := *suffixOpt
	shouldDisplay := *showOpt
	shouldCopy := *clipOpt
	shouldFetchFromAPI := *apiOpt

	configFile := os.Getenv("HOME") + "/.shapass"

	var passwordConfig PasswordConfig

	fmt.Print("Enter master password: ")
	masterPassword, err := getPassword()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	passwordConfig.MasterPassword = masterPassword

	if shouldFetchFromAPI {
		if flag.NArg() > 1 {
			fmt.Println("Error: Provide at most 1 service as last argument")
			os.Exit(1)
		}

		var email string
		if configExists(configFile) {
			config, err := loadConfig(configFile)
			if err == nil {
				fmt.Printf("Should use email '%s'? [Y/n] ", config.Email)
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')

				shouldUseEmailInput := string([]byte(input)[0])
				shouldUseEmail := strings.TrimSuffix(strings.ToLower(shouldUseEmailInput), "\n")

				if strings.ToLower(shouldUseEmail) == "y" || shouldUseEmail == "" {
					email = config.Email
				} else {
					fmt.Print("Email: ")
					fmt.Scanln(&email)
				}
			}
		} else {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}

		apiPasswordConfig := PasswordConfig{"shapass", passwordConfig.MasterPassword, "", "", 32}
		apiPassword := generatePassword(apiPasswordConfig)

		rules, err := fetchRulesFromAPI(email, apiPassword)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		storeConfig(configFile, ShapassCLIConfig{Email: email})

		var rule Rule
		if flag.NArg() == 0 {
			fmt.Printf("Choose a service [1-%d]:\n", len(rules))

			for idx, r := range rules {
				fmt.Printf("[%d] %s\n", idx+1, r.Name)
			}

			var serviceIdx int
			fmt.Scanln(&serviceIdx)

			if serviceIdx < 1 || serviceIdx > len(rules) {
				fmt.Println("Error: invalid service number")
				os.Exit(1)
			}

			rule = rules[serviceIdx-1]
		} else {
			serviceName := flag.Args()[0]
			match := false
			for _, r := range rules {
				if r.Name == serviceName {
					match = true
					rule = r
				}
			}

			if !match {
				fmt.Printf("Error: Service '%s' not found in shapass API\n", serviceName)
				os.Exit(1)
			}
		}

		passwordConfig.Service = rule.Name
		passwordConfig.Prefix = rule.Prefix
		passwordConfig.Suffix = rule.Suffix
		passwordConfig.Length = rule.Length
	} else {
		if flag.NArg() != 1 {
			fmt.Println("Error: Provide at least 1 service as last argument")
			os.Exit(1)
		}

		if outputLength < 1 || outputLength > 44 {
			fmt.Printf("Error: Invalid output length [%d]. It must be a number from 1 to 44.\n", outputLength)
			os.Exit(1)
		}

		passwordConfig.Service = flag.Args()[0]
		passwordConfig.Prefix = prefix
		passwordConfig.Suffix = suffix
		passwordConfig.Length = outputLength
	}

	outputPasswd := generatePassword(passwordConfig)

	if shouldDisplay {
		fmt.Println(outputPasswd)
	}

	if shouldCopy {
		clipboard.WriteAll(string(outputPasswd))
	}
}
