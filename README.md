# nmealogger - log NMEA data for offline analysis

Reads NMEA sentences from the network, adds timestamps and logs to files. The intended use is to capture instrument data
from a sailing session and store it for later analysis. Assumes an NMEA network server running on port `10110` on `localhost`.
[Kplex](https://www.stripydog.com/kplex/index.html) works well, alternatively SignalK NMEA 0183 over IP should also work.

Example log data:

```log
2024-07-15T13:09:48.683+0000    $IIRMC,130900,A,5930.975,N,02446.310,E,05.9,161,150724,00,E,A*0A
2024-07-15T13:09:48.746+0000    $IIVHW,,,117,M,05.7,N,,*61
2024-07-15T13:09:49.129+0000    $GPRMC,130949,A,5930.970,N,02446.315,E,05.7,160,150724,00,E,A*1F
2024-07-15T13:09:49.157+0000    $GPGLL,5930.970,N,02446.315,E,130949,A,A*43
2024-07-15T13:09:49.217+0000    $IIVLW,09452,N,030.8,N*52
2024-07-15T13:09:49.267+0000    $IIVWR,154,R,05.5,N,,,,*61
```

The log files will be created in `/data` after every 5 minutes. If an Internet connection is available the `loguploader` daemon will
attempt to upload the finalized log files to Google Drive. Uploaded log files are renamed to have an `.uploaded` suffix and deleted
from `/data` after a while.

Binaries built from the `cmd` directory:

* `nmealogger` - the logging daemon
* `logupload` - upload log files to Google drive from Pi
* `logdownload` - fetch log files in bulk from Google Drive and optionally clean up the Drive.
* `nmeareplay` - replay the log files from a network server. Enables offline use of tools such as NMEAremote.

## Installation

`make build` builds a Debian package for `arm32`. This has been confirmed to work on both the earlier generation and later 64 bit
Raspberry Pi-s. For other architectures the `Dockerfile` needs to be adjusted.

Copy the resulting `.deb` file to the Pi and install with `dpkg -i`. This will create the systemd services for
NMEA logger, log uploader and cleanup. The services will be started automaticaly at system startup.

The `/data` directory needs to be created manually.

## Uploading to Google Drive

Using Drive as log destination is convenient as ~everyone has it, it doesn't necessarily cost anything and has a decent API
(if somewhat poorly documented). The setup can be a bit involved though: a cloud project needs to be created and an IAM service
account. Use the Google Developer [console](https://console.cloud.google.com). Download the service account credentials - these
need to be configured for `loguploader` and `logdownloader`.

Create the destination folder in the Drive and note the folder ID from URL (this will need to be configured for the
`loguploader`). Share the folder with the service account.

