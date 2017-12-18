package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

func main() {
	app := kingpin.New("cloudkey", "Encrypt and decrypt files with key in the cloud.")
	app.Version("0.0.1")

	config := app.Command("config", "Create configuration file").Alias("c")

	configGcp := config.Command("gcp", "Create configuration file for GCP").Alias("g")
	configGcpGcloudAccount(configGcp)
	configGcpServiceAccount(configGcp)

	configAws := config.Command("aws", "Create configuration file for AWS").Alias("a")
	configAwsStaticCreds(configAws)
	configAwsSharedCreds(configAws)

	encrypt := app.Command("encrypt", "Encrypt file").Alias("en")
	encryptGcp(encrypt)
	encryptAws(encrypt)

	decrypt := app.Command("decrypt", "Decrypt file").Alias("de")
	decryptGcp(decrypt)
	decryptAws(decrypt)

	reencrypt := app.Command("re-encrypt", "Re-encrypt file, if encrypted file exists").Alias("ren")
	reEncryptGcp(reencrypt)
	reEncryptAws(reencrypt)

	redecrypt := app.Command("re-decrypt", "Re-decrypt file, if encrypted file exists").Alias("rde")
	reDecryptGcp(redecrypt)
	reDecryptAws(redecrypt)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func configGcpGcloudAccount(parent *kingpin.CmdClause) {
	cmd := parent.Command("gcloud-account", "Create configuration file for GCP using gcloud account").Alias("g")
	config := cmd.Arg("config", "config file").Required().String()
	project := cmd.Flag("project", "project name").Short('p').Required().String()
	location := cmd.Flag("location", "keyring location").Short('l').Required().String()
	keyring := cmd.Flag("keyring", "keyring name").Short('r').Required().String()
	key := cmd.Flag("key", "key name").Short('k').Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		return createGCPConfigForGcloudAccount(*config, *project, *location, *keyring, *key)
	})
}

func configGcpServiceAccount(parent *kingpin.CmdClause) {
	cmd := parent.Command("service-account", "Create configuration file for GCP using service account").Alias("g")
	config := cmd.Arg("config", "config file").Required().String()
	project := cmd.Flag("project", "project name").Short('p').Required().String()
	location := cmd.Flag("location", "keyring location").Short('l').Required().String()
	keyring := cmd.Flag("keyring", "keyring name").Short('r').Required().String()
	key := cmd.Flag("key", "key name").Short('k').Required().String()
	serviceAccountKey := cmd.Flag("service-account-key", "service account json key name").Short('s').String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		return createGCPConfigForServiceAccount(*config, *project, *location, *keyring, *key, *serviceAccountKey)
	})
}

func configAwsStaticCreds(parent *kingpin.CmdClause) {
	cmd := parent.Command("static-creds", "Create configuration file for AWS using static credentials").Alias("st")
	config := cmd.Arg("config", "config file").Required().String()
	accessKeyID := cmd.Flag("access-key-id", "access key id").Short('a').Required().String()
	secretAccessKey := cmd.Flag("secret-access-key", "secret access key").Short('s').Required().String()
	accessToken := cmd.Flag("access-token", "access token").Short('t').Required().String()
	region := cmd.Flag("region", "region name").Short('r').Required().String()
	key := cmd.Flag("key", "crypto key id").Short('k').Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		return createAWSConfigForStaticCredentials(*config, *accessKeyID, *secretAccessKey, *accessToken, *region, *key)
	})
}

func configAwsSharedCreds(parent *kingpin.CmdClause) {
	cmd := parent.Command("shared-creds", "Create configuration file for AWS using shared credentials").Alias("st")
	config := cmd.Arg("config", "config file").Required().String()
	file := cmd.Flag("cred-file", "credential file").Short('f').Required().String()
	profile := cmd.Flag("profile", "profile name").Short('p').Required().String()
	region := cmd.Flag("region", "region name").Short('r').Required().String()
	key := cmd.Flag("key", "crypto key id").Short('k').Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		return createAWSConfigForSharedCredentials(*config, *file, *profile, *region, *key)
	})
}

func encryptGcp(parent *kingpin.CmdClause) {
	cmd := parent.Command("gcp", "Encrypt file using GCP").Alias("g")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	file := cmd.Arg("file", "target file").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.EncryptGCP(*file, *extension); err != nil {
			return err
		}
		return nil
	})
}

func encryptAws(parent *kingpin.CmdClause) {
	cmd := parent.Command("aws", "Encrypt file using AWS").Alias("a")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	file := cmd.Arg("file", "target file").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.EncryptAWS(*file, *extension); err != nil {
			return err
		}
		return nil
	})
}

func decryptGcp(parent *kingpin.CmdClause) {
	cmd := parent.Command("gcp", "Decrypt file using GCP").Alias("g")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	file := cmd.Arg("file", "target file").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.DecryptGCP(*file, *extension); err != nil {
			return err
		}
		return nil
	})
}

func decryptAws(parent *kingpin.CmdClause) {
	cmd := parent.Command("aws", "Decrypt file using AWS").Alias("a")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	file := cmd.Arg("file", "target file").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.DecryptAWS(*file, *extension); err != nil {
			return err
		}
		return nil
	})
}

func reEncryptGcp(parent *kingpin.CmdClause) {
	cmd := parent.Command("gcp", "Encrypt files recursively, if encrypted file is exists using GCP KMS").Alias("g")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	dir := cmd.Arg("dir", "target dir").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.ReEncryptGCP(*dir, *extension); err != nil {
			return err
		}
		return nil
	})
}

func reEncryptAws(parent *kingpin.CmdClause) {
	cmd := parent.Command("aws", "Encrypt files recursively, if encrypted file is exists using AWS KMS").Alias("a")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	dir := cmd.Arg("dir", "target dir").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.ReEncryptAWS(*dir, *extension); err != nil {
			return err
		}
		return nil
	})
}

func reDecryptGcp(parent *kingpin.CmdClause) {
	cmd := parent.Command("gcp", "Decrypt files recursively, if encrypted file is exists using GCP KMS").Alias("g")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	dir := cmd.Arg("dir", "target dir").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.ReDecryptGCP(*dir, *extension); err != nil {
			return err
		}
		return nil
	})
}

func reDecryptAws(parent *kingpin.CmdClause) {
	cmd := parent.Command("aws", "Decrypt files recursively, if encrypted file is exists using AWS KMS").Alias("a")
	config := cmd.Flag("config", "config file").Short('c').Required().String()
	extension := cmd.Flag("extension", "encrypted files extension name").Short('e').Default(".crypted").String()
	dir := cmd.Arg("dir", "target dir").Required().String()
	cmd.Action(func(context *kingpin.ParseContext) error {
		ckey, err := NewCloudKey(*config)
		if err != nil {
			return err
		}
		if err := ckey.ReDecryptAWS(*dir, *extension); err != nil {
			return err
		}
		return nil
	})
}
