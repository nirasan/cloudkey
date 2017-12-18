package main

import (
	"bytes"
	"errors"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type Config struct {
	GcpConfig GcpConfig
	AwsConfig AwsConfig
}

type GcpConfig struct {
	UseGcloudAccount  bool
	UseServiceAccount bool
	Project           string
	Location          string
	Keyring           string
	Cryptokey         string
	ServiceAccountKey string
}

type AwsConfig struct {
	UseStaticCreds  bool
	UseSharedCreds  bool
	AccessKeyID     string
	SecretAccessKey string
	AccessToken     string
	Region          string
	CryptoKeyID     string
	CredFile        string
	Profile         string
}

func LoadConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil {
		return &Config{}, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := &Config{}
	if _, err := toml.Decode(string(data), conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func SaveConfig(path string, conf *Config) error {
	var b bytes.Buffer
	e := toml.NewEncoder(&b)
	err := e.Encode(conf)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, b.Bytes(), 0600); err != nil {
		return err
	}
	return nil
}

func createGCPConfigForGcloudAccount(config, project, location, keyring, cryptokey string) error {
	if config == "" || project == "" || location == "" || keyring == "" || cryptokey == "" {
		return errors.New("invalid args")
	}

	c, err := LoadConfig(config)
	if err != nil {
		return err
	}

	c.GcpConfig = GcpConfig{
		UseGcloudAccount: true,
		Project:          project,
		Location:         location,
		Keyring:          keyring,
		Cryptokey:        cryptokey,
	}

	if err := SaveConfig(config, c); err != nil {
		return err
	}

	return nil
}

func createGCPConfigForServiceAccount(config, project, location, keyring, cryptokey, serviceAccountKey string) error {
	if config == "" || project == "" || location == "" || keyring == "" || cryptokey == "" || serviceAccountKey == "" {
		return errors.New("invalid args")
	}

	c, err := LoadConfig(config)
	if err != nil {
		return err
	}

	c.GcpConfig = GcpConfig{
		UseServiceAccount: true,
		Project:           project,
		Location:          location,
		Keyring:           keyring,
		Cryptokey:         cryptokey,
		ServiceAccountKey: serviceAccountKey,
	}

	if err := SaveConfig(config, c); err != nil {
		return err
	}

	return nil
}

func createAWSConfigForStaticCredentials(config, accessKeyID, secretAccessKey, accessToken, region, cryptoKeyID string) error {
	if config == "" || (accessKeyID == "" && secretAccessKey == "" && accessToken == "") || region == "" || cryptoKeyID == "" {
		return errors.New("invalid args")
	}

	c, err := LoadConfig(config)
	if err != nil {
		return err
	}

	c.AwsConfig = AwsConfig{
		UseStaticCreds:  true,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		AccessToken:     accessToken,
		Region:          region,
		CryptoKeyID:     cryptoKeyID,
	}

	if err := SaveConfig(config, c); err != nil {
		return err
	}

	return nil
}

func createAWSConfigForSharedCredentials(config, credFile, profile, region, cryptoKeyID string) error {
	if config == "" || region == "" || cryptoKeyID == "" {
		return errors.New("invalid args")
	}

	c, err := LoadConfig(config)
	if err != nil {
		return err
	}

	c.AwsConfig = AwsConfig{
		UseSharedCreds: true,
		CredFile:       credFile,
		Profile:        profile,
		Region:         region,
		CryptoKeyID:    cryptoKeyID,
	}

	if err := SaveConfig(config, c); err != nil {
		return err
	}

	return nil
}
