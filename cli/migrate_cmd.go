package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/ugent-library/muts/migrate"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:       "migrate [up|down]",
	Short:     "Run database migrations",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"up", "down"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "up":
			return migrate.Up(context.Background(), config.Store.Conn)
		case "down":
			return migrate.Down(context.Background(), config.Store.Conn)
		}
		return nil
	},
}
