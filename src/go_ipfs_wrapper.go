package main

import (
	"context"
	"fmt"
	"os"
	"io"
	"path"
	"bytes"
	"unsafe"
	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	"gx/ipfs/QmPTfgFTo9PFr1PvPKyKoeMgBvYPh6cX3aDP7DHKVbnCbi/go-ipfs-cmds/cli"
	"gx/ipfs/QmTyiSs9VgdVb4pnzdjtKhcfdTkHFEaNn6xnCbZq4DTFRt/go-ipfs-config"
	commands "github.com/ipfs/go-ipfs/core/commands"
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

type Api struct {
	api coreiface.CoreAPI
	node *core.IpfsNode
	ipfsConfig *config.Config
	ctx context.Context
	cancel context.CancelFunc
}

var api *Api = nil

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

func createErrorCallback(err error, callback unsafe.Pointer) {
	var e = err.Error()
	C.callback(callback, C.CString(e), nil, C.size_t(len(e)))
}

//export ipfs_start
func ipfs_start(repo_root *C.char) *C.char {
	var a Api
	repoRoot := C.GoString(repo_root)

	ctx, cancel := context.WithCancel(context.Background())

	a.ctx = ctx
	a.cancel = cancel

	if err := checkRepo(repoRoot); err != nil {
		return C.CString(err.Error())
	}

	var conf *config.Config

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

	node, err := core.NewNode(a.ctx, &core.BuildCfg{
		Online: true,
		Permanent: true,
		Repo: repo,
	})

	node.SetLocal(false)
	a.api = coreapi.NewCoreAPI(node)
	a.node = node
	a.ipfsConfig = conf
	api = &a

	return nil
}

//export ipfs_stop
func ipfs_stop() {
	api.cancel()
}

//export ipfs_add_bytes
func ipfs_add_bytes(data unsafe.Pointer, size C.size_t, callback unsafe.Pointer) {
	content := C.GoBytes(data, C.int(size))

	go func() {
		ctx := context.Background()

		p, err := api.api.Unixfs().Add(ctx, bytes.NewReader(content))

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		C.callback(callback, nil, C.CString(p.String()), C.size_t(len(p.String())))
	}()
}

// - go-ipfs coreapi does not support adding by path yet so this method is implemented via other internal apis
// - all files / directories are wrapped to container folder to preserve filenames
// - if a directory is given, files are added recursively
// - nocopy is not enabled because it's expiremental
//export ipfs_add_path_or_file
func ipfs_add_path_or_file(path *C.char, callback unsafe.Pointer) {
	go func() {
		ctx := context.Background()
		pathString := C.GoString(path)

		// emulate command line args
		var args []string
		args = append(args, "add")

		if info, err := os.Stat(pathString); err == nil && info.IsDir() {
			args = append(args, "-r")
		}

		args = append(args, pathString)
		args = append(args, "-w")

		// parse args to get Request object
		req, err := cli.Parse(ctx, args, nil, commands.Root)
		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		outChan := make(chan interface{}, 8)

		fileAdder, err := coreunix.NewAdder(ctx, api.node.Pinning, api.node.Blockstore, api.node.DAG)
		if err != nil {
			createErrorCallback(err, callback);
			return;
		}

		fileAdder.Out = outChan
		fileAdder.Wrap = true

		for {
			file, err := req.Files.NextFile()

			if err == io.EOF {
				// Finished the list of files.
				break
			} else if err != nil {
				createErrorCallback(err, callback)
				return;
			}

			if err := fileAdder.AddFile(file); err != nil {
				createErrorCallback(err, callback)
				return;
			}
		}

		nd, e := fileAdder.Finalize()
		if e != nil {
			createErrorCallback(e, callback)
			return;
		}

		if err := fileAdder.PinRoot(); err != nil {
			createErrorCallback(err, callback)
			return;
		}

		defer close(outChan)

		ndstr := nd.String()

		C.callback(callback, nil, C.CString(ndstr), C.size_t(len(ndstr)))
	}()
}