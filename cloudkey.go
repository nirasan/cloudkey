package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudkms/v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type CloudKey struct {
	Config     *Config
	GCPService *cloudkms.Service
	AWSService *kms.KMS
}

func NewCloudKey(configName string) (*CloudKey, error) {
	if configName == "" {
		return nil, errors.New("config name required")
	}

	c, err := LoadConfig(configName)
	if err != nil {
		return nil, err
	}

	return &CloudKey{Config: c}, nil
}

func (c *CloudKey) UseGCP() error {
	if c.GCPService != nil {
		return nil
	}

	var client *http.Client
	var err error

	if c.Config.GcpConfig.UseGcloudAccount {
		client, err = google.DefaultClient(context.Background(), cloudkms.CloudPlatformScope)
		if err != nil {
			return err
		}
	} else {
		data, err := ioutil.ReadFile(c.Config.GcpConfig.ServiceAccountKey)
		if err != nil {
			return err
		}
		conf, err := google.JWTConfigFromJSON(data, cloudkms.CloudPlatformScope)
		if err != nil {
			return err
		}
		client = conf.Client(context.Background())
	}

	s, err := cloudkms.New(client)
	if err != nil {
		return err
	}

	c.GCPService = s
	return nil
}

func (c *CloudKey) UseAWS() error {
	if c.AWSService != nil {
		return nil
	}

	ac := c.Config.AwsConfig

	var cred *credentials.Credentials
	if ac.UseStaticCreds {
		cred = credentials.NewStaticCredentials(ac.AccessKeyID, ac.SecretAccessKey, ac.AccessToken)
	} else {
		cred = credentials.NewSharedCredentials(ac.CredFile, ac.Profile)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      &ac.Region,
		Credentials: cred,
	})
	if err != nil {
		return err
	}
	c.AWSService = kms.New(sess)

	return nil
}

func (c *CloudKey) EncryptGCP(path, extension string) error {
	if path == "" {
		return errors.New("path required")
	}

	if err := c.UseGCP(); err != nil {
		return err
	}

	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	req := &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString(plaintext),
	}
	resp, err := c.GCPService.Projects.Locations.KeyRings.CryptoKeys.Encrypt(c.GCPCryptoKeyResource(), req).Do()
	if err != nil {
		return err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(resp.Ciphertext)
	if err != nil {
		return err
	}

	return c.OutputEncrypted(ciphertext, path, extension)
}

func (c *CloudKey) DecryptGCP(path, extension string) error {
	if path == "" {
		return errors.New("file required")
	}

	if err := c.UseGCP(); err != nil {
		return err
	}

	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	req := &cloudkms.DecryptRequest{
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}
	resp, err := c.GCPService.Projects.Locations.KeyRings.CryptoKeys.Decrypt(c.GCPCryptoKeyResource(), req).Do()
	if err != nil {
		return err
	}
	plaintext, err := base64.StdEncoding.DecodeString(resp.Plaintext)
	if err != nil {
		return err
	}

	return c.OutputDecrypted(plaintext, path, extension)
}

func (c *CloudKey) GCPCryptoKeyResource() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		c.Config.GcpConfig.Project, c.Config.GcpConfig.Location, c.Config.GcpConfig.Keyring, c.Config.GcpConfig.Cryptokey)
}

func (c *CloudKey) EncryptAWS(path, extension string) error {
	if path == "" {
		return errors.New("file required")
	}

	if err := c.UseAWS(); err != nil {
		return err
	}

	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	result, err := c.AWSService.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(c.Config.AwsConfig.CryptoKeyID),
		Plaintext: []byte(plaintext),
	})
	if err != nil {
		return err
	}

	return c.OutputEncrypted(result.CiphertextBlob, path, extension)
}

func (c *CloudKey) DecryptAWS(path, extension string) error {
	if path == "" {
		return errors.New("file required")
	}

	if err := c.UseAWS(); err != nil {
		return err
	}

	ciphertext, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	result, err := c.AWSService.Decrypt(&kms.DecryptInput{
		CiphertextBlob: ciphertext,
	})
	if err != nil {
		return err
	}

	return c.OutputDecrypted(result.Plaintext, path, extension)
}

func (c *CloudKey) OutputEncrypted(ciphertext []byte, path, extension string) error {
	if extension == "" {
		extension = ".crypted"
	}
	output := path + extension
	if err := ioutil.WriteFile(output, ciphertext, 0600); err != nil {
		return err
	}

	log.Printf("Encrypted file created: %s", output)

	return nil
}

func (c *CloudKey) OutputDecrypted(plaintext []byte, path, extension string) error {
	output := path
	if extension == "" {
		extension = ".crypted"
	}
	if i := strings.Index(output, extension); i != -1 {
		output = output[:i]
	}
	if err := ioutil.WriteFile(output, plaintext, 0600); err != nil {
		return err
	}

	log.Printf("Decrypted file created: %s", output)

	return nil
}

func (c *CloudKey) ReEncryptGCP(dir, extension string) error {
	if dir == "" {
		return errors.New("dir required")
	}

	if err := c.UseGCP(); err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			_, err := os.Stat(path + extension)
			if err == nil {
				if err := c.EncryptGCP(path, extension); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (c *CloudKey) ReEncryptAWS(dir, extension string) error {
	if dir == "" {
		return errors.New("dir required")
	}

	if err := c.UseAWS(); err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			_, err := os.Stat(path + extension)
			if err == nil {
				if err := c.EncryptAWS(path, extension); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (c *CloudKey) ReDecryptGCP(dir, extension string) error {
	if dir == "" {
		return errors.New("dir required")
	}

	if err := c.UseGCP(); err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(path, extension) {
				if err := c.DecryptGCP(path, extension); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (c *CloudKey) ReDecryptAWS(dir, extension string) error {
	if dir == "" {
		return errors.New("dir required")
	}

	if err := c.UseAWS(); err != nil {
		return err
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if strings.HasSuffix(path, extension) {
				if err := c.DecryptAWS(path, extension); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
