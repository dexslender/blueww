package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

func TestList(t *testing.T) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to system bus:", err)
		os.Exit(1)
	}
	defer conn.Close()

	var s []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&s)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of owned names:", err)
		os.Exit(1)
	}

	fmt.Println("Currently owned names on the system bus:")
	for _, v := range s {
		fmt.Println(v)
	}
}

func TestBluez(t *testing.T) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	obj := conn.Object("org.bluez", "/")

	t.Log(obj)

	// ----CodeSnipets
	// 	var s map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	// 	err = obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&s)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Println(s)

	// // var s string
	// v, err := conn.Object("org.bluez", "/").GetProperty("sender")
	//
	//	if err != nil {
	//		panic(err)
	//	}
	//
	// fmt.Println(v)
	// // fmt.Println(s)
}

func TestWeird(t *testing.T) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	node, err := introspect.Call(conn.Object(":1.4", "/"))
	data, _ := json.MarshalIndent(node, "", "  ")
	os.Stdout.Write(data)
}

func TestBluezObj(t *testing.T) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	obj := conn.Object("org.bluez", "/")

	var s map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err = obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&s)
	if err != nil {
		panic(err)
	}
	for path, v := range s {
		fmt.Printf("\t%s :: %T\n", path, v)
		for path, v := range v {
			fmt.Printf("\t\t%s :: %T\n", path, v)
			for k, v := range v {
				fmt.Printf("\t\t\t%s :: %s\n", k, v.String())
			}
		}
	}
}
