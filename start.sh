#!/bin/bash
/etc/init.d/dbus start
/usr/libexec/bluetooth/bluetoothd --debug &
CompileDaemon --build="go build ." --command=./HomeHub