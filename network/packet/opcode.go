package packet

type Opcode uint32

const (
	CMSGAuthSRP6Begin             Opcode = 0x0033
	CMSGAuthSRP6Proof             Opcode = 0x0034
	CMSGAuthSRP6Recode            Opcode = 0x0035
	CMSGCharCreate                Opcode = 0x0036
	CMSGCharEnum                  Opcode = 0x0037
	CMSGCharDelete                Opcode = 0x0038
	SMSGAuthSRP6Response          Opcode = 0x0039
	SMSGCharCreate                Opcode = 0x003a
	SMSGCharEnum                  Opcode = 0x003b
	SMSGCharDelete                Opcode = 0x003c
	CMSGPlayerLogin               Opcode = 0x003d
	SMSGNewWorld                  Opcode = 0x003e
	MSGMoveWorldportAck           Opcode = 0x00d9
	SMSGCharacterLoginFailed      Opcode = 0x0041
	SMSGLoginSetTimeSpeed         Opcode = 0x0042
	SMSGAuthChallenge             Opcode = 0x01dd
	CMSGPlayerLogout              Opcode = 0x004a
	CMSGLogoutRequest             Opcode = 0x004b
	SMSGLogoutResponse            Opcode = 0x004c
	SMSGLogoutComplete            Opcode = 0x004d
	CMSGLogoutCancel              Opcode = 0x004e
	SMSGLogoutCancelAck           Opcode = 0x004f
	CMSGNameQuery                 Opcode = 0x0050
	SMSGNameQueryResponse         Opcode = 0x0051
	CMSGItemQuerySingle           Opcode = 0x0056
	CMSGItemQueryMultiple         Opcode = 0x0057
	SMSGItemQuerySingleResponse   Opcode = 0x0058
	SMSGItemQueryMultipleResponse Opcode = 0x0059
	CMSGPageTextQuery             Opcode = 0x005a
	SMSGPageTextQueryResponse     Opcode = 0x005b
	CMSGQuestQuery                Opcode = 0x005c
	SMSGQuestQueryResponse        Opcode = 0x005d
	CMSGGameObjectQuery           Opcode = 0x005e
	SMSGGameObjectQueryResponse   Opcode = 0x005f
	CMSGCreatureQuery             Opcode = 0x0060
	SMSGCreatureQueryResponse     Opcode = 0x0061
	CMSGMessageChat               Opcode = 0x0095
	SMSGMessageChat               Opcode = 0x0096
	CMSGWho                       Opcode = 0x0062
	SMSGWho                       Opcode = 0x0063
	SMSGInitializeFactions        Opcode = 0x0115
	SMSGActionButtons             Opcode = 0x011c
	SMSGInitialSpells             Opcode = 0x011d
	SMSGUpdateObject              Opcode = 0x00a9
	CMSGQueryTime                 Opcode = 0x01bf
	SMSGQueryTimeResponse         Opcode = 0x01c0
	CMSGZoneUpdate                Opcode = 0x01e5
	CMSGPing                      Opcode = 0x01cd
	SMSGPong                      Opcode = 0x01ce
	CMSGAuthSession               Opcode = 0x01de
	SMSGAuthResponse              Opcode = 0x01df
	SMSGCompressedUpdateObject    Opcode = 0x01e7
)

var opcodeNames = map[Opcode]string{
	CMSGAuthSRP6Begin:             "CMSG_AUTH_SRP6_BEGIN",
	CMSGAuthSRP6Proof:             "CMSG_AUTH_SRP6_PROOF",
	CMSGAuthSRP6Recode:            "CMSG_AUTH_SRP6_RECODE",
	CMSGCharCreate:                "CMSG_CHAR_CREATE",
	CMSGCharEnum:                  "CMSG_CHAR_ENUM",
	CMSGCharDelete:                "CMSG_CHAR_DELETE",
	SMSGAuthSRP6Response:          "SMSG_AUTH_SRP6_RESPONSE",
	SMSGCharCreate:                "SMSG_CHAR_CREATE",
	SMSGCharEnum:                  "SMSG_CHAR_ENUM",
	SMSGCharDelete:                "SMSG_CHAR_DELETE",
	CMSGPlayerLogin:               "CMSG_PLAYER_LOGIN",
	SMSGNewWorld:                  "SMSG_NEW_WORLD",
	SMSGCharacterLoginFailed:      "SMSG_CHARACTER_LOGIN_FAILED",
	SMSGLoginSetTimeSpeed:         "SMSG_LOGIN_SETTIMESPEED",
	SMSGAuthChallenge:             "SMSG_AUTH_CHALLENGE",
	CMSGPlayerLogout:              "CMSG_PLAYER_LOGOUT",
	CMSGLogoutRequest:             "CMSG_LOGOUT_REQUEST",
	CMSGNameQuery:                 "CMSG_NAME_QUERY",
	SMSGNameQueryResponse:         "SMSG_NAME_QUERY_RESPONSE",
	CMSGItemQuerySingle:           "CMSG_ITEM_QUERY_SINGLE",
	CMSGItemQueryMultiple:         "CMSG_ITEM_QUERY_MULTIPLE",
	SMSGItemQuerySingleResponse:   "SMSG_ITEM_QUERY_SINGLE_RESPONSE",
	SMSGItemQueryMultipleResponse: "SMSG_ITEM_QUERY_MULTIPLE_RESPONSE",
	CMSGPageTextQuery:             "CMSG_PAGE_TEXT_QUERY",
	SMSGPageTextQueryResponse:     "SMSG_PAGE_TEXT_QUERY_RESPONSE",
	CMSGQuestQuery:                "CMSG_QUEST_QUERY",
	SMSGQuestQueryResponse:        "SMSG_QUEST_QUERY_RESPONSE",
	CMSGGameObjectQuery:           "CMSG_GAMEOBJECT_QUERY",
	SMSGGameObjectQueryResponse:   "SMSG_GAMEOBJECT_QUERY_RESPONSE",
	CMSGCreatureQuery:             "CMSG_CREATURE_QUERY",
	SMSGCreatureQueryResponse:     "SMSG_CREATURE_QUERY_RESPONSE",
	CMSGMessageChat:               "CMSG_MESSAGECHAT",
	SMSGMessageChat:               "SMSG_MESSAGECHAT",
	CMSGWho:                       "CMSG_WHO",
	SMSGWho:                       "SMSG_WHO",
	SMSGInitializeFactions:        "SMSG_INITIALIZE_FACTIONS",
	SMSGActionButtons:             "SMSG_ACTION_BUTTONS",
	SMSGInitialSpells:             "SMSG_INITIAL_SPELLS",
	SMSGUpdateObject:              "SMSG_UPDATE_OBJECT",
	CMSGPing:                      "CMSG_PING",
	SMSGPong:                      "SMSG_PONG",
	CMSGAuthSession:               "CMSG_AUTH_SESSION",
	SMSGAuthResponse:              "SMSG_AUTH_RESPONSE",
	SMSGCompressedUpdateObject:    "SMSG_COMPRESSED_UPDATE_OBJECT",
}

func IsMovement(opcode Opcode) bool { return opcode >= 0x00b5 && opcode <= 0x00e9 }

func (o Opcode) String() string {
	if name, ok := opcodeNames[o]; ok {
		return name
	}
	return "UNKNOWN"
}
