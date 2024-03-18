/*
Executor: list

# Description

# List valid Object Storage credentials

import "magalu.cloud/lib/products/object_storage/api_key"
*/
package apiKey

import (
	"context"

	mgcCore "magalu.cloud/core"
	mgcClient "magalu.cloud/lib"
	mgcHelpers "magalu.cloud/lib/helpers"
)

type ListResultItem struct {
	Description   string `json:"description"`
	EndValidity   string `json:"end_validity,omitempty"`
	KeyPairId     string `json:"key_pair_id"`
	KeyPairSecret string `json:"key_pair_secret"`
	Name          string `json:"name"`
	RevokedAt     string `json:"revoked_at,omitempty"`
	StartValidity string `json:"start_validity"`
	TenantName    string `json:"tenant_name,omitempty"`
	Uuid          string `json:"uuid"`
}

type ListResult []ListResultItem

func List(
	client *mgcClient.Client,
	ctx context.Context,
) (
	result ListResult,
	err error,
) {
	exec, ctx, err := mgcHelpers.PrepareExecutor("List", mgcCore.RefPath("/object-storage/api-key/list"), client, ctx)
	if err != nil {
		return
	}

	var p mgcCore.Parameters

	var c mgcCore.Configs

	r, err := exec.Execute(ctx, p, c)
	if err != nil {
		return
	}
	return mgcHelpers.ConvertResult[ListResult](r)
}

// TODO: links
// TODO: related
