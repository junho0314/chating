package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// Key should be a 16, 24 or 32 bytes long.
var aesEncryptKey = []byte(")H@McQfTjWnZr4u7x!A%C*F-JaNdRgUk")
var aseIv = []byte("3504956013256078")

func EncryptPassword(password string) (string, error) {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(encryptedPassword), err
}

func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func unpad(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func EncryptAES(plaintext string) string {

	if len(plaintext) <= 0 {
		return ""
	}
	block, err := aes.NewCipher(aesEncryptKey)
	if err != nil {
		panic(err)
	}
	// Create a new cipher block chaining (CBC) cipher using the AES cipher block and IV
	mode := cipher.NewCBCEncrypter(block, aseIv)
	// Pad the plaintext to the next multiple of the block size
	paddedPlaintext := pad([]byte(plaintext), aes.BlockSize)
	// Create a buffer for the ciphertext
	ciphertext := make([]byte, len(paddedPlaintext))
	// Encrypt the padded plaintext into the ciphertext buffer
	mode.CryptBlocks(ciphertext, paddedPlaintext)
	// Encode the ciphertext and IV as base64 strings for transport
	ciphertextString := base64.StdEncoding.EncodeToString(ciphertext)
	return ciphertextString
}

func DecryptAES(plaintext string) string {

	if len(plaintext) <= 0 {
		return ""
	}
	block, err := aes.NewCipher(aesEncryptKey)
	if err != nil {
		panic(err)
	}
	decodedCiphertext, err := base64.StdEncoding.DecodeString(plaintext)
	if err != nil {
		return plaintext
	}
	// Create a new CBC cipher using the AES cipher block and decoded IV
	decryptMode := cipher.NewCBCDecrypter(block, aseIv)
	// Create a buffer for the decrypted plaintext
	decryptedPlaintext := make([]byte, len(decodedCiphertext))
	// Decrypt the ciphertext into the decrypted plaintext buffer
	decryptMode.CryptBlocks(decryptedPlaintext, decodedCiphertext)
	// Remove the padding from the decrypted plaintext
	plaintextWithoutPadding := unpad(decryptedPlaintext)
	return string(plaintextWithoutPadding)
}

func IsInvalidPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err != nil
}
