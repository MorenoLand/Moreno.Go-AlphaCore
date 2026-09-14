package packet

type Opcode uint32

const (
	CMSGWorldTeleport             Opcode = 0x0008
	CMSGZoneMap                   Opcode = 0x000a
	CMSGEnablePVP                 Opcode = 0x0030
	CMSGPVPPort                   Opcode = 0x0032
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
	MSGMoveStartForward           Opcode = 0x00b5
	MSGMoveStartBackward          Opcode = 0x00b6
	MSGMoveStop                   Opcode = 0x00b7
	MSGMoveStartStrafeLeft        Opcode = 0x00b8
	MSGMoveStartStrafeRight       Opcode = 0x00b9
	MSGMoveStopStrafe             Opcode = 0x00ba
	MSGMoveJump                   Opcode = 0x00bb
	MSGMoveStartTurnLeft          Opcode = 0x00bc
	MSGMoveStartTurnRight         Opcode = 0x00bd
	MSGMoveStopTurn               Opcode = 0x00be
	MSGMoveStartPitchUp           Opcode = 0x00bf
	MSGMoveStartPitchDown         Opcode = 0x00c0
	MSGMoveStopPitch              Opcode = 0x00c1
	MSGMoveSetRunMode             Opcode = 0x00c2
	MSGMoveSetWalkMode            Opcode = 0x00c3
	MSGMoveToggleLogging          Opcode = 0x00c4
	MSGMoveTeleport               Opcode = 0x00c5
	MSGMoveTeleportCheat          Opcode = 0x00c6
	MSGMoveTeleportAck            Opcode = 0x00c7
	MSGMoveToggleFallLogging      Opcode = 0x00c8
	MSGMoveCollideRedirect        Opcode = 0x00c9
	MSGMoveCollideStuck           Opcode = 0x00ca
	MSGMoveStartSwim              Opcode = 0x00cb
	MSGMoveStopSwim               Opcode = 0x00cc
	MSGMoveSetRunSpeedCheat       Opcode = 0x00cd
	MSGMoveSetRunSpeed            Opcode = 0x00ce
	MSGMoveSetWalkSpeedCheat      Opcode = 0x00cf
	MSGMoveSetWalkSpeed           Opcode = 0x00d0
	MSGMoveSetSwimSpeedCheat      Opcode = 0x00d1
	MSGMoveSetSwimSpeed           Opcode = 0x00d2
	MSGMoveSetAllSpeedCheat       Opcode = 0x00d3
	MSGMoveSetTurnRateCheat       Opcode = 0x00d4
	MSGMoveSetTurnRate            Opcode = 0x00d5
	MSGMoveToggleCollisionCheat   Opcode = 0x00d6
	MSGMoveSetFacing              Opcode = 0x00d7
	MSGMoveSetPitch               Opcode = 0x00d8
	MSGMoveWorldportAck           Opcode = 0x00d9
	SMSGMonsterMove               Opcode = 0x00da
	SMSGForceSpeedChange          Opcode = 0x00df
	CMSGForceSpeedChangeAck       Opcode = 0x00e0
	SMSGForceSwimSpeedChange      Opcode = 0x00e1
	CMSGForceSwimSpeedChangeAck   Opcode = 0x00e2
	SMSGForceMoveRoot             Opcode = 0x00e3
	CMSGForceMoveRootAck          Opcode = 0x00e4
	SMSGForceMoveUnroot           Opcode = 0x00e5
	CMSGForceMoveUnrootAck        Opcode = 0x00e6
	MSGMoveRoot                   Opcode = 0x00e7
	MSGMoveUnroot                 Opcode = 0x00e8
	MSGMoveHeartbeat              Opcode = 0x00e9
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
	CMSGGameObjectUse             Opcode = 0x00b1
	CMSGAutostoreLootItem         Opcode = 0x00fb
	CMSGInitiateTrade             Opcode = 0x0109
	CMSGBeginTrade                Opcode = 0x010a
	CMSGAcceptTrade               Opcode = 0x010d
	CMSGUnacceptTrade             Opcode = 0x010e
	CMSGCancelTrade               Opcode = 0x010f
	CMSGSetTradeItem              Opcode = 0x0110
	CMSGClearTradeItem            Opcode = 0x0111
	CMSGSetTradeGold              Opcode = 0x0112
	SMSGTradeStatus               Opcode = 0x0113
	SMSGTradeStatusExtended       Opcode = 0x0114
	CMSGLoot                      Opcode = 0x0150
	CMSGLootMoney                 Opcode = 0x0151
	CMSGLootRelease               Opcode = 0x0152
	SMSGLootResponse              Opcode = 0x0153
	SMSGLootReleaseResponse       Opcode = 0x0154
	SMSGLootRemoved               Opcode = 0x0155
	SMSGLootMoneyNotify           Opcode = 0x0156
	SMSGLootItemNotify            Opcode = 0x0157
	SMSGLootClearMoney            Opcode = 0x0158
	CMSGPetitionShowlist          Opcode = 0x01ad
	SMSGPetitionShowlist          Opcode = 0x01ae
	CMSGPetitionBuy               Opcode = 0x01af
	CMSGPetitionShowSignatures    Opcode = 0x01b0
	SMSGPetitionShowSignatures    Opcode = 0x01b1
	CMSGPetitionSign              Opcode = 0x01b2
	SMSGPetitionSignResults       Opcode = 0x01b3
	CMSGOfferPetition             Opcode = 0x01b4
	CMSGTurnInPetition            Opcode = 0x01b5
	SMSGTurnInPetitionResults     Opcode = 0x01b6
	CMSGPetitionQuery             Opcode = 0x01b7
	SMSGPetitionQueryResponse     Opcode = 0x01b8
	SMSGGameObjectQueryResponse   Opcode = 0x005f
	CMSGCreatureQuery             Opcode = 0x0060
	SMSGCreatureQueryResponse     Opcode = 0x0061
	CMSGUseItem                   Opcode = 0x00ab
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
	CMSGGuildQuery                Opcode = 0x0054
	SMSGGuildQueryResponse        Opcode = 0x0055
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
	CMSGLootMethod                Opcode = 0x007a
	CMSGGroupDisband              Opcode = 0x007b
	SMSGGroupDestroyed            Opcode = 0x007c
	SMSGGroupList                 Opcode = 0x007d
	SMSGPartyMemberStats          Opcode = 0x007e
	SMSGPartyCommandResult        Opcode = 0x007f
	CMSGGuildCreate               Opcode = 0x0081
	CMSGGuildInvite               Opcode = 0x0082
	SMSGGuildInvite               Opcode = 0x0083
	CMSGGuildAccept               Opcode = 0x0084
	CMSGGuildDecline              Opcode = 0x0085
	SMSGGuildDecline              Opcode = 0x0086
	CMSGGuildInfo                 Opcode = 0x0087
	SMSGGuildInfo                 Opcode = 0x0088
	CMSGGuildRoster               Opcode = 0x0089
	SMSGGuildRoster               Opcode = 0x008a
	CMSGGuildPromote              Opcode = 0x008b
	CMSGGuildDemote               Opcode = 0x008c
	CMSGGuildLeave                Opcode = 0x008d
	CMSGGuildRemove               Opcode = 0x008e
	CMSGGuildDisband              Opcode = 0x008f
	CMSGGuildLeader               Opcode = 0x0090
	CMSGGuildMOTD                 Opcode = 0x0091
	SMSGGuildEvent                Opcode = 0x0092
	SMSGGuildCommandResult        Opcode = 0x0093
	CMSGJoinChannel               Opcode = 0x0097
	CMSGLeaveChannel              Opcode = 0x0098
	SMSGChannelNotify             Opcode = 0x0099
	CMSGChannelList               Opcode = 0x009a
	SMSGChannelList               Opcode = 0x009b
	CMSGChannelPassword           Opcode = 0x009c
	CMSGChannelSetOwner           Opcode = 0x009d
	CMSGChannelOwner              Opcode = 0x009e
	CMSGChannelModerator          Opcode = 0x009f
	CMSGChannelUnmoderator        Opcode = 0x00a0
	CMSGChannelMute               Opcode = 0x00a1
	CMSGChannelUnmute             Opcode = 0x00a2
	CMSGChannelInvite             Opcode = 0x00a3
	CMSGChannelKick               Opcode = 0x00a4
	CMSGChannelBan                Opcode = 0x00a5
	CMSGChannelUnban              Opcode = 0x00a6
	CMSGChannelAnnouncements      Opcode = 0x00a7
	CMSGChannelModerate           Opcode = 0x00a8
	SMSGInitializeFactions        Opcode = 0x0115
	CMSGSetActionButton           Opcode = 0x011b
	SMSGActionButtons             Opcode = 0x011c
	SMSGInitialSpells             Opcode = 0x011d
	SMSGLearnedSpell              Opcode = 0x011e
	SMSSupersededSpell            Opcode = 0x011f
	CMSGNewSpellSlot              Opcode = 0x0120
	CMSGCastSpell                 Opcode = 0x0121
	CMSGCancelCast                Opcode = 0x0122
	SMSGCastResult                Opcode = 0x0123
	SMSGSpellStart                Opcode = 0x0124
	SMSGSpellGo                   Opcode = 0x0125
	SMSGSpellFailure              Opcode = 0x0126
	SMSGSpellCooldown             Opcode = 0x0127
	SMSGCooldownEvent             Opcode = 0x0128
	CMSGCancelAura                Opcode = 0x0129
	SMSGUpdateAuraDuration        Opcode = 0x012a
	SMSGPetCastFailed             Opcode = 0x012b
	MSGChannelStart               Opcode = 0x012c
	MSGChannelUpdate              Opcode = 0x012d
	CMSGCancelChannelling         Opcode = 0x012e
	SMSGClearCooldown             Opcode = 0x01cf
	CMSGMountSpecialAnim          Opcode = 0x0164
	SMSGMountSpecialAnim          Opcode = 0x0165
	CMSGListInventory             Opcode = 0x016e
	SMSGListInventory             Opcode = 0x016f
	SMSGItemPushResult            Opcode = 0x0159
	CMSGSellItem                  Opcode = 0x0170
	SMSGSellItem                  Opcode = 0x0171
	CMSGBuyItem                   Opcode = 0x0172
	CMSGBuyItemInSlot             Opcode = 0x0173
	SMSGBuyItem                   Opcode = 0x0174
	SMSGBuyFailed                 Opcode = 0x0175
	CMSGQuestGiverStatusQuery     Opcode = 0x017e
	SMSGQuestGiverStatus          Opcode = 0x017f
	CMSGQuestGiverHello           Opcode = 0x0180
	SMSGQuestGiverQuestList       Opcode = 0x0181
	CMSGQuestGiverQueryQuest      Opcode = 0x0182
	CMSGQuestGiverAcceptQuest     Opcode = 0x0185
	CMSGQuestGiverCompleteQuest   Opcode = 0x0186
	SMSGQuestGiverRequestItems    Opcode = 0x0187
	CMSGQuestGiverRequestReward   Opcode = 0x0188
	SMSGQuestGiverOfferReward     Opcode = 0x0189
	CMSGQuestGiverChooseReward    Opcode = 0x018a
	SMSGQuestGiverQuestInvalid    Opcode = 0x018b
	CMSGQuestGiverCancel          Opcode = 0x018c
	SMSGQuestGiverQuestComplete   Opcode = 0x018d
	SMSGQuestGiverQuestFailed     Opcode = 0x018e
	CMSGQuestLogRemoveQuest       Opcode = 0x0190
	SMSGQuestLogFull              Opcode = 0x0191
	CMSGQuestConfirmAccept        Opcode = 0x0196
	SMSGQuestConfirmAccept        Opcode = 0x0197
	CMSGTaxiClearAllNodes         Opcode = 0x0198
	CMSGTaxiEnableAllNodes        Opcode = 0x0199
	CMSGTaxiShowNodes             Opcode = 0x019a
	SMSGShowTaxiNodes             Opcode = 0x019b
	CMSGTaxiNodeStatusQuery       Opcode = 0x019c
	SMSGTaxiNodeStatus            Opcode = 0x019d
	CMSGTaxiQueryAvailableNodes   Opcode = 0x019e
	CMSGActivateTaxi              Opcode = 0x019f
	SMSGActivateTaxiReply         Opcode = 0x01a0
	SMSGNewTaxiPath               Opcode = 0x01a1
	CMSGBinderActivate            Opcode = 0x01a7
	SMSGPlayerBindError           Opcode = 0x01a8
	CMSGBankerActivate            Opcode = 0x01a9
	SMSGShowBank                  Opcode = 0x01aa
	CMSGBuyBankSlot               Opcode = 0x01ab
	SMSGBuyBankSlotResult         Opcode = 0x01ac
	CMSGTrainerList               Opcode = 0x01a2
	SMSGTrainerList               Opcode = 0x01a3
	CMSGTrainerBuySpell           Opcode = 0x01a4
	SMSGTrainerBuySucceeded       Opcode = 0x01a5
	SMSGTrainerBuyFailed          Opcode = 0x01a6
	SMSGQuestGiverQuestDetails    Opcode = 0x0184
	SMSGQuestUpdateFailed         Opcode = 0x0192
	SMSGQuestUpdateComplete       Opcode = 0x0193
	SMSGQuestUpdateAddKill        Opcode = 0x0194
	SMSGQuestUpdateAddItem        Opcode = 0x0195
	CMSGSetSelection              Opcode = 0x0130
	CMSGSetTarget                 Opcode = 0x0131
	CMSGAttackSwing               Opcode = 0x0134
	CMSGAttackStop                Opcode = 0x0135
	SMSGAttackStart               Opcode = 0x0136
	SMSGAttackStop                Opcode = 0x0137
	SMSGAttackSwingNotInRange     Opcode = 0x0138
	SMSGAttackSwingBadFacing      Opcode = 0x0139
	SMSGAttackSwingNotStanding    Opcode = 0x013a
	SMSGAttackSwingDeadTarget     Opcode = 0x013b
	SMSGAttackSwingCantAttack     Opcode = 0x013c
	SMSGAttackerStateUpdate       Opcode = 0x013d
	SMSGVictimStateUpdateObsolete Opcode = 0x013e
	SMSGDamageDone                Opcode = 0x013f
	SMSGDamageTaken               Opcode = 0x0140
	SMSGCancelCombat              Opcode = 0x0141
	CMSGStandStateChange          Opcode = 0x00f4
	CMSGTextEmote                 Opcode = 0x00f7
	SMSGEmote                     Opcode = 0x00f6
	SMSGTextEmote                 Opcode = 0x00f8
	SMSGDestroyObject             Opcode = 0x00aa
	CMSGOpenItem                  Opcode = 0x00ac
	CMSGInspect                   Opcode = 0x0107
	SMSGInspect                   Opcode = 0x0108
	CMSGSetDeathBindPoint         Opcode = 0x0147
	SMSGBindPointUpdate           Opcode = 0x0148
	CMSGGetDeathBindZone          Opcode = 0x0149
	SMSGBindZoneReply             Opcode = 0x014a
	SMSGPlayerBound               Opcode = 0x014b
	SMSGResurrectRequest          Opcode = 0x014e
	CMSGRepopRequest              Opcode = 0x014d
	CMSGResurrectResponse         Opcode = 0x014f
	CMSGReclaimCorpse             Opcode = 0x01c3
	CMSGReadItem                  Opcode = 0x00ad
	SMSGReadItemOK                Opcode = 0x00ae
	SMSGReadItemFailed            Opcode = 0x00af
	CMSGAutoequipItem             Opcode = 0x00fd
	CMSGAutostoreBagItem          Opcode = 0x00fe
	CMSGSwapItem                  Opcode = 0x00ff
	CMSGSwapInvItem               Opcode = 0x0100
	CMSGSplitItem                 Opcode = 0x0101
	CMSGDestroyItem               Opcode = 0x0104
	CMSGWrapItem                  Opcode = 0x01c4
	SMSGInventoryChangeFailure    Opcode = 0x0105
	SMSGUpdateObject              Opcode = 0x00a9
	CMSGQueryTime                 Opcode = 0x01bf
	SMSGQueryTimeResponse         Opcode = 0x01c0
	CMSGPlayedTime                Opcode = 0x01bd
	SMSGPlayedTime                Opcode = 0x01be
	CMSGSetWeaponMode             Opcode = 0x01d1
	CMSGPlayerMacro               Opcode = 0x01d4
	SMSGPlayerMacro               Opcode = 0x01d5
	MSGMinimapPing                Opcode = 0x01c6
	CMSGZoneUpdate                Opcode = 0x01e5
	CMSGAreaTrigger               Opcode = 0x00b4
	CMSGPing                      Opcode = 0x01cd
	SMSGPong                      Opcode = 0x01ce
	MSGRandomRoll                 Opcode = 0x01ec
	MSGLookingForGroup            Opcode = 0x01f0
	CMSGSetLookingForGroup        Opcode = 0x01f1
	CMSGAuthSession               Opcode = 0x01de
	SMSGAuthResponse              Opcode = 0x01df
	MSGSaveGuildEmblem            Opcode = 0x01e2
	MSGTabardVendorActivate       Opcode = 0x01e3
	CMSGBug                       Opcode = 0x01bb
	SMSGCompressedUpdateObject    Opcode = 0x01e7
)

