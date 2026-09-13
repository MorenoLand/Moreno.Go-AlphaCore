package packet

type Opcode uint32

const (
	CMSGAuthSRP6Begin          Opcode = 0x0033
	CMSGAuthSRP6Proof          Opcode = 0x0034
	CMSGAuthSRP6Recode         Opcode = 0x0035
	CMSGCharCreate             Opcode = 0x0036
	CMSGCharEnum               Opcode = 0x0037
	CMSGCharDelete             Opcode = 0x0038
	SMSGAuthSRP6Response       Opcode = 0x0039
	SMSGCharCreate             Opcode = 0x003a
	SMSGCharEnum               Opcode = 0x003b
	SMSGCharDelete             Opcode = 0x003c
	CMSGPlayerLogin            Opcode = 0x003d
	SMSGNewWorld               Opcode = 0x003e
	SMSGAuthChallenge          Opcode = 0x01dd
	CMSGPlayerLogout           Opcode = 0x004a
	CMSGLogoutRequest          Opcode = 0x004b
	SMSGUpdateObject           Opcode = 0x00a9
	CMSGPing                   Opcode = 0x01cd
	SMSGPong                   Opcode = 0x01ce
	CMSGAuthSession            Opcode = 0x01de
	SMSGAuthResponse           Opcode = 0x01df
	SMSGCompressedUpdateObject Opcode = 0x01e7
)

var opcodeNames = map[Opcode]string{
	CMSGAuthSRP6Begin:          "CMSG_AUTH_SRP6_BEGIN",
	CMSGAuthSRP6Proof:          "CMSG_AUTH_SRP6_PROOF",
	CMSGAuthSRP6Recode:         "CMSG_AUTH_SRP6_RECODE",
	CMSGCharCreate:             "CMSG_CHAR_CREATE",
	CMSGCharEnum:               "CMSG_CHAR_ENUM",
	CMSGCharDelete:             "CMSG_CHAR_DELETE",
	SMSGAuthSRP6Response:       "SMSG_AUTH_SRP6_RESPONSE",
	SMSGCharCreate:             "SMSG_CHAR_CREATE",
	SMSGCharEnum:               "SMSG_CHAR_ENUM",
	SMSGCharDelete:             "SMSG_CHAR_DELETE",
	CMSGPlayerLogin:            "CMSG_PLAYER_LOGIN",
	SMSGNewWorld:               "SMSG_NEW_WORLD",
	SMSGAuthChallenge:          "SMSG_AUTH_CHALLENGE",
	CMSGPlayerLogout:           "CMSG_PLAYER_LOGOUT",
	CMSGLogoutRequest:          "CMSG_LOGOUT_REQUEST",
	SMSGUpdateObject:           "SMSG_UPDATE_OBJECT",
	CMSGPing:                   "CMSG_PING",
	SMSGPong:                   "SMSG_PONG",
	CMSGAuthSession:            "CMSG_AUTH_SESSION",
	SMSGAuthResponse:           "SMSG_AUTH_RESPONSE",
	SMSGCompressedUpdateObject: "SMSG_COMPRESSED_UPDATE_OBJECT",
}

func (o Opcode) String() string {
	if name, ok := opcodeNames[o]; ok {
		return name
	}
	return "UNKNOWN"
}
