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
			h := haws.New(
				viper.GetString("prefix"),
				viper.GetString("region"),
				viper.GetString("record"),
				viper.GetString("zone_id"),
				viper.GetString("bucket_path"),
				dryRun,
			)

			if err := h.GetStackOutput("certificate", h.NewCertificate()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("bucket", h.NewBucket()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("cloudfront", h.NewCdn()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("user", h.NewIamUser()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			h.GenerateHugoConfig()
		},
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
}
