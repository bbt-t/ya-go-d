package pkg

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func GenerateRandom() string {
	/*
		Random string generation.
	*/
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(randomBytes)[:8]
}

func MakeCookie(userID int, secretKey string) []byte {
	sessionID := GenerateRandom()
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(fmt.Sprintf("%v%v", sessionID, userID)))
	cookie := append([]byte(sessionID), h.Sum(nil)...)
	cookie = append(cookie, []byte(fmt.Sprint(userID))...)
	return cookie
}
