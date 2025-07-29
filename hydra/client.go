package hydra

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

type HydraClient struct {
	Instance string
	JobSet   string
	Job      string
	Project  string
}

// see https://github.com/NixOS/hydra/blob/master/hydra-api.yaml
// These are partial implementations, just grabbing what I need.

type Build struct {
	// 1 is finished, else not
	Finished int `json:"finished"`
	// may be nil if not finished, 1 is success, else not
	BuildStatus int `json:"buildstatus"`
	// should be length 1
	JobSetEvals []int `json:"jobsetevals"`
}

type Eval struct {
	// flake specification for a specific git commit
	Flake string `json:"flake"`
}

/*
Gets a the latest build. These are host toplevel derivations in this
use case.
*/
func (client HydraClient) GetLatestBuild() Build {
	httpClient := http.Client{}

	requestUrl, err := url.JoinPath(client.Instance, "job", client.Project, client.JobSet, client.Job, "latest")
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	slog.Debug("GetLatestBuild",
		slog.String("body", string(body)),
		slog.String("url", requestUrl))

	var build Build
	err = json.Unmarshal(body, &build)
	if err != nil {
		panic(err)
	}

	slog.Debug(fmt.Sprintf("%+v", build))
	return build
}

/*
Gets a specific evaluation. This includes the flake that includes the
job / build.
*/
func (client HydraClient) GetEval(build Build) Eval {
	httpClient := http.Client{}

	requestUrl, err := url.JoinPath(client.Instance, "eval", strconv.Itoa(build.JobSetEvals[0]))
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	slog.Debug("GetEval",
		slog.String("body", string(body)),
		slog.String("url", requestUrl))

	var eval Eval
	err = json.Unmarshal(body, &eval)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", eval)
	return eval
}
