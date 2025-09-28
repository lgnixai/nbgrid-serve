package main

import (
	"fmt"
	"teable-go-backend/pkg/utils"
)

func main() {
	fmt.Println("Testing ID generation:")
	fmt.Println("User ID:", utils.GenerateUserID())
	fmt.Println("Space ID:", utils.GenerateSpaceID())
	fmt.Println("Account ID:", utils.GenerateAccountID())
}
