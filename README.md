# DotFiles CLI

> DotFiles CLI (**dotfiles**) is a powerful command-line tool designed to seamlessly manage and install your dotfiles into your home directory.

## Features

- **Installation Methods**: Choose between copying or creating symbolic links for your dotfiles.
- **Rule-Based Installation**: Install files based on rules, e.g., only install files for installed software. [Configuration](#configuration)
- **Theme Support**: Install different configuration files based on defined themes. [Theme Support](#theme-support)
- **Template Processing**: Leverage Go templating for dynamic file content. [Template Processing](#template-processing)
- **Automatic Cleanup**: Automatically remove files that are not tracked anymore, keeping your home directory clean.

## Installation

Download the binary from the [GitHub Releases](https://github.com/PhilippHeuer/dotfiles-cli/releases).

```bash
curl -L -o ~/.local/bin/dotfiles https://github.com/PhilippHeuer/dotfiles-cli/releases/latest/download/linux_amd64
chmod +x ~/.local/bin/dotfiles
```

## Usage

| Command                                      | Description                                                        |
|----------------------------------------------|--------------------------------------------------------------------|
| `dotfiles install ~/dotfiles --mode symlink` | Installs files by creating symlinks                                |
| `dotfiles install ~/dotfiles --mode copy`    | Installs files by making copies                                    |
| `dotfiles query FontFamily`                  | Queries the state file for app information (e.g. theme properties) |
| `dotfiles clean`                             | Cleans all tracked files, keeping directories (from state)         |

After the first installation, you can run the `dotfiles install` command without the source directory as it is stored in the app state.

> The `--mode` flag is optional, the default is `copy`.

## Configuration

Your `~/dotfiles` repository needs to contain a `dotfiles.yaml` file, which defines the configuration for all directories.
You can make use of rules to only install files based on the installed software.

### Structure Overview

```yaml
# include additional config files (merged, supports ~ and env vars)
includes:
  - ~/.dotfiles-overlay.yaml

# directories to install
directories:
  - path: config/alacritty                  # source path (relative to dotfiles source)
    target: $HOME/.config/alacritty          # destination path
    mode: symlink                            # optional: override global mode (copy, symlink)
    rules:
    - rule: inPath("alacritty")
    templateFiles:                           # optional: files to process with Go templates
      - config/alacritty/alacritty.toml
    themeFiles:                              # optional: theme-dependent file symlinks
      - target: $HOME/.config/alacritty/themes/current.toml
        sources:
          catppuccin-mocha: themes/catppuccin-mocha.toml
          rose-pine: themes/rose-pine.toml
    linkFiles:                               # optional: individual file symlinks with fallback paths
      - paths:
          - ~/.config/oracle/personal/tnsnames.ora
          - oracle/global/tnsnames.ora
        target: tnsnames.ora
        mode: symlink
```

For all available rules, see the [Rule Reference](#rule-reference).

### `linkFiles` — File Symlinks with Fallback

Define individual file symlinks with ordered fallback sources. Paths starting with `~/` or `/` are used as-is; relative paths are resolved relative to the directory's source path (for `paths`) or target path (for `target`):

```yaml
- path: oracle
  target: $HOME/.config/oracle
  linkFiles:
    - paths:
        - ~/.config/oracle/personal/tnsnames.ora   # absolute (home dir override)
        - global/tnsnames.ora                        # relative to oracle/ source dir
      target: tnsnames.ora                           # relative to $HOME/.config/oracle
      mode: symlink
```

### `includes` — Config Merging

Include and merge additional YAML config files (absolute path or relative to the config file's directory):

```yaml
includes:
  - machine-specific.yaml
  - /home/user/.dotfiles/secrets.yaml
```

## Theme Support

You can specify themes for your dotfiles, which can be used to copy/link files based on the selected theme.
The following example shows a simple theme configuration for Alacritty, you can import `$HOME/.config/alacritty/themes/current.toml` in your Alacritty configuration.

```yaml
themes:
- name: catppuccin-mocha
  font-family: JetBrainsMono Nerd Font Mono
  font-size: 11
  properties:
    yourCustomProperty: value

directories:
- path: config/alacritty
  target: $HOME/.config/alacritty
  rules:
  - rule: inPath("alacritty")
  themeFiles:
  # allows you to import the themes/current.toml from your config and symlink the content based on the theme
  - target: $HOME/.config/alacritty/themes/current.toml
    sources:
      catppuccin-mocha: themes/catppuccin-mocha.toml
      rose-pine: themes/rose-pine.toml
      nord: themes/nord.toml
      tokyo-night: themes/tokyo-night.toml
```

## Template Processing

You can toggle template processing by setting the `templateFiles` property in your configuration, files will always be copied regardless of the mode (`copy`, `symlink`, ...).

```yaml
themes:
- name: catppuccin-mocha
  font-family: JetBrainsMono Nerd Font Mono
  font-size: 11
  properties:
    yourCustomProperty: value

directories:
- path: config/alacritty
  target: $HOME/.config/alacritty
  rules:
  - rule: inPath("alacritty")
  templateFiles: # allows the use of {{ .FontFamily }}, {{ .YourCustomProperty }} and other variables in your config
   - config/alacritty/alacritty.toml
```

The following values are available for templating: `Name`, `ColorScheme`, `WallpaperDir`, `FontFamily`, `FontSize`, `GtkTheme`, `CosmicTheme`, `IconTheme`, `CursorTheme`.
Additionally, any value you define in the theme properties will be available (in CamelCase).

## Rule Reference

The rules make use of [cel-go](https://github.com/google/cel-go) expressions, additionally the following functions are available:

- `inPath("alacritty")`: Checks if path contains the given executable.
- TODO: document all functions

## License

Released under the [MIT license](./LICENSE).
