// +build debug

package config

import (
	"fmt"
)

const DEBUG = true

func Initialize() {
	fmt.Println("Running in debug environment")
}
