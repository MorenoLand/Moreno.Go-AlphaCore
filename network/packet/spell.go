package packet

type SpellTargetMask uint16

const (
	SpellTargetSelf           SpellTargetMask = 0
	SpellTargetUnit           SpellTargetMask = 0x0002
	SpellTargetPlayer         SpellTargetMask = 0x0008
	SpellTargetItem           SpellTargetMask = 0x0010
	SpellTargetSource         SpellTargetMask = 0x0020
	SpellTargetDestination    SpellTargetMask = 0x0040
	SpellTargetEnemies        SpellTargetMask = 0x0080
	SpellTargetUnitSelf       SpellTargetMask = 0x0100
	SpellTargetDead           SpellTargetMask = 0x0400
	SpellTargetGameObject     SpellTargetMask = 0x0800
	SpellTargetTradeItem      SpellTargetMask = 0x1000
	SpellTargetString         SpellTargetMask = 0x2000
	SpellTargetGameObjectItem SpellTargetMask = 0x4000
)

const (
	SpellTargetUnitMask  SpellTargetMask = SpellTargetUnit | SpellTargetGameObject
	SpellTargetItemMask  SpellTargetMask = SpellTargetItem | SpellTargetTradeItem
	SpellTargetTerrain   SpellTargetMask = SpellTargetSource | SpellTargetDestination
	SpellTargetObject    SpellTargetMask = SpellTargetEnemies | SpellTargetGameObject
	SpellTargetUnitTypes SpellTargetMask = SpellTargetUnit | SpellTargetPlayer | SpellTargetUnitSelf | SpellTargetDead | SpellTargetGameObject
)

type SpellCastResult byte

const (
	SpellFailedAffectingCombat SpellCastResult = 0x00
	SpellFailedError           SpellCastResult = 0x0e
	SpellFailedBadImplicit     SpellCastResult = 0x05
	SpellFailedBadTargets      SpellCastResult = 0x06
	SpellFailedCasterDead      SpellCastResult = 0x0a
	SpellFailedInterrupted     SpellCastResult = 0x11
	SpellFailedLevel           SpellCastResult = 0x16
	SpellFailedMoving          SpellCastResult = 0x1a
	SpellFailedNotKnown        SpellCastResult = 0x23
	SpellFailedNotReady        SpellCastResult = 0x25
	SpellFailedNoPower         SpellCastResult = 0x2c
	SpellFailedPacified        SpellCastResult = 0x37
	SpellFailedSilenced        SpellCastResult = 0x3a
	SpellFailedStunned         SpellCastResult = 0x3e
	SpellFailedOutOfRange      SpellCastResult = 0x36
	SpellFailedTooClose        SpellCastResult = 0x4a
	SpellFailedReagents        SpellCastResult = 0x38
	SpellFailedSpellProgress   SpellCastResult = 0x3b
	SpellFailedUnavailable     SpellCastResult = 0x3d
	SpellFailedTargetsDead     SpellCastResult = 0x3f
	SpellFailedTargetEnemy     SpellCastResult = 0x42
	SpellFailedTargetFriendly  SpellCastResult = 0x44
	SpellFailedTargetNotDead   SpellCastResult = 0x46
	SpellFailedTargetNoPockets SpellCastResult = 0x48
	SpellFailedNoComboPoints   SpellCastResult = 0x52
	SpellFailedTargetDueling   SpellCastResult = 0x54
	SpellFailedUnknown         SpellCastResult = 0x57
	SpellFailedDontReport      SpellCastResult = 0x0b
	SpellNoError               SpellCastResult = 0xff
)

type SpellCastStatus byte

const (
	SpellCastSuccess SpellCastStatus = 0
	SpellCastFailed  SpellCastStatus = 2
)

type SpellState byte

const (
	SpellStatePreparing SpellState = iota
	SpellStateCasting
	SpellStateFinished
	SpellStateDelayed
	SpellStateActive
)

type SpellAttributes uint32

