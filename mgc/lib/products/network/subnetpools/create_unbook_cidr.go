/*
Executor: create-unbook-cidr

# Summary

# Unbook Subnetpool

# Description

# Unbooking a CIDR range from a subnetpool

Version: 1.141.3

import "magalu.cloud/lib/products/network/subnetpools"
*/
package subnetpools

import (
	"context"

	mgcCore "magalu.cloud/core"
	mgcHelpers "magalu.cloud/lib/helpers"
)

type CreateUnbookCidrParameters struct {
	Cidr         *string `json:"cidr,omitempty"`
	SubnetpoolId string  `json:"subnetpool_id"`
}

type CreateUnbookCidrConfigs struct {
	Env       *string `json:"env,omitempty"`
	Region    *string `json:"region,omitempty"`
	ServerUrl *string `json:"serverUrl,omitempty"`
}

func (s *service) CreateUnbookCidr(
	parameters CreateUnbookCidrParameters,
	configs CreateUnbookCidrConfigs,
) (
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("CreateUnbookCidr", mgcCore.RefPath("/network/subnetpools/create-unbook-cidr"), s.client, s.ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[CreateUnbookCidrParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[CreateUnbookCidrConfigs](configs); err != nil {
		return
	}

	_, err = exec.Execute(ctx, p, c)
	return
}

// Context from caller is used to allow cancellation of long-running requests
func (s *service) CreateUnbookCidrContext(
	ctx context.Context,
	parameters CreateUnbookCidrParameters,
	configs CreateUnbookCidrConfigs,
) (
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("CreateUnbookCidr", mgcCore.RefPath("/network/subnetpools/create-unbook-cidr"), s.client, ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[CreateUnbookCidrParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[CreateUnbookCidrConfigs](configs); err != nil {
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
