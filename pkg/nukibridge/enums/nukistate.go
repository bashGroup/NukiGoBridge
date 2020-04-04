package enums

type NukiState uint8

const (
	NukiStateUninitalized    NukiState = 0x00
	NukiStatePairingMode     NukiState = 0x01
	NukiStateDoorMode        NukiState = 0x02
	NukiStateMaintenanceMode NukiState = 0x04
)
