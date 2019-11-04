/*
 * Copyright (c) 2018 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
 *
 * WSO2 Inc. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http:www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package util

import (
	"io"
	"sync"

	"github.com/tj/go-spin"
	"gopkg.in/cheggaaa/pb.v1"
)

type Spinner struct {
	mux            sync.Mutex
	core           *spin.Spinner
	action         string
	previousAction string
	isRunning      bool
	isSpinning     bool
	error          bool
}

type Gcp struct {
	Compute GcpCompute `json:"compute"`
	Core    GcpCore    `json:"core"`
}

type GcpCompute struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
}

type GcpCore struct {
	Account string `json:"account"`
	Project string `json:"project"`
}

type RegistryCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type progressWriter struct {
	writer  io.WriterAt
	size    int64
	bar     *pb.ProgressBar
	display bool
}

func (pw *progressWriter) init(s3ObjectSize int64) {
	if pw.display {
		pw.bar = pb.StartNew(int(s3ObjectSize))
		pw.bar.ShowSpeed = true
		pw.bar.Format("[=>_]")
		pw.bar.SetUnits(pb.U_BYTES_DEC)
	}
}

func (pw *progressWriter) finish() {
	if pw.display {
		pw.bar.Finish()
	}
}

func (pw *progressWriter) setProgress(length int64) {
	pw.bar.Set64(length)
}
