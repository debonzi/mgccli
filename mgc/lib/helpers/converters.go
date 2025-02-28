package helpers

import (
	"fmt"
	"io"

	mgcCore "magalu.cloud/core"
	mgcUtils "magalu.cloud/core/utils"
)

func ConvertResult[T any](r mgcCore.Result) (result T, err error) {
	rValue, ok := mgcCore.ResultAs[mgcCore.ResultWithValue](r)
	if !ok {
		err = &mgcUtils.ChainedError{Name: "result", Err: fmt.Errorf("result does not contain a value: %#v", r)}
		return
	}
	if err = mgcUtils.DecodeValue(rValue.Value(), &result); err != nil {
		err = &mgcUtils.ChainedError{Name: "result", Err: err}
		return
	}
	return
}

func ConvertResultReader[T any](r mgcCore.Result) (result T, err error) {
	rValue, ok := mgcCore.ResultAs[mgcCore.ResultWithReader](r)
	if !ok {
		err = &mgcUtils.ChainedError{Name: "result", Err: fmt.Errorf("result does not contain a value: %#v", r)}
		return
	}
	if closer, ok := rValue.(io.Closer); ok {
		defer closer.Close()
	}

	bytes, err := io.ReadAll(rValue.Reader())

	if err != nil {
		err = &mgcUtils.ChainedError{Name: "result", Err: err}
	}

	if err = mgcUtils.DecodeValue(bytes, &result); err != nil {
		err = &mgcUtils.ChainedError{Name: "result", Err: err}
		return
	}
	return
}

func convertInput[I any, O map[string]any](input I, kind string) (output O, err error) {
	var v any
	if v, err = mgcUtils.SimplifyAny(input); err != nil {
		err = &mgcUtils.ChainedError{Name: kind, Err: err}
		return
	}

	var ok bool
	if output, ok = v.(O); !ok {
		err = &mgcUtils.ChainedError{Name: kind, Err: fmt.Errorf("invalid type: %#v", v)}
		return
	}

	return
}

func ConvertParameters[T any](parameters T) (p mgcCore.Parameters, err error) {
	return convertInput[T, mgcCore.Parameters](parameters, "parameters")
}

func ConvertConfigs[T any](configs T) (c mgcCore.Configs, err error) {
	return convertInput[T, mgcCore.Configs](configs, "configs")
}
