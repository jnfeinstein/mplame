// +build !debug

package config

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gorelic"
	"os"
)

const DEBUG = false

func Initialize(m *martini.ClassicMartini) {
	fmt.Println("Running in production environment")

	gorelic.InitNewrelicAgent(os.Getenv("NEW_RELIC_LICENSE_KEY"), "mpLAME", true)
	m.Use(gorelic.Handler)
}
