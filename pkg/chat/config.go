package chat

type ServerConfig struct {
	Port       int
	MaxClients int `json:"max-clients"`
	Database   DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Username string
	Password string
	Name     string
}
