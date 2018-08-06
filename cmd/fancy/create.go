package main

import (
	"fmt"
	"io"
	"os"

	"time"

	"github.com/birdayz/fancyfs/blobstore"
	"github.com/birdayz/fancyfs/cas"
	"github.com/birdayz/fancyfs/schema"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resource",
	Long:  "Create a resource",
}

var createFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Create file",
	Long:  "Create file",
	Args:  cobra.ExactArgs(1),
	// Ignore errcheck errors as os.Stderr error is not checked
	// nolint: errcheck
	Run: func(cmd *cobra.Command, args []string) {
		minio, err := blobstore.NewMinio(minioEndpoint, minioAccessKey, minioSecretKey, false)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to create blobstore: ", err)
			os.Exit(1)
		}

		blobstore := minio
		f := cas.NewFile(blobstore, defaultBlobSize, "tmp") // TODO fixme "generate" permanode id before
		schemaStore := schema.Storage{
			Blobstore: blobstore,
		}

		filePath := args[0]
		input, err := os.Open(filePath) // nolint: gosec
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		stat, err := input.Stat()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		n, err := io.Copy(f, input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to copy all:", err)
			fmt.Println(n)
			os.Exit(1)
		}
		if n != stat.Size() {
			fmt.Fprintln(os.Stderr, "Failed to copy properly..dafuq", err)
			os.Exit(1)
		}

		blobs, size := f.GetSnapshot()

		if size != stat.Size() {
			fmt.Fprintln(os.Stderr, "Didnt copy enough bytes?!")
			os.Exit(1)
		}

		schemaBlob := &schema.FileNode{
			Meta: &schema.PermanodeMeta{
				Rnd:             "", // TODO
				CreateTimestamp: time.Now().UnixNano(),
			},
			Filename: stat.Name(),
			Size:     stat.Size(),
			BlobSize: defaultBlobSize,
			BlobRefs: blobs,
		}

		id, _, err := schemaStore.Put(schemaBlob)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("sha256-" + id)

	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createFileCmd)
}
