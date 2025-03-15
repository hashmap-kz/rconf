# rconf

**rconf** is a command-line tool that executes local shell-scripts on multiple remote hosts via SSH.

---

## Features

- Execute multiple shell scripts on multiple remote hosts via SSH
- No complex configs, no intricate YAML, no DSLs - just plain shell and a single binary
- Concurrent execution with worker limits
- Structured logging
- Secure authentication using SSH private keys
- Automatic script upload and execution
- Summary table of execution results

---

## Installation

1. Download the latest binary for your platform from
   the [Releases page](https://github.com/hashmap-kz/rconf/releases).
2. Place the binary in your system's `PATH` (e.g., `/usr/local/bin`).

### Example installation script for Unix-Based OS _(requirements: tar, curl, jq)_:

```bash
(
set -euo pipefail

OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed -e 's/x86_64/amd64/' -e 's/\(arm\)\(64\)\?.*/\1\2/' -e 's/aarch64$/arm64/')"
TAG="$(curl -s https://api.github.com/repos/hashmap-kz/rconf/releases/latest | jq -r .tag_name)"

curl -L "https://github.com/hashmap-kz/rconf/releases/download/${TAG}/rconf_${TAG}_${OS}_${ARCH}.tar.gz" |
tar -xzf - -C /usr/local/bin && \
chmod +x /usr/local/bin/rconf
)
```

---

## Usage

```sh
rconf \
  --user myuser \
  --key /path/to/private_key \
  --scripts /path/to/script1.sh,/path/to/script-dir/ \
  --hosts 10.40.240.193,10.40.240.189 \
  --workers 5 \
  --log execution.log
```

### Flags

| Flag        | Short | Description                                                    |
|-------------|-------|----------------------------------------------------------------|
| `--user`    | `-u`  | SSH username (required)                                        |
| `--key`     | `-k`  | Path to SSH private key (required)                             |
| `--scripts` | `-s`  | Comma-separated list of script paths or directories (required) |
| `--hosts`   | `-H`  | Comma-separated list of remote hosts (required)                |
| `--workers` | `-w`  | Maximum concurrent SSH connections (default: 2)                |
| `--log`     | `-l`  | Log file path (default: `ssh_execution.log`)                   |

## How It Works

1. The tool reads the provided scripts into memory.
2. It establishes SSH and SFTP connections to each host.
3. The scripts are uploaded to the remote host's `/tmp/` directory.
4. The scripts are executed remotely using `sudo`.
5. Execution results are stored and displayed in a summary table.

---

## Example Output

```plaintext
🚀 Starting script execution...
[HOST: 10.40.240.189] 🔄 Connecting...
[HOST: 10.40.240.193] 🔄 Connecting...
[HOST: 10.40.240.193] ⏳ Uploading scripts\00-packages.sh...
[HOST: 10.40.240.193] 🚀 Executing scripts\00-packages.sh...
[HOST: 10.40.240.189] ⏳ Uploading scripts\00-packages.sh...
[HOST: 10.40.240.189] 🚀 Executing scripts\00-packages.sh...
[HOST: 10.40.240.193] ✅ Successfully executed scripts\00-packages.sh
[HOST: 10.40.240.193] ⏳ Uploading scripts\01-timezone.sh...
[HOST: 10.40.240.193] 🚀 Executing scripts\01-timezone.sh...
[HOST: 10.40.240.189] ✅ Successfully executed scripts\00-packages.sh
[HOST: 10.40.240.189] ⏳ Uploading scripts\01-timezone.sh...
[HOST: 10.40.240.189] 🚀 Executing scripts\01-timezone.sh...
[HOST: 10.40.240.193] ✅ Successfully executed scripts\01-timezone.sh
[HOST: 10.40.240.193] ⏳ Uploading scripts\02-locale.sh...
[HOST: 10.40.240.193] 🚀 Executing scripts\02-locale.sh...
[HOST: 10.40.240.189] ✅ Successfully executed scripts\01-timezone.sh
[HOST: 10.40.240.189] ⏳ Uploading scripts\02-locale.sh...
[HOST: 10.40.240.189] 🚀 Executing scripts\02-locale.sh...
[HOST: 10.40.240.189] ✅ Successfully executed scripts\02-locale.sh
[HOST: 10.40.240.189] 🔄 Disconnecting...
[HOST: 10.40.240.193] ✅ Successfully executed scripts\02-locale.sh
[HOST: 10.40.240.193] 🔄 Disconnecting...

=== Execution Summary ===
HOST            RESULT
10.40.240.189  Success
10.40.240.193  Success
```

---

## Logging

All execution details, including errors, are logged to the specified log file (`ssh_execution.log`).

---

## Requirements

- Go 1.18+
- SSH access to remote hosts
- Scripts must be shell scripts (`.sh`)

---

## **Contributing**

We welcome contributions! To contribute: see the [Contribution](CONTRIBUTING.md) guidelines.

---

## **License**

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## TODO

- Support for `ssh://user:pass@host:port` connection strings
- Support for password authentication
- Configurable `sudo` behavior
- Parallel execution optimizations
- Integration tests
- github-actions (CI, release, etc...)

