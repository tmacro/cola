# COLA - Container Linux Appliance

COLA is set of tools to build customized Flatcar Container Linux configurations and images.
It includes a transpiler to convert a high-level HCL based configuration to a low-level Ignition configuration.
It is still in early development and breaking changes are to be expected.

## Example

```hcl
# machine.hcl
system {
  hostname = "my-machine"
}

user "admin" {
  groups = ["sudo"]
  ssh_authorized_keys = [
    "ssh-ed25519 AAAAAAAA...",
  ]
}

container "nginx" {
  image = "nginx:latest"

  volume "/usr/share/nginx/html" {
    source = "/var/www"
  }
}

extension "tailscale" {
  version = "1.76.6"
  arch = "x86-64"
  bakery_url = "https://github.com/flatcar/sysext-bakery/releases/download/latest/"
}
```

## Installation

**Download** a release from the [GitHub Releases](https://github.com/tmacro/cola/releases) page, or **build from source**:
   ```bash
   go install github.com/tmacro/cola/cmd/cola@latest
   ```

Confirm it’s installed:
   ```bash
   cola --help
   ```
   You should see usage information for **cola**.

## Usage

### Global Flags

All **cola** commands support the following global flags:

- `--log-level="debug"`
  Sets the log level. Possible values might include `debug`, `info`, `warn`, `error`.

- `--log-format="text"`
  Sets the log output format. Possible values are `text` or `json`.

### Commands

#### `generate`

Generates an Ignition config.

```
Usage: cola generate [flags]

Generate an Ignition config.

Flags:
  -h, --help                    Show help information.
      --log-level="debug"       Set the log level.
      --log-format="text"       Set the log format. (json, text)
  -c, --config=CONFIG,...       Path to the configuration file or directory.
  -o, --output=STRING           Output file.
  -b, --bundled-extensions      Assume extensions will be bundled into the image.
  -e, --extension-dir=STRING    Directory containing sysexts.
```

**Example**:
```bash
cola generate \
  --config=machine.hcl \
  --config=extra-configs/ \
  --output=machine.ign \
  --extension-dir=./extensions
```

#### `bundle`

Bundles sysexts and an Ignition config into a self-contained Flatcar Linux image.

```
Usage: cola bundle --image=STRING [flags]

Bundle sysexts and an Ignition config with a Flatcar Linux image.

Flags:
  -h, --help                    Show help information.
      --log-level="debug"       Set the log level.
      --log-format="text"       Set the log format. (json, text)
  -c, --config=CONFIG,...       Path to the configuration file or directory.
      --base=BASE,...           Use this config as a base to extend from.
  -f, --image=STRING            Path to the Flatcar Linux image.
  -g, --gen-ignition            Generate the Ignition config. (cannot be used with --ignition)
  -i, --ignition=STRING         Path to the Ignition config.
  -o, --output=STRING           Output file.
  -e, --extension-dir=STRING    Directory containing sysexts.
```

**Example**:
```bash
cola bundle \
  --image=flatcar.img \
  --config=machine.hcl \
  --extension-dir=./extensions \
  --output=bundled-flatcar.img \
  --log-level=info
```
