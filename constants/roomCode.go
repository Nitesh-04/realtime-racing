package constants

import (
	"math/rand"
	"strconv"
	"time"
)

// GenerateRoomCode generates a unique 6-digit room code

func GenerateRoomCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := r.Intn(900000) + 100000
	return strconv.Itoa(code)
}