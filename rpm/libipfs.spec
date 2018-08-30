Name:           libipfs
Version:        0.1
Release:        0
Summary:        C wrapper for go-ipfs
License:        MIT
Group:          Development/Libraries
URL:        	https://github.com/skvark/libipfs
Source0:    	%{name}-%{version}.tar.gz
Provides:       libipfs-devel = %{name}%{version}
Obsoletes:      libipfs < %{name}%{version}
BuildRequires:  git
BuildRequires:  tar
BuildRequires:  go
ExclusiveArch:  %ix86 x86_64 %arm

%description
Simple C wrapper for go-ipfs.

%prep

%define _sfos_version %{getenv:SFOS_VERSION}
%define _target %{getenv:TARGET}
%define _abi %{getenv:ABI}

%ifarch %ix86
# 386 (a.k.a. x86 or x86-32)
%define _goarch 386
%endif

%ifarch %arm
# arm (a.k.a. ARM)
%define _goarch arm
%endif

%setup -q -n %{name}-%{version}

%build

export GOROOT=/srv/mer/targets/SailfishOS-%{_sfos_version}-%{_target}/usr/local/go
export PATH=$PATH:$GOROOT/bin

go env
go get -u -d github.com/ipfs/go-ipfs
cd $HOME/go/src/github.com/ipfs/go-ipfs
make deps

cd $HOME/libipfs/src

export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/srv/mer/toolings/SailfishOS-%{_sfos_version}/usr/lib
export CC=/srv/mer/toolings/SailfishOS-%{_sfos_version}/opt/cross/bin/%{_target}-meego-linux-%{_abi}-gcc
export CXX=/srv/mer/toolings/SailfishOS-%{_sfos_version}/opt/cross/bin/%{_target}-meego-linux-%{_abi}-g++
export GOOS=linux
export GOARCH=%{_goarch}

%if "%{_goarch}" == arm
export GOARM=7
%endif

export CGO_ENABLED=1
export CGO_CFLAGS_ALLOW=.*
export CGO_CXXFLAGS_ALLOW=.*
export CGO_LDFLAGS_ALLOW=.*
export CPATH=/srv/mer/targets/SailfishOS-%{_sfos_version}-%{_target}/usr/include
export LIBRARY_PATH=/srv/mer/targets/SailfishOS-%{_sfos_version}-%{_target}/usr/lib
export CGO_LDFLAGS=--sysroot=/srv/mer/targets/SailfishOS-%{_sfos_version}-%{_target}/

go build -x -v -ldflags=all="-s -w" -o libipfs.so -buildmode=c-shared go_ipfs_wrapper.go

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}%{_libdir}/
mkdir -p %{buildroot}%{_includedir}/
cp src/libipfs.so %{buildroot}%{_libdir}/libipfs.so
cp src/libipfs.h %{buildroot}%{_includedir}/libipfs.h

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%{_libdir}/libipfs.so
%{_includedir}/libipfs.h