/*
Executor: create

# Summary

# Create Public IP XaaS

# Description

Create async Public IP in a VPC with provided vpc_id and x_tenant_id for XaaS

Version: 1.124.1

import "magalu.cloud/lib/products/network/xaas_public_ip/vpcs_public_ips"
*/
package vpcsPublicIps

import (
	mgcCore "magalu.cloud/core"
	mgcHelpers "magalu.cloud/lib/helpers"
)

type CreateParameters struct {
	Description   *string `json:"description,omitempty"`
	ProjectType   string  `json:"project_type"`
	ValidateQuota *bool   `json:"validate_quota,omitempty"`
	VpcId         string  `json:"vpc_id"`
	Wait          *bool   `json:"wait,omitempty"`
	WaitTimeout   *int    `json:"wait_timeout,omitempty"`
}

type CreateConfigs struct {
	Env       *string `json:"env,omitempty"`
	Region    *string `json:"region,omitempty"`
	ServerUrl *string `json:"serverUrl,omitempty"`
}

type CreateResult struct {
	Id string `json:"id"`
}

func (s *service) Create(
	parameters CreateParameters,
	configs CreateConfigs,
) (
	result CreateResult,
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("Create", mgcCore.RefPath("/network/xaas_public_ip/vpcs-public-ips/create"), s.client, s.ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters
	if p, err = mgcHelpers.ConvertParameters[CreateParameters](parameters); err != nil {
		return
	}

	var c mgcCore.Configs
	if c, err = mgcHelpers.ConvertConfigs[CreateConfigs](configs); err != nil {
		return
	}

	r, err := exec.Execute(ctx, p, c)
	if err != nil {
		return
	}
	return mgcHelpers.ConvertResult[CreateResult](r)
}

// TODO: links
// TODO: related
