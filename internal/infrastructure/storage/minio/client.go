package minio

import (
	"errors"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Region          string
	Bucket          string
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Endpoint) == "" {
		return errors.New("minio endpoint is required")
	}
	if strings.TrimSpace(c.AccessKeyID) == "" {
		return errors.New("minio access key is required")
	}
	if strings.TrimSpace(c.SecretAccessKey) == "" {
		return errors.New("minio secret key is required")
	}
	if strings.TrimSpace(c.Bucket) == "" {
		return errors.New("object storage bucket is required")
	}
	return nil
}

func New(cfg Config) (*minio.Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
