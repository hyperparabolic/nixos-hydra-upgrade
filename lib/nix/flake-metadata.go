package nix

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

type FlakeMetadata struct {
	// unix timestamp
	LastModified int64 `json:"lastModified"`
	// flake url
	OriginalUrl string `json:"originalUrl"`
}

func GetFlakeMetadata(flake string) FlakeMetadata {
	cmd := exec.Command("nix", "flake", "metadata", flake, "--json")

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	var metadata FlakeMetadata
	err = json.Unmarshal(output, &metadata)
	if err != nil {
		panic(err)
	}

	slog.Debug(fmt.Sprintf("%+v", metadata))
	return metadata
}
