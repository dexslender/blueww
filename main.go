package main

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus/v5"
)

type Bluetooth struct {
	Powered bool
}

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data, _ := json.Marshal(GetInitialState(conn))
	fmt.Printf("%s\n", data)

	err = conn.AddMatchSignal(
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchSender(":1.4"),
	)
	if err != nil {
		panic(err)
	}
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		update := v.Body[1].(map[string]dbus.Variant)
		if mode, ok := update["Powered"]; ok {
			data, _ := json.Marshal(Bluetooth{mode.Value().(bool)})
			fmt.Printf("%s\n", data)
		}
	}
}

func GetInitialState(conn *dbus.Conn) Bluetooth {
	obj := conn.Object("org.bluez", "/")
	var bluez map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&bluez)
	if err != nil {
		panic(err)
	}
	return Bluetooth{bluez["/org/bluez/hci0"]["org.bluez.Adapter1"]["Powered"].Value().(bool)}
}