const (
	SpellAttributeRanged              SpellAttributes = 0x00000002
	SpellAttributeAbility             SpellAttributes = 0x00000010
	SpellAttributeTrade               SpellAttributes = 0x00000020
	SpellAttributePassive             SpellAttributes = 0x00000040
	SpellAttributeDayOnly             SpellAttributes = 0x00001000
	SpellAttributeNightOnly           SpellAttributes = 0x00002000
	SpellAttributeIndoorOnly          SpellAttributes = 0x00004000
	SpellAttributeOutdoorOnly         SpellAttributes = 0x00008000
	SpellAttributeNotShapeshift       SpellAttributes = 0x00010000
	SpellAttributeStealthed           SpellAttributes = 0x00020000
	SpellAttributeDoNotStealth        SpellAttributes = 0x00040000
	SpellAttributeLevelDamage         SpellAttributes = 0x00080000
	SpellAttributeStopAttack          SpellAttributes = 0x00100000
	SpellAttributeAllowDead           SpellAttributes = 0x00800000
	SpellAttributeAllowMounted        SpellAttributes = 0x01000000
	SpellAttributeDisabledWhileActive SpellAttributes = 0x02000000
	SpellAttributeAuraDebuff          SpellAttributes = 0x04000000
	SpellAttributeAllowSitting        SpellAttributes = 0x08000000
	SpellAttributeCantCombat          SpellAttributes = 0x10000000
	SpellAttributeCantCancel          SpellAttributes = 0x80000000
)

type SpellAttributesEx uint32

const (
	SpellAttributeExDrainAllPower      SpellAttributesEx = 0x00000002
	SpellAttributeExChanneled          SpellAttributesEx = 0x00000004
	SpellAttributeExNoSkillIncrease    SpellAttributesEx = 0x00000010
	SpellAttributeExNotBreakStealth    SpellAttributesEx = 0x00000020
	SpellAttributeExNegative           SpellAttributesEx = 0x00000080
	SpellAttributeExNoThreat           SpellAttributesEx = 0x00000400
	SpellAttributeExUniqueAura         SpellAttributesEx = 0x00000800
	SpellAttributeExCantTargetSelf     SpellAttributesEx = 0x00080000
	SpellAttributeExRequireTargetCombo SpellAttributesEx = 0x00100000
	SpellAttributeExRequireCombo       SpellAttributesEx = 0x00400000
	SpellAttributeExCastWhenLearned    SpellAttributesEx = 0x80000000
)

type SpellCastFlags byte

const (
	SpellCastFlagNone    SpellCastFlags = 0
	SpellCastFlagProc    SpellCastFlags = 0x01
	SpellCastFlagArea    SpellCastFlags = 0x08
	SpellCastFlagHasAmmo SpellCastFlags = 0x10
)

const (
	SpellInterruptMovement     int64 = 0x01
	SpellAuraInterruptMovement int64 = 0x08
	SpellAuraInterruptTurning  int64 = 0x10
)

const (
	AuraFlagCancelable AuraFlags = 0x01
	AuraFlagEffect2    AuraFlags = 0x02
	AuraFlagEffect1    AuraFlags = 0x04
	AuraFlagEffect0    AuraFlags = 0x08
)

type AuraFlags byte

type SpellEffect int64

const (
	SpellEffectNone               SpellEffect = 0
	SpellEffectInstantKill        SpellEffect = 1
	SpellEffectSchoolDamage       SpellEffect = 2
	SpellEffectDummy              SpellEffect = 3
	SpellEffectDodge              SpellEffect = 20
	SpellEffectParry              SpellEffect = 22
	SpellEffectBlock              SpellEffect = 23
	SpellEffectWeapon             SpellEffect = 25
	SpellEffectDefense            SpellEffect = 26
	SpellEffectPersistentAreaAura SpellEffect = 27
	SpellEffectSkillStep          SpellEffect = 44
	SpellEffectTeleportUnits      SpellEffect = 5
	SpellEffectApplyAura          SpellEffect = 6
	SpellEffectPowerDrain         SpellEffect = 8
	SpellEffectHealthLeech        SpellEffect = 9
	SpellEffectHeal               SpellEffect = 10
	SpellEffectBind               SpellEffect = 11
	SpellEffectQuestComplete      SpellEffect = 16
	SpellEffectWeaponDamage       SpellEffect = 17
	SpellEffectWeaponDamagePlus   SpellEffect = 58
	SpellEffectResurrect          SpellEffect = 18
	SpellEffectCreateItem         SpellEffect = 24
	SpellEffectSummon             SpellEffect = 28
	SpellEffectSummonWild         SpellEffect = 41
	SpellEffectSummonGuardian     SpellEffect = 42
	SpellEffectEnergize           SpellEffect = 30
	SpellEffectOpenLock           SpellEffect = 33
	SpellEffectSummonMount        SpellEffect = 34
	SpellEffectApplyAreaAura      SpellEffect = 35
	SpellEffectLearnSpell         SpellEffect = 36
	SpellEffectDispel             SpellEffect = 38
	SpellEffectLanguage           SpellEffect = 39
	SpellEffectDualWield          SpellEffect = 40
	SpellEffectLearnPetSpell      SpellEffect = 57
	SpellEffectOpenLockItem       SpellEffect = 59
	SpellEffectProficiency        SpellEffect = 60
	SpellEffectSendEvent          SpellEffect = 61
	SpellEffectThreat             SpellEffect = 63
	SpellEffectStealth            SpellEffect = 48
	SpellEffectDetect             SpellEffect = 49
	SpellEffectSummonObject       SpellEffect = 50
	SpellEffectSummonObjectWild   SpellEffect = 76
	SpellEffectCreateHouse        SpellEffect = 81
	SpellEffectEnchantPermanent   SpellEffect = 53
	SpellEffectEnchantTemporary   SpellEffect = 54
	SpellEffectTameCreature       SpellEffect = 55
	SpellEffectSummonPet          SpellEffect = 56
	SpellEffectPowerBurn          SpellEffect = 62
	SpellEffectTriggerSpell       SpellEffect = 64
	SpellEffectHealMaxHealth      SpellEffect = 67
	SpellEffectInterruptCast      SpellEffect = 68
	SpellEffectPickpocket         SpellEffect = 71
	SpellEffectAddFarsight        SpellEffect = 72
	SpellEffectSummonPossessed    SpellEffect = 73
	SpellEffectSummonTotem        SpellEffect = 74
	SpellEffectScript             SpellEffect = 77
	SpellEffectDistract           SpellEffect = 69
	SpellEffectPull               SpellEffect = 70
	SpellEffectAttack             SpellEffect = 78
	SpellEffectSanctuary          SpellEffect = 79
	SpellEffectAddComboPoints     SpellEffect = 80
	SpellEffectDuel               SpellEffect = 83
	SpellEffectStuck              SpellEffect = 84
	SpellEffectSummonPlayer       SpellEffect = 85
	SpellEffectActivateObject     SpellEffect = 86
)

