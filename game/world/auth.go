package world

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"Moreno.AlphaCore/network/packet"
)

func verifyPassword(stored, password string) packet.AuthCode {
	hash := sha256.Sum256([]byte(password))
	if subtle.ConstantTimeCompare([]byte(stored), []byte(hex.EncodeToString(hash[:]))) == 1 {
		return packet.AuthOK
	}
	return packet.AuthIncorrectPassword
}

func hexDecode(value string) ([]byte, error) { return hex.DecodeString(value) }

func equal(left, right []byte) bool {
	return len(left) == len(right) && subtle.ConstantTimeCompare(left, right) == 1
}
