package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// Version is injected at build time
	Version = "dev"
	// BuildTime is injected at build time
	BuildTime = "unknown"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "veo3",
		Short: "Google Veo 3.1 Video Generation CLI",
		Long: `A comprehensive command-line utility for Google's Veo 3.1 video generation API,
enabling developers and creators to generate AI videos with native audio,
extend existing videos, use reference images, and perform frame interpolation from the terminal.`,
		Version: fmt.Sprintf("%s (built at %s)", Version, BuildTime),
	}

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/veo3/config.yaml)")
	cmd.PersistentFlags().String("api-key", "", "Google Gemini API key")
	cmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	cmd.PersistentFlags().Bool("verbose", false, "Enable debug logging")
	cmd.PersistentFlags().Bool("quiet", false, "Suppress progress output")

	viper.BindPFlag("api-key", cmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("json", cmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("verbose", cmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", cmd.PersistentFlags().Lookup("quiet"))

	// Add subcommands
	cmd.AddCommand(newGenerateCmd())
	cmd.AddCommand(newAnimateCmd())
	cmd.AddCommand(newInterpolateCmd())
	cmd.AddCommand(newExtendCmd())

	return cmd
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.config/veo3")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("VEO3")

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
