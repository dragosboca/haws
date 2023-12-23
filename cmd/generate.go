package cmd

import (
	"fmt"
	"os"

	"github.com/dragosboca/haws/pkg/haws"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate configs",
		Long:  "Generate various config files and print them on the screen",

		Run: func(cmd *cobra.Command, args []string) {

			h := haws.New(dryRun,
				viper.GetString("prefix"),
				viper.GetString("region"),
				viper.GetString("zone_id"),
				viper.GetString("bucket_path"),
				viper.GetString("record"),
			)

			stacks := []string{"certificate", "bucket", "cloudfront", "user"}
			for _, stack := range stacks {
				if err := h.GetStackOutput(stack); err != nil {
					fmt.Printf("%v\n", err)
					os.Exit(1)
				}

			}
			h.GenerateHugoConfig(viper.GetString("region"), viper.GetString("bucket_path"))
		},
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
}
