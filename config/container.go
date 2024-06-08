package config

import (
	"log"

	"github.com/containers/image/v5/types"
)

type DockerConfiguration struct {
	Registry      string
	Repository    string
	SystemContext *types.SystemContext
}

var containersConfig *DockerConfiguration

func ContainersConfig() DockerConfiguration {
	if containersConfig == nil {
		containersConfig = readContainersConfig()
	}

	return *containersConfig
}

func readContainersConfig() *DockerConfiguration {
	registry, err := ReadEnv("REGISTRY")
	if err != nil {
		log.Fatalf("failed to load REGISTRY")
	}

	repo, err := ReadEnv("REPOSITORY")
	if err != nil {
		log.Fatalf("failed to load REPOSITORY")
	}

	dockerAuthConfig := &types.DockerAuthConfig{
		Username: ReadEnvWithDefault("REPO_USERNAME", "repoUser"),
		Password: ReadEnvWithDefault("REPO_PASSWORD", "repoPass"),
	}

	return &DockerConfiguration{
		Repository: repo,
		Registry:   registry,
		SystemContext: &types.SystemContext{
			DockerAuthConfig:          dockerAuthConfig,
			DockerCertPath:            ReadEnvWithDefault("DOCKER_CERT_PATH", ""),
			DockerBearerRegistryToken: ReadEnvWithDefault("REGISTRY_BEARER_TOKEN", ""),
		},
	}
}
