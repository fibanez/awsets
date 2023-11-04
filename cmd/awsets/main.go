package main

import "github.com/fibanez/awsets/cmd/awsets/cmd"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(map[string]string{
		"version": version,
		"commit":  commit,
		"date":    date,
	})
}
