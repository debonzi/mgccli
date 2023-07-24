package sdk

import "magalu.cloud/core"

func newStaticRootVpcPort() *core.StaticGroup {
	return core.NewStaticGroup(
		"port",
		"",
		"",
		[]core.Descriptor{
			newStaticRootVpcPortStatic(), // cmd: vpc port static
		},
	)
}
