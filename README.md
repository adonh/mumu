<div align="center">

# mumu

**Save and restore window-to-Space layouts on macOS. From the terminal.**

[![Go Version](https://img.shields.io/github/go-mod/go-version/adonh/mumu?style=flat-square&logo=go)](https://github.com/adonh/mumu)
[![License](https://img.shields.io/github/license/adonh/mumu?style=flat-square)](LICENSE)
[![Early Development](https://img.shields.io/badge/status-early%20dev-orange?style=flat-square)](#)

</div>

---

Reconnect your monitors and get every window back on the Space it was on — without relaunching anything.

```bash
mumu save        # remember what's where, for this display setup
mumu restore     # put matching, already-open windows back
```

> **Early development** — CLI and behavior may change between releases.

> [!NOTE]
> mumu's core is based on [mimi](https://github.com/y3owk1n/mimi) but with a focus on layout instead of
> individual windows and spaces.

---

## Install

```bash
brew tap adonh/tap
brew install --cask adonh/tap/mumu
```

Grant **Accessibility** and **Screen Recording** in **System Settings → Privacy & Security**, then start using it immediately. There's no daemon, no background process, and no config file — `mumu` only ever does something when you run one of its commands.

Other options (Nix flake, build from source) → [Installation Guide](docs/INSTALLATION.md)

---

## What mumu does

| You want to…                     | Command             |
| :-------------------------------- | :------------------- |
| Save the current layout           | `mumu save`          |
| Restore the saved layout          | `mumu restore`       |
| List all saved layouts            | `mumu list`          |
| Preview a saved layout            | `mumu show`          |
| Delete a saved layout             | `mumu delete`        |
| Check permission status           | `mumu status`        |

Layouts are saved and looked up by the number of currently connected displays, so plugging in (or unplugging) a monitor automatically finds the right saved arrangement. Restore only ever moves windows belonging to apps that are already running — it never launches anything, and never creates or deletes Spaces. Space numbers count left to right across all your displays, matching how you actually see them laid out, regardless of which display macOS considers "primary."

Full reference → [CLI Guide](docs/CLI.md)

---

## How it works

Window enumeration uses `CGWindowListCopyWindowInfo` (so it sees windows on every Space, not just the currently displayed one), and window-to-Space moves use the private SkyLight API for instant, animation-free relocation. Everything else goes through public Accessibility APIs. No SIP disable is required.

→ [Architecture Guide](docs/ARCHITECTURE.md)

---

## Documentation

| Guide                                      | What's in it                     |
| :----------------------------------------- | :-------------------------------- |
| [Installation](docs/INSTALLATION.md)       | Homebrew, Nix, source, permissions |
| [CLI](docs/CLI.md)                         | Every command and flag            |
| [Architecture](docs/ARCHITECTURE.md)       | How the pieces fit                |
| [Troubleshooting](docs/TROUBLESHOOTING.md) | Common issues and fixes           |
| [Contributing](CONTRIBUTING.md)            | PRs and bug reports               |

---

## Contributing

```bash
just build && just lint && just test
```

→ [Development Guide](docs/DEVELOPMENT.md)

---

## License

MIT — see [LICENSE](LICENSE).

<div align="center">
<br/>

**Try it. Two commands and you're running.**

```bash
brew install --cask adonh/tap/mumu && mumu save
```

</div>
