package main

import (
	"fmt"
	"io"
	"os"

	"strings"

	"github.com/birdayz/fancyfs/blobstore"
	"github.com/birdayz/fancyfs/cas"
	"github.com/birdayz/fancyfs/schema"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resource",
	Long:  "Get a resource",
}

var getFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Create file",
	Long:  "Create file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		minio, err := blobstore.NewMinio(minioEndpoint, minioAccessKey, minioSecretKey, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create blobstore: ", err) // nolint: errcheck
			os.Exit(1)
		}

		blobstore := minio
		// f := cas.NewFile(blobstore, defaultBlobSize)
		schemaStore := schema.Storage{
			Blobstore: blobstore,
		}
		fileNode, err := schemaStore.Get(strings.TrimPrefix(args[0], "sha256-"))
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get schema blob", err) // nolint: errcheck
		}

		f := cas.NewFileFromSchemaBlob(minio, fileNode.GetBlobSize(), fileNode.GetBlobRefs(), fileNode.GetSize())

		n, err := io.Copy(os.Stdout, f)
		if err != nil {
			panic(err)
		}
		if n != fileNode.GetSize() {
			panic("didnt copy all")
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getFileCmd)
}
