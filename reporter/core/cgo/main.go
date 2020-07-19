// +build !js

package main

import "C"
import (
	"flag"

	"git.yiad.am/productimon/reporter/core/config"
	"git.yiad.am/productimon/reporter/core/reporter"
)

var r *reporter.Reporter

//export ProdCoreReadConfig
func ProdCoreReadConfig() {
	flag.Parse()
	r = reporter.NewReporter(config.NewConfig())
}

func loginAndRun(server, username, password, deviceName string) bool {
	return r.Login(server, username, password, deviceName) && r.Run()
}

//export ProdCoreInitReporterInteractive
func ProdCoreInitReporterInteractive() bool {
	if r.Run() {
		return true
	}
	return loginAndRun(interactiveScanCreds(r.Config.Server))
}

//export ProdCoreInitReporterByCert
func ProdCoreInitReporterByCert() bool {
	return r.Run()
}

//export ProdCoreInitReporterByCreds
func ProdCoreInitReporterByCreds(server, username, password, deviceName *C.char) bool {
	return loginAndRun(C.GoString(server), C.GoString(username), C.GoString(password), C.GoString(deviceName))
}

//export ProdCoreSwitchWindow
func ProdCoreSwitchWindow(programName *C.char) {
	r.SwitchWindow(C.GoString(programName))
}

//export ProdCoreStartTracking
func ProdCoreStartTracking() {
	r.StartTracking()
}

//export ProdCoreStopTracking
func ProdCoreStopTracking() {
	r.StopTracking()
}

//export ProdCoreHandleMouseClick
func ProdCoreHandleMouseClick() {
	r.HandleMouseClick()
}

//export ProdCoreHandleKeystroke
func ProdCoreHandleKeystroke() {
	r.HandleKeystroke()
}

//export ProdCoreQuitReporter
func ProdCoreQuitReporter(isTracking C.char) {
	r.Quit(isTracking != 0)
}

func main() {
	panic("You shouldn't run core as a separate binary!")
}
