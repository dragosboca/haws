package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/dragosboca/haws/pkg/haws"
	"github.com/spf13/cobra"
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the cloudformation stacks",
		Long:  "Deploy all stacks",

		Run: func(cmd *cobra.Command, args []string) {
			h := haws.New(
				viper.GetString("prefix"),
				viper.GetString("region"),
				viper.GetString("record"),
				viper.GetString("zone_id"),
				viper.GetString("bucket_path"),
				dryRun,
			).WithDefaults()

			err := h.Deploy()
			if err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
			h.GenerateHugoConfig()
		},
	}
)

func init() {
	deployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Only Simulate the actions")

	rootCmd.AddCommand(deployCmd)
}
