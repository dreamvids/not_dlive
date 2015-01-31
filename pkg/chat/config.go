package chat

type ServerConfig struct {
	Port        int
	MaxClients  int    `json:"max-clients"`
	ModoRank    int    `json:"modo-rank"`
	AdminRank   int    `json:"admin-rank"`
	MuteMessage string `json:"mute-message"`
	Database    DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Name     string
}
