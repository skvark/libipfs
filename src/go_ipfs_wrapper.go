package main

import (
	"context"
	"fmt"
	"os"
	"io"
	"io/ioutil"
	"path"
	"bytes"
	"unsafe"
	"encoding/json"
	"encoding/base64"
	core "github.com/ipfs/go-ipfs/core"
	coreapi "github.com/ipfs/go-ipfs/core/coreapi"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	corerepo "github.com/ipfs/go-ipfs/core/corerepo"
	coreunix "github.com/ipfs/go-ipfs/core/coreunix"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	identify "gx/ipfs/QmY51bqSM5XgxQZqsBrQcRkKTnCb8EKpJpR9K6Qax7Njco/go-libp2p/p2p/protocol/identify"
	ic "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	"gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds/cli"
	"github.com/ipfs/go-ipfs/repo/config"
	commands "github.com/ipfs/go-ipfs/core/commands"
)

//#include <stdlib.h>
//
//static void callback(void* func, char* error, char* data, size_t size, int fid, void* instance)
//{
//    ((void(*)(char*, char*, size_t, int, void*)) func)(error, data, size, fid, instance);
//}
//
//enum functions {
//    f_ipfs_add_bytes,
//    f_ipfs_add_path_or_file,
//    f_ipfs_ls,
//    f_ipfs_cat,
//    f_ipfs_unpin,
//    f_ipfs_gc,
//    f_ipfs_peers,
//    f_ipfs_id,
//    f_ipfs_repo_stats,
//    f_ipfs_config
//};
//
import "C"

const (
	nBitsForKeypair = 2048
)

type NodeInfo struct {
	ID string
	PublicKey string
	Addresses []string
	AgentVersion string
	ProtocolVersion string
}

type Peer struct {
	Pid string
	Addr string
}

type Dir struct {
	Cid string
	Name string
}

type Api struct {
	api coreiface.CoreAPI
	node *core.IpfsNode
	ipfsConfig *config.Config
	ctx context.Context
	cancel context.CancelFunc
}

var api *Api = nil

// Given that this wrapper is used via C++ classes,
// this variable holds the class instance pointer
// to be used in the callback method of the class
var callback_class_instance unsafe.Pointer = nil

func main() { }

//export register_callback_class_instance
func register_callback_class_instance(instance unsafe.Pointer) {
	callback_class_instance = instance
}

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

	data := []byte(e)
	cdata := C.CBytes(data)
	defer C.free(cdata)

	C.callback(callback, (*C.char)(cdata), nil, C.size_t(len(e)), -1, callback_class_instance)
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
		p, err := api.api.Unixfs().Add(api.ctx, bytes.NewReader(content))

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		path := p.String()
		data := []byte(path)
		cdata := C.CBytes(data)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(path)), C.f_ipfs_add_bytes, callback_class_instance)
	}()
}

// - go-ipfs coreapi does not support adding by path yet so this method is implemented via other internal apis
// - all files / directories are wrapped to container folder to preserve filenames
// - if a directory is given, files are added recursively
// - nocopy is not enabled because it's expiremental
//export ipfs_add_path_or_file
func ipfs_add_path_or_file(path *C.char, callback unsafe.Pointer) {
	go func() {
		pathString := C.GoString(path)

		// emulate command line args
		var args []string
		args = append(args, "add")

		if info, err := os.Stat(pathString); err == nil && info.IsDir() {
			args = append(args, "-r")
			args = append(args, pathString)
		} else {
			args = append(args, pathString)
			args =append(args, "-w")
		}

		// parse args to get Request object
		req, err := cli.Parse(api.ctx, args, nil, commands.Root)
		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		outChan := make(chan interface{}, 8)

		fileAdder, err := coreunix.NewAdder(api.ctx, api.node.Pinning, api.node.Blockstore, api.node.DAG)
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

		data := []byte(nd.String())
		cdata := C.CBytes(data)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(data)), C.f_ipfs_add_path_or_file, callback_class_instance)
	}()
}

//export ipfs_ls
func ipfs_ls(cid *C.char, callback unsafe.Pointer) {
	cid_string := C.GoString(cid)
	path, err := coreiface.ParsePath("/ipfs/" + cid_string)

	if err != nil {
		createErrorCallback(err, callback)
		return;
	}

	go func() {
		dir, err := api.api.Unixfs().Ls(api.ctx, path)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		dirs := []Dir{}

		for _, d := range dir {
			ds := Dir {
				Cid: d.Cid.String(),
				Name: d.Name,
			}

			dirs = append(dirs, ds)
		}

		jsonStr, err := json.Marshal(dirs)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		cdata := C.CBytes(jsonStr)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(jsonStr)), C.f_ipfs_ls, callback_class_instance)
	}()
}

//export ipfs_cat
func ipfs_cat(cid *C.char, callback unsafe.Pointer) {
	cid_string := C.GoString(cid)
	path, err := coreiface.ParsePath(cid_string)

	if err != nil {
		createErrorCallback(err, callback)
		return;
	}

	go func() {
		r, err := api.api.Unixfs().Cat(api.ctx, path)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		bytes, err := ioutil.ReadAll(r)
		cdata := C.CBytes(bytes)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(bytes)), C.f_ipfs_cat, callback_class_instance)
	}()
}

//export ipfs_unpin
func ipfs_unpin(cid *C.char, callback unsafe.Pointer) {
	cid_string := C.GoString(cid)
	path, err := coreiface.ParsePath(cid_string)

	if err != nil {
		createErrorCallback(err, callback)
		return;
	}

	go func() {
		err := api.api.Pin().Rm(api.ctx, path)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		data := []byte(cid_string)
		cdata := C.CBytes(data)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(cid_string)), C.f_ipfs_unpin, callback_class_instance)
	}()
}

//export ipfs_gc
func ipfs_gc(callback unsafe.Pointer) {
	go func() {
		if err := corerepo.GarbageCollect(api.node, api.ctx); err != nil {
			createErrorCallback(err, callback)
			return;
		}
		C.callback(callback, nil, nil, 0, C.f_ipfs_gc, callback_class_instance)
	}()
}

//export ipfs_peers
func ipfs_peers(callback unsafe.Pointer) {
	go func() {
		conns := api.node.PeerHost.Network().Conns()

		peers := []Peer{}

		for _, c := range conns {
			pid := c.RemotePeer()
			addr := c.RemoteMultiaddr()

			p := Peer {
				Pid: pid.Pretty(),
				Addr: addr.String(),
			}

			peers = append(peers, p)
		}

		jsonStr, err := json.Marshal(peers)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		cdata := C.CBytes(jsonStr)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(jsonStr)), C.f_ipfs_peers, callback_class_instance)
	}()
}

//export ipfs_id
func ipfs_id(id *C.char, callback unsafe.Pointer) {
	go func() {
		id_string := C.GoString(id)
		nodeInfo := new(NodeInfo)

		if len(id_string) == 0 {
			nodeInfo.ID = api.node.Identity.Pretty()

			if api.node.PrivateKey == nil {
				if err := api.node.LoadPrivateKey(); err != nil {
					createErrorCallback(err, callback)
				}
			}

			pk := api.node.PrivateKey.GetPublic()
			pkb, err := ic.MarshalPublicKey(pk)

			if err != nil {
				createErrorCallback(err, callback)
				return;
			}

			nodeInfo.PublicKey = base64.StdEncoding.EncodeToString(pkb)

			if api.node.PeerHost != nil {
				for _, a := range api.node.PeerHost.Addrs() {
					s := a.String() + "/ipfs/" + nodeInfo.ID
					nodeInfo.Addresses = append(nodeInfo.Addresses, s)
				}
			}

			nodeInfo.ProtocolVersion = identify.LibP2PVersion
			nodeInfo.AgentVersion = identify.ClientVersion

		} else {

			pid, err := peer.IDB58Decode(id_string)
			if err != nil {
				createErrorCallback(err, callback)
				return;
			}

			p, err := api.node.Routing.FindPeer(api.ctx, pid)

			if err != nil {
				createErrorCallback(err, callback)
				return;
			}

			if pk := api.node.Peerstore.PubKey(p.ID); pk != nil {
				pkb, err := ic.MarshalPublicKey(pk)
				if err != nil {
					createErrorCallback(err, callback)
					return;
				}
				nodeInfo.PublicKey = base64.StdEncoding.EncodeToString(pkb)
			}

			for _, a := range api.node.Peerstore.Addrs(p.ID) {
				nodeInfo.Addresses = append(nodeInfo.Addresses, a.String())
			}

			if v, err := api.node.Peerstore.Get(p.ID, "ProtocolVersion"); err == nil {
				if vs, ok := v.(string); ok {
					nodeInfo.ProtocolVersion = vs
				}
			}

			if v, err := api.node.Peerstore.Get(p.ID, "AgentVersion"); err == nil {
				if vs, ok := v.(string); ok {
					nodeInfo.AgentVersion = vs
				}
			}
		}

		jsonStr, err := json.Marshal(nodeInfo)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		cdata := C.CBytes(jsonStr)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(jsonStr)), C.f_ipfs_id, callback_class_instance)
	}()
}

//export ipfs_repo_stats
func ipfs_repo_stats(callback unsafe.Pointer) {
	go func() {
		stat, err := corerepo.RepoStat(api.ctx, api.node)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		jsonStr, err := json.Marshal(stat)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		cdata := C.CBytes(jsonStr)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(jsonStr)), C.f_ipfs_repo_stats, callback_class_instance)
	}()
}

//export ipfs_config
func ipfs_config(callback unsafe.Pointer) {
	go func() {
		cfg, err := api.node.Repo.Config()

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		jsonStr, err := json.Marshal(cfg)

		if err != nil {
			createErrorCallback(err, callback)
			return;
		}

		cdata := C.CBytes(jsonStr)
		defer C.free(cdata)

		C.callback(callback, nil, (*C.char)(cdata), C.size_t(len(jsonStr)), C.f_ipfs_config, callback_class_instance)
	}()
}