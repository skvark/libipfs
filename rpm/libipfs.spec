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
%setup -q -n %{name}-%{version}

%build
export PATH=$PATH:/usr/local/go/bin
go get -u -d github.com/ipfs/go-ipfs
ls $HOME
cd $HOME/go/src/github.com/ipfs/go-ipfs
go --version
make deps
cd %{buildroot}/src
go build -o libipfs.so -buildmode=c-shared go_ipfs_wrapper.go

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