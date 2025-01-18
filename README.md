Nomad CDROM Plugin
==================

Fingerprint optical drives listed in `/proc/sys/dev/cdrom/info`

## Install

1. Download an appropriate artifact from (Releases)[./releases] for the target platform
2. Unzip and move the plugin to the configured Nomad `plugin_dir`

    ```
    $ unzip nomad-device-cdrom_0.0.1_linux_amd64.zip
    Archive:  ./nomad-device-cdrom_act-build_linux_amd64.zip
    inflating: nomad-device-cdrom
    inflating: LICENSE

    ## With `plugin_dir = "/usr/local/libexec/nomad"`. Path will vary by Nomad configuration
    $ sudo mv nomad-device-cdrom /usr/local/libexec/nomad/cdrom
    ```

3. Configure plugin and restart Nomad client

## Plugin Configuration

```hcl
plugin "cdrom" {
  config {
    ## Allow up to 4 tasks to reserve generic/cdrom/readonly devices on this client
    readonly_seats = 4
  }
}
```

## Task Usage

```hcl
job "something" {
  group "stuff" {
    task "shared" {
      resources {
        device "generic/cdrom/readonly" {
          count = 1
        }
      }

      ...
    }

    task "exclusive" {
      resources {
        device "generic/cdrom/read_write" {
          count = 1
        }
      }
    }
  }
}
```
