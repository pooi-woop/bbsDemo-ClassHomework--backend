package main

import (
	"fmt"
	"bbsDemo/utils"
)

func main() {
	for i := 0; i < 10; i++ {
		code := utils.GenerateVerificationCode()
		fmt.Printf("验证码 %d: %s\n", i+1, code)
	}
}
