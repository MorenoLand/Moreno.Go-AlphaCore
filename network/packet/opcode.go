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
	CMSGFriendList                Opcode = 0x0066
	SMSGFriendList                Opcode = 0x0067
	SMSGFriendStatus              Opcode = 0x0068
	CMSGAddFriend                 Opcode = 0x0069
	CMSGDelFriend                 Opcode = 0x006a
	SMSGIgnoreList                Opcode = 0x006b
	CMSGAddIgnore                 Opcode = 0x006c
	CMSGDelIgnore                 Opcode = 0x006d
	CMSGGroupInvite               Opcode = 0x006e
	SMSGGroupInvite               Opcode = 0x006f
	CMSGGroupAccept               Opcode = 0x0072
	CMSGGroupDecline              Opcode = 0x0073
	SMSGGroupDecline              Opcode = 0x0074
	CMSGGroupUninvite             Opcode = 0x0075
	CMSGGroupUninviteGUID         Opcode = 0x0076
	SMSGGroupUninvite             Opcode = 0x0077
	CMSGGroupSetLeader            Opcode = 0x0078
	SMSGGroupSetLeader            Opcode = 0x0079
	CMSGGroupDisband              Opcode = 0x007b
	SMSGGroupDestroyed            Opcode = 0x007c
	SMSGGroupList                 Opcode = 0x007d
	SMSGPartyMemberStats          Opcode = 0x007e
	SMSGPartyCommandResult        Opcode = 0x007f
	SMSGInitializeFactions        Opcode = 0x0115
	CMSGSetActionButton           Opcode = 0x011b
	SMSGActionButtons             Opcode = 0x011c
	SMSGInitialSpells             Opcode = 0x011d
	CMSGNewSpellSlot              Opcode = 0x0120
	CMSGSetSelection              Opcode = 0x0130
	CMSGSetTarget                 Opcode = 0x0131
	CMSGStandStateChange          Opcode = 0x00f4
	CMSGTextEmote                 Opcode = 0x00f7
	SMSGEmote                     Opcode = 0x00f6
	SMSGTextEmote                 Opcode = 0x00f8
	SMSGDestroyObject             Opcode = 0x00aa
	CMSGOpenItem                  Opcode = 0x00ac
	CMSGReadItem                  Opcode = 0x00ad
	SMSGReadItemOK                Opcode = 0x00ae
	SMSGReadItemFailed            Opcode = 0x00af
	CMSGAutoequipItem             Opcode = 0x00fd
	CMSGAutostoreBagItem          Opcode = 0x00fe
	CMSGSwapItem                  Opcode = 0x00ff
	CMSGSwapInvItem               Opcode = 0x0100
	CMSGSplitItem                 Opcode = 0x0101
	CMSGDestroyItem               Opcode = 0x0104
	SMSGInventoryChangeFailure    Opcode = 0x0105
	SMSGUpdateObject              Opcode = 0x00a9
	CMSGQueryTime                 Opcode = 0x01bf
	SMSGQueryTimeResponse         Opcode = 0x01c0
	CMSGPlayedTime                Opcode = 0x01bd
	SMSGPlayedTime                Opcode = 0x01be
	CMSGZoneUpdate                Opcode = 0x01e5
	CMSGPing                      Opcode = 0x01cd
	SMSGPong                      Opcode = 0x01ce
	MSGRandomRoll                 Opcode = 0x01ec
	MSGLookingForGroup            Opcode = 0x01f0
	CMSGSetLookingForGroup        Opcode = 0x01f1
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
	CMSGFriendList:                "CMSG_FRIEND_LIST",
	SMSGFriendList:                "SMSG_FRIEND_LIST",
	SMSGFriendStatus:              "SMSG_FRIEND_STATUS",
	CMSGAddFriend:                 "CMSG_ADD_FRIEND",
	CMSGDelFriend:                 "CMSG_DEL_FRIEND",
	SMSGIgnoreList:                "SMSG_IGNORE_LIST",
	CMSGAddIgnore:                 "CMSG_ADD_IGNORE",
	CMSGDelIgnore:                 "CMSG_DEL_IGNORE",
	CMSGGroupInvite:               "CMSG_GROUP_INVITE",
	SMSGGroupInvite:               "SMSG_GROUP_INVITE",
	CMSGGroupAccept:               "CMSG_GROUP_ACCEPT",
	CMSGGroupDecline:              "CMSG_GROUP_DECLINE",
	SMSGGroupDecline:              "SMSG_GROUP_DECLINE",
	CMSGGroupUninvite:             "CMSG_GROUP_UNINVITE",
	CMSGGroupUninviteGUID:         "CMSG_GROUP_UNINVITE_GUID",
	SMSGGroupUninvite:             "SMSG_GROUP_UNINVITE",
	CMSGGroupSetLeader:            "CMSG_GROUP_SET_LEADER",
	SMSGGroupSetLeader:            "SMSG_GROUP_SET_LEADER",
	CMSGGroupDisband:              "CMSG_GROUP_DISBAND",
	SMSGGroupDestroyed:            "SMSG_GROUP_DESTROYED",
	SMSGGroupList:                 "SMSG_GROUP_LIST",
	SMSGPartyMemberStats:          "SMSG_PARTY_MEMBER_STATS",
	SMSGPartyCommandResult:        "SMSG_PARTY_COMMAND_RESULT",
	SMSGInitializeFactions:        "SMSG_INITIALIZE_FACTIONS",
	CMSGSetActionButton:           "CMSG_SET_ACTION_BUTTON",
	SMSGActionButtons:             "SMSG_ACTION_BUTTONS",
	SMSGInitialSpells:             "SMSG_INITIAL_SPELLS",
	CMSGNewSpellSlot:              "CMSG_NEW_SPELL_SLOT",
	CMSGSetSelection:              "CMSG_SET_SELECTION",
	CMSGSetTarget:                 "CMSG_SET_TARGET",
	CMSGStandStateChange:          "CMSG_STANDSTATECHANGE",
	CMSGTextEmote:                 "CMSG_TEXT_EMOTE",
	SMSGEmote:                     "SMSG_EMOTE",
	SMSGTextEmote:                 "SMSG_TEXT_EMOTE",
	SMSGDestroyObject:             "SMSG_DESTROY_OBJECT",
	CMSGOpenItem:                  "CMSG_OPEN_ITEM",
	CMSGReadItem:                  "CMSG_READ_ITEM",
	SMSGReadItemOK:                "SMSG_READ_ITEM_OK",
	SMSGReadItemFailed:            "SMSG_READ_ITEM_FAILED",
	CMSGAutoequipItem:             "CMSG_AUTOEQUIP_ITEM",
	CMSGAutostoreBagItem:          "CMSG_AUTOSTORE_BAG_ITEM",
	CMSGSwapItem:                  "CMSG_SWAP_ITEM",
	CMSGSwapInvItem:               "CMSG_SWAP_INV_ITEM",
	CMSGSplitItem:                 "CMSG_SPLIT_ITEM",
	CMSGDestroyItem:               "CMSG_DESTROYITEM",
	SMSGInventoryChangeFailure:    "SMSG_INVENTORY_CHANGE_FAILURE",
	SMSGUpdateObject:              "SMSG_UPDATE_OBJECT",
	CMSGPing:                      "CMSG_PING",
	SMSGPong:                      "SMSG_PONG",
	CMSGPlayedTime:                "CMSG_PLAYED_TIME",
	SMSGPlayedTime:                "SMSG_PLAYED_TIME",
	MSGRandomRoll:                 "MSG_RANDOM_ROLL",
	MSGLookingForGroup:            "MSG_LOOKING_FOR_GROUP",
	CMSGSetLookingForGroup:        "CMSG_SET_LOOKING_FOR_GROUP",
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
