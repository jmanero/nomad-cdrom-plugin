// Example configuration snippet
plugin "cdrom" {
  config {
    fingerprint_interval = "1m"
    cdrom_info_path      = "/proc/sys/dev/cdrom/info"

    // Create a generic/cdrom/readonly device-group with 4 devices for each
    // detected optical drive. This allows multiple tasks to declare a
    // `device "generic/cdrom/readonly"` requirement and share access to the
    // same physical device on a host
    readonly_seats = 4
  }
}
