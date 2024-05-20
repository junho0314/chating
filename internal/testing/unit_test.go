package testing

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	switch 1 {
	case 1:
		GetPassword()
	}
}

func GetPassword() {
	key := "hMG1HNqSMGzQLuUOGaV9mx"
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		fmt.Println("Base64 decoding error:", err)
	} else {
		fmt.Println("Base64 decoded key:", decodedKey)
	}
}
