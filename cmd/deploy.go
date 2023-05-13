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
			)

			if err := h.DeployStack("certificate", h.NewCertificate()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.DeployStack("bucket", h.NewBucket()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.DeployStack("cloudfront", h.NewCdn()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}

			if err := h.DeployStack("user", h.NewIamUser()); err != nil {
				fmt.Printf("%v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	deployCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Only Simulate the actions")

	rootCmd.AddCommand(deployCmd)
}
