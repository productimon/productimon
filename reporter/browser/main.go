package main

import (
	"log"
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

func isTracking() bool {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	return isRunning && r.IsTracking()
}

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

func login(server, username, password, deviceName string) bool {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	_stop()
	r.Login(server, username, password, deviceName)
	_run()
	return isRunning
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
	// update config, pass in config json
	js.Global().Set("setConfig", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			setConfig(args[0].String())
		}()
		return nil
	}))
	// start tracking, optionally pass in a callback, will return boolean indicating whether it succeeded
	js.Global().Set("startTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			start()
			if len(args) > 0 {
				args[0].Invoke(isTracking())
			}
		}()
		return nil
	}))
	// stop tracking, optionally pass in a callback, will be called after tracking is stopped
	js.Global().Set("stopTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			stop()
			if len(args) > 0 {
				args[0].Invoke()
			}
		}()
		return nil
	}))
	// login, pass in serverName, username, password, deviceName, and optionally a callback function indicating success
	js.Global().Set("login", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			if len(args) < 4 {
				log.Println("not enough arguments")
				return
			}
			ret := login(args[0].String(), args[1].String(), args[2].String(), args[3].String())
			if len(args) > 4 {
				args[4].Invoke(ret)
			}
		}()
		return nil
	}))
	// switch to passed in url
	js.Global().Set("switchUrl", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			if len(args) == 0 {
				log.Println("not enough arguments")
				return
			}
			switchUrl(args[0].String())
		}()
		return nil
	}))
	// return isTracking in a callback function
	js.Global().Set("isTracking", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go func() {
			args[0].Invoke(isTracking())
		}()
		return nil
	}))
	js.Global().Call("onCoreLoaded", isRunning)
}

func main() {
	log.Println("Productimon wasm init")
	c := make(chan struct{}, 0)
	r = reporter.NewReporter(config.NewConfig())
	run()
	registerCallbacks()
	<-c // block this goroutine from quiting
}
