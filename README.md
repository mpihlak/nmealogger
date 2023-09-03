# nmealogger - write now, analyze later

Reads NMEA sentences from network, add timestamps and write to log files. The intended use is to capture instrument data
from a sailing session and store it for later analysis. Assumes an NMEA network server such as [Kplex](https://www.stripydog.com/kplex/index.html)
running on `127.0.0.1:10110`

Example data:
```
2023-09-02T12:41:35.168370+00:00        $IIHDG,097,,,00,E*1C
2023-09-02T12:41:35.168370+00:00        $IIMTW,+15.0,C*3C
2023-09-02T12:41:35.168370+00:00        $IIMWV,032,R,03.9,N,A*18
2023-09-02T12:41:35.168370+00:00        $IIMWV,032,T,03.9,N,A*1E
2023-09-02T12:41:35.168370+00:00        $IIRMC,171300,A,5928.094,N,02449.269,E,00.1,036,010923,00,E,A*08
2023-09-02T12:41:35.168370+00:00        $IIVHW,,,097,M,00.0,N,,*6A
2023-09-02T12:41:35.168370+00:00        $IIVLW,09200,N,000.0,N*58
```
