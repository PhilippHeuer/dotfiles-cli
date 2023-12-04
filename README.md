# DotFiles CLI

> DotFiles CLI (**dotfiles**) is a simple cli to install your dotfiles into your home directory.

## Features

- Automatically cleans up files that are not tracked anymore (e.g. deleted from dotfiles)
- Support for copying or symlinking files
- Support for custom directories
- Clean command to remove all files that are tracked

## Installation

Download the binary from the [GitHub Releases](https://github.com/PhilippHeuer/dotfiles-cli/releases).

```bash
curl -L -o ~/.local/bin/dotfiles https://github.com/PhilippHeuer/dotfiles-cli/releases/download/v0.1.0/linux_amd64
chmod +x ~/.local/bin/dotfiles
```

## Usage

| Command                                               | Description                                                              |
|-------------------------------------------------------|--------------------------------------------------------------------------|
| `dotfiles install ~/git/your-dotfiles --mode symlink` | Installs all files by creating symlinks                                  |
| `dotfiles install ~/git/your-dotfiles --mode copy`    | Installs all files by making copies                                      |
| `dotfiles clean`                                      | Cleans all files tracked in dotfiles install state (keeping directories) |

> The `--mode` flag is optional, the default is `copy`.

## Configuration

Your `~/git/your-dotfiles` should contain a `dotfiles.yaml` file, which defines the files to install.

```yaml
directories:
- path: config
  target: $HOME/.config
- path: scripts
  target: $HOME/.local/bin
```

## License

Released under the [MIT license](./LICENSE).
