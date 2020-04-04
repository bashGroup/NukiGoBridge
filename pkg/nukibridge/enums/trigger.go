package enums

type Trigger uint8

const (
	TriggerSystem    Trigger = 0x00
	TriggerManual    Trigger = 0x01
	TriggerButton    Trigger = 0x02
	TriggerAutomatic Trigger = 0x03
	TriggerAutoLock  Trigger = 0x06
)
