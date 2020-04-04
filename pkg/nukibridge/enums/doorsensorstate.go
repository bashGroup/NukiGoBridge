package enums

type DoorSensorState uint8

const (
	DoorSensorStateUnavailable      DoorSensorState = 0x00
	DoorSensorStateDeactivated      DoorSensorState = 0x01
	DoorSensorStateDoorClosed       DoorSensorState = 0x02
	DoorSensorStateDoorOpened       DoorSensorState = 0x03
	DoorSensorStateDoorStateUnknown DoorSensorState = 0x04
	DoorSensorStateCalibrating      DoorSensorState = 0x05
)