var opcodeNames = map[Opcode]string{
	CMSGWorldTeleport:             "CMSG_WORLD_TELEPORT",
	CMSGZoneMap:                   "CMSG_ZONE_MAP",
	CMSGEnablePVP:                 "CMSG_ENABLE_PVP",
	CMSGPVPPort:                   "CMSG_PVP_PORT",
	CMSGAuthSRP6Begin:             "CMSG_AUTH_SRP6_BEGIN",
	CMSGAuthSRP6Proof:             "CMSG_AUTH_SRP6_PROOF",
	CMSGAuthSRP6Recode:            "CMSG_AUTH_SRP6_RECODE",
	CMSGCharCreate:                "CMSG_CHAR_CREATE",
	CMSGCharEnum:                  "CMSG_CHAR_ENUM",
	CMSGCharDelete:                "CMSG_CHAR_DELETE",
	CMSGUseItem:                   "CMSG_USE_ITEM",
	SMSGAuthSRP6Response:          "SMSG_AUTH_SRP6_RESPONSE",
	SMSGCharCreate:                "SMSG_CHAR_CREATE",
	SMSGCharEnum:                  "SMSG_CHAR_ENUM",
	SMSGCharDelete:                "SMSG_CHAR_DELETE",
	CMSGPlayerLogin:               "CMSG_PLAYER_LOGIN",
	SMSGNewWorld:                  "SMSG_NEW_WORLD",
	MSGMoveStartForward:           "MSG_MOVE_START_FORWARD",
	MSGMoveStartBackward:          "MSG_MOVE_START_BACKWARD",
	MSGMoveStop:                   "MSG_MOVE_STOP",
	MSGMoveStartStrafeLeft:        "MSG_MOVE_START_STRAFE_LEFT",
	MSGMoveStartStrafeRight:       "MSG_MOVE_START_STRAFE_RIGHT",
	MSGMoveStopStrafe:             "MSG_MOVE_STOP_STRAFE",
	MSGMoveJump:                   "MSG_MOVE_JUMP",
	MSGMoveStartTurnLeft:          "MSG_MOVE_START_TURN_LEFT",
	MSGMoveStartTurnRight:         "MSG_MOVE_START_TURN_RIGHT",
	MSGMoveStopTurn:               "MSG_MOVE_STOP_TURN",
	MSGMoveStartPitchUp:           "MSG_MOVE_START_PITCH_UP",
	MSGMoveStartPitchDown:         "MSG_MOVE_START_PITCH_DOWN",
	MSGMoveStopPitch:              "MSG_MOVE_STOP_PITCH",
	MSGMoveSetRunMode:             "MSG_MOVE_SET_RUN_MODE",
	MSGMoveSetWalkMode:            "MSG_MOVE_SET_WALK_MODE",
	MSGMoveToggleLogging:          "MSG_MOVE_TOGGLE_LOGGING",
	MSGMoveTeleport:               "MSG_MOVE_TELEPORT",
	MSGMoveTeleportCheat:          "MSG_MOVE_TELEPORT_CHEAT",
	MSGMoveTeleportAck:            "MSG_MOVE_TELEPORT_ACK",
	MSGMoveToggleFallLogging:      "MSG_MOVE_TOGGLE_FALL_LOGGING",
	MSGMoveCollideRedirect:        "MSG_MOVE_COLLIDE_REDIRECT",
	MSGMoveCollideStuck:           "MSG_MOVE_COLLIDE_STUCK",
	MSGMoveStartSwim:              "MSG_MOVE_START_SWIM",
	MSGMoveStopSwim:               "MSG_MOVE_STOP_SWIM",
	MSGMoveSetRunSpeedCheat:       "MSG_MOVE_SET_RUN_SPEED_CHEAT",
	MSGMoveSetRunSpeed:            "MSG_MOVE_SET_RUN_SPEED",
	MSGMoveSetWalkSpeedCheat:      "MSG_MOVE_SET_WALK_SPEED_CHEAT",
	MSGMoveSetWalkSpeed:           "MSG_MOVE_SET_WALK_SPEED",
	MSGMoveSetSwimSpeedCheat:      "MSG_MOVE_SET_SWIM_SPEED_CHEAT",
	MSGMoveSetSwimSpeed:           "MSG_MOVE_SET_SWIM_SPEED",
	MSGMoveSetAllSpeedCheat:       "MSG_MOVE_SET_ALL_SPEED_CHEAT",
	MSGMoveSetTurnRateCheat:       "MSG_MOVE_SET_TURN_RATE_CHEAT",
	MSGMoveSetTurnRate:            "MSG_MOVE_SET_TURN_RATE",
	MSGMoveToggleCollisionCheat:   "MSG_MOVE_TOGGLE_COLLISION_CHEAT",
	MSGMoveSetFacing:              "MSG_MOVE_SET_FACING",
	MSGMoveSetPitch:               "MSG_MOVE_SET_PITCH",
	MSGMoveWorldportAck:           "MSG_MOVE_WORLDPORT_ACK",
	SMSGMonsterMove:               "SMSG_MONSTER_MOVE",
	CMSGAreaTrigger:               "CMSG_AREATRIGGER",
	SMSGForceSpeedChange:          "SMSG_FORCE_SPEED_CHANGE",
	CMSGForceSpeedChangeAck:       "CMSG_FORCE_SPEED_CHANGE_ACK",
	SMSGForceSwimSpeedChange:      "SMSG_FORCE_SWIM_SPEED_CHANGE",
	CMSGForceSwimSpeedChangeAck:   "CMSG_FORCE_SWIM_SPEED_CHANGE_ACK",
	SMSGForceMoveRoot:             "SMSG_FORCE_MOVE_ROOT",
	CMSGForceMoveRootAck:          "CMSG_FORCE_MOVE_ROOT_ACK",
	SMSGForceMoveUnroot:           "SMSG_FORCE_MOVE_UNROOT",
	CMSGForceMoveUnrootAck:        "CMSG_FORCE_MOVE_UNROOT_ACK",
	MSGMoveRoot:                   "MSG_MOVE_ROOT",
	MSGMoveUnroot:                 "MSG_MOVE_UNROOT",
	MSGMoveHeartbeat:              "MSG_MOVE_HEARTBEAT",
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
	CMSGGameObjectUse:             "CMSG_GAMEOBJ_USE",
	CMSGAutostoreLootItem:         "CMSG_AUTOSTORE_LOOT_ITEM",
	CMSGInitiateTrade:             "CMSG_INITIATE_TRADE",
	CMSGBeginTrade:                "CMSG_BEGIN_TRADE",
	CMSGAcceptTrade:               "CMSG_ACCEPT_TRADE",
	CMSGUnacceptTrade:             "CMSG_UNACCEPT_TRADE",
	CMSGCancelTrade:               "CMSG_CANCEL_TRADE",
	CMSGSetTradeItem:              "CMSG_SET_TRADE_ITEM",
	CMSGClearTradeItem:            "CMSG_CLEAR_TRADE_ITEM",
	CMSGSetTradeGold:              "CMSG_SET_TRADE_GOLD",
	SMSGTradeStatus:               "SMSG_TRADE_STATUS",
	SMSGTradeStatusExtended:       "SMSG_TRADE_STATUS_EXTENDED",
	CMSGLoot:                      "CMSG_LOOT",
	CMSGLootMoney:                 "CMSG_LOOT_MONEY",
	CMSGLootRelease:               "CMSG_LOOT_RELEASE",
	SMSGLootResponse:              "SMSG_LOOT_RESPONSE",
	SMSGLootReleaseResponse:       "SMSG_LOOT_RELEASE_RESPONSE",
	SMSGLootRemoved:               "SMSG_LOOT_REMOVED",
	SMSGLootMoneyNotify:           "SMSG_LOOT_MONEY_NOTIFY",
	SMSGLootItemNotify:            "SMSG_LOOT_ITEM_NOTIFY",
	SMSGLootClearMoney:            "SMSG_LOOT_CLEAR_MONEY",
	CMSGPetitionShowlist:          "CMSG_PETITION_SHOWLIST",
	SMSGPetitionShowlist:          "SMSG_PETITION_SHOWLIST",
	CMSGPetitionBuy:               "CMSG_PETITION_BUY",
	CMSGPetitionShowSignatures:    "CMSG_PETITION_SHOW_SIGNATURES",
	SMSGPetitionShowSignatures:    "SMSG_PETITION_SHOW_SIGNATURES",
	CMSGPetitionSign:              "CMSG_PETITION_SIGN",
	SMSGPetitionSignResults:       "SMSG_PETITION_SIGN_RESULTS",
	CMSGOfferPetition:             "CMSG_OFFER_PETITION",
	CMSGTurnInPetition:            "CMSG_TURN_IN_PETITION",
	SMSGTurnInPetitionResults:     "SMSG_TURN_IN_PETITION_RESULTS",
	CMSGPetitionQuery:             "CMSG_PETITION_QUERY",
	SMSGPetitionQueryResponse:     "SMSG_PETITION_QUERY_RESPONSE",
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
	CMSGGuildQuery:                "CMSG_GUILD_QUERY",
	SMSGGuildQueryResponse:        "SMSG_GUILD_QUERY_RESPONSE",
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
	CMSGLootMethod:                "CMSG_LOOT_METHOD",
	CMSGGroupDisband:              "CMSG_GROUP_DISBAND",
	SMSGGroupDestroyed:            "SMSG_GROUP_DESTROYED",
	SMSGGroupList:                 "SMSG_GROUP_LIST",
	SMSGPartyMemberStats:          "SMSG_PARTY_MEMBER_STATS",
	SMSGPartyCommandResult:        "SMSG_PARTY_COMMAND_RESULT",
	CMSGGuildCreate:               "CMSG_GUILD_CREATE",
	CMSGGuildInvite:               "CMSG_GUILD_INVITE",
	SMSGGuildInvite:               "SMSG_GUILD_INVITE",
	CMSGGuildAccept:               "CMSG_GUILD_ACCEPT",
	CMSGGuildDecline:              "CMSG_GUILD_DECLINE",
	SMSGGuildDecline:              "SMSG_GUILD_DECLINE",
	CMSGGuildInfo:                 "CMSG_GUILD_INFO",
	SMSGGuildInfo:                 "SMSG_GUILD_INFO",
	CMSGGuildRoster:               "CMSG_GUILD_ROSTER",
	SMSGGuildRoster:               "SMSG_GUILD_ROSTER",
	CMSGGuildPromote:              "CMSG_GUILD_PROMOTE",
	CMSGGuildDemote:               "CMSG_GUILD_DEMOTE",
	CMSGGuildLeave:                "CMSG_GUILD_LEAVE",
	CMSGGuildRemove:               "CMSG_GUILD_REMOVE",
	CMSGGuildDisband:              "CMSG_GUILD_DISBAND",
	CMSGGuildLeader:               "CMSG_GUILD_LEADER",
	CMSGGuildMOTD:                 "CMSG_GUILD_MOTD",
	SMSGGuildEvent:                "SMSG_GUILD_EVENT",
	SMSGGuildCommandResult:        "SMSG_GUILD_COMMAND_RESULT",
	CMSGJoinChannel:               "CMSG_JOIN_CHANNEL",
	CMSGLeaveChannel:              "CMSG_LEAVE_CHANNEL",
	SMSGChannelNotify:             "SMSG_CHANNEL_NOTIFY",
	CMSGChannelList:               "CMSG_CHANNEL_LIST",
	SMSGChannelList:               "SMSG_CHANNEL_LIST",
	CMSGChannelPassword:           "CMSG_CHANNEL_PASSWORD",
	CMSGChannelSetOwner:           "CMSG_CHANNEL_SET_OWNER",
	CMSGChannelOwner:              "CMSG_CHANNEL_OWNER",
	CMSGChannelModerator:          "CMSG_CHANNEL_MODERATOR",
	CMSGChannelUnmoderator:        "CMSG_CHANNEL_UNMODERATOR",
	CMSGChannelMute:               "CMSG_CHANNEL_MUTE",
	CMSGChannelUnmute:             "CMSG_CHANNEL_UNMUTE",
	CMSGChannelInvite:             "CMSG_CHANNEL_INVITE",
	CMSGChannelKick:               "CMSG_CHANNEL_KICK",
	CMSGChannelBan:                "CMSG_CHANNEL_BAN",
	CMSGChannelUnban:              "CMSG_CHANNEL_UNBAN",
	CMSGChannelAnnouncements:      "CMSG_CHANNEL_ANNOUNCEMENTS",
	CMSGChannelModerate:           "CMSG_CHANNEL_MODERATE",
	SMSGInitializeFactions:        "SMSG_INITIALIZE_FACTIONS",
	CMSGSetActionButton:           "CMSG_SET_ACTION_BUTTON",
	SMSGActionButtons:             "SMSG_ACTION_BUTTONS",
	SMSGInitialSpells:             "SMSG_INITIAL_SPELLS",
	SMSGLearnedSpell:              "SMSG_LEARNED_SPELL",
	SMSSupersededSpell:            "SMSG_SUPERCEDED_SPELL",
	CMSGNewSpellSlot:              "CMSG_NEW_SPELL_SLOT",
	CMSGCastSpell:                 "CMSG_CAST_SPELL",
	CMSGCancelCast:                "CMSG_CANCEL_CAST",
	SMSGCastResult:                "SMSG_CAST_RESULT",
	SMSGSpellStart:                "SMSG_SPELL_START",
	SMSGSpellGo:                   "SMSG_SPELL_GO",
	SMSGSpellFailure:              "SMSG_SPELL_FAILURE",
	SMSGSpellCooldown:             "SMSG_SPELL_COOLDOWN",
	SMSGCooldownEvent:             "SMSG_COOLDOWN_EVENT",
	CMSGCancelAura:                "CMSG_CANCEL_AURA",
	SMSGUpdateAuraDuration:        "SMSG_UPDATE_AURA_DURATION",
	SMSGPetCastFailed:             "SMSG_PET_CAST_FAILED",
	MSGChannelStart:               "MSG_CHANNEL_START",
	MSGChannelUpdate:              "MSG_CHANNEL_UPDATE",
	CMSGCancelChannelling:         "CMSG_CANCEL_CHANNELLING",
	SMSGClearCooldown:             "SMSG_CLEAR_COOLDOWN",
	CMSGMountSpecialAnim:          "CMSG_MOUNTSPECIAL_ANIM",
	SMSGMountSpecialAnim:          "SMSG_MOUNTSPECIAL_ANIM",
	CMSGListInventory:             "CMSG_LIST_INVENTORY",
	SMSGListInventory:             "SMSG_LIST_INVENTORY",
	SMSGItemPushResult:            "SMSG_ITEM_PUSH_RESULT",
	CMSGSellItem:                  "CMSG_SELL_ITEM",
	SMSGSellItem:                  "SMSG_SELL_ITEM",
	CMSGBuyItem:                   "CMSG_BUY_ITEM",
	CMSGBuyItemInSlot:             "CMSG_BUY_ITEM_IN_SLOT",
	SMSGBuyItem:                   "SMSG_BUY_ITEM",
	SMSGBuyFailed:                 "SMSG_BUY_FAILED",
	CMSGQuestGiverStatusQuery:     "CMSG_QUESTGIVER_STATUS_QUERY",
	SMSGQuestGiverStatus:          "SMSG_QUESTGIVER_STATUS",
	CMSGQuestGiverHello:           "CMSG_QUESTGIVER_HELLO",
	SMSGQuestGiverQuestList:       "SMSG_QUESTGIVER_QUEST_LIST",
	CMSGQuestGiverQueryQuest:      "CMSG_QUESTGIVER_QUERY_QUEST",
	CMSGQuestGiverAcceptQuest:     "CMSG_QUESTGIVER_ACCEPT_QUEST",
	CMSGQuestGiverCompleteQuest:   "CMSG_QUESTGIVER_COMPLETE_QUEST",
	SMSGQuestGiverRequestItems:    "SMSG_QUESTGIVER_REQUEST_ITEMS",
	CMSGQuestGiverRequestReward:   "CMSG_QUESTGIVER_REQUEST_REWARD",
	SMSGQuestGiverOfferReward:     "SMSG_QUESTGIVER_OFFER_REWARD",
	CMSGQuestGiverChooseReward:    "CMSG_QUESTGIVER_CHOOSE_REWARD",
	SMSGQuestGiverQuestInvalid:    "SMSG_QUESTGIVER_QUEST_INVALID",
	CMSGQuestGiverCancel:          "CMSG_QUESTGIVER_CANCEL",
	SMSGQuestGiverQuestComplete:   "SMSG_QUESTGIVER_QUEST_COMPLETE",
	SMSGQuestGiverQuestFailed:     "SMSG_QUESTGIVER_QUEST_FAILED",
	CMSGQuestLogRemoveQuest:       "CMSG_QUESTLOG_REMOVE_QUEST",
	SMSGQuestLogFull:              "SMSG_QUESTLOG_FULL",
	CMSGQuestConfirmAccept:        "CMSG_QUEST_CONFIRM_ACCEPT",
	SMSGQuestConfirmAccept:        "SMSG_QUEST_CONFIRM_ACCEPT",
	CMSGTaxiClearAllNodes:         "CMSG_TAXICLEARALLNODES",
	CMSGTaxiEnableAllNodes:        "CMSG_TAXIENABLEALLNODES",
	CMSGTaxiShowNodes:             "CMSG_TAXISHOWNODES",
	SMSGShowTaxiNodes:             "SMSG_SHOWTAXINODES",
	CMSGTaxiNodeStatusQuery:       "CMSG_TAXINODE_STATUS_QUERY",
	SMSGTaxiNodeStatus:            "SMSG_TAXINODE_STATUS",
	CMSGTaxiQueryAvailableNodes:   "CMSG_TAXIQUERYAVAILABLENODES",
	CMSGActivateTaxi:              "CMSG_ACTIVATETAXI",
	SMSGActivateTaxiReply:         "SMSG_ACTIVATETAXIREPLY",
	SMSGNewTaxiPath:               "SMSG_NEW_TAXI_PATH",
	CMSGBinderActivate:            "CMSG_BINDER_ACTIVATE",
	SMSGPlayerBindError:           "SMSG_PLAYERBINDERROR",
	CMSGBankerActivate:            "CMSG_BANKER_ACTIVATE",
	SMSGShowBank:                  "SMSG_SHOW_BANK",
	CMSGBuyBankSlot:               "CMSG_BUY_BANK_SLOT",
	SMSGBuyBankSlotResult:         "SMSG_BUY_BANK_SLOT_RESULT",
	CMSGTrainerList:               "CMSG_TRAINER_LIST",
	SMSGTrainerList:               "SMSG_TRAINER_LIST",
	CMSGTrainerBuySpell:           "CMSG_TRAINER_BUY_SPELL",
	SMSGTrainerBuySucceeded:       "SMSG_TRAINER_BUY_SUCCEEDED",
	SMSGTrainerBuyFailed:          "SMSG_TRAINER_BUY_FAILED",
	SMSGQuestGiverQuestDetails:    "SMSG_QUESTGIVER_QUEST_DETAILS",
	SMSGQuestUpdateFailed:         "SMSG_QUESTUPDATE_FAILED",
	SMSGQuestUpdateComplete:       "SMSG_QUESTUPDATE_COMPLETE",
	SMSGQuestUpdateAddKill:        "SMSG_QUESTUPDATE_ADD_KILL",
	SMSGQuestUpdateAddItem:        "SMSG_QUESTUPDATE_ADD_ITEM",
	CMSGSetSelection:              "CMSG_SET_SELECTION",
	CMSGSetTarget:                 "CMSG_SET_TARGET",
	CMSGAttackSwing:               "CMSG_ATTACKSWING",
	CMSGAttackStop:                "CMSG_ATTACKSTOP",
	SMSGAttackStart:               "SMSG_ATTACKSTART",
	SMSGAttackStop:                "SMSG_ATTACKSTOP",
	SMSGAttackSwingNotInRange:     "SMSG_ATTACKSWING_NOTINRANGE",
	SMSGAttackSwingBadFacing:      "SMSG_ATTACKSWING_BADFACING",
	SMSGAttackSwingNotStanding:    "SMSG_ATTACKSWING_NOTSTANDING",
	SMSGAttackSwingDeadTarget:     "SMSG_ATTACKSWING_DEADTARGET",
	SMSGAttackSwingCantAttack:     "SMSG_ATTACKSWING_CANT_ATTACK",
	SMSGAttackerStateUpdate:       "SMSG_ATTACKERSTATEUPDATE",
	SMSGVictimStateUpdateObsolete: "SMSG_VICTIMSTATEUPDATE_OBSOLETE",
	SMSGDamageDone:                "SMSG_DAMAGE_DONE",
	SMSGDamageTaken:               "SMSG_DAMAGE_TAKEN",
	SMSGCancelCombat:              "SMSG_CANCEL_COMBAT",
	CMSGStandStateChange:          "CMSG_STANDSTATECHANGE",
	CMSGTextEmote:                 "CMSG_TEXT_EMOTE",
	SMSGEmote:                     "SMSG_EMOTE",
	SMSGTextEmote:                 "SMSG_TEXT_EMOTE",
	SMSGDestroyObject:             "SMSG_DESTROY_OBJECT",
	CMSGOpenItem:                  "CMSG_OPEN_ITEM",
	CMSGInspect:                   "CMSG_INSPECT",
	SMSGInspect:                   "SMSG_INSPECT",
	CMSGSetDeathBindPoint:         "CMSG_SETDEATHBINDPOINT",
	SMSGBindPointUpdate:           "SMSG_BINDPOINTUPDATE",
	CMSGGetDeathBindZone:          "CMSG_GETDEATHBINDZONE",
	SMSGBindZoneReply:             "SMSG_BINDZONEREPLY",
	SMSGPlayerBound:               "SMSG_PLAYERBOUND",
	SMSGResurrectRequest:          "SMSG_RESURRECT_REQUEST",
	CMSGRepopRequest:              "CMSG_REPOP_REQUEST",
	CMSGResurrectResponse:         "CMSG_RESURRECT_RESPONSE",
	CMSGReclaimCorpse:             "CMSG_RECLAIM_CORPSE",
	CMSGReadItem:                  "CMSG_READ_ITEM",
	SMSGReadItemOK:                "SMSG_READ_ITEM_OK",
	SMSGReadItemFailed:            "SMSG_READ_ITEM_FAILED",
	CMSGAutoequipItem:             "CMSG_AUTOEQUIP_ITEM",
	CMSGAutostoreBagItem:          "CMSG_AUTOSTORE_BAG_ITEM",
	CMSGSwapItem:                  "CMSG_SWAP_ITEM",
	CMSGSwapInvItem:               "CMSG_SWAP_INV_ITEM",
	CMSGSplitItem:                 "CMSG_SPLIT_ITEM",
	CMSGDestroyItem:               "CMSG_DESTROYITEM",
	CMSGWrapItem:                  "CMSG_WRAP_ITEM",
	SMSGInventoryChangeFailure:    "SMSG_INVENTORY_CHANGE_FAILURE",
	SMSGUpdateObject:              "SMSG_UPDATE_OBJECT",
	CMSGPing:                      "CMSG_PING",
	SMSGPong:                      "SMSG_PONG",
	CMSGPlayedTime:                "CMSG_PLAYED_TIME",
	SMSGPlayedTime:                "SMSG_PLAYED_TIME",
	CMSGSetWeaponMode:             "CMSG_SETWEAPONMODE",
	CMSGPlayerMacro:               "CMSG_PLAYER_MACRO",
	SMSGPlayerMacro:               "SMSG_PLAYER_MACRO",
	MSGMinimapPing:                "MSG_MINIMAP_PING",
	MSGRandomRoll:                 "MSG_RANDOM_ROLL",
	MSGLookingForGroup:            "MSG_LOOKING_FOR_GROUP",
	CMSGSetLookingForGroup:        "CMSG_SET_LOOKING_FOR_GROUP",
	CMSGAuthSession:               "CMSG_AUTH_SESSION",
	SMSGAuthResponse:              "SMSG_AUTH_RESPONSE",
	MSGSaveGuildEmblem:            "MSG_SAVE_GUILD_EMBLEM",
	MSGTabardVendorActivate:       "MSG_TABARDVENDOR_ACTIVATE",
	CMSGBug:                       "CMSG_BUG",
	SMSGCompressedUpdateObject:    "SMSG_COMPRESSED_UPDATE_OBJECT",
}

func IsMovement(opcode Opcode) bool { return opcode >= 0x00b5 && opcode <= 0x00e9 }

func (o Opcode) String() string {
	if name, ok := opcodeNames[o]; ok {
		return name
	}
	return "UNKNOWN"
}
