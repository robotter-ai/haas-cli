
# HAAS CLI Tool

A command-line interface tool for automating HAAS (Hardware-as-a-Service) related tasks.

## Prerequisites

- Aleph CLI tool must be installed
- sevctl must be installed

## Installation

1. Clone the repository
2. Run `make` to install dependencies
3. Build the CLI tool

## Usage

### Available Commands

1. **Attach Account**
```bash
haas-cli attach-account --private-key <your-private-key> [--verbose]

# Options:
#   -k, --private-key string   Private key for the account (required)
#   -v, --verbose             Show detailed output
```

2. **Create and Start Instance**
```bash
haas-cli create-and-start --name <instance-name> --secret <vm-secret> [--verbose]

# Options:
#   -n, --name string         Name for the VM instance (required)
#   -s, --secret string       Secret for the VM instance (required)
#   -v, --verbose            Show detailed output
```

3. **Restart Instance**
```bash
haas-cli restart-instance <vm-hash> <secret> [--verbose]

# Options:
#   -v, --verbose            Show detailed output
```

4. **Stop Instance**
```bash
haas-cli stop <vm-hash> [--verbose]

# Options:
#   -v, --verbose            Show detailed output
```

5. **Delete Instance**
```bash
haas-cli delete <vm-hash> [--verbose]

# Options:
#   -v, --verbose            Show detailed output
```

### Examples

1. Attach an account:
```bash
haas-cli attach-account -k "your-private-key-here"
```

2. Create and start a new VM instance:
```bash
haas-cli create-and-start -n "my-instance" -s "my-secret"
```

3. Restart an existing VM instance:
```bash
haas-cli restart-instance abc123hash xyz789secret
```

4. Stop a VM instance:
```bash
haas-cli stop abc123hash
```

5. Delete a VM instance:
```bash
haas-cli delete abc123hash
```

## Default Configuration

The CLI uses the following default values for VM creation:
- Rootfs Hash: eca1a77a489ef53ecb6f6febad0735103c85b6e0176ff707390b5fee939ea824
- Rootfs Size: 10240
- VCPUs: 1
- Memory Size: 2048

## Verbose Mode

Add the `-v` or `--verbose` flag to any command to see detailed output and debug information.

## Error Handling

If any command fails, the CLI will display an error message and exit with status code 1.
