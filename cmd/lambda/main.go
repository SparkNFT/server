package main

import (
	"context"
	"encoding/json"
	"os"

	my_config "github.com/SparkNFT/key_server/config"
	"github.com/SparkNFT/key_server/controller"
	"github.com/SparkNFT/key_server/model"
	"github.com/akrylysov/algnhsa"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"golang.org/x/xerrors"
)

// Get environment value
func get_env(env_key string) string {
	result := os.Getenv(env_key)
	if len(result) == 0 {
		panic(xerrors.Errorf("ENV %s must be given! Abort.", env_key))
	}
	return result
}

func init_config_from_aws_secret() {
	secret_name := get_env("SECRET_NAME")
	region := get_env("SECRET_REGION")

	//Create a Secrets Manager client
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		panic(xerrors.Errorf("error when loading SDK config: %w", err))
	}

	client := secretsmanager.NewFromConfig(cfg)
	input := secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secret_name),
		VersionStage: aws.String("AWSCURRENT"),
	}
	result, err := client.GetSecretValue(context.Background(), &input)
	if err != nil {
		panic(xerrors.Errorf("error when getting secret value: %w", err))
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	if result.SecretString == nil {
		panic(xerrors.New("error when getting secret string"))
	}
	secret_string := *result.SecretString

	err = json.Unmarshal([]byte(secret_string), &my_config.C)
	if err != nil {
		panic(xerrors.Errorf("error when parsing config JSON: %w", err))
	}
}

func init() {
	init_config_from_aws_secret()
	model.Init()
	controller.Init()
}

func main() {
	algnhsa.ListenAndServe(controller.Engine, nil)
}
