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

## Build & Install

```bash
git clone https://github.com/yannlambret/tfs.git && cd tfs
go mod tidy
go build -o build/tfs
```
Alternatively, you can download a prebuilt binary for your platform from the [Releases page](https://github.com/yannlambret/tfs/releases). \
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

### 🎯 Respect Terraform version constraints (including `~>`)

When you run `tfs` without specifying a version, it inspects Terraform configuration files
in the current directory to infer which version to activate.

`tfs` understands standard comparison operators (`=`, `!=`, `>`, `>=`, `<`, `<=`) and also
supports the Terraform / HashiCorp pessimistic operator `~>` (sometimes called the "compatible with" operator).

Supported forms and their expansion:

```
~> 1        => >=1.0.0, <2.0.0
~> 1.2      => >=1.2.0, <1.3.0
~> 1.2.3    => >=1.2.3, <1.3.0
```

Multiple constraints are ANDed (`,` separator) and any `||` segments (if present) act as OR, following the semantics of the underlying `hashicorp/go-version` library.

Examples:

```
# Accept any 1.2.x (but not 1.3.0):
required_version = "~> 1.2"

# Accept patch upgrades starting at 1.2.5 (but still < 1.3.0):
required_version = "~> 1.2, >= 1.2.5"

# Accept any 1.x:
required_version = "~> 1"

# Mixed with other operators:
required_version = ">= 1.5.0, ~> 1.6"
```

Invalid uses (e.g. `~> 1.alpha`, `~> 1..2`, `~> ~> 1.2`) are rejected and will cause `tfs` to report an error instead of choosing a wrong version silently.

The `~>` constraints are internally expanded before being passed to the version resolver; this ensures consistent behavior without pulling additional parsing libraries.

> Tip: If no constraint is found, `tfs` simply activates the most recently downloaded Terraform version.

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
