package main

import (
	"fmt"
	"os"

	"os/signal"

	"github.com/birdayz/fancyfs"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

func main() {
	root := &fancyfs.Node{Node: nodefs.NewDefaultNode()}

	fs := &fancyfs.Fs{
		Root: root,
	}

	_ = fs

	folder := "/tmp/mnt26"

	err := os.Mkdir(folder, 0750)
	if err != nil {
		panic(err)
	}
	server, _, err := nodefs.MountRoot(folder, root, nodefs.NewOptions())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Mounting %v\n", folder)
	go server.Serve()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	<-signals

	fmt.Println("Exiting")
	err = server.Unmount()
	if err != nil {
		fmt.Printf("Failed to unmount: %v", err)
	}
}
