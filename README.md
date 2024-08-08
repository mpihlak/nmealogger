# nmealogger - log NMEA data for offline analysis

Reads NMEA sentences from the network, adds timestamps and writes to log files. The intended use is to capture instrument data
from a sailing session and store it for later analysis. Assumes an NMEA network server such as [Kplex](https://www.stripydog.com/kplex/index.html)
running on `127.0.0.1:10110`

I'm using this on a Raspberry Pi with a RS422 hat, connected to the output of Tacktick T122 NMEA interface. The Pi also has a 4G
modem attached to it and provides a Wifi hotspot so that logs can be uploaded ASAP and that the raw NMEA data is available on the
boat network.

Example log data:

```log
2024-07-15T13:09:48.683+0000    $IIRMC,130900,A,5930.975,N,02446.310,E,05.9,161,150724,00,E,A*0A
2024-07-15T13:09:48.746+0000    $IIVHW,,,117,M,05.7,N,,*61
2024-07-15T13:09:49.129+0000    $GPRMC,130949,A,5930.970,N,02446.315,E,05.7,160,150724,00,E,A*1F
2024-07-15T13:09:49.157+0000    $GPGLL,5930.970,N,02446.315,E,130949,A,A*43
2024-07-15T13:09:49.217+0000    $IIVLW,09452,N,030.8,N*52
2024-07-15T13:09:49.267+0000    $IIVWR,154,R,05.5,N,,,,*61
```

The log files will be created in `/data`. If an Internet connection is available the `loguploader` daemon will attempt to upload
the files to Google Drive. Uploaded log files are renamed to have an `.uploaded` suffix and deleted after a while.

Binaries built from the `cmd` directory:

* `nmealogger` - the logging daemon
* `logupload` - upload log files to Google drive from Pi
* `logdownload` - fetch log files in bulk from Google Drive and optionally clean up the Drive.
* `nmeareplay` - replay the log files from a network server. Enables offline use of tools such as NMEAremote.

## Installation

`make build` builds a Debian package for `arm32` Pi (first generation?). For other architectures the `Dockerfile` needs to be adjusted.
Installing the package creates systemd services for NMEA logger, log uploader and cleanup.

## Uploading to Google Drive

Using Drive as log destination is convenient as everyone has it, it doesn't necessarily cost anything and has a decent API (if somewhat poorly documented). The setup can be a bit involved though:

First a Cloud Project needs to be created. Then a service account. All this in the Google Developer [console](https://console.cloud.google.com/iam-admin/serviceaccounts?project=foo). The destination log folder in Drive needs to be shared with that service account.

The `loguploader` and `logdownloader` utilities need the JSON credentials file for the service account. Also the folder ID must be obtained
from somewhere as Google Drive APIs operate on folder ID, not name. The easiest is probably to extract it from the Google Drive URL for the
folder.
