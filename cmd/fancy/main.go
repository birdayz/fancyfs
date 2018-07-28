package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	minioEndpoint  string
	minioAccessKey string
	minioSecretKey string
)

func init() {
	minioEndpoint = os.Getenv("MINIO_ENDPOINT")
	minioAccessKey = os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey = os.Getenv("MINIO_SECRET_KEY")
}

const defaultBlobSize = 2 * 1024 * 1024

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "fancy",
	Short: "fancy is the CLI tool for fancyfs",
	Long:  `The best FS EVER.`,
}
