// +build !heroku

package config

import (
	"fmt"
	"github.com/go-martini/martini"
)

const DEBUG = true

func Initialize(m *martini.ClassicMartini) {
	fmt.Println("Running in debug environment")
}
