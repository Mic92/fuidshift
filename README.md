# Fuidshift: Move Filesystem ownership into other subordinated uid ranges

## Motivation

When using unprivileged [lxc](https://linuxcontainers.org/)/[systemd-nspawn](http://www.freedesktop.org/software/systemd/man/systemd-nspawn.html)
containers the container process is shifted into subuids ranges.
This requires however that container filesystem use the same range. Most
distribution installer/bootstrap tools does provide options to achieve this.
`fuidshift` allow to migrate the os filesystem tree later on. fuidshift is part
of [lxd](https://linuxcontainers.org/lxd/). This repo however removed all unneeded dependencies,
which comes with lxd, so it can be build and installed with a single `go get`.

## Installation

1. Install the go compiler
2. Get fuidshift:

  $ go get github.com/Mic92/fuidshift

## Usage

This shift uid/guid range use:

    $ fuidshift path/to/rootfs/ b:0:100000:65536

and reverse it with:

    $ fuidshift -r path/to/rootfs/ b:0:100000:65536
