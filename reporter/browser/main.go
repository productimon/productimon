package main

import (
	"fmt"
	"net/url"
	"sync"
	"syscall/js"

	"git.yiad.am/productimon/reporter/core/config"
	"git.yiad.am/productimon/reporter/core/reporter"
)

var (
	r *reporter.Reporter

	stateMutex sync.Mutex
	isRunning  bool
)

func _run() {
	if isRunning {
		return
	}
	isRunning = r.Run()
}

func _quit() {
	if isRunning {
		r.Quit()
		isRunning = false
	}
}

func _stop() {
	if isRunning {
		r.StopTracking()
	}
}

func _start() {
	_run()
	r.StartTracking()
}

func run() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_run()
}

func stop() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_stop()
}

func start() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_start()
}

func quit() {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_quit()
}

func setConfig(cfg string) {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_quit()
	// TODO: update config here
	_run()
}

func login(server, username, password, deviceName string) {
	stop()
	r.Login(server, username, password, deviceName)
	run()
}

func switchUrl(newurl string) {
	u, err := url.Parse(newurl)
	if err != nil {
		newurl = "Unknown"
	} else {
		newurl = u.Host
	}
	// core would discard this if not tracking
	r.SwitchWindow(newurl)
}

// we can't block in js callback so everything is async
func registerCallbacks() {
	js.Global().Set("setConfig", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			setConfig(args[0].String())
		}()
		return nil
	}))
	js.Global().Set("startTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			start()
		}()
		return nil
	}))
	js.Global().Set("stopTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			stop()
		}()
		return nil
	}))
	js.Global().Set("login", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			login(args[0].String(), args[1].String(), args[2].String(), args[3].String())
		}()
		return nil
	}))
	js.Global().Set("switchUrl", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			switchUrl(args[0].String())
		}()
		return nil
	}))
	js.Global().Set("isTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			args[0].Invoke(r.IsTracking())
		}()
		return nil
	}))
	js.Global().Call("onCoreLoaded")
}

func main() {
	fmt.Println("Productimon wasm init")
	c := make(chan struct{}, 0)
	r = reporter.NewReporter(config.NewConfig())
	run()
	registerCallbacks()
	<-c // block this goroutine from quiting
}
