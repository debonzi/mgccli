/*
Executor: retype

# Summary

Change the type of a volume.

# Description

Change the Volume Type of an existing Volume for the currently

	authenticated tenant.

The Volume retype will be completed when the Volume status returns to

	"completed".

#### Rules
- The Volume state must be "available".
- The Volume status must be "completed" or "retype_error".
- The new Volume Type must belong to the same region as the Volume.

#### Notes
  - Utilize the **block-storage volume list** command to retrieve a list of all
    Volumes and obtain the ID of the Volume you want to retype.
  - Utilize the **block-storage volume-types list** command to retrieve a list of
    all Volume Types and obtain the ID of the Volume Type you want to use.

Version: v1

import "magalu.cloud/lib/products/block_storage/volumes"
*/
package volumes

import (
	"context"

	mgcCore "magalu.cloud/core"
	mgcHelpers "magalu.cloud/lib/helpers"
)

type RetypeParameters struct {
	Id      string                  `json:"id"`
	NewType RetypeParametersNewType `json:"new_type"`
}

// any of: RetypeParametersNewType
type RetypeParametersNewType struct {
	DiskType *string                      `json:"disk_type,omitempty"`
	Id       string                       `json:"id"`
	Iops     *RetypeParametersNewTypeIops `json:"iops,omitempty"`
	Name     *string                      `json:"name,omitempty"`
	Status   *string                      `json:"status,omitempty"`
}

type RetypeParametersNewTypeIops struct {
	Read  int `json:"read"`
	Total int `json:"total"`
	Write int `json:"write"`
}

type RetypeConfigs struct {
	Env       *string `json:"env,omitempty"`
	Region    *string `json:"region,omitempty"`
	ServerUrl *string `json:"serverUrl,omitempty"`
}

func (s *service) Retype(
	parameters RetypeParameters,
	configs RetypeConfigs,
) (
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("Retype", mgcCore.RefPath("/block-storage/volumes/retype"), s.client, s.ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[RetypeParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[RetypeConfigs](configs); err != nil {
		return
	}

	_, err = exec.Execute(ctx, p, c)
	return
}

// Context from caller is used to allow cancellation of long-running requests
func (s *service) RetypeContext(
	ctx context.Context,
	parameters RetypeParameters,
	configs RetypeConfigs,
) (
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("Retype", mgcCore.RefPath("/block-storage/volumes/retype"), s.client, ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[RetypeParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[RetypeConfigs](configs); err != nil {
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

	_, err = exec.Execute(ctx, p, c)
	return
}

// TODO: links
// TODO: related
