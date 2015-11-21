package main

import (
	"os"

	"github.com/itchio/butler/bio"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version    = "head" // set by command-line on CI release builds
	app        = kingpin.New("butler", "Your very own itch.io helper")
	jsonOutput = app.Flag("json", "Enable machine-readable JSON-lines output").Short('j').Bool()
	quiet      = app.Flag("quiet", "Hide progress indicators & other extra info").Short('q').Bool()

	dlCmd  = app.Command("dl", "Download a file (resumes if can, checks hashes)")
	dlUrl  = dlCmd.Arg("url", "Address to download from").Required().String()
	dlDest = dlCmd.Arg("dest", "File to write downloaded data to").Required().String()

	pushCmd  = app.Command("push", "Upload a new version of something to itch.io")
	pushSrc  = pushCmd.Arg("src", "Directory or archive to upload").Required().ExistingFileOrDir()
	pushRepo = pushCmd.Arg("repo", "Repository to push to, e.g. leafo/xmoon:win64").Required().String()
)

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)
	app.VersionFlag.Short('V')

	cmd, err := app.Parse(os.Args[1:])
	bio.JsonOutput = *jsonOutput
	bio.Quiet = *quiet

	switch kingpin.MustParse(cmd, err) {
	case dlCmd.FullCommand():
		dl(*dlUrl, *dlDest)

	case pushCmd.FullCommand():
		push(*pushSrc, *pushRepo)
	}
}