package enums

type LockAction uint8

const (
	LockActionUnlock         LockAction = 0x01
	LockActionLock           LockAction = 0x02
	LockActionUnlatch        LockAction = 0x03
	LockActionLocknGo        LockAction = 0x04
	LockActionLocknGoUnlatch LockAction = 0x05
	LockActionFullLock       LockAction = 0x06
	LockActionFobAction1     LockAction = 0x81
	LockActionFobAction2     LockAction = 0x82
	LockActionFobAction3     LockAction = 0x83
)
