package cli

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/B0lg0r0v/sliver_modified/server/assets"
	"github.com/B0lg0r0v/sliver_modified/server/c2"
	"github.com/B0lg0r0v/sliver_modified/server/certs"
	"github.com/B0lg0r0v/sliver_modified/server/configs"
	"github.com/B0lg0r0v/sliver_modified/server/cryptography"
	"github.com/B0lg0r0v/sliver_modified/server/daemon"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Force start server in daemon mode",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		force, err := cmd.Flags().GetBool(forceFlagStr)
		if err != nil {
			fmt.Printf("Failed to parse --%s flag %s\n", forceFlagStr, err)
			return
		}
		lhost, err := cmd.Flags().GetString(lhostFlagStr)
		if err != nil {
			fmt.Printf("Failed to parse --%s flag %s\n", lhostFlagStr, err)
			return
		}
		lport, err := cmd.Flags().GetUint16(lportFlagStr)
		if err != nil {
			fmt.Printf("Failed to parse --%s flag %s\n", lportFlagStr, err)
			return
		}

		appDir := assets.GetRootAppDir()
		logFile := initConsoleLogging(appDir)
		defer logFile.Close()

		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic:\n%s", debug.Stack())
				fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
				os.Exit(99)
			}
		}()

		assets.Setup(force, false)
		certs.SetupCAs()
		certs.SetupWGKeys()
		cryptography.ECCServerKeyPair()
		cryptography.TOTPServerSecret()
		cryptography.MinisignServerPrivateKey()

		serverConfig := configs.GetServerConfig()
		c2.StartPersistentJobs(serverConfig)

		daemon.Start(lhost, uint16(lport))
	},
}
