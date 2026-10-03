// Command devtoken prints a 24-hour access token for a user or admin id, signed
// with the keys in .env. LOCAL DEVELOPMENT ONLY — it exists so the rewards flow
// can be exercised as a customer before the customer app's purchase screen is
// built. Lives only on the local/rewards-test branch; do not ship it.
//
//	go run ./cmd/devtoken user  usr_xxx
//	go run ./cmd/devtoken admin adm_xxx
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
)

func main() {
	if len(os.Args) != 3 || (os.Args[1] != "user" && os.Args[1] != "admin") {
		log.Fatal("usage: devtoken user|admin <id>")
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	usedFor := token.User
	if os.Args[1] == "admin" {
		usedFor = token.Admin
	}
	res, err := token.NewTokenService(cfg).GenerateToken(token.GenerateTokenRequest{
		UserID:   os.Args[2],
		UsedFor:  usedFor,
		ExpireAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		log.Fatalf("generate token: %v", err)
	}
	fmt.Println(res.TokenString)
}
