package api

import (
	api "code/tech-test/application"

	"github.com/spf13/cobra"
)

// Command creates cobra command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Start Users API",
		RunE:  Run(),
	}

	return cmd
}

func Run() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		api.SetupAPI()

		return nil
	}
}
