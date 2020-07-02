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

type configItem struct {
	Name     string
	Prompt   string
	Default  string
	Secret   bool
	Validate func(string) error
}

var inDocker bool

func init() {
	inDocker = checkForDocker()

	fmt.Println("Config stuff...")
	viper.SetConfigName(".config.yaml")
	viper.SetConfigType("yaml")

	var repos = "repos"
	if inDocker {
		repos = "/repos"
	}
	viper.AddConfigPath(repos)
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
	err := checkForConfigValues()
	if err != nil {
		log.Panic(err)
	}
	viper.Set("path", repos)
}

func checkForDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func checkForConfigValues() error {
	for _, k := range configKeys {
		if v := viper.Get(k.Name); v != nil {
			fmt.Printf("Key \"%v\" already configured, skipping\n", k.Name)
		} else {
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

			viper.Set(k.Name, result)
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func nonEmptyValidate(key string) error {
	if key == "" {
		return errors.New("Cannot be blank")
	}
	return nil
}

func ghKeyValidate(key string) error {
	if len(key) != 40 {
		return errors.New("GitHub Tokens must be 40 characters")
	}
	return nil
}

func glKeyValidate(key string) error {
	if len(key) != 20 {
		return errors.New("GitLab Tokens must be 20 characters")
	}
	return nil
}

func bbKeyValidate(key string) error {
	if len(key) != 20 {
		return errors.New("Bitbucket Tokens must be 20 characters")
	}
	return nil
}

func azKeyValidate(key string) error {
	if len(key) != 52 {
		return errors.New("Azure DevOps Tokens must be 52 characters")
	}
	return nil
}
