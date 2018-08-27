# libipfs [![Build Status](https://travis-ci.org/skvark/libipfs.svg?branch=master)](https://travis-ci.org/skvark/libipfs)

Work in Progress! More information and features will be added when the project proceeds further.

Simple C wrapper for [go-ipfs](https://github.com/ipfs/go-ipfs).

``go-ipfs`` is usually run as a daemon with e.g. ``ipfs daemon``. This is not possible in many systems (such as in Sailfish OS based mobile devices) which have more restricted environments and tighter rules for application deployment (no multiple executables allowed etc.).

The main purpose of this library is to be used as a dependency for actual applications in systems where separate daemons are not allowed. The CI (Travis) has been set up so that it produces RPM packages for Sailfish OS which is also the primary target of this library.

## Build Process

RPM packages are built on Travis with the help of [SailfishOS SDK Docker image](https://github.com/CODeRUS/docker-sailfishos-sdk-local). Unfortunately it's not possible to build these packages on [Mer Open Build Service](https://build.merproject.org/) since network connectivity is missing in the OBS environments. ``go-ipfs``'s as well as ``go``'s package management depends heavily on network availability.

The ``.spec`` file and packaging setup are somewhat unusual because ``go-ipfs`` and the wrapper library have to be cross-compiled with ``cgo`` for ARM platform with Sailfish OS SDK toolchains. Please see the ``.spec`` file and ``.travis.yml`` for the changes which have been made to be able to execute the build process successfully.

The ``go`` compiler is installed to the build environment via [Mer OBS](https://build.merproject.org/package/show/home:skvark/go).

## go-ipfs Wrapper

The wrapper code itself is in the ``src`` folder. It's just a very initial prototype, but it is able to spin up a full IPFS node, add content to the IPFS and shut down the node. The code will utilize callbacks in some cases because long running tasks are executed as ``goroutine``s.

See the ``test`` folder for very basic command line app example.
