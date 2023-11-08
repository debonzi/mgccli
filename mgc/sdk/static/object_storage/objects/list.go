package objects

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"magalu.cloud/core"
	"magalu.cloud/core/pipeline"
	"magalu.cloud/core/utils"
	"magalu.cloud/sdk/static/object_storage/common"
)

var listObjectsLogger = utils.NewLazyLoader(func() *zap.SugaredLogger {
	return logger().Named("list")
})

type ListObjectsParams struct {
	Destination      string           `json:"dst" jsonschema:"description=Path of the bucket to list objects from,example=s3://bucket1/"`
	PaginationParams `json:",squash"` // nolint
}

type PaginationParams struct {
	MaxItems          int    `json:"max-items,omitempty" jsonschema:"description=Limit of items to be listed,default=1000,minimum=1,=example=1000"`
	ContinuationToken string `json:"continuation-token,omitempty" jsonschema:"description=Token of result page to continue from"`
}

type prefix struct {
	Path string `xml:"Prefix"`
}

type ListObjectsResponse struct {
	Name                   string           `xml:"Name"`
	Contents               []*BucketContent `xml:"Contents"`
	CommonPrefixes         []*prefix        `xml:"CommonPrefixes" json:"SubDirectories"`
	PaginationResponseInfo `json:",squash"` // nolint
}

type PaginationResponseInfo struct {
	NextContinuationToken string `xml:"NextContinuationToken"`
	IsTruncated           bool   `xml:"IsTruncated"`
}

type BucketContent struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ContentSize  int64  `xml:"Size"`
}

type BucketContentDirEntry = *pipeline.SimpleWalkDirEntry[*BucketContent]

func (b *BucketContent) ModTime() time.Time {
	modTime, err := time.Parse(time.RFC3339, b.LastModified)
	if err != nil {
		listObjectsLogger().Named("BucketContent.ModTime()").Errorw("failed to parse time", "err", err, "key", b.Key, "lastModified", b.LastModified)
		modTime = time.Time{}
	}
	return modTime
}

func (b *BucketContent) Mode() fs.FileMode {
	return utils.FILE_PERMISSION
}

func (b *BucketContent) Size() int64 {
	return b.ContentSize
}

func (b *BucketContent) Sys() any {
	return nil
}

func (b *BucketContent) Info() (fs.FileInfo, error) {
	return b, nil
}

func (b *BucketContent) IsDir() bool {
	return false
}

func (b *BucketContent) Name() string {
	return b.Key
}

func (b *BucketContent) Type() fs.FileMode {
	return utils.FILE_PERMISSION
}

var _ fs.DirEntry = (*BucketContent)(nil)
var _ fs.FileInfo = (*BucketContent)(nil)

func newListRequest(ctx context.Context, cfg common.Config, bucket string, page PaginationParams) (*http.Request, error) {
	parsedUrl, err := parseURL(cfg, bucket)
	if err != nil {
		return nil, err
	}

	listReqQuery := parsedUrl.Query()
	listReqQuery.Set("list-type", "2")
	if page.ContinuationToken != "" {
		listReqQuery.Set("continuation-token", page.ContinuationToken)
	}
	if page.MaxItems <= 0 {
		return nil, fmt.Errorf("invalid item limit MaxItems, must be higher than zero: %d", page.MaxItems)
	} else if page.MaxItems > common.ApiLimitMaxItems {
		page.MaxItems = common.ApiLimitMaxItems
	}
	listReqQuery.Set("max-keys", fmt.Sprint(page.MaxItems))
	parsedUrl.RawQuery = listReqQuery.Encode()

	return http.NewRequestWithContext(ctx, http.MethodGet, parsedUrl.String(), nil)
}

var getList = utils.NewLazyLoader[core.Executor](newList)

func newList() core.Executor {
	return core.NewStaticExecute(
		core.DescriptorSpec{
			Name:        "list",
			Description: "List all objects from a bucket",
		},
		List,
	)
}

func parseURL(cfg common.Config, bucketURI string) (*url.URL, error) {
	// Bucket URI cannot end in '/' as this makes it search for a
	// non existing directory
	bucketURI = strings.TrimSuffix(bucketURI, "/")
	dirs := strings.Split(bucketURI, "/")
	path, err := url.JoinPath(common.BuildHost(cfg), dirs[0])
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if len(dirs) <= 1 {
		return u, nil
	}
	q := u.Query()
	delimiter := "/"
	prefixQ := strings.Join(dirs[1:], delimiter)
	lastChar := string(prefixQ[len(prefixQ)-1])
	if lastChar != delimiter {
		prefixQ += delimiter
	}
	q.Set("prefix", prefixQ)
	q.Set("delimiter", delimiter)
	q.Set("encoding-type", "url")
	u.RawQuery = q.Encode()
	return u, nil
}

func List(ctx context.Context, params ListObjectsParams, cfg common.Config) (result ListObjectsResponse, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	objChan := ListGenerator(ctx, params, cfg)

	entries, err := pipeline.SliceItemLimitedConsumer[[]BucketContentDirEntry](ctx, params.MaxItems, objChan)
	if err != nil {
		return result, err
	}

	contents := make([]*BucketContent, 0, len(entries))
	for _, entry := range entries {
		if entry.Err() != nil {
			return result, entry.Err()
		}

		contents = append(contents, entry.Object)
	}

	result = ListObjectsResponse{
		Contents: contents,
	}
	return result, nil
}

func ListGenerator(ctx context.Context, params ListObjectsParams, cfg common.Config) (outputChan <-chan BucketContentDirEntry) {
	ch := make(chan BucketContentDirEntry)
	outputChan = ch

	logger := listObjectsLogger().Named("ListGenerator").With(
		"params", params,
		"cfg", cfg,
	)

	generator := func() {
		defer func() {
			close(ch)
			logger.Info("closed output channel")
		}()

		page := params.PaginationParams
		var requestedItems int
		bucket, _ := strings.CutPrefix(params.Destination, common.URIPrefix)
		for {
			requestedItems = 0

			req, err := newListRequest(ctx, cfg, bucket, page)
			var result ListObjectsResponse
			if err == nil {
				result, _, err = common.SendRequest[ListObjectsResponse](ctx, req)
			}

			if err != nil {
				logger.Errorw("list request failed", "err", err, "req", req)
				select {
				case <-ctx.Done():
					logger.Debugw("context.Done()", "err", err)
				case ch <- pipeline.NewSimpleWalkDirEntry[*BucketContent](params.Destination, nil, err):
				}
				return
			}

		listObjectsResponseLoop:
			for _, content := range result.Contents {
				dirEntry := pipeline.NewSimpleWalkDirEntry(
					path.Join(params.Destination, content.Key),
					content,
					nil,
				)

				select {
				case <-ctx.Done():
					logger.Debugw("context.Done()", "err", ctx.Err())
					return
				case ch <- dirEntry:
					requestedItems++
					if requestedItems >= page.MaxItems {
						logger.Infow("item limit reached", "limit", params.PaginationParams.MaxItems)
						break listObjectsResponseLoop
					}
				}
			}

			page.ContinuationToken = result.NextContinuationToken
			page.MaxItems = page.MaxItems - requestedItems
			if !result.IsTruncated || page.MaxItems <= 0 {
				logger.Info("finished reading contents")
				break
			}
		}
	}

	logger.Info("list generation start")
	go generator()
	return
}
