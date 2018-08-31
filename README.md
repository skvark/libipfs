# libipfs [![Build Status](https://travis-ci.org/skvark/libipfs.svg?branch=master)](https://travis-ci.org/skvark/libipfs)

Work in Progress! More information and features will be added when the project proceeds further.

Simple C wrapper for [go-ipfs](https://github.com/ipfs/go-ipfs).

``go-ipfs`` is usually run as a daemon with e.g. ``ipfs daemon``. This is not possible in many systems (such as in Sailfish OS based mobile devices) which have more restricted environments and tighter rules for application deployment (no multiple executables allowed etc.).

The main purpose of this library is to be used as a dependency for actual applications in systems where separate daemons are not allowed. The CI (Travis) has been set up so that it produces RPM packages for Sailfish OS which is also the primary target of this library.

## Build Process (Sailfish OS)

RPM packages are built on Travis with the help of [SailfishOS SDK Docker image](https://github.com/CODeRUS/docker-sailfishos-sdk-local). Unfortunately it's not possible to build these packages on [Mer Open Build Service](https://build.merproject.org/) since network connectivity is missing in the OBS environments. ``go-ipfs``'s as well as ``go``'s package management depends heavily on network availability.

The ``.spec`` file and packaging setup are somewhat unusual because ``go-ipfs`` and the wrapper library have to be cross-compiled with ``cgo`` for ARM platform with Sailfish OS SDK toolchains. Please see the ``.spec`` file and ``.travis.yml`` for the changes which have been made to be able to execute the build process successfully.

The ``go`` compiler is installed to the build environment during build.

#### Manual builds in other systems

While the primary target of this repository is Sailfish OS, the wrapper can be built for any system which is supported by ``go`` and ``go-ipfs``. 

1. Install ``go`` if it's not installed.
2. Fetch [go-ipfs](https://github.com/ipfs/go-ipfs) sources according to the guide in the ``go-ipfs`` README and run ``make deps`` in the source folder.
3. Run ``CGO_ENABLED=1 go build -o libipfs.so -buildmode=c-shared src/go_ipfs_wrapper.go``

You might want to set GOOS and GOARCH environment variables if you are targeting some other OS and architecture than your host system (cross-compilation). However, this requires the target system toolchain (compilers, linkers etc.) to work properly. For example see the ``rpm/libipfs.spec`` file.

## Documentation for go-ipfs Wrapper

The wrapper code itself is in the ``src`` folder. It's just a very initial prototype, but it is able to spin up a full IPFS node, add content to the IPFS and shut down the node. The code utilizes callbacks in the case of long running tasks. These tasks are executed as ``goroutine``s.

See the ``test`` folder for very basic ``C`` command line app example.

Further documentation will be written when the wrapper stabilizes. This depends partially from ``go-ipfs`` development since the ``go-ipfs`` internal coreapi is not very stable and is missing a lot of features.

### Installation in Sailfish OS SDK

To be able to use this library as a dependency, it needs to be installed to the Sailfish OS build engine. This can be done by logging into the build engine via SSH and running following commmands (change the package name according to the version you wish to install):

**i486 target**

1. Download the package: ``curl -O -L https://github.com/skvark/libipfs/releases/download/0.1.0/libipfs-0.1.0-0.i486.rpm``
2. Install the package: ``sb2 -t SailfishOS-2.2.0.29-i486 -m sdk-install -R rpm -i libipfs-0.1.0-0.i486.rpm``

**armv7hl target**

1. Download the package: ``curl -O -L https://github.com/skvark/libipfs/releases/download/0.1.0/libipfs-0.1.0-0.armv7hl.rpm``
2. Install the package: ``sb2 -t SailfishOS-2.2.0.29-armv7hl -m sdk-install -R rpm -i libipfs-0.1.0-0.armv7hl.rpm``

After installation, go to Sailfish OS-> Targets -> target settings -> Sync. Do this for both targets.

In the project which depends on this library, set the include path in ``.pro`` file (so that the ``libipfs.h`` header is found). For example on Windows and for armv7hl target:

``INCLUDEPATH += C:\SailfishOS\mersdk\targets\SailfishOS-2.2.0.29-armv7hl\usr\include\``

Additionally, you'll need to copy the ``libipfs.so`` file in the project ``.spec`` file to the ``rpm`` package and ship it with your application. 
