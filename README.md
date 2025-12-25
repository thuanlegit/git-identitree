# Git Identitree

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://golang.org)

A powerful CLI tool to manage multiple Git identities and automatically switch between them based on directory context using Git's native `includeIf` mechanism.

## Features

✨ **Multiple Profiles** - Create and manage unlimited Git profiles (work, personal, open-source, etc.)

🔄 **Automatic Switching** - Profiles activate automatically based on directory context

🔑 **SSH Key Management** - Integrated SSH key loading/unloading per profile

🔐 **GPG Signing Support** - Configure GPG keys for commit signing

📝 **Custom Author Names** - Separate profile names from Git author names

🎨 **Interactive UI** - Beautiful TUI for profile creation and viewing

⚡ **Shell Completion** - Tab completion for bash, zsh, fish, and PowerShell

🛡️ **Safe Operations** - Prevents accidental deletion of mapped profiles with auto-unmap option

## Installation

### Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/thuanlegit/git-identitree/releases):

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/thuanlegit/git-identitree/releases/latest/download/gidtree-darwin-arm64 -o gidtree
chmod +x gidtree
sudo mv gidtree /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/thuanlegit/git-identitree/releases/latest/download/gidtree-darwin-amd64 -o gidtree
chmod +x gidtree
sudo mv gidtree /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/thuanlegit/git-identitree/releases/latest/download/gidtree-linux-amd64 -o gidtree
chmod +x gidtree
sudo mv gidtree /usr/local/bin/
```

#### Windows
Download `gidtree-windows-amd64.exe` from the [releases page](https://github.com/thuanlegit/git-identitree/releases) and add it to your PATH.

### Build from Source

Requires Go 1.24 or later:

```bash
go install github.com/thuanlegit/git-identitree/cmd/gidtree@latest
```

Or clone and build:

```bash
git clone https://github.com/thuanlegit/git-identitree.git
cd git-identitree
go build -o gidtree ./cmd/gidtree
sudo mv gidtree /usr/local/bin/
```

## Quick Start

### 1. Initialize Git Identitree

```bash
gidtree init
```

This creates the `~/.gidtree/` directory and `profiles.yaml` file.

### 2. Create Your First Profile

```bash
gidtree profile create
```

You'll be prompted for:
- **Profile Name** (required) - e.g., "work"
- **Email** (required) - e.g., "you@company.com"
- **Author Name** (optional) - defaults to profile name
- **SSH Key Path** (optional) - e.g., "~/.ssh/id_rsa_work"
- **GPG Key ID** (optional) - for signed commits

### 3. Map Profile to a Directory

```bash
gidtree map work ~/projects/work
```

### 4. Verify It Works

```bash
cd ~/projects/work
git config user.email  # Shows: you@company.com
git config user.name   # Shows: work (or your custom author name)
```

That's it! Git will now automatically use your work profile whenever you're in `~/projects/work` or any subdirectory.

## Usage

### Profile Management

#### Create a Profile
```bash
gidtree profile create
```

Interactive form with autocomplete for SSH key paths.

#### List All Profiles
```bash
gidtree profile list
```

Beautiful TUI showing all profiles with their settings.

#### Update a Profile
```bash
gidtree profile update <name>
```

Update an existing profile with pre-populated values.

#### Delete a Profile
```bash
gidtree profile delete <name>
```

If the profile is mapped to directories, you'll be prompted to automatically unmap them first.

### Directory Mapping

#### Map Profile to Directory
```bash
gidtree map <profile> <directory>
```

Example:
```bash
gidtree map work ~/projects/work
gidtree map personal ~/projects/personal
gidtree map opensource ~/oss
```

#### Unmap a Directory
```bash
gidtree unmap <directory>
```

#### View Status
```bash
gidtree status
```

Shows all mappings and which profile is active in the current directory.

### SSH Key Management

#### Load SSH Key for Profile
```bash
gidtree ssh load <profile>
```

#### Unload SSH Key
```bash
gidtree ssh unload <profile>
```

#### Auto-Activate
```bash
gidtree activate
```

Detects current directory and loads the appropriate SSH key automatically.

### Shell Completion

Enable tab completion for your shell:

#### Bash
```bash
gidtree completion bash > /etc/bash_completion.d/gidtree
```

#### Zsh
```bash
gidtree completion zsh > "${fpath[1]}/_gidtree"
```

#### Fish
```bash
gidtree completion fish > ~/.config/fish/completions/gidtree.fish
```

#### PowerShell
```powershell
gidtree completion powershell | Out-String | Invoke-Expression
```

### Version
```bash
gidtree version
```

## How It Works

Git Identitree uses Git's native `includeIf` conditional include feature to automatically switch profiles based on directory context.

When you map a profile to a directory, the tool:

1. **Creates a profile-specific config** at `~/.gitconfig-<profile>` containing:
   - `user.name` (from author name or profile name)
   - `user.email`
   - `user.signingkey` (if GPG key is configured)
   - `core.sshCommand` (if SSH key is configured)

2. **Adds a conditional include** to `~/.gitconfig`:
   ```ini
   [includeIf "gitdir/i:/absolute/path/to/directory/"]
       path = ~/.gitconfig-<profile>
   ```

Git automatically loads the appropriate config when you're working in that directory tree.

## File Structure

```
~/.gidtree/
└── profiles.yaml          # All profile definitions

~/.gitconfig               # Main Git config (with includeIf blocks)
~/.gitconfig-work          # Work profile settings
~/.gitconfig-personal      # Personal profile settings
```

## Examples

### Multiple Profiles Setup

```bash
# Initialize
gidtree init

# Create work profile
gidtree profile create
# Name: work
# Email: john@company.com
# Author Name: John Doe
# SSH Key: ~/.ssh/id_rsa_work
# GPG Key: ABC123DEF456

# Create personal profile
gidtree profile create
# Name: personal
# Email: john@personal.com
# Author Name: John Doe
# SSH Key: ~/.ssh/id_rsa_personal

# Create open-source profile
gidtree profile create
# Name: opensource
# Email: john@opensource.org
# Author Name: John Doe (OSS Contributor)
# SSH Key: ~/.ssh/id_rsa_oss

# Map directories
gidtree map work ~/work
gidtree map personal ~/personal
gidtree map opensource ~/oss

# View everything
gidtree profile list
gidtree status
```

### Using Different Author Names

```bash
# Create a profile with custom author name
gidtree profile create
# Name: company-work
# Email: john.doe@company.com
# Author Name: John A. Doe
# (Profile name is 'company-work' but commits show 'John A. Doe')
```

### SSH Key Auto-Loading (Optional)

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Auto-load SSH keys when changing directories
cd() {
  builtin cd "$@" && gidtree activate 2>/dev/null
}
```

## Safety Features

- ✅ Profile deletion is blocked if the profile is mapped to any directories
- ✅ Option to automatically unmap all directories when deleting
- ✅ Path normalization ensures consistent operation across shells
- ✅ SSH key paths support `~` expansion and are validated before use
- ✅ Git config modifications are made safely with proper error handling
- ✅ Profile name cannot be changed after creation (prevents mapping conflicts)

## Advanced Usage

### Update Profile Settings

```bash
# Update email or SSH key for existing profile
gidtree profile update work
```

### Manage Mappings

```bash
# Map multiple directories to same profile
gidtree map work ~/projects/client1
gidtree map work ~/projects/client2

# View all mappings
gidtree status

# Unmap specific directory
gidtree unmap ~/projects/client1
```

### Delete Profile with Mappings

```bash
gidtree profile delete work
# Profile 'work' is mapped to the following directories:
#   - /home/user/projects/client1/
#   - /home/user/projects/client2/
#
# Do you want to unmap all directories and delete the profile? (y/N): y
#
# Unmapping directories...
#   ✓ Unmapped: /home/user/projects/client1/
#   ✓ Unmapped: /home/user/projects/client2/
#
# ✓ Profile 'work' deleted successfully
```

## Troubleshooting

### Profile Not Switching

1. Check if directory is mapped:
   ```bash
   gidtree status
   ```

2. Verify Git config:
   ```bash
   cat ~/.gitconfig | grep includeIf
   ```

3. Test with:
   ```bash
   cd /path/to/mapped/directory
   git config user.email
   ```

### SSH Key Not Loading

1. Ensure SSH agent is running:
   ```bash
   eval "$(ssh-agent -s)"
   ```

2. Manually load key to test:
   ```bash
   gidtree ssh load <profile>
   ```

3. Check key exists:
   ```bash
   ls -la ~/.ssh/
   ```

### Tilde (~) Path Issues

Git Identitree automatically expands `~` in paths. Always use:
- ✅ `~/.ssh/id_rsa_work`
- ❌ `/home/username/.ssh/id_rsa_work`

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- UI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Huh](https://github.com/charmbracelet/huh)
- Inspired by the need to manage multiple Git identities seamlessly

## Support

- 🐛 [Report a Bug](https://github.com/thuanlegit/git-identitree/issues)
- 💡 [Request a Feature](https://github.com/thuanlegit/git-identitree/issues)
- 📖 [Documentation](https://github.com/thuanlegit/git-identitree)

---

**Note:** Replace `thuanlegit` in all GitHub URLs with your actual GitHub username before publishing.
