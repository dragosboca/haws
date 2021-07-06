package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/dragosboca/haws/pkg/haws"
	"github.com/spf13/cobra"
)

var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate configs",
		Long:  "Generate various config files and print them on the screen",

		Run: func(cmd *cobra.Command, args []string) {
			h = haws.New(
				viper.GetString("prefix"),
				viper.GetString("region"),
				viper.GetString("record"),
				viper.GetString("zone_id"),
				viper.GetString("bucket_path"),
				dryRun,
			)

			if err := h.GetStackOutput("certificate", haws.NewCertificate(&h)); err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("bucket", haws.NewBucket(&h)); err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("cloudfront", haws.NewCdn(&h)); err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			if err := h.GetStackOutput("user", haws.NewIamUser(&h)); err != nil {
				fmt.Printf("%s\n", err)
				os.Exit(1)
			}

			h.GenerateHugoConfig()
		},
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
}
