# Git Identitree

A CLI tool to manage multiple Git identities and automatically switch between them based on directory context using Git's native `includeIf` mechanism.

## Installation

```bash
go build -o gidtree ./cmd/gidtree
```

Or install directly:

```bash
go install ./cmd/gidtree
```

## Usage

### Initialization

First, initialize Git Identitree:

```bash
gidtree init
```

This creates the `~/.gidtree/` directory and `profiles.yaml` file.

### Profile Management

#### Create a Profile

Interactively create a new profile:

```bash
gidtree profile create
```

You'll be prompted for:
- Profile Name (required)
- Email (required)
- SSH Key Path (optional)
- GPG Key ID (optional)

#### List Profiles

Display all stored profiles:

```bash
gidtree profile list
```

#### Delete a Profile

Delete a profile (will fail if the profile is mapped to any directories):

```bash
gidtree profile delete <name>
```

### Directory Mapping

#### Map a Profile to a Directory

Associate a profile with a directory:

```bash
gidtree map <profile> <directory>
```

This will:
1. Create or update `~/.gitconfig-<profile>` with the profile's settings
2. Add an `includeIf` block to `~/.gitconfig` that activates the profile when working in that directory

Example:
```bash
gidtree map work ~/projects/work
```

#### Unmap a Directory

Remove the association between a directory and its profile:

```bash
gidtree unmap <directory>
```

### Status

View current mappings and active profile:

```bash
gidtree status
```

This shows:
- Current directory and active profile (if any)
- All directory mappings
- Git config file status

### SSH Key Management

#### Load SSH Key

Manually load the SSH key for a profile:

```bash
gidtree ssh load <profile>
```

#### Unload SSH Key

Manually unload the SSH key for a profile:

```bash
gidtree ssh unload <profile>
```

### Auto-Activation

Automatically detect the current directory and load the appropriate SSH key:

```bash
gidtree activate
```

This command:
1. Detects the current working directory
2. Finds the mapped profile (if any)
3. Loads the associated SSH key into the SSH agent

You can add this to your shell profile to run automatically when changing directories.

## How It Works

Git Identitree uses Git's native `includeIf` conditional include feature to automatically switch profiles based on directory context.

When you map a profile to a directory, the tool:

1. Creates a profile-specific config file at `~/.gitconfig-<profile>` containing:
   - `user.name`
   - `user.email`
   - `user.signingkey` (if GPG key is configured)
   - `core.sshCommand` (if SSH key is configured)

2. Adds a conditional include to `~/.gitconfig`:
   ```ini
   [includeIf "gitdir/i:/absolute/path/to/directory/"]
       path = ~/.gitconfig-<profile>
   ```

Git automatically loads the appropriate config when you're working in that directory tree.

## File Structure

- `~/.gidtree/profiles.yaml` - Stores all profile definitions
- `~/.gitconfig` - Main Git config (modified with includeIf blocks)
- `~/.gitconfig-<profile>` - Profile-specific config files

## Safety Features

- Profile deletion is prevented if the profile is mapped to any directories
- Path normalization ensures consistent operation across shells
- SSH key paths are validated before use
- Git config modifications are made safely

## Examples

### Work and Personal Profiles

```bash
# Create work profile
gidtree profile create
# Name: work
# Email: john@company.com
# SSH Key: ~/.ssh/id_rsa_work
# GPG Key: ABC123DEF456

# Create personal profile
gidtree profile create
# Name: personal
# Email: john@personal.com
# SSH Key: ~/.ssh/id_rsa_personal

# Map work profile to work directory
gidtree map work ~/projects/work

# Map personal profile to personal directory
gidtree map personal ~/projects/personal

# Check status
gidtree status
```

Now, when you `cd` into `~/projects/work`, Git will automatically use your work profile. When you `cd` into `~/projects/personal`, it will use your personal profile.

## License

MIT

