package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
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

type Service struct {
	Name      string
	Length    int `json:"outputLength"`
	Prefix    string
	Suffix    string
	Algorithm string
}

type EncryptedData struct {
	IV   []byte `json:"iv"`
	Data []byte `json:"data"`
}

type LoadResponse struct {
	Status        string
	EncryptedData string
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

func load(token string) (EncryptedData, error) {
	listRequest := TokenRequest{Token: token}
	reqBody := new(bytes.Buffer)
	json.NewEncoder(reqBody).Encode(listRequest)

	resp, err := http.Post(apiUrl+"/load", "application/json", reqBody)
	if err != nil {
		return EncryptedData{}, err
	}
	defer resp.Body.Close()

	var loadResponse LoadResponse
	json.NewDecoder(resp.Body).Decode(&loadResponse)

	var encryptedData EncryptedData
	json.Unmarshal([]byte(loadResponse.EncryptedData), &encryptedData)

	return encryptedData, nil
}

func decryptPayload(masterPassword string, iv []byte, data []byte) ([]byte, error) {
	apiPassword := generatePassword(PasswordConfig{"shapass", masterPassword, "", "", 32})
	tokenPassword := generatePassword(PasswordConfig{masterPassword, apiPassword, "", "", 32})

	var key [32]byte
	copy(key[:], tokenPassword)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	if len(data)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	if len(iv) != aes.BlockSize {
		errorMsg := fmt.Sprintf("IV (size: %d) must have size equal to the block size (%d)",
			len(iv), aes.BlockSize)

		return nil, errors.New(errorMsg)
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	mode.CryptBlocks(data, data)

	return data, nil
}

func isEmpty(ed EncryptedData) bool {
	return (len(ed.IV) == 0 || len(ed.Data) == 0)
}

func fetchServicesFromAPI(email string, masterPassword string) ([]Service, error) {
	apiPasswordConfig := PasswordConfig{"shapass", masterPassword, "", "", 32}
	apiPassword := generatePassword(apiPasswordConfig)

	token, err := login(email, apiPassword)
	if err != nil {
		return nil, err
	}

	encryptedData, err := load(token)
	if err != nil {
		return nil, err
	}

	if isEmpty(encryptedData) {
		return nil, errors.New("No service data to load from API")
	}

	payloadJSON, err := decryptPayload(masterPassword, encryptedData.IV, encryptedData.Data)
	if err != nil {
		return nil, err
	}

	var payload map[string]Service
	payload = make(map[string]Service)

	err = json.Unmarshal(bytes.Trim(payloadJSON, "\x00"), &payload)
	if err != nil {
		return nil, err
	}

	services := make([]Service, 0, len(payload))
	for _, service := range payload {
		services = append(services, service)
	}

	return services, nil
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

		services, err := fetchServicesFromAPI(email, passwordConfig.MasterPassword)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		storeConfig(configFile, ShapassCLIConfig{Email: email})

		var service Service
		if flag.NArg() == 0 {
			fmt.Printf("Choose a service [1-%d]:\n", len(services))

			for idx, r := range services {
				fmt.Printf("[%d] %s\n", idx+1, r.Name)
			}

			var serviceIdx int
			fmt.Scanln(&serviceIdx)

			if serviceIdx < 1 || serviceIdx > len(services) {
				fmt.Println("Error: invalid service number")
				os.Exit(1)
			}

			service = services[serviceIdx-1]
		} else {
			serviceName := flag.Args()[0]
			match := false
			for _, r := range services {
				if r.Name == serviceName {
					match = true
					service = r
				}
			}

			if !match {
				fmt.Printf("Error: Service '%s' not found in shapass API\n", serviceName)
				os.Exit(1)
			}
		}

		passwordConfig.Service = service.Name
		passwordConfig.Prefix = service.Prefix
		passwordConfig.Suffix = service.Suffix
		passwordConfig.Length = service.Length
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
		fmt.Println("Password copied to clipboard successfully!")
	}
}
