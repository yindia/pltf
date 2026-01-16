package backend

import (
	"context"
	"fmt"
	"strings"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"pltf/pkg/config"
)

type s3Backend struct{}

func (s s3Backend) Type() string { return "s3" }

func (s s3Backend) Resolve(envCfg *config.EnvironmentConfig, envEntry config.EnvironmentEntry) (Config, error) {
	bucket := strings.TrimSpace(envCfg.Backend.Bucket)
	if bucket == "" {
		bucket = defaultBackendBucket("s3", envCfg.Metadata.Org, envCfg.Metadata.Name)
	}
	region := strings.TrimSpace(envCfg.Backend.Region)
	if region == "" {
		region = envEntry.Region
	}
	return Config{
		Type:          "s3",
		Bucket:        bucket,
		Region:        region,
		Container:     strings.TrimSpace(envCfg.Backend.Container),
		ResourceGroup: strings.TrimSpace(envCfg.Backend.ResourceGroup),
		Profile:       strings.TrimSpace(envCfg.Backend.Profile),
	}, nil
}

func (s s3Backend) Ensure(ctx context.Context, cfg Config) error {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return fmt.Errorf("s3 backend bucket is required")
	}
	opts := []func(*awscfg.LoadOptions) error{}
	if cfg.Region != "" {
		opts = append(opts, awscfg.WithRegion(cfg.Region))
	}
	if cfg.Profile != "" {
		opts = append(opts, awscfg.WithSharedConfigProfile(cfg.Profile))
	}
	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(awsCfg)
	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: &cfg.Bucket,
	})
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "NotFound") && !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("failed to check bucket %s: %w", cfg.Bucket, err)
	}

	createInput := &s3.CreateBucketInput{
		Bucket: &cfg.Bucket,
	}
	if cfg.Region != "" && cfg.Region != "us-east-1" {
		createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(cfg.Region),
		}
	}
	if _, err := client.CreateBucket(ctx, createInput); err != nil {
		return fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
	}
	return nil
}

func init() {
	RegisterProvider(s3Backend{})
}
