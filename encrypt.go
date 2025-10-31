package main

import (
	"fmt"
)

// generateRoundKey creates a slightly modified version of the key for each round.
func generateRoundKey(baseKey []byte, round int) []byte {
	roundKey := make([]byte, len(baseKey))
	for i := range baseKey {
		roundKey[i] = baseKey[i] + byte(round)

	}
	return roundKey
}

// xorEncrypt performs XOR encryption for multiple rounds.
func xorEncrypt(plaintext, key []byte, rounds int) []byte {
	data := make([]byte, len(plaintext))
	copy(data, plaintext)

	for r := 0; r < rounds; r++ {
		roundKey := generateRoundKey(key, r)
		for i := 0; i < len(data); i++ {
			data[i] ^= roundKey[i%len(roundKey)]
		}
	}
	return data
}

// xorDecrypt reverses the encryption process by applying the rounds in reverse.
func xorDecrypt(ciphertext, key []byte, rounds int) []byte {
	data := make([]byte, len(ciphertext))
	copy(data, ciphertext)

	// Reverse the rounds
	for r := rounds - 1; r >= 0; r-- {
		roundKey := generateRoundKey(key, r)
		for i := 0; i < len(data); i++ {
			data[i] ^= roundKey[i%len(roundKey)]
		}
	}
	return data
}

func main() {
	plaintext := []byte("HELLO WORLD")
	key := []byte("KEY")
	rounds := 5

	// Encrypt
	ciphertext := xorEncrypt(plaintext, key, rounds)
	fmt.Printf("Ciphertext (hex): %x\n", ciphertext)

	// Decrypt
	decrypted := xorDecrypt(ciphertext, key, rounds)
	fmt.Println("Decrypted:", string(decrypted))
}
