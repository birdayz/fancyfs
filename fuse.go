package fancyfs

import (
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
)

type Fs struct {
	Root *Node
}

type Node struct {
	nodefs.Node
	inode *nodefs.Inode
}

func (n *Node) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) (code fuse.Status) {
	println("get", file)

	if n.Inode().IsDir() {
		out.Mode = fuse.S_IFDIR | 0777
		return fuse.OK
	}
	return fuse.OK
}

func (n *Node) Inode() *nodefs.Inode {
	return n.inode
}

func (n *Node) SetInode(node *nodefs.Inode) {
	n.inode = node
}

func (n *Node) Mknod(name string, mode uint32, dev uint32, context *fuse.Context) (newNode *nodefs.Inode, code fuse.Status) {
	println("Mknod")
	return
}

func (n *Node) OpenDir(context *fuse.Context) (dir []fuse.DirEntry, status fuse.Status) {
	return dir, fuse.OK
}

func (n *Node) Lookup(out *fuse.Attr, name string, context *fuse.Context) (inode *nodefs.Inode, status fuse.Status) {
	println("Lookup ", name)
	return nil, fuse.ENOENT
}

func (n *Node) Deletable() bool {
	return true
}

func (n *Node) OnMount(c *nodefs.FileSystemConnector) {
	println("OnMount")
}

func (n *Node) OnUnmount() {

}
