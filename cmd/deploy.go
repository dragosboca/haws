package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dragosboca/haws/pkg/haws"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the cloudformation stacks",
		Long:  "Deploy all stacks",

		Run: func(cmd *cobra.Command, args []string) {
			h := haws.New(dryRun,
				viper.GetString("prefix"),
				viper.GetString("region"),
				viper.GetString("zone_id"),
				viper.GetString("bucket_path"),
				viper.GetString("record"),
			)

			if err := h.Deploy(); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	deployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate the actions")

	rootCmd.AddCommand(deployCmd)
}
