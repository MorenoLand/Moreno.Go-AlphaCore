package packet

import (
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"math/big"
	"strings"
)

const Generator uint64 = 7

var (
	modulus, _ = new(big.Int).SetString("894B645E89E1535BBDAD5B8B290650530801B18EBFBF5E8FAB3C82872A3E9BB7", 16)
	XorNg      = []byte{221, 123, 176, 58, 56, 172, 115, 17, 3, 152, 124, 90, 80, 111, 202, 150, 108, 123, 194, 167}
)

func GeneratorBytes() []byte { return []byte{byte(Generator)} }

func ModulusBytes() []byte { return bytesLE(modulus, 32) }

func ValidPublicKey(value []byte) bool {
	if len(value) != 32 {
		return false
	}
	return new(big.Int).Mod(intLE(value), modulus).Sign() != 0
}

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	return salt, err
}

func CalculateX(username, password string, salt []byte) []byte {
	interim := sha1.Sum([]byte(strings.ToUpper(username) + ":" + strings.ToUpper(password)))
	value := make([]byte, 0, len(salt)+len(interim))
	value = append(value, salt...)
	value = append(value, interim[:]...)
	result := sha1.Sum(value)
	return result[:]
}

func PasswordVerifier(username, password string, salt []byte) []byte {
	return bytesLE(pow(new(big.Int).SetUint64(Generator), intLE(CalculateX(username, password, salt))), 32)
}

func ServerPublicKey(verifier, privateKey []byte) []byte {
	value := new(big.Int).Mul(new(big.Int).SetInt64(3), intLE(verifier))
	value.Add(value, pow(new(big.Int).SetUint64(Generator), intLE(privateKey)))
	value.Mod(value, modulus)
	return bytesLE(value, 32)
}

func ClientPublicKey(privateKey []byte) []byte {
	return bytesLE(pow(new(big.Int).SetUint64(Generator), intLE(privateKey)), 32)
}

func ScramblingParameter(clientPublicKey, serverPublicKey []byte) []byte {
	return digest(clientPublicKey, serverPublicKey)
}

func ClientSKey(clientPrivateKey, serverPublicKey, x, u []byte) []byte {
	base := new(big.Int).Sub(intLE(serverPublicKey), new(big.Int).Mul(new(big.Int).SetInt64(3), pow(new(big.Int).SetUint64(Generator), intLE(x))))
	base.Mod(base, modulus)
	exponent := new(big.Int).Add(intLE(clientPrivateKey), new(big.Int).Mul(intLE(u), intLE(x)))
	return bytesLE(pow(base, exponent), 32)
}

func ServerSKey(clientPublicKey, verifier, u, privateKey []byte) []byte {
	base := new(big.Int).Mul(intLE(clientPublicKey), pow(intLE(verifier), intLE(u)))
	base.Mod(base, modulus)
	return bytesLE(pow(base, intLE(privateKey)), 32)
}

func Interleaved(secret []byte) ([]byte, error) {
	value := append([]byte(nil), secret...)
	for len(value) > 0 && value[0] == 0 {
		if len(value) < 2 {
			return nil, errors.New("invalid SRP shared secret")
		}
		value = value[2:]
	}
	if len(value) == 0 {
		return nil, errors.New("invalid SRP shared secret")
	}
	even := make([]byte, (len(value)+1)/2)
	odd := make([]byte, len(value)/2)
	for index, item := range value {
		if index%2 == 0 {
			even[index/2] = item
		} else {
			odd[index/2] = item
		}
	}
	first := sha1.Sum(even)
	second := sha1.Sum(odd)
	result := make([]byte, 0, len(first)+len(second))
	for index := range first {
		result = append(result, first[index], second[index])
	}
	return result, nil
}

func ClientProof(x []byte, username string, sessionKey, clientPublicKey, serverPublicKey, salt []byte) []byte {
	usernameHash := sha1.Sum([]byte(strings.ToUpper(username)))
	return digest(x, usernameHash[:], salt, clientPublicKey, serverPublicKey, sessionKey)
}

func ServerProof(clientPublicKey, clientProof, sessionKey []byte) []byte {
	return digest(clientPublicKey, clientProof, sessionKey)
}

func WorldServerProof(username string, clientSeed, serverSeed, sessionKey []byte) []byte {
	zero := []byte{0, 0, 0, 0}
	return digest([]byte(strings.ToUpper(username)), zero, clientSeed, zero, serverSeed, sessionKey)
}

func intLE(value []byte) *big.Int {
	reversed := make([]byte, len(value))
	for index := range value {
		reversed[len(value)-index-1] = value[index]
	}
	return new(big.Int).SetBytes(reversed)
}

func bytesLE(value *big.Int, size int) []byte {
	encoded := value.Bytes()
	if len(encoded) > size {
		encoded = encoded[len(encoded)-size:]
	}
	result := make([]byte, size)
	for index := range encoded {
		result[index] = encoded[len(encoded)-index-1]
	}
	return result
}

func pow(base, exponent *big.Int) *big.Int {
	base = new(big.Int).Mod(new(big.Int).Set(base), modulus)
	return new(big.Int).Exp(base, exponent, modulus)
}

func digest(parts ...[]byte) []byte {
	hash := sha1.New()
	for _, part := range parts {
		hash.Write(part)
	}
	return hash.Sum(nil)
}
