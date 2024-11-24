package sessions

import (
	"github.com/B0lg0r0v/sliver_modified/client/console"
	"github.com/desertbit/grumble"
)

// BackgroundCmd - Background the active session
func BackgroundCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	con.ActiveTarget.Background()
	con.PrintInfof("Background ...\n")
}
