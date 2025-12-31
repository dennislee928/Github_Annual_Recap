package config

import (
	"fmt"
	"os"
)

type Config struct {
	Token      string
	APIBase    string
	GraphQLEnd string
}

func FromEnv() (Config, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("GITHUB_TOKEN is required (set it in env or .env)")
	}

	api := os.Getenv("GITHUB_API_BASE")
	if api == "" {
		api = "https://api.github.com"
	}
	gql := os.Getenv("GITHUB_GRAPHQL")
	if gql == "" {
		gql = "https://api.github.com/graphql"
	}

	return Config{
		Token:      token,
		APIBase:    api,
		GraphQLEnd: gql,
	}, nil
}
