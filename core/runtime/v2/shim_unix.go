//go:build !windows

/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package v2

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/containerd/fifo"
	"golang.org/x/sys/unix"
)

func openShimLog(ctx context.Context, bundle *Bundle, _ func(string, time.Duration) (net.Conn, error)) (io.ReadCloser, error) {
	return fifo.OpenFifo(ctx, filepath.Join(bundle.Path, "log"), unix.O_RDWR|unix.O_CREAT|unix.O_NONBLOCK, 0700)
}

func checkCopyShimLogError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	
	// Always ignore these expected errors during cleanup regardless of context state
	if err == fifo.ErrReadClosed || 
	   errors.Is(err, os.ErrClosed) ||
	   strings.Contains(err.Error(), "file already closed") ||
	   strings.Contains(err.Error(), "broken pipe") ||
	   strings.Contains(err.Error(), "connection reset") {
		return nil
	}
	
	// If context is done, only log unexpected errors
	select {
	case <-ctx.Done():
		return nil
	default:
		return err
	}
}
