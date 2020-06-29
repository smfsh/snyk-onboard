package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

var configKeys = []configItem{
	{
		name:   "ghUser",
		prompt: "GitHub Username",
	},
	{
		name:     "ghKey",
		prompt:   "GitHub API Token",
		secret:   true,
		validate: ghKeyValidate,
	},
}

type configItem struct {
	name     string
	prompt   string
	secret   bool
	validate func(string) error
}

func init() {
	fmt.Println("Config stuff...")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.WriteConfigAs("config.yaml")
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
}

func checkForConfigValues() error {
	for _, k := range configKeys {
		if v := viper.Get(k.name); v != nil {
			fmt.Printf("Key \"%v\" already configured, skipping\n", k.name)
		} else {
			prompt := promptui.Prompt{
				Label: k.prompt,
			}
			if k.secret == true {
				prompt.Mask = '*'
			}
			if k.validate != nil {
				prompt.Validate = k.validate
			}
			result, err := prompt.Run()
			if err != nil {
				return err
			}

			viper.Set(k.name, result)
			err = viper.WriteConfig()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ghKeyValidate(key string) error {
	if len(key) != 40 {
		return errors.New("GitHub Tokens must be 40 characters")
	}
	return nil
}
