# Installation Guide

This guide covers installation methods for `mumu` on macOS.

---

## Table of Contents

- [Requirements](#requirements)
- [Method 1: Homebrew (Recommended)](#method-1-homebrew-recommended)
- [Method 2: Nix Flake](#method-2-nix-flake)
- [Method 3: From Source](#method-3-from-source)
- [Post-Installation](#post-installation)
- [Shell Completions](#shell-completions)
- [Troubleshooting](#troubleshooting)
- [Uninstallation](#uninstallation)

---

## Requirements

- macOS 14.0 or later
- Accessibility permissions (granted during setup)
- Screen Recording permissions (required for `mumu save`/`restore`)

There is no daemon, background process, or config file to set up — `mumu` is a plain CLI that only runs when you invoke a command.

---

## Method 1: Homebrew (Recommended)

> [!NOTE]
> The Homebrew tap is maintained in another repo: [adonh/homebrew-tap](https://github.com/adonh/homebrew-tap).
> If there's a problem with the tap, please open an issue in that repo or even better, a PR.

```bash
brew tap adonh/tap
brew install --cask adonh/tap/mumu
```

---

## Method 2: Nix Flake

`mumu` is available as a Nix flake exposing a plain package (no service modules, since there's no daemon to run).

> `pkgs.mumu` uses the published release zip; `pkgs.mumu-source` builds from source.

### Add Flake Input

```nix
# flake.nix
{
  inputs = {
     # ... other inputs
     mumu.url = "github:adonh/mumu"; # or "https://flakehub.com/f/adonh/mumu/0.1"
     # ... other inputs
  };
}
```

### Install via Overlay

```nix
{
  outputs = { self, nixpkgs, mumu, ... }: {
     darwinConfigurations.your-hostname = nix-darwin.lib.darwinSystem {
       modules = [
         {
           nixpkgs.overlays = [ mumu.overlays.default ];
           environment.systemPackages = [ pkgs.mumu ];
         }
       ];
     };
  };
}
```

Or install the package output directly:

```nix
{
  outputs = { self, nixpkgs, mumu, ... }: {
     darwinConfigurations.your-hostname = nix-darwin.lib.darwinSystem {
       modules = [
         {
           environment.systemPackages = [
             mumu.packages.aarch64-darwin.default
           ];
         }
       ];
     };
  };
}
```

Or with home-manager:

```nix
{
  home.packages = [ mumu.packages.${system}.default ];
}
```

### Updating

```bash
nix flake update mumu
# Then rebuild your system/home configuration
```

---

## Method 3: From Source

### Requirements

- Go 1.26+
- Xcode Command Line Tools
- [just](https://github.com/casey/just) command runner

### Build

```bash
git clone https://github.com/adonh/mumu.git
cd mumu

# Build CLI
just release
mv ./bin/mumu /usr/local/bin/mumu

# Or build app bundle
just bundle
mv ./build/Mumu.app /Applications/Mumu.app
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed build options.

---

## Post-Installation

### 1. Grant Permissions

Open **System Settings → Privacy & Security → Accessibility → Add mumu**, then **System Settings → Privacy & Security → Screen Recording → Add mumu**. Both are required for `mumu save` and `mumu restore` (Screen Recording lets it read window titles reliably for matching). Check either permission's status anytime with:

```bash
mumu status
```

### 2. Verify

```bash
mumu --version
mumu save
```

That's it — there's no service to start and nothing else to configure.

---

## Shell Completions

`mumu` provides shell completions for bash, zsh, and fish.

### Bash

```bash
mumu completion bash > /usr/local/etc/bash_completion.d/mumu
```

### Zsh

```bash
mumu completion zsh > "${fpath[1]}/_mumu"
```

### Fish

```bash
mumu completion fish > ~/.config/fish/completions/mumu.fish
```

---

## Troubleshooting

### "mumu wants to control this computer using accessibility features"

This is normal. Click **OK** and grant permissions in System Settings.

### Command not found: mumu

If using the CLI build, ensure the binary is in your PATH:

```bash
# Add to ~/.zshrc or ~/.bashrc
export PATH="/usr/local/bin:$PATH"
```

### Permission denied

Make the binary executable:

```bash
chmod +x /usr/local/bin/mumu
```

### App won't open (macOS quarantine)

macOS may quarantine apps from unidentified developers:

```bash
xattr -cr /Applications/Mumu.app
```

Then try opening again.

### Nix build fails

Ensure you're on an Apple Silicon Mac (arm64). For Intel Macs, change the URL to:

```nix
url = "https://github.com/adonh/mumu/releases/download/v${version}/mumu-darwin-amd64.zip";
```

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for permission and restore-matching issues.

---

## Uninstallation

### Homebrew

```bash
brew uninstall --cask adonh/tap/mumu
```

### Manual

```bash
# Remove app bundle
rm -rf /Applications/Mumu.app

# Remove CLI
rm /usr/local/bin/mumu

# Remove saved layouts and settings
rm -rf ~/Library/Application\ Support/mumu
# (or $XDG_DATA_HOME/mumu and $XDG_CONFIG_HOME/mumu, if those are set)
```

### Nix

Remove the input/overlay from your configuration and rebuild.