type AuraType int64

const (
	AuraModStun                 AuraType = 12
	AuraModConfuse              AuraType = 5
	AuraModFear                 AuraType = 7
	AuraModStealth              AuraType = 16
	AuraModPacify               AuraType = 25
	AuraModSilence              AuraType = 27
	AuraModIncreaseHealth       AuraType = 34
	AuraModIncreaseMana         AuraType = 35
	AuraModRoot                 AuraType = 26
	AuraModIncreaseMountedSpeed AuraType = 32
	AuraModSchoolImmunity       AuraType = 39
	AuraModDamageImmunity       AuraType = 40
	AuraModDisarm               AuraType = 67
	AuraPeriodicDamage          AuraType = 3
	AuraPeriodicHeal            AuraType = 8
	AuraPeriodicTriggerSpell    AuraType = 23
	AuraPeriodicEnergize        AuraType = 24
	AuraPeriodicLeech           AuraType = 53
	AuraPeriodicManaFunnel      AuraType = 63
	AuraPeriodicManaLeech       AuraType = 64
	AuraModMounted              AuraType = 78
)

type SpellImplicitTarget int64

const (
	SpellImplicitInitial            SpellImplicitTarget = 0
	SpellImplicitSelf               SpellImplicitTarget = 1
	SpellImplicitAroundCasterParty  SpellImplicitTarget = 20
	SpellImplicitAllAroundCaster    SpellImplicitTarget = 22
	SpellImplicitAllEnemyInArea     SpellImplicitTarget = 15
	SpellImplicitAllEnemyInstant    SpellImplicitTarget = 16
	SpellImplicitAllFriendlyAround  SpellImplicitTarget = 30
	SpellImplicitAllFriendlyInArea  SpellImplicitTarget = 31
	SpellImplicitAllParty           SpellImplicitTarget = 33
	SpellImplicitUnitNearCaster     SpellImplicitTarget = 4
	SpellImplicitPet                SpellImplicitTarget = 5
	SpellImplicitEnemyUnit          SpellImplicitTarget = 6
	SpellImplicitAreaEnemy          SpellImplicitTarget = 15
	SpellImplicitAreaEnemyInstant   SpellImplicitTarget = 16
	SpellImplicitSelectedFriend     SpellImplicitTarget = 21
	SpellImplicitSelectedGameObject SpellImplicitTarget = 23
	SpellImplicitInFront            SpellImplicitTarget = 24
	SpellImplicitUnit               SpellImplicitTarget = 25
	SpellImplicitGameObjectItem     SpellImplicitTarget = 26
	SpellImplicitMaster             SpellImplicitTarget = 27
	SpellImplicitPartyAroundCaster  SpellImplicitTarget = 20
	SpellImplicitHostileSelection   SpellImplicitTarget = 36
	SpellImplicitSelfFishing        SpellImplicitTarget = 39
)
