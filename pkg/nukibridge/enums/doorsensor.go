package enums

type DoorSensor uint8

const (
	DoorSensorDoorOpened   DoorSensor = 0x00
	DoorSensorDoorClosed   DoorSensor = 0x01
	DoorSensorSensorJammed DoorSensor = 0x02
)
