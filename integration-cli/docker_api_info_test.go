package main

import (
	"net/http"

	"fmt"

	"github.com/docker/docker/client"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/go-check/check"
	"golang.org/x/net/context"
)

func (s *DockerSuite) TestInfoAPI(c *check.C) {
	cli, err := client.NewEnvClient()
	c.Assert(err, checker.IsNil)
	info, err := cli.Info(context.Background())
	c.Assert(err, checker.IsNil)

	// always shown fields
	stringsToCheck := []string{
		"ID",
		"Containers",
		"ContainersRunning",
		"ContainersPaused",
		"ContainersStopped",
		"Images",
		"LoggingDriver",
		"OperatingSystem",
		"NCPU",
		"OSType",
		"Architecture",
		"MemTotal",
		"KernelVersion",
		"Driver",
		"ServerVersion",
		"SecurityOptions"}

	out := fmt.Sprintf("%+v", info)
	for _, linePrefix := range stringsToCheck {
		c.Assert(out, checker.Contains, linePrefix)
	}
}

func (s *DockerSuite) TestInfoAPIVersioned(c *check.C) {
	var httpClient *http.Client
	cli, err := client.NewClient(daemonHost(), "v1.20", httpClient, nil)
	c.Assert(err, checker.IsNil)
	testRequires(c, DaemonIsLinux) // Windows only supports 1.25 or later

	info, err := cli.Info(context.Background())
	c.Assert(err, checker.IsNil)

	out := fmt.Sprintf("%+v", info)
	c.Assert(out, checker.Contains, "not supported")
}
