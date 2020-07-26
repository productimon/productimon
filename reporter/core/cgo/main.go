// +build !js

package main

import "C"
import (
	"git.yiad.am/productimon/internal"
	"git.yiad.am/productimon/reporter/core/config"
	"git.yiad.am/productimon/reporter/core/reporter"
	"unsafe"
)

var r *reporter.Reporter

//export ProdCoreReadConfig
func ProdCoreReadConfig() {
	internal.ParseFlags()
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

//export ProdCoreIsTracking
func ProdCoreIsTracking() bool {
	return r.IsTracking()
}

// expect copts to be a nullptr-terminated c-string array
//export ProdCoreSetOptions
func ProdCoreSetOptions(copts **C.char) {
	var opts []string
	curr := copts
	for *curr != nil {
		opt := C.GoString(*curr)
		opts = append(opts, opt)
		curr = (**C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(curr)) + unsafe.Sizeof(*curr)))
	}
	r.SetOptions(opts...)
}

//export ProdCoreGetServer
func ProdCoreGetServer() *C.char {
	return C.CString(r.Config.Server)
}

//export ProdCoreIsOptionEnabled
func ProdCoreIsOptionEnabled(opt *C.char) bool {
	return r.IsOptionEnabled(C.GoString(opt))
}

//export ProdCoreSaveConfig
func ProdCoreSaveConfig() bool {
	return r.SaveConfig() == nil
}

//export ProdCoreQuitReporter
func ProdCoreQuitReporter() {
	r.Quit()
}

func main() {
	panic("You shouldn't run core as a separate binary!")
}
