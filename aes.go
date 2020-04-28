package main

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func decryptConfigPayload(masterPassword string, iv []uint8, data []uint8) {
	apiPasswordConfig := PasswordConfig{"shapass", masterPassword, "", "", 32}
	apiPassword := generatePassword(apiPasswordConfig)

	var key [32]byte
	copy(key[:], apiPassword)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	if len(data)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(data, data)

	fmt.Printf("decrypted: %v\n", string(data))
}

func main() {
	iv := []uint8{75, 92, 20, 32, 75, 63, 3, 54, 48, 33, 12, 13, 47, 34, 90, 23}

	data := []uint8{176, 172, 28, 117, 207, 163, 140, 7, 113, 8, 116, 23, 20, 128, 84, 138, 182, 103, 145, 249, 149, 147, 164, 81, 62, 120, 145, 232, 45, 171, 66, 99, 236, 7, 92, 16, 175, 176, 157, 217, 165, 40, 83, 212, 4, 31, 196, 165, 67, 75, 178, 35, 130, 204, 167, 43, 219, 212, 52, 113, 193, 212, 103, 135, 204, 144, 231, 63, 218, 187, 159, 135, 167, 22, 250, 217, 164, 17, 251, 225, 92, 53, 174, 120, 165, 122, 124, 159, 199, 7, 152, 203, 194, 182, 52, 31, 231, 73, 216, 58, 5, 75, 176, 53, 146, 69, 140, 115, 229, 70, 127, 25, 188, 117, 103, 145, 235, 239, 125, 212, 95, 99, 171, 234, 92, 157, 247, 187, 173, 31, 92, 209, 223, 197, 191, 100, 204, 108, 17, 64, 176, 184, 112, 193, 58, 107, 24, 159, 33, 125, 162, 255, 124, 64, 223, 45, 234, 131, 200, 202, 19, 175, 148, 118, 200, 39, 40, 234, 171, 6, 142, 101, 147, 254, 148, 12, 132, 235, 115, 199, 0, 148, 95, 58, 93, 82, 65, 187, 215, 105, 90, 193, 122, 219, 237, 122, 177, 140, 225, 145, 106, 79, 50, 108, 214, 105, 131, 119, 235, 53, 138, 1, 167, 110, 182, 87, 162, 161, 168, 226, 235, 2, 21, 242, 2, 126, 91, 63, 182, 236, 103, 156, 84, 177, 191, 146, 78, 177, 34, 199, 90, 241, 206, 7, 241, 40, 20, 107, 71, 245, 242, 36, 50, 172, 126, 194, 36, 183, 107, 9, 140, 94, 141, 115, 230, 5, 136, 131, 30, 124, 13, 163, 177, 226, 212, 161, 214, 23, 126, 252, 250, 15, 255, 50, 213, 41, 6, 2, 248, 237, 57, 138, 78, 72, 137, 130, 228, 36, 69, 137, 7, 255, 211, 252, 254, 69, 136, 46, 238, 112, 253, 242, 5, 212, 74, 226, 159, 43, 155, 144, 132, 86, 33, 118, 247, 70, 109, 8, 141, 151, 52, 91, 235, 173, 167, 50, 251, 49, 123, 102, 205, 148, 14, 143, 110, 5, 252, 31, 189, 78, 169, 47, 77, 143, 194, 221, 20, 112, 195, 64, 244, 93, 134, 61, 106, 8, 89, 236}

	_ = "H8TAc/fO3M0MgD9D0vfk55AK4YQkeRUZ"
	masterPassword := "NjOSK77vzdZmhkplbNgBfYC0w5LzM8XU"
	var key [32]byte
	copy(key[:], masterPassword)

	fmt.Printf("key: %v\n", key)

	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}

	if len(data)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(data, data)

	fmt.Printf("decrypted: %v\n", string(data))

	decryptConfigPayload("Metasploitl33t123!", iv, data)
}
