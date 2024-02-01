package objects

import (
	"context"
	"fmt"
	"strings"

	"magalu.cloud/core"
	"magalu.cloud/core/utils"
	"magalu.cloud/sdk/static/object_storage/common"
)

var getCopy = utils.NewLazyLoader[core.Executor](func() core.Executor {
	executor := core.NewStaticExecute(
		core.DescriptorSpec{
			Name:        "copy",
			Description: "Copy an object from a bucket to another",
		},
		copy,
	)

	return core.NewExecuteResultOutputOptions(executor, func(exec core.Executor, result core.Result) string {
		return "template=Copied from {{.src}} to {{.dst}}\n"
	})
})

func copy(ctx context.Context, p common.CopyObjectParams, cfg common.Config) (result core.Value, err error) {
	_, err = common.HeadFile(ctx, cfg, p.Source)
	if err != nil {
		return nil, fmt.Errorf("error validating source: %w", err)
	}

	fileName := p.Source.Filename()
	if fileName == "" {
		return nil, core.UsageError{Err: fmt.Errorf("source must be a URI to an object")}
	}

	fullDstPath := p.Destination
	if fullDstPath == "" {
		return nil, core.UsageError{Err: fmt.Errorf("destination cannot be empty")}
	}

	if strings.HasSuffix(fullDstPath.String(), "/") || p.Destination.IsRoot() {
		// If it isn't a file path, don't rename, just append source with bucket URI
		fullDstPath = fullDstPath.JoinPath(fileName)
	}

	err = common.CopySingleFile(ctx, cfg, p.Source, fullDstPath)
	if err != nil {
		return nil, err
	}

	return common.CopyObjectParams{Source: p.Source, Destination: fullDstPath}, err
}
