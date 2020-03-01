package helpers

import (
	"fmt"
	"runtime"
)

// PrintMemUsage prints memory usage.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	fmt.Printf(" TotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	fmt.Printf(" Sys = %v MiB", m.Sys/1024/1024)
	fmt.Printf(" NumGC = %v\n", m.NumGC)
}
