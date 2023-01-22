# tfs

`tfs` is a command line tool that helps for using different version of Terraform
on a daily basis. It has been inspired by [this project](https://github.com/warrensbox/terraform-switcher).

`tfs` is simple and ligthweight, and rely on [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)
conventions.

It works out of the box for every GNU/Linux distribution and macOS.

## Build

```text
git clone https://github.com/yannlambret/tfs.git && cd tfs
go build -o dist/tfs
```

Then put the resulting binary somewhere in your PATH.

## Usage

Here are a few ways to use the tool (assuming the local folder contains some
Terraform manifest files).

* Downloading and using the right Terraform version while honoring a Terraform
version constrainst:

```
$ tfs
```

If there is no specific constrainst, `tfs` will activate the most recent
Terraform version that has been downloaded so far.

* Using a specific Terraform version:

```
$ tfs 1.3.2
```

* Cleaning up the cache:

```
$ tfs prune
```

* Cleaning up useless versions:

```
$ tfs prune-until 1.3.0
```

By default, Terraform binaries will be stored in `${XDG_CACHE_HOME}/tfs`, else in
`${HOME}/.cache/tfs`.

A symbolic link to the active Terraform binary is created in `${HOME}/.local/bin`.