# SMPP Tester

This utility is for load testing SMPP servers. It can send SubmitSMs with fixed speed

It is based on go-smpp library https://github.com/linxGnu/gosmpp

Usage is below

```
Usage:
smpp-tester [OPTIONS]

Application Options:
-s, --speed=           rate per second (default: 50)
-h, --host=            smpp server host (default: localhost)
-P, --port=            smpp server port (default: 2775)
-u, --system_id=       SMPP systemId
-p, --password=        SMPP password
-y, --skip-confirm
-t, --text=            SMS text (default: load-test)
-m, --max-count=       Maximum SMS number to send (default: -1)
-w, --wait-deliver-sm= Wait in seconds for deliver_sm after sending' (default: 10)

Help Options:
-h, --help             Show this help message
```