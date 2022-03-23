package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	prefix string
	region string
	dryRun bool

	record string
	zoneId string
	path   string

	rootCmd = &cobra.Command{
		Use:   "haws",
		Short: "Hugo on AWS",
		Long:  "A cloudformation and template generator for running Hugo on AWS",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&prefix, "prefix", "", "Prefix for resources created. Can not be empty")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .haws.toml in current directory)")
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "AWS region for the bucket and cloudfront distribution")

	rootCmd.PersistentFlags().StringVar(&record, "record", "", "Record name to be added to R53 zone")
	rootCmd.PersistentFlags().StringVar(&zoneId, "zone-id", "", "AWS Id of the zone used for SSL certificate validation and where the record should be added")
	rootCmd.PersistentFlags().StringVar(&path, "bucket-path", "", "Path prefix that will be appended by cloudfront to all requests (it should correspond to a sub-folder in the bucket)")

	if err := viper.BindPFlag("prefix", rootCmd.PersistentFlags().Lookup("prefix")); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region")); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("record", rootCmd.PersistentFlags().Lookup("record")); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("zone_id", rootCmd.PersistentFlags().Lookup("zone-id")); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("bucket_path", rootCmd.PersistentFlags().Lookup("bucket-path")); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		viper.AddConfigPath(cwd)
		viper.SetConfigType("toml")
		viper.SetConfigName(".haws")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			fmt.Printf("Fatal error config file: %v\n", err)
			os.Exit(1)
		}
	}
}
