package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

type Bluetooth struct {
	Device
	Powered  bool
	Pairable bool
	Devices  []Device
}

type Device struct {
	Alias     string
	Address   string
	Connected bool
}

type DeviceInfo struct {
	Address   string
	Alias     string
	Interface string
}

var bluetooth Bluetooth

var Listening map[dbus.ObjectPath]DeviceInfo = make(map[dbus.ObjectPath]DeviceInfo)

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data, _ := json.Marshal(GetInitial(conn))
	fmt.Printf("%s\n", data)

	for path := range Listening {
		// fmt.Println("// listening path:", path, "interface:", iface)
		err := conn.AddMatchSignal(
			dbus.WithMatchObjectPath(path),
			dbus.WithMatchMember("PropertiesChanged"),
		)
		if err != nil {
			fmt.Println("// unexpected err:", err)
		}
	}
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		// fmt.Println("// Received signal data:", v)
		// fmt.Println("// Path interface check:", Listening[v.Path] == v.Body[0])
		if Listening[v.Path].Interface != v.Body[0] {
			continue
		}
		update := v.Body[1].(map[string]dbus.Variant)
		if mode, ok := update["Powered"]; ok {
			bluetooth.Powered = mode.Value().(bool)
			Update()
		} else if connected, ok := update["Connected"]; ok {
			if connected.Value() == true {
				bluetooth.Connected = true
				dev := Listening[v.Path]
				for i := range bluetooth.Devices {
					if bluetooth.Devices[i].Address == dev.Address {
						bluetooth.Devices[i].Connected = true
					}
				}
				Update()
			} else {
				bluetooth.Connected = false
				dev := Listening[v.Path]
				for i := range bluetooth.Devices {
					if bluetooth.Devices[i].Address == dev.Address {
						bluetooth.Devices[i].Connected = false
					}
				}
				Update()
			}
		} else if pair, ok := update["Pairable"]; ok {
			bluetooth.Pairable = pair.Value().(bool)
			Update()
		}
	}
}

func GetInitial(conn *dbus.Conn) Bluetooth {
	obj := conn.Object("org.bluez", "/")
	var bluez map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&bluez)
	if err != nil {
		panic(err)
	}

	for path, vPath := range bluez {
		for iface, prop := range vPath {
			if alias, ok := prop["Alias"]; ok {
				if roles, ok := prop["Roles"]; ok {
					if slices.Contains(roles.Value().([]string), "central") {
						bluetooth.Device = Device{
							Alias:   Get[string](alias),
							Address: Get[string](prop["Address"]),
						}
						bluetooth.Powered = Get[bool](prop["Powered"])
						bluetooth.Pairable = Get[bool](prop["Pairable"])
					}
				} else {
					if ok := Get[bool](prop["Connected"]); ok {
						bluetooth.Connected = ok
					}
					bluetooth.Devices = append(bluetooth.Devices, Device{
						Alias:     Get[string](alias),
						Address:   Get[string](prop["Address"]),
						Connected: Get[bool](prop["Connected"]),
					})
				}
				Listening[path] = DeviceInfo{
					Address:   Get[string](prop["Address"]),
					Interface: iface,
					Alias:     Get[string](alias),
				}
			}
		}
	}
	return bluetooth
}

func Get[T any](v dbus.Variant) T {
	return v.Value().(T)
}

func Update() {
	data, _ := json.Marshal(bluetooth)
	fmt.Printf("%s\n", data)
}
