package enums

type LockState uint8

const (
	LockStateUncalibrated  LockState = 0x00
	LockStateLocked        LockState = 0x01
	LockStateUnlocking     LockState = 0x02
	LockStateUnlocked      LockState = 0x03
	LockStateLocking       LockState = 0x04
	LockStateUnlatched     LockState = 0x05
	LockStateLocknGoActive LockState = 0x06
	LockStateUnlatching    LockState = 0x07
	LockStateCalibration   LockState = 0xFC
	LockStateBootRun       LockState = 0xFD
	LockStateMotorBlocked  LockState = 0xFE
	LockStateUndefined     LockState = 0xFF
)
