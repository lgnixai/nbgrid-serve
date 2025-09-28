package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 数据库中的密码哈希
	hashedPassword := "$2a$10$N/zWItBNTHwEX7e51LnAL.942SDm.G5JPKxMaw5KB5qTePxnUgDey"
	
	// 测试不同的密码
	passwords := []string{
		"password123",
		"test123",
		"admin123",
		"123456",
		"password",
		"test",
	}
	
	for _, password := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err == nil {
			fmt.Printf("密码匹配: %s\n", password)
			return
		}
	}
	
	fmt.Println("没有找到匹配的密码")
}
