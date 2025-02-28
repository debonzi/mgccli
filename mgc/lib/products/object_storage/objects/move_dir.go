/*
Executor: move-dir

# Summary

# Moves objects from source to destination

# Description

Moves objects from a source to a destination.
They can be either local or remote but not both local (Local -> Remote, Remote -> Local, Remote -> Remote)

import "magalu.cloud/lib/products/object_storage/objects"
*/
package objects

import (
	"context"

	mgcCore "magalu.cloud/core"
	mgcHelpers "magalu.cloud/lib/helpers"
)

type MoveDirParameters struct {
	BatchSize *int   `json:"batch_size,omitempty"`
	Dst       string `json:"dst"`
	Src       string `json:"src"`
}

type MoveDirConfigs struct {
	ChunkSize *int    `json:"chunkSize,omitempty"`
	Region    *string `json:"region,omitempty"`
	ServerUrl *string `json:"serverUrl,omitempty"`
	Workers   *int    `json:"workers,omitempty"`
}

type MoveDirResult struct {
	BatchSize *int   `json:"batch_size,omitempty"`
	Dst       string `json:"dst"`
	Src       string `json:"src"`
}

func (s *service) MoveDir(
	parameters MoveDirParameters,
	configs MoveDirConfigs,
) (
	result MoveDirResult,
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("MoveDir", mgcCore.RefPath("/object-storage/objects/move-dir"), s.client, s.ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[MoveDirParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[MoveDirConfigs](configs); err != nil {
		return
	}

	r, err := exec.Execute(ctx, p, c)
	if err != nil {
		return
	}
	return mgcHelpers.ConvertResult[MoveDirResult](r)
}

// Context from caller is used to allow cancellation of long-running requests
func (s *service) MoveDirContext(
	ctx context.Context,
	parameters MoveDirParameters,
	configs MoveDirConfigs,
) (
	result MoveDirResult,
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("MoveDir", mgcCore.RefPath("/object-storage/objects/move-dir"), s.client, ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[MoveDirParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[MoveDirConfigs](configs); err != nil {
		return
	}

	sdkConfig := s.client.Sdk().Config().TempConfig()
	if c["serverUrl"] == nil && sdkConfig["serverUrl"] != nil {
		c["serverUrl"] = sdkConfig["serverUrl"]
	}

	if c["env"] == nil && sdkConfig["env"] != nil {
		c["env"] = sdkConfig["env"]
	}

	if c["region"] == nil && sdkConfig["region"] != nil {
		c["region"] = sdkConfig["region"]
	}

	r, err := exec.Execute(ctx, p, c)
	if err != nil {
		return
	}
	return mgcHelpers.ConvertResult[MoveDirResult](r)
}

// TODO: links
// TODO: related
