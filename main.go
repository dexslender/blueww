package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/godbus/dbus/v5"
)

var info Bluetooth
var defPath dbus.ObjectPath
var defAdap string

type Bluetooth struct {
	Powered   bool
	Connected bool
	Pairable  bool
	Device    struct {
		Alias   string
		Address string
	}
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
		dbus.WithMatchSender(":1.3"),
	)
	if err != nil {
		panic(err)
	}
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		update := v.Body[1].(map[string]dbus.Variant)
		if mode, ok := update["Powered"]; ok {
			info.Powered = mode.Value().(bool)
			data, _ := json.Marshal(info)
			fmt.Printf("%s\n", data)
		} else if connected, ok := update["Connected"]; ok {
			if connected.Value() == true {
				info.Connected = true
				UpdateWithPath(conn, v.Path, v.Body[0].(string))
			} else {
				info.Connected = false
				DefaultAdapter(conn)
			}
		} else if pair, ok := update["Pairable"]; ok {
			info.Pairable = pair.Value().(bool)
			data, _ := json.Marshal(info)
			fmt.Printf("%s\n", data)
		}
	}
}

func DefaultAdapter(conn *dbus.Conn) {
	obj := conn.Object("org.bluez", "/")
	var bluez map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&bluez)
	if err != nil {
		panic(err)
	}

	if blu, ok := bluez[defPath][defAdap]; ok {
		info.Device.Alias = blu["Alias"].Value().(string)
		info.Device.Address = blu["Address"].Value().(string)
		info.Powered = blu["Powered"].Value().(bool)

		data, _ := json.Marshal(info)
		fmt.Printf("%s\n", data)
	}
}

func UpdateWithPath(conn *dbus.Conn, path dbus.ObjectPath, device string) {
	obj := conn.Object("org.bluez", "/")
	var bluez map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&bluez)
	if err != nil {
		panic(err)
	}

	if v, ok := bluez[path][device]; ok {
		if _, ok := v["Alias"]; ok {
			info.Device.Alias = v["Alias"].Value().(string)
			info.Device.Address = v["Address"].Value().(string)
			data, _ := json.Marshal(info)
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
	for p, path := range bluez {
		for intf, iface := range path {
			if _, ok := iface["Alias"]; ok {
				if v, ok := iface["Connected"]; ok {
					if v.Value() == true {
						info.Connected = iface["Connected"].Value().(bool)
						info.Device.Alias = iface["Alias"].Value().(string)
						info.Device.Address = iface["Address"].Value().(string)
					}
				} else if slices.Contains(iface["Roles"].Value().([]string), "central") {
					defPath = p
					defAdap = intf
					info.Powered = iface["Powered"].Value().(bool)
					info.Device.Alias = iface["Alias"].Value().(string)
					info.Device.Address = iface["Address"].Value().(string)
					info.Pairable = iface["Pairable"].Value().(bool)
				}
			}
		}
	}
	// return Bluetooth{bluez["/org/bluez/hci0"]["org.bluez.Adapter1"]["Powered"].Value().(bool)}
	return info
}
