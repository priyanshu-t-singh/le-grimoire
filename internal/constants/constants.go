package constants

const (
	DefaultHost = "127.0.0.1"
	DefaultPort = 8321
)

// NOTE: Injected at build time using ldflags, default is `dev` for local development
var Version = "v1.2.1-dev"

const (
	KavitaAPIBaseURL = "http://kavita:5000"
)

// TODO: Make this configurable for each device type, and/or make it auto-detectable
// Display Configuration (4.2" e-paper = 400x300)
const (
	DisplayHeight = 300
	DisplayWidth  = 400
)

// $HOME/.config/le-grimoire/config.toml
const (
	ConfigFileName = "config.toml"
)
