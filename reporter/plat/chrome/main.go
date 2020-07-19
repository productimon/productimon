package main

import (
	"fmt"

	"git.yiad.am/productimon/reporter/core/config"
	"git.yiad.am/productimon/reporter/core/reporter"
)

func main() {
	fmt.Println("hello world")
	r := reporter.NewReporter(config.NewConfig())
	r.Login("127.0.0.1:4201", "test@productimon.com", "test", "web")

	if !r.Run() {
		fmt.Println("failed to init")
	}
	r.StartTracking()
	r.StopTracking()
	fmt.Println("that's all from me")
}
