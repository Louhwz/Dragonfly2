/*
 *     Copyright 2020 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
//go:generate mockgen -destination ./mock/mock_source_client.go -package mock d7y.io/dragonfly/v2/pkg/source ResourceClient

package source

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

var clients = make(map[string]ResourceClient)

type Header map[string]string

func Register(schema string, resourceClient ResourceClient) {
	clients[strings.ToLower(schema)] = resourceClient
}

func UnRegister(schema string) {
	delete(clients, strings.ToLower(schema))
}

// ResourceClient supply apis that interact with the source.
type ResourceClient interface {

	// GetContentLength get length of resource content
	// return -l if request fail
	// return -1 if response status is not StatusOK and StatusPartialContent
	GetContentLength(ctx context.Context, url string, header Header) (int64, error)

	// IsSupportRange checks if resource supports breakpoint continuation
	IsSupportRange(ctx context.Context, url string, header Header) (bool, error)

	// IsExpired checks if a resource received or stored is the same.
	IsExpired(ctx context.Context, url string, header Header, expireInfo map[string]string) (bool, error)

	// Download download from source
	Download(ctx context.Context, url string, header Header) (io.ReadCloser, error)

	// DownloadWithExpire download from source and get expire info
	DownloadWithExpire(ctx context.Context, url string, header Header) (io.ReadCloser, map[string]string, error)

	// GetExpireInfo get expire info of resource
	GetExpireInfo(ctx context.Context, url string, header Header) (map[string]string, error)
}

func GetContentLength(ctx context.Context, url string, header Header) (int64, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return -1, err
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	return sourceClient.GetContentLength(ctx, url, header)
}

func IsSupportRange(ctx context.Context, url string, header Header) (bool, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return false, err
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	return sourceClient.IsSupportRange(ctx, url, header)
}

func IsExpired(ctx context.Context, url string, header Header, expireInfo map[string]string) (bool, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return false, err
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	return sourceClient.IsExpired(ctx, url, header, expireInfo)
}

func Download(ctx context.Context, url string, header Header) (io.ReadCloser, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return nil, err
	}
	return sourceClient.Download(ctx, url, header)
}

func DownloadWithExpire(ctx context.Context, url string, header Header) (io.ReadCloser, map[string]string, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return nil, nil, err
	}
	return sourceClient.DownloadWithExpire(ctx, url, header)
}

func GetExpireInfo(ctx context.Context, url string, header Header) (map[string]string, error) {
	sourceClient, err := getSourceClient(url)
	if err != nil {
		return nil, err
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
	}
	return sourceClient.GetExpireInfo(ctx, url, header)
}

// getSourceClient get a source client from source manager with specified schema.
func getSourceClient(rawURL string) (ResourceClient, error) {
	url, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	client, ok := clients[strings.ToLower(url.Scheme)]
	if !ok || client == nil {
		return nil, fmt.Errorf("can not find client for supporting url %s", rawURL)
	}
	return client, nil
}
