package config

type Config struct {
	//Port for the server to listen on
	Port string
	//Spotify client ID // Allowing the user to provide this in a config file removes the need to prompt the user for it
	ClientID string
	// Later then we might want to look into adding more configuration options
	// Especially incase we would like to run the hacky spotify reverse engineered server
}

func New() *Config {
	return &Config{
		Port:     ":8080",
		ClientID: "",
	}
}
