package enums

type CompletionStatus uint8

const (
	CompletionStatusSuccess           CompletionStatus = 0x00
	CompletionStatusMotorBlocked      CompletionStatus = 0x01
	CompletionStatusCanceled          CompletionStatus = 0x02
	CompletionStatusTooRecent         CompletionStatus = 0x03
	CompletionStatusBusy              CompletionStatus = 0x04
	CompletionStatusLowMotorVoltage   CompletionStatus = 0x05
	CompletionStatusClutchFailure     CompletionStatus = 0x06
	CompletionStatusMotorPowerFailure CompletionStatus = 0x07
	CompletionStatusIncompleteFailure CompletionStatus = 0x08
	CompletionStatusOtherError        CompletionStatus = 0xFE
	CompletionStatusUnknown           CompletionStatus = 0xFF
)
