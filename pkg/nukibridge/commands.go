package nukibridge

type Command uint16

const (
	CmdRequestData                 Command = 0x0001
	CmdPublicKey                   Command = 0x0003
	CmdChallenge                   Command = 0x0004
	CmdAuthorizationAuthenticator  Command = 0x0005
	CmdAuthorizationData           Command = 0x0006
	CmdAuthorizationID             Command = 0x0007
	CmdRemoveUserAuthorization     Command = 0x0008
	CmdRequestAuthorizationEntries Command = 0x0009
	CmdRequestConfig               Command = 0x0014
	CmdConfig                      Command = 0x0015
	CmdKeyturnerStates             Command = 0x000C
	CmdLockAction                  Command = 0x000D
	CmdStatus                      Command = 0x000E
	CmdErrorReport                 Command = 0x0012
	CmdAuthorizationIDConfirmation Command = 0x001E
	CmdRequestLogEntries           Command = 0x0031
	CmdLogEntry                    Command = 0x0032
	// ...
)
