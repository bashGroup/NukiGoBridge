package enums

type LogType uint8

const (
	LogTypeLogging           LogType = 0x01
	LogTypeLockAction        LogType = 0x02
	LogTypeCalibration       LogType = 0x03
	LogTypeInitializationRun LogType = 0x04
	LogTypeKeypadAction      LogType = 0x05
	LogTypeDoorSensor        LogType = 0x06
	LogTypeDoorSensorLogging LogType = 0x07
)
