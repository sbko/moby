package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/docker/docker/integration-cli/request"
	"github.com/docker/docker/pkg/testutil"
	"github.com/go-check/check"
	"golang.org/x/net/context"
)

func (s *DockerSuite) TestAPICreateWithNotExistImage(c *check.C) {
	cli, err := client.NewEnvClient()
	c.Assert(err, checker.IsNil)
	name := "test"
	config := container.Config{
		Image:   "test456:v1",
		Volumes: map[string]struct{}{"/tmp": {}},
	}

	_, err = cli.ContainerCreate(context.Background(), &config, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected := "No such image: test456:v1"
	c.Assert(err.Error(), checker.Contains, expected)

	config2 := container.Config{
		Image:   "test456",
		Volumes: map[string]struct{}{"/tmp": {}},
	}

	_, err = cli.ContainerCreate(context.Background(), &config2, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected = "No such image: test456"
	c.Assert(err.Error(), checker.Contains, expected)

	config3 := container.Config{
		Image: "sha256:0cb40641836c461bc97c793971d84d758371ed682042457523e4ae701efeaaaa",
	}

	_, err = cli.ContainerCreate(context.Background(), &config3, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected = "No such image: sha256:0cb40641836c461bc97c793971d84d758371ed682042457523e4ae701efeaaaa"
	c.Assert(err.Error(), checker.Contains, expected)

}

// Test for #25099
func (s *DockerSuite) TestAPICreateEmptyEnv(c *check.C) {
	cli, err := client.NewEnvClient()
	c.Assert(err, checker.IsNil)
	name := "test1"

	config := container.Config{
		Image: "busybox",
		Env:   []string{"", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		Cmd:   []string{"true"},
	}

	_, err = cli.ContainerCreate(context.Background(), &config, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected := "invalid environment variable:"
	c.Assert(err.Error(), checker.Contains, expected)

	name = "test2"
	config = container.Config{
		Image: "busybox",
		Env:   []string{"=", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		Cmd:   []string{"true"},
	}

	_, err = cli.ContainerCreate(context.Background(), &config, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected = "invalid environment variable: ="
	c.Assert(err.Error(), checker.Contains, expected)

	name = "test3"
	config = container.Config{
		Image: "busybox",
		Env:   []string{"=foo", "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		Cmd:   []string{"true"},
	}

	_, err = cli.ContainerCreate(context.Background(), &config, &container.HostConfig{}, &network.NetworkingConfig{}, name)
	expected = "invalid environment variable: =foo"
	c.Assert(err.Error(), checker.Contains, expected)
}

func (s *DockerSuite) TestAPICreateWithInvalidHealthcheckParams(c *check.C) {
	// test invalid Interval in Healthcheck: less than 0s
	name := "test1"
	config := map[string]interface{}{
		"Image": "busybox",
		"Healthcheck": map[string]interface{}{
			"Interval": -10 * time.Millisecond,
			"Timeout":  time.Second,
			"Retries":  int(1000),
		},
	}

	res, body, err := request.Post("/containers/create?name="+name, request.JSONBody(config))
	c.Assert(err, check.IsNil)
	c.Assert(res.StatusCode, check.Equals, http.StatusInternalServerError)

	buf, err := testutil.ReadBody(body)
	c.Assert(err, checker.IsNil)

	expected := fmt.Sprintf("Interval in Healthcheck cannot be less than %s", container.MinimumDuration)
	c.Assert(getErrorMessage(c, buf), checker.Contains, expected)

	// test invalid Interval in Healthcheck: larger than 0s but less than 1ms
	name = "test2"
	config = map[string]interface{}{
		"Image": "busybox",
		"Healthcheck": map[string]interface{}{
			"Interval": 500 * time.Microsecond,
			"Timeout":  time.Second,
			"Retries":  int(1000),
		},
	}
	res, body, err = request.Post("/containers/create?name="+name, request.JSONBody(config))
	c.Assert(err, check.IsNil)

	buf, err = testutil.ReadBody(body)
	c.Assert(err, checker.IsNil)

	c.Assert(res.StatusCode, check.Equals, http.StatusInternalServerError)
	c.Assert(getErrorMessage(c, buf), checker.Contains, expected)

	// test invalid Timeout in Healthcheck: less than 1ms
	name = "test3"
	config = map[string]interface{}{
		"Image": "busybox",
		"Healthcheck": map[string]interface{}{
			"Interval": time.Second,
			"Timeout":  -100 * time.Millisecond,
			"Retries":  int(1000),
		},
	}
	res, body, err = request.Post("/containers/create?name="+name, request.JSONBody(config))
	c.Assert(err, check.IsNil)
	c.Assert(res.StatusCode, check.Equals, http.StatusInternalServerError)

	buf, err = testutil.ReadBody(body)
	c.Assert(err, checker.IsNil)

	expected = fmt.Sprintf("Timeout in Healthcheck cannot be less than %s", container.MinimumDuration)
	c.Assert(getErrorMessage(c, buf), checker.Contains, expected)

	// test invalid Retries in Healthcheck: less than 0
	name = "test4"
	config = map[string]interface{}{
		"Image": "busybox",
		"Healthcheck": map[string]interface{}{
			"Interval": time.Second,
			"Timeout":  time.Second,
			"Retries":  int(-10),
		},
	}
	res, body, err = request.Post("/containers/create?name="+name, request.JSONBody(config))
	c.Assert(err, check.IsNil)
	c.Assert(res.StatusCode, check.Equals, http.StatusInternalServerError)

	buf, err = testutil.ReadBody(body)
	c.Assert(err, checker.IsNil)

	expected = "Retries in Healthcheck cannot be negative"
	c.Assert(getErrorMessage(c, buf), checker.Contains, expected)

	// test invalid StartPeriod in Healthcheck: not 0 and less than 1ms
	name = "test3"
	config = map[string]interface{}{
		"Image": "busybox",
		"Healthcheck": map[string]interface{}{
			"Interval":    time.Second,
			"Timeout":     time.Second,
			"Retries":     int(1000),
			"StartPeriod": 100 * time.Microsecond,
		},
	}
	res, body, err = request.Post("/containers/create?name="+name, request.JSONBody(config))
	c.Assert(err, check.IsNil)
	c.Assert(res.StatusCode, check.Equals, http.StatusInternalServerError)

	buf, err = testutil.ReadBody(body)
	c.Assert(err, checker.IsNil)

	expected = fmt.Sprintf("StartPeriod in Healthcheck cannot be less than %s", container.MinimumDuration)
	c.Assert(getErrorMessage(c, buf), checker.Contains, expected)
}
