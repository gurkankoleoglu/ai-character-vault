package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func IsValidTurkishPhone(phone string) bool {
	if len(phone) != 10 || phone[0] != '5' {
		return false
	}
	for _, char := range phone {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}
