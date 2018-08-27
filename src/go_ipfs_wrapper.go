package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"bytes"
	"unsafe"
	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	"gx/ipfs/QmTyiSs9VgdVb4pnzdjtKhcfdTkHFEaNn6xnCbZq4DTFRt/go-ipfs-config"
)

//#include <stdlib.h>
//
//static void callback(void* func, char* error, char* data, size_t size)
//{
//    ((void(*)(char*, char*, size_t)) func)(error, data, size);
//}
import "C"

const (
	nBitsForKeypair = 2048
)

type Node struct {
	node *core.IpfsNode
	api coreiface.CoreAPI
	ctx context.Context
	cancel context.CancelFunc
}

var node *Node = nil

func main() { }

// see: https://github.com/ipfs/go-ipfs/blob/master/cmd/ipfs/init.go
func checkRepo(repo_root string) error {
	_, err := os.Stat(repo_root)

	if err == nil { // dir exists
		testfile := path.Join(repo_root, "test")
		fi, err := os.Create(testfile)
		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("%s is not writeable by the current user", repo_root)
			}
			return fmt.Errorf("unexpected error while checking writeablility of repo root: %s", err)
		}
		fi.Close()
		return os.Remove(testfile)
	}

	if os.IsNotExist(err) {
		return os.Mkdir(repo_root, 0775) // create the root dir if possible
	}

	if os.IsPermission(err) {
		return fmt.Errorf("cannot write to %s, incorrect permissions", err)
	}

	return err
}

//export ipfs_start
func ipfs_start(repo_root *C.char) *C.char {

	var n Node
	repoRoot := C.GoString(repo_root)

	n.ctx, n.cancel = context.WithCancel(context.Background())

	if err := checkRepo(repoRoot); err != nil {
		return C.CString(err.Error())
	}

	// check that there is no existing repo in the repoRoot
	// create if no repo exists
	if !fsrepo.IsInitialized(repoRoot) {

		conf, err := config.Init(os.Stdout, nBitsForKeypair)

		if err != nil {
			return C.CString(err.Error())
		}

		if err := fsrepo.Init(repoRoot, conf); err != nil {
			return C.CString(err.Error())
		}
	}

	// try to open the repo
	repo, err := fsrepo.Open(repoRoot);

	if err != nil {
		return C.CString(err.Error())
	}

	n.node, err = core.NewNode(n.ctx, &core.BuildCfg{
		Online: true,
		Permanent: true,
		Repo: repo,
	})

	n.node.SetLocal(false)
	n.api = coreapi.NewCoreAPI(n.node)
	node = &n

	return nil
}

//export ipfs_stop
func ipfs_stop() {
	node.cancel()
}

//export ipfs_add
func ipfs_add(data unsafe.Pointer, size C.size_t, callback unsafe.Pointer) {
	content := C.GoBytes(data, C.int(size))

	go func() {
		ctx := context.Background()

		p, err := node.api.Unixfs().Add(ctx, bytes.NewReader(content))

		if err != nil {
			var e = err.Error()
			C.callback(callback, C.CString(e), nil, C.size_t(len(e)))
			return;
		}

		C.callback(callback, nil, C.CString(p.String()), C.size_t(len(p.String())))
	}()
}