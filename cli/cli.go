package cli

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v8"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
)

var (
	// version Version
	config Config
	logger *slog.Logger

	rootCmd = &cobra.Command{
		Use:   "muts",
		Short: "muts CLI",
	}
)

func init() {
	cobra.OnInitialize(initVersion, initConfig, initLogger)
}

func initConfig() {
	cobra.CheckErr(env.ParseWithOptions(&config, env.Options{Prefix: "MUTS_"}))
}

func initVersion() {
	// cobra.CheckErr(env.Parse(&version))
}

func initLogger() {
	// if config.Env == "local" {
	// 	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	// } else {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// }
}

func Run() error {
	return rootCmd.Execute()
}
