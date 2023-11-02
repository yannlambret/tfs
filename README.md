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

When downloading a Terraform binary, an SHA256 checksum of the file is
automatically performed based on the information given by HashiCorp
(releases are by default fetched from https://releases.hashicorp.com).

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

A symbolic link to the active Terraform binary is created in `${HOME}/.local/bin`,
so this directory should be added to your PATH environment variable.

## Configuration

You can change the behavior of the tool by setting up a configuration file.
By default, the configuration file PATH will be equivalent to
`$XDG_CONFIG_HOME/tfs/config.yaml`, `$HOME/.config/tfs/config.yaml` otherwise.

Here is a configuration template with the supported values:

```yaml
# -- Download URL

# Change this if you need to download release files from a specific location.
#terraform_download_url: "https://releases.hashicorp.com" # default value

# -- Cache management

# Cache directory for Terraform release files.
# Default value: "${XDG_CACHE_HOME}/tfs"
# Fallback value: "${HOME}/.cache/tfs"
#cache_directory: <CUSTOM_PATH>

# Keep a limited number of release files in the cache.
cache_auto_clean: true # default value

# Number of Terraform releases that you want to keep.
# Most recent releases will be kept in the cache.
cache_history: 8 # default value

# Slightly more sophisticated cache management.
# Keep a specific number of Terraform releases
# per minor version (as usual, most recent ones
# will be kept).
# So for instance, with the values defined below,
# the cache could contain the folloging releases:
#   * 1.3.6
#   * 1.3.8
#   * 1.4.5
#   * 1.4.6
#   * 1.5.0
# When these two directives are commented out,
# the option 'cache_history' is ignored.
#cache_minor_version_nb: 3
#cache_patch_version_nb: 2
```

Sometimes, an additional version can be kept in the cache, if you intend to use
a release that should have been removed otherwise.

Consider this scenario (using the same configuration settings as above):

```
$ tfs prune # cache is now empty
```

```
$ tfs 1.5.2
$ tfs 1.5.3
```

The cache now contains the following releases:

```
$ ls -1 ~/.cache/tfs
terraform_1.5.2
terraform_1.5.3
```

Now let's download the 1.5.4 version:

```
$ tfs 1.5.4
```

As we should expect, the version 1.5.2 has been removed because we want to keep
at most two patch versions:

```
$ ls -1 ~/.cache/tfs
terraform_1.5.3
terraform_1.5.4
```

Now suppose that you need to work with Terraform 1.5.1:

```
$ tfs 1.5.1
```

In this specific scenario, the command should download the 1.5.1 binary and
remove it immediately before exiting. Obviously, this is not what you want
and this is not what happens:

```
$ ls -1 ~/.cache/tfs
terraform_1.5.1
terraform_1.5.3
terraform_1.5.4
```

When downloading a Terraform release, `tfs` will flag it as active to prevent
it from being removed, whatever the cache content may be. If you switch back
later to Terraform 1.5.4, the 1.5.1 version will be removed as usual:

```
$ tfs 1.5.4
$ ls -1 ~/.cache/tfs
terraform_1.5.3
terraform_1.5.4
```