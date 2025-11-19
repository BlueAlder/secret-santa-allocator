package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/BlueAlder/secret-santa-allocator/pkg/allocator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	outputFile   string
	outputFormat string
)

var rootCmd = &cobra.Command{
	Use:   "secret-santa-allocator",
	Short: "A tool to allocate secret santas",
	Long:  `A tool to allocate secret santas with support for exclusion rules and history.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAllocation()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.secret-santa-allocator.yaml)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "file to write allocation to")
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "format to write output file to (json/yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".secret-santa-allocator")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runAllocation() {
	// Logic adapted from original main.go
	// We need to load the config from the file specified or viper
	// For now, let's assume the config file passed to viper contains the allocation config
	// OR we keep the original behavior of passing a specific config file for the allocation data.
	
	// The original main.go took a config file for the allocation rules/names.
	// Viper might be used for app-level config, but here the "config" IS the allocation data.
	// So we can use viper to unmarshal it.

	var config allocator.Config
	if err := viper.Unmarshal(&config); err != nil {
		slog.Error("Unable to decode into struct", "error", err)
		os.Exit(1)
	}

	// Validate config manually if needed, or trust allocator
	// The allocator.LoadConfigFromYaml did validation.
	// We might need to call validate.
	
	// Since we are using viper, we might have already read the config.
	// But wait, the original main.go read a specific file.
	// If cfgFile is set, viper read it.
	
	// We need to handle the case where no config is provided.
	if viper.ConfigFileUsed() == "" {
		slog.Error("No configuration file provided")
		os.Exit(1)
	}

	a, err := allocator.NewFromConfig(&config)
	if err != nil {
		slog.Error("Error creating allocator", "error", err)
		os.Exit(1)
	}

	fmt.Println("trying to find a suitable allocation...")
	alcc, err := a.Allocate()
	if err != nil {
		slog.Error("Unable to allocate names", "error", err)
		os.Exit(1)
	}
	fmt.Println("found a suitable allocation! ✅")

	if outputFile != "" {
		fmt.Printf("printing resulting allocation to file: %s in %s format\n", outputFile, outputFormat)
		err := a.OutputToFile(alcc, outputFile, outputFormat)
		if err != nil {
			slog.Error("Error writing to file", "error", err)
		} else {
			fmt.Printf("successfully wrote to %s!\n", outputFile)
		}
	} else {
		fmt.Println()
		alcc.PrintAliases()
		fmt.Println()
		alcc.PrintNameToName()
		fmt.Println()
		alcc.PrintNameToPassword()
	}
}
