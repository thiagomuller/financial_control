//go:build ignore

package main

import (
	"fmt"
	"github.com/thiago/my_finances_api/internal/auth"
)

func main() {
	stored := "c7b8d0170847d74801502a40e4b11dab:627b84876eff307dab8e9a84bb2edd38d6373d1c78d0cc471da9096b0cbb1f70"
	for _, p := range []string{"NewPass123", "newpass123"} {
		fmt.Printf("%q => %v\n", p, auth.CheckPassword(p, stored))
	}
	
	// Also hash NewPass123 fresh to see what we'd store
	h, _ := auth.HashPassword("NewPass123")
	fmt.Println("fresh hash:", h)
	fmt.Println("verify fresh:", auth.CheckPassword("NewPass123", h))
}
