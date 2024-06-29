package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/godbus/dbus/v5"
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
	err = conn.AddMatchSignal(
		dbus.WithMatchMember("PropertiesChanged"),
		dbus.WithMatchSender(":1.4"),
	)
	if err != nil {
		t.Fatal(err)
	}
	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)
	for v := range c {
		t.Logf("%+v", v)
	}
	
	// obj := conn.Object("org.bluez", "/")

	// t.Log(obj)

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

func TestIntrospect(t *testing.T) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	obj := conn.Object(":1.4", "/")

	call := obj.AddMatchSignal(
		"org.freedesktop.DBus.ObjectManager",
		"InterfacesAdded",
	)
	
	c := make(chan *dbus.Signal, 10)
	call.Store(&c)
	for v := range c {
		t.Logf("%+v", v)
	}
	
	// node, err := introspect.Call(obj)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// data, _ := json.MarshalIndent(node, "", "  ")
	// os.Stdout.Write(data)
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
