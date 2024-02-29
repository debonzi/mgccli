package buckets

import (
	"magalu.cloud/core"
	"magalu.cloud/core/utils"
)

var GetGroup = utils.NewLazyLoader[core.Grouper](func() core.Grouper {
	return core.NewStaticGroup(
		core.DescriptorSpec{
			Name:        "buckets",
			Description: "Bucket operations for Object Storage API",
		},
		func() []core.Descriptor {
			return []core.Descriptor{
				getCreate(),    // object-storage buckets create
				getDelete(),    // object-storage buckets delete
				getList(),      // object-storage buckets list
				getPublicUrl(), // object-storage objects public-url
			}
		},
	)
})
