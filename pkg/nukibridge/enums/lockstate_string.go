// Code generated by "stringer -type LockState -trimprefix LockState"; DO NOT EDIT.

package enums

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LockStateUncalibrated-0]
	_ = x[LockStateLocked-1]
	_ = x[LockStateUnlocking-2]
	_ = x[LockStateUnlocked-3]
	_ = x[LockStateLocking-4]
	_ = x[LockStateUnlatched-5]
	_ = x[LockStateLocknGoActive-6]
	_ = x[LockStateUnlatching-7]
	_ = x[LockStateCalibration-252]
	_ = x[LockStateBootRun-253]
	_ = x[LockStateMotorBlocked-254]
	_ = x[LockStateUndefined-255]
}

const (
	_LockState_name_0 = "UncalibratedLockedUnlockingUnlockedLockingUnlatchedLocknGoActiveUnlatching"
	_LockState_name_1 = "CalibrationBootRunMotorBlockedUndefined"
)

var (
	_LockState_index_0 = [...]uint8{0, 12, 18, 27, 35, 42, 51, 64, 74}
	_LockState_index_1 = [...]uint8{0, 11, 18, 30, 39}
)

func (i LockState) String() string {
	switch {
	case i <= 7:
		return _LockState_name_0[_LockState_index_0[i]:_LockState_index_0[i+1]]
	case 252 <= i && i <= 255:
		i -= 252
		return _LockState_name_1[_LockState_index_1[i]:_LockState_index_1[i+1]]
	default:
		return "LockState(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
