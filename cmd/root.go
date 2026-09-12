package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"planeshift/cmd/get"
	"planeshift/common"
	"planeshift/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "gopkg.in/yaml.v3"
	// "sigs.k8s.io/yaml"
)

type (
	Project struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Path   string `json:"path"`
		SSHURL string `json:"ssh_url_to_repo"`
	}
)

var (
	cfgFile   string
	semVer    string
	gitCommit string
	gitRef    string
	buildDate string

	// semVerReg - gets the semVer portion only, cutting off any other release details
	semVerReg = regexp.MustCompile(`(v[0-9]+\.[0-9]+\.[0-9]+).*`)

	c = &config.Config{}
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "planeshift",
	Short: c.CmdHelp.RootCmd.Short(),
	Long:  c.CmdHelp.RootCmd.Long(),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		logFile, _ := cmd.Flags().GetString("log-file")
		logLevel, _ := cmd.Flags().GetString("log-level")
		ll := "Warning"
		switch strings.ToLower(logLevel) {
		case "trace":
			ll = "Trace"
		case "debug":
			ll = "Debug"
		case "info":
			ll = "Info"
		case "warning":
			ll = "Warning"
		case "error":
			ll = "Error"
		case "fatal":
			ll = "Fatal"
		}

		common.NewLogger(ll, logFile)

		c.VersionDetail.SemVer = semVer
		c.VersionDetail.BuildDate = buildDate
		c.VersionDetail.GitCommit = gitCommit
		c.VersionDetail.GitRef = gitRef
		c.VersionJSON = fmt.Sprintf("{\"SemVer\": \"%s\", \"BuildDate\": \"%s\", \"GitCommit\": \"%s\", \"GitRef\": \"%s\"}", semVer, buildDate, gitCommit, gitRef)
		if c.OutputFormat != "" {
			c.FormatOverridden = true
			c.NoHeaders = false
			c.OutputFormat = strings.ToLower(c.OutputFormat)
			switch c.OutputFormat {
			case "json", "gron", "yaml", "text", "table", "raw":
				break
			default:
				fmt.Println("Valid options for -o are [json|gron|text|table|yaml|raw]")
				os.Exit(1)
			}
		}

		// if os.Args[1] != "version" {
		// }
	},
}

func buildRootCmd() *cobra.Command {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.splicectl/config.yml)")
	RootCmd.PersistentFlags().StringVarP(&c.OutputFormat, "output", "o", "", "output types: json, text, yaml, gron, raw")
	RootCmd.PersistentFlags().BoolVar(&c.NoHeaders, "no-headers", false, "Suppress header output in Text output")
	// RootCmd.PersistentFlags().BoolVar(&c., "no-headers", false, "Suppress header output in Text output")
	RootCmd.PersistentFlags().StringVarP(&c.LogLevel, "log-level", "v", "", "Set the logging level: trace,debug,info,warning,error,fatal")
	RootCmd.PersistentFlags().StringVar(&c.LogFile, "log-file", "", "Set the logging level: trace,debug,info,warning,error,fatal")

	return RootCmd
}

func addSubCommands() {
	RootCmd.AddCommand(
		// from 'import planeshift/cmd/<subcommand:package>'
		// <package>.InitSubCommands(c),
		get.InitSubCommands(c),
	)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	buildRootCmd()
	cobra.OnInitialize(initConfig)
	addSubCommands()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	home, herr := os.UserHomeDir()
	cobra.CheckErr(herr)
	configHome := fmt.Sprintf("%s/.config", home)
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		configHome = os.Getenv("XDG_CONFIG_HOME")
	}
	confDir := fmt.Sprintf("%s/planeshift", configHome)
	envCfgFile := os.Getenv("PLANESHIFT_CONFIG")
	if envCfgFile != "" {
		logrus.Debug("Using PLANESHIFT_CONFIG")
		configFile := fmt.Sprintf("%s/%s", confDir, envCfgFile)
		created := createRestrictedConfigFile(configFile)
		viper.SetConfigFile(configFile)
		if created {
			setConfigFileVersion()
		}
	} else {
		if cfgFile != "" {
			// Use config file from the flag.
			viper.SetConfigFile(cfgFile)
		} else {
			// Find home directory.
			home, err := os.UserHomeDir()
			cobra.CheckErr(err)

			workDir := fmt.Sprintf("%s/.config/planeshift", home)
			if _, err := os.Stat(workDir); err != nil {
				if os.IsNotExist(err) {
					mkerr := os.MkdirAll(workDir, os.ModePerm)
					if mkerr != nil {
						logrus.Fatal("Error creating ~/.config/planeshift directory", mkerr)
					}
				}
			}
			if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
				configFile := fmt.Sprintf("%s/%s", workDir, "config.yaml")
				createRestrictedConfigFile(configFile)
				viper.SetConfigFile(configFile)
			} else {
				logrus.Info("The ~/.config/planeshift path is a file and not a directory, please remove the 'planeshift' file.")
				os.Exit(1)
			}
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		logrus.Warn("Failed to read viper config file.")
	}
}

// returns true if the file was created, false if it already exists
func createRestrictedConfigFile(fileName string) bool {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			file, ferr := os.Create(fileName)
			if ferr != nil {
				logrus.Fatal("Unable to create the configfile.")
			}
			mode := int(0600)
			if cherr := file.Chmod(os.FileMode(mode)); cherr != nil {
				logrus.Warn("Chmod for config file failed, please set the mode to 0600.")
			}
			return true
		}
	}
	return false
}

func setConfigFileVersion() {
	verParts, verr := config.ParseSemver(semVer)
	if verr != nil {
		logrus.WithError(verr).Fatal("Failed to parse the semver")
	}
	viper.Set("configVersion", fmt.Sprintf("v%d", verParts.Major))
	c.ConfigVersion = fmt.Sprintf("v%d", verParts.Major)
	werr := viper.WriteConfig()
	if werr != nil {
		logrus.WithError(werr).Info("Failed to write config")
	}
}

// ClientSemVer - returns the full semVer as the first string and the numerical
// portion as the second string, they may be identical. One example where they
// would not be is:
//
//	semVer: v0.1.1-cacert -> (v0.1.1-cacert, v0.1.1).
func ClientSemVer() (string, string) {
	submatches := semVerReg.FindStringSubmatch(semVer)
	if submatches == nil || len(submatches) < 2 {
		logrus.Fatalf("the semver in the current build is not valid: %s", semVer)
	}
	return submatches[0], submatches[1]
}
