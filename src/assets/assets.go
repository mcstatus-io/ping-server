package assets

import _ "embed"

var (
	//go:embed icon.png
	DefaultIcon []byte
	//go:embed favicon.ico
	Favicon []byte
)
