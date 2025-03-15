# rconf

**rconf** is a command-line tool that executes local shell-scripts on multiple remote hosts via SSH.

## Features

- Execute shell scripts on multiple remote hosts via SSH
- Concurrent execution with worker limits
- Structured logging
- Secure authentication using SSH private keys
- Automatic script upload and execution
- Summary table of execution results

## Installation

TODO

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

## Logging

All execution details, including errors, are logged to the specified log file (`ssh_execution.log`).

## Requirements

- Go 1.18+
- SSH access to remote hosts
- Scripts must be shell scripts (`.sh`)

## License

TODO

## Contributing

TODO

## TODO

- Support for `ssh://user:pass@host:port` connection strings
- Support for password authentication
- Configurable `sudo` behavior
- Parallel execution optimizations
- Integration tests
- github-actions (CI, release, etc...)

