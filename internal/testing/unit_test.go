package testing

import (
	"chating_service/internal/utils"
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
	pwd, _ := utils.EncryptPassword("1234")
	fmt.Println(pwd)
}
