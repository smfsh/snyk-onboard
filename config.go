// Initialize configuration used for the app. Most config
// is set during init() which is run before main().

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

// Each configuration item used over the course of the
// app's processes must be defined here early. Masking
// and validation functions can be included.
var configKeys = []configItem{
	{
		Name:     "ghUser",
		Prompt:   "GitHub Username",
		Validate: nonEmptyValidate,
	},
	{
		Name:   "ghOrg",
		Prompt: "GitHub Organization (optional)",
	},
	{
		Name:     "ghKey",
		Prompt:   "GitHub API Token",
		Secret:   true,
		Validate: ghKeyValidate,
	},
	{
		Name:     "glUser",
		Prompt:   "GitLab Username",
		Validate: nonEmptyValidate,
	},
	{
		Name:     "glKey",
		Prompt:   "GitLab API Token",
		Secret:   true,
		Validate: glKeyValidate,
	},
	{
		Name:     "bbUser",
		Prompt:   "Bitbucket Username",
		Validate: nonEmptyValidate,
	},
	{
		Name:     "bbKey",
		Prompt:   "Bitbucket API Token",
		Secret:   true,
		Validate: bbKeyValidate,
	},
	{
		Name:     "azOrg",
		Prompt:   "Azure DevOps Organization",
		Validate: nonEmptyValidate,
	},
	{
		Name:     "azKey",
		Prompt:   "Azure DevOps API Token",
		Secret:   true,
		Validate: azKeyValidate,
	},
}

// A struct representing a single piece of configuration.
// Name and prompt are typically the minimum fields
// required. Default, secret, and validate are optional.
type configItem struct {
	Name     string
	Prompt   string
	Default  string
	Secret   bool
	Validate func(string) error
}

// Global variable used to determine running scope.
var inDocker bool

// Init() function, called before anything else in the
// application. Used for initializing configuration and
// runtime parameters.
func init() {
	// Check for whether the app is inside a container.
	inDocker = checkForDocker()

	fmt.Println("Initializing configuration...")
	viper.SetConfigName(".config.yaml")
	viper.SetConfigType("yaml")

	// Set repo path hard ("/repos") if in a container.
	var repos = "repos"
	if inDocker {
		repos = "/repos"
	}
	viper.AddConfigPath(repos)

	// Attempt to read in previous configuration. If
	// config doesn't exist, create initial config
	// file and set permissions appropriately.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			out := filepath.Join(repos, ".config.yaml")
			err = viper.WriteConfigAs(out)
			if err != nil {
				log.Panic(err)
			}
			err = os.Chmod(out, 0700)
			if err != nil {
				log.Panic(err)
			}
		} else {
			log.Panic(err)
		}
	}
	// Place all config values into memory for later access.
	err := checkForConfigValues()
	if err != nil {
		log.Panic(err)
	}
	// Add a final config item into memory outside of the
	// normal access scope, ensuring we do not write this
	// to the config file. This allows the application to
	// transparently switch between container and direct
	// execution without resetting the path.
	viper.Set("path", repos)
}

// Returns boolean true or false if running in a
// Docker container or not. This is not a comprehensive
// check for all container types.
func checkForDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// Loop through each config value needed. Prompt for the value
// or skip if it is already in the config file.
func checkForConfigValues() error {
	for _, k := range configKeys {
		if v := viper.Get(k.Name); v != nil { // We have config already.
			fmt.Printf("Key \"%v\" already configured, skipping\n", k.Name)
		} else { // We don't have the config already, prompt for it.
			prompt := promptui.Prompt{
				Label: k.Prompt,
			}
			if k.Default != "" {
				prompt.Default = k.Default
			}
			if k.Secret == true {
				prompt.Mask = '*'
			}
			if k.Validate != nil {
				prompt.Validate = k.Validate
			}
			result, err := prompt.Run()
			if err != nil {
				return err
			}

			// Set the config value into memory and write out
			// to the config file.
			viper.Set(k.Name, result)
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Validation function to determine if the key is empty.
func nonEmptyValidate(key string) error {
	if key == "" {
		return errors.New("Cannot be blank")
	}
	return nil
}

// Validation function to determine, based on length, if the
// key appears to be a valid GitHub credential.
func ghKeyValidate(key string) error {
	if len(key) != 40 {
		return errors.New("GitHub Tokens must be 40 characters")
	}
	return nil
}

// Validation function to determine, based on length, if the
// key appears to be a valid GitLab credential.
func glKeyValidate(key string) error {
	if len(key) != 20 {
		return errors.New("GitLab Tokens must be 20 characters")
	}
	return nil
}

// Validation function to determine, based on length, if the
// key appears to be a valid Bitbucket credential.
func bbKeyValidate(key string) error {
	if len(key) != 20 {
		return errors.New("Bitbucket Tokens must be 20 characters")
	}
	return nil
}

// Validation function to determine, based on length, if the
// key appears to be a valid Azure credential.
func azKeyValidate(key string) error {
	if len(key) != 52 {
		return errors.New("Azure DevOps Tokens must be 52 characters")
	}
	return nil
}
