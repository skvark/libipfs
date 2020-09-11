package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"

	config "github.com/ipfs/go-ipfs-config"
	libp2p "github.com/ipfs/go-ipfs/core/node/libp2p"
	icore "github.com/ipfs/interface-go-ipfs-core"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/plugin/loader" // This package is needed so that all the preloaded plugins are loaded automatically
	"github.com/ipfs/go-ipfs/repo/fsrepo"
)

//#include <stdlib.h>
//
//static void callback(void* func, char* error, char* data, size_t size, int fid, void* instance)
//{
//    ((void(*)(char*, char*, size_t, int, void*)) func)(error, data, size, fid, instance);
//}
//
//enum functions {
//    f_ipfs_start,
//    f_ipfs_add_bytes,
//    f_ipfs_add_path_or_file,
//    f_ipfs_ls,
//    f_ipfs_cat,
//    f_ipfs_unpin,
//    f_ipfs_gc,
//    f_ipfs_peers,
//    f_ipfs_id,
//    f_ipfs_repo_stats,
//    f_ipfs_config,
//    f_ipfs_files_cp,
//    f_ipfs_files_ls,
//    f_ipfs_files_mkdir,
//    f_ipfs_files_stat
//};
//
import "C"

const (
	nBitsForKeypair = 2048
)

type Api struct {
	ipfs   icore.CoreAPI
	ctx    context.Context
	cancel context.CancelFunc
}

var api *Api = nil

// Given that this wrapper is used via C++ classes,
// this variable holds the class instance pointer
// to be used in the callback method of the class
var callback_class_instance unsafe.Pointer = nil

func main() {}

//export register_callback_class_instance
func register_callback_class_instance(instance unsafe.Pointer) {
	callback_class_instance = instance
}

func createErrorCallback(err error, callback unsafe.Pointer) {
	var e = err.Error()

	data := []byte(e)
	cdata := C.CBytes(data)
	defer C.free(cdata)

	C.callback(callback, (*C.char)(cdata), nil, C.size_t(len(e)), -1, callback_class_instance)
}

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func createRepo(repoRoot string) error {
	if _, err := os.Stat(repoRoot); os.IsNotExist(err) {
		os.Mkdir(repoRoot, os.ModeDir)
	} else {
		return nil
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		return err
	}

	// Create the repo with the config
	err = fsrepo.Init(repoRoot, cfg)
	if err != nil {
		return fmt.Errorf("Failed to init node: %s", err)
	}

	return nil
}

func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, error) {
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
	}

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, err
	}

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

//export ipfs_start
func ipfs_start(repo_root *C.char, callback unsafe.Pointer) {
	repoRoot := C.GoString(repo_root)
	api = new(Api)

	ctx, cancel := context.WithCancel(context.Background())

	api.ctx = ctx
	api.cancel = cancel

	go func() {
		if err := createRepo(repoRoot); err != nil {
			createErrorCallback(err, callback)
			return
		}

		defaultPath, err := config.PathRoot()
		if err != nil {
			createErrorCallback(err, callback)
			return
		}

		if err := setupPlugins(defaultPath); err != nil {
			createErrorCallback(err, callback)
			return
		}

		ipfs, err := createNode(ctx, defaultPath)

		api.ipfs = ipfs

		C.callback(callback, nil, nil, 0, C.f_ipfs_start, callback_class_instance)
	}()
}

//export ipfs_stop
func ipfs_stop() {
	api.cancel()
}
