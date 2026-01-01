// Example configuration snippet
plugin "cdrom" {
  config {
    fingerprint_interval = "5m"
    cdrom_info_path      = "/proc/sys/dev/cdrom/info"
  }
}
