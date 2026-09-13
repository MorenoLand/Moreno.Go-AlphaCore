package packet

type AuthCode byte

const (
	AuthOK                AuthCode = 0x0c
	AuthFailed            AuthCode = 0x0d
	AuthIncorrectPassword AuthCode = 0x16
	AuthUnknownAccount    AuthCode = 0x15
)

type SRP6Response byte

const (
	SRP6Challenge SRP6Response = 0
	SRP6Proof     SRP6Response = 1
)
