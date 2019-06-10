[![Build Status](https://travis-ci.org/ganto/packer-builder-ovirt.svg?branch=master)](https://travis-ci.org/ganto/packer-builder-ovirt)

# oVirt packer.io builder

This builder plugin extends [packer.io](https://packer.io) to support building
images for [oVirt](https://www.ovirt.org).

## Development

### Prerequisites

To compile this plugin you must have a working Go compiler setup. Follow the
[official instructions](https://golang.org/doc/install) or use your local
package manager to install Go on your system.

### Compile the plugin

```shell
cd $GOPATH
mkdir -p src/github.com/ganto
cd src/github.com/ganto
git clone https://github.com/ganto/packer-builder-ovirt.git
cd packer-builder-ovirt
PACKER_DEV=1 make bin
```

If the build was successful, you should now have the `packer-builder-ovirt`
binary in your `$GOPATH/bin` directory.

In order to do a cross-compile, run the following build command:

```shell
XC_OS="linux" XC_ARCH="386 amd64" make bin
```

This builds 32 and 64 bit binaries for Linux. Native binaries will be installed
in `$GOPATH/bin` as above, and cross-compiled ones in the `pkg/` directory.
