package system

import "ntduncan.com/typer/utils"

type Config struct {
	TopScore string         `json:"TopScore"`
	Mode     utils.TestMode `json:"Mode"`
	Size     int            `json:"Size"`
}
