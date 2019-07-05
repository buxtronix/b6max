package b6max
====
This is a Go library and CLI for interfacing with SkyRC iMAX style chargers
over USB.

Basic library usage
----
```go
package main

import (
  "fmt"
  "github.com/buxtronix/b6max
)

func main() {
  device, err := b6max.OpenFirst()
  if err != nil {
    fmt.Error(err.Error())
    return
  }
  defer device.Close()

  di, err := device.ReadDevInfo()
  if err != nil {
    fmt.Error(err.Error())
    return
  }
  fmt.Printf("Device info:\n%s", di)

  si, err := device.ReadSysInfo()
  if err != nil {
    fmt.Error(err.Error())
    return
  }
  fmt.Printf("System info:\n%s", si)
}
```

CLI basic usage
----

```
$ ./main 
Control program for SkyRC IMAX battery chargers.

Allows for querying and controlling these battery chargers
including charging, discharging, repeaking, rebalancing and more.

Supports chargers with direct USB connection, tested with
SkyRC IMAX B6AC V2.

Usage:
  b6max [command]

Available Commands:
  balance     Rebalance the cells of a battery (Li* only)
  charge      Run a battery charge program
  discharge   Run a battery discharge cycle
  help        Help about any command
  info        Query and display info about the attached charger
  repeak      Run a repeak cycle of the battery (NiMh/NiCd only)
  stop        Stop the currently running program
  storage     Charge or discharge the battery to a storage voltage (Li* only)
  watch       Watch status of currently running program

Flags:
      --config string   config file (default is $HOME/.cli.yaml)
  -d, --device string   Open specific USB device
  -h, --help            help for b6max
      --verbose         verbose logging

Use "b6max [command] --help" for more information about a command.
$ b6max info
Core Type : 100084
Versions  : hw: 1.00    sw: 0.11
Time Limit     : 300m (enabled: false)
Capacity Limit : 8000mV (enabled: false)
Beeps          : key (true), system (true)
Vin low limit  : 11.00v
Temp limit     : 80C
Current voltage: 0.98v
Cell voltages  : C1:0.53v C2:0.27v C3:0.28v C4:0.28v C5:0.28v C6:0.28v
State      Time    mAh       Voltage  Current   TempExt  TempInt  Imp    C1     C2      C3      C4     C5     C6
Idle           0       0    8.000v     0.26A     42C       248C      0Ω   20.483v  53.506v  4.865v  3.841v  5.633v  7.425v
$
```

USB device access
----
To access the USB device, you need to either run the binary as 'root' (discouraged),
or add the following udev rule:

```udev
SUBSYSTEM=="usb", ATTRS{idVendor}=="0000", ATTRS{idProduct}=="0001", MODE:="666", GROUP="plugdev"
KERNEL=="hidraw*", ATTRS{idVendor}=="0000", ATTRS{idProduct}=="0001", MODE="0666", GROUP="plugdev"
```

TODO
----
 * Include standard battery profiles
 * Add a web server with graphs, etc.
 * Write settings back to device
 * Log charge data for analysis, etc
 * Unit tests

Etc...
----
Copyright 2019 Ben Buxton <bbuxton@gmail.com>.

Licenced under the Apache 2.0 licence.

PRs gladly accepted!

