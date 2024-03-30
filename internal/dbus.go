package internal

import (
	"fmt"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
)

const (
	FDN_PATH            = "/org/freedesktop/Notifications"
	FDN_IFAC            = "org.freedesktop.Notifications"
	FDN_NAME            = "org.freedesktop.Notifications"
	INTROSPECTABLE_IFAC = "org.freedesktop.DBus.Introspectable"

	FDN_SPEC_VERSION = "1.2"
)

const DBUS_XML = `<node name="` + FDN_PATH + `">
  <interface name="` + FDN_IFAC + `">

      <method name="GetCapabilities">
          <arg direction="out" name="capabilities"    type="as" />
      </method>

      <method name="Notify">
          <arg direction="in"  name="app_name"        type="s"/>
          <arg direction="in"  name="replaces_id"     type="u"/>
          <arg direction="in"  name="app_icon"        type="s"/>
          <arg direction="in"  name="summary"         type="s"/>
          <arg direction="in"  name="body"            type="s"/>
          <arg direction="in"  name="actions"         type="as"/>
          <arg direction="in"  name="hints"           type="a{sv}"/>
          <arg direction="in"  name="expire_timeout"  type="i"/>
          <arg direction="out" name="id"              type="u"/>
      </method> 

      <method name="GetServerInformation">
          <arg direction="out" name="name"            type="s"/>
          <arg direction="out" name="vendor"          type="s"/>
          <arg direction="out" name="version"         type="s"/>
          <arg direction="out" name="spec_version"    type="s"/>
      </method>

      <method name="CloseNotification">
          <arg direction="in"  name="id"              type="u"/>
      </method>

     <signal name="NotificationClosed">
          <arg name="id"         type="u"/>
          <arg name="reason"     type="u"/>
      </signal>

      <signal name="ActionInvoked">
          <arg name="id"         type="u"/>
          <arg name="action_key" type="s"/>
      </signal>
  </interface>
` + introspect.IntrospectDataString + `
</node>`

var (
	conn     *dbus.Conn
	hyprsock HyprConn
)

type Notifications string

func (n Notifications) GetCapabilities() ([]string, *dbus.Error) {
	var cap []string
	return cap, nil
}

func (n Notifications) Notify(
	app_name string,
	replaces_id uint32,
	app_icon string,
	summary string,
	body string,
	actions []string,
	hints map[string]dbus.Variant,
	expire_timeout int32,
) (uint32, *dbus.Error) {
	if expire_timeout == -1 {
		expire_timeout = 5000
	}

	hyprsock.Notify(
		hyprsock.icon.INFO,
		expire_timeout,
		hyprsock.color.GREEN,
		summary,
		hyprsock.DEFAULT_FONT_SIZE,
	)
	go SendCloseSignal(expire_timeout, 1)
	return 1, nil
}

func (n Notifications) CloseNotification(id uint32) *dbus.Error {
	hyprsock.DismissNotify(-1)

	go SendCloseSignal(0, 3)
	return nil
}

func (n Notifications) GetServerInformation() (string, string, string, string, *dbus.Error) {
	return PACKAGE, AUTHOR, VERSION, FDN_SPEC_VERSION, nil
}

func SendCloseSignal(timeout int32, reason uint32) {
	d := time.Duration(int64(timeout)) * time.Millisecond
	time.Sleep(d)
	conn.Emit(
		FDN_PATH,
		"org.freedesktop.Notifications.NotificationClosed",
		uint32(0),
		reason,
	)
}

func InitDBus() {
	var err error
	conn, err = dbus.ConnectSessionBus()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	GetHyprSocket(&hyprsock)

	n := Notifications(PACKAGE)
	conn.Export(n, FDN_PATH, FDN_IFAC)
	conn.Export(introspect.Introspectable(DBUS_XML), FDN_PATH, INTROSPECTABLE_IFAC)

	reply, err := conn.RequestName(FDN_NAME, dbus.NameFlagDoNotQueue)
	if err != nil {
		panic(err)
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
		os.Exit(1)
	}
	select {}
}
