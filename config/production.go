// +build !debug

package config

import (
	"fmt"
	"github.com/yvasiyarov/gorelic"
	"os"
)

const DEBUG = false

func Initialize() {
	fmt.Println("Running in production environment")

	agent := gorelic.NewAgent()
	agent.Verbose = true
	agent.NewrelicLicense = os.Getenv("NEW_RELIC_LICENSE_KEY")
	agent.Run()
}
