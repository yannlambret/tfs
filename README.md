![go test](https://github.com/yannlambret/tfs/actions/workflows/test.yml/badge.svg)
[![go report](https://goreportcard.com/badge/github.com/yannlambret/tfs)](https://goreportcard.com/report/github.com/yannlambret/tfs)
![go version](https://img.shields.io/github/go-mod/go-version/yannlambret/tfs?label=go%20version)
[![release](https://img.shields.io/github/v/release/yannlambret/tfs?label=release)](https://github.com/yannlambret/tfs/releases)

# tfs

`tfs` is a command-line tool that helps you manage multiple versions of Terraform efficiently.\
It was inspired by [this project](https://github.com/warrensbox/terraform-switcher).

`tfs` is simple, lightweight, and follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

It works out of the box on all GNU/Linux distributions and macOS.

---

## Build

```bash
git clone https://github.com/yannlambret/tfs.git && cd tfs
go build -o dist/tfs
```

Then place the resulting binary somewhere in your `PATH`.

---

## Usage

Here are a few examples of how to use the tool (assuming the current directory contains some Terraform manifest files):

### 🔄 Automatically use the appropriate Terraform version

```bash
tfs
```

Note: `tfs` now uses HashiCorp’s `hc-install` library to automatically download and install Terraform.\
It is thus no longer possible to download Terraform from alternative sources.

If no version constraint is detected, `tfs` will activate the most recently downloaded Terraform version.

### 📌 Use a specific Terraform version
If no Terraform version constraint is specified in the configuration, you can manually select
the version to use:

```bash
tfs 1.10.1
```

### 📂 List cached versions

```bash
tfs list
```

### 🧹 Clear the entire cache

```bash
tfs prune
```

### 🗑️ Remove versions older than a specific one

```bash
tfs prune-until 1.8.0
```

---

## Caching & Paths

By default, Terraform binaries are stored in `${XDG_CACHE_HOME}/tfs`\
If `XDG_CACHE_HOME` is not set, it defaults to `${HOME}/.cache/tfs`.

A symbolic link to the active Terraform binary is created at `${HOME}/.local/bin/terraform`,\
so make sure this directory is added to your `PATH`.

---

## Configuration

`tfs` supports a configuration file to customize behavior.\
By default, the configuration file is located at:

```
$XDG_CONFIG_HOME/tfs/config.yaml
```

If `XDG_CONFIG_HOME` is not set, it falls back to:

```
$HOME/.config/tfs/config.yaml
```

### Configuration Template

```yaml
# -- Cache Management

# Custom path for the Terraform cache directory.
# Default: "${XDG_CACHE_HOME}/tfs"
# Fallback: "${HOME}/.cache/tfs"
#cache_directory: <CUSTOM_PATH>

# Enable automatic cache cleanup.
cache_auto_clean: true # default value

# Maximum number of releases to keep in the cache (fallback mode).
cache_history: 8 # default value

# Advanced cache management:
# Keep a limited number of releases per minor version,
# and a limited number of patch versions within each minor.
#
# For example, with the config below, you might keep:
#   * 1.9.3
#   * 1.9.5
#   * 1.10.2
#   * 1.10.3
#   * 1.11.0
#
# When both values are defined, cache_history is ignored.
#cache_minor_version_nb: 3
#cache_patch_version_nb: 2
```

---

## Advanced Cache Behavior

`tfs` tries to preserve the current version in use — even if it would otherwise be cleaned up.

Consider this scenario (using the config above):

```bash
$ tfs prune  # The cache is now empty

$ tfs 1.10.2
$ tfs 1.10.3
```

Now the cache contains:

```bash
$ tfs list
1.10.2
1.10.3 (active)
```

Let’s install one more version:

```bash
$ tfs 1.10.4
```

Since you want to keep at most two patch versions, `1.10.2` is removed:

```bash
$ tfs list
1.10.3
1.10.4 (active)
```

Now, suppose you need to work with `1.10.1`:

```bash
$ tfs 1.10.1
```

Even though `1.10.1` falls outside the patch retention window, `tfs` will **not** delete it immediately, because it’s now the **active** version.

```bash
$ tfs list
1.10.1 (active)
1.10.3
1.10.4
```

But as soon as you switch again:

```bash
$ tfs 1.10.4
$ tfs list
1.10.3
1.10.4 (active)
```

Now `1.10.1` is cleaned up, as expected.

---

## License

MIT © [Yann Lambret](https://github.com/yannlambret)
