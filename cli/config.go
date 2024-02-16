package cli

//go:generate go run github.com/g4s8/envdoc@v0.1.2 --output ../CONFIG.md --all

// Application configuration
type Config struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT" envDefault:"3000"`
	// Store configuration
	Store struct {
		// Database connection string
		Conn string `env:"CONN,notEmpty"`
	} `envPrefix:"STORE_"`
}
