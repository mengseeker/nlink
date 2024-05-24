package main

import (
	"github.com/mengseeker/nlink/cmd"
)

func main() {
	// go startpprof()
	cmd.Execute()
}

// func startpprof() {
// 	cpuf, err := os.Create("cpu.pprof")
// 	if err != nil {
// 		panic(err)
// 	}
// 	pprof.StartCPUProfile(cpuf)

// 	sigs := make(chan os.Signal, 1)
// 	signal.Notify(sigs, os.Interrupt)
// 	go func() {
// 		<-sigs

// 		pprof.StopCPUProfile()
// 		cpuf.Close()
// 		os.Exit(0)
// 	}()
// }
