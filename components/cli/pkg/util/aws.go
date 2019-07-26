/*
 * Copyright (c) 2019 WSO2 Inc. (http:www.wso2.org) All Rights Reserved.
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
 *
 */

package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/cellery-io/sdk/components/cli/pkg/constants"
)

func DownloadFromS3Bucket(bucket, item, path string, displayProgressBar bool) {
	file, err := os.Create(filepath.Join(path, item))
	if err != nil {
		ExitWithErrorMessage("Failed to create file path "+path, err)
	}

	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(constants.AWS_REGION), Credentials: credentials.AnonymousCredentials},
	)

	// Get the object size
	s3ObjectSize := GetS3ObjectSize(bucket, item)

	// Create a downloader with the session and custom options
	downloader := s3manager.NewDownloader(sess, func(d *s3manager.Downloader) {
		d.PartSize = 64 * 1024 * 1024 // 64MB per part
		d.Concurrency = 6
	})

	writer := &progressWriter{writer: file, size: s3ObjectSize}
	writer.display = displayProgressBar

	writer.init(s3ObjectSize)
	_, err = downloader.Download(writer,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		ExitWithErrorMessage(fmt.Sprintf("Failed to download %s from s3 bucket %s", item, bucket), err)
	}

	writer.finish()
	fmt.Println("Completed downloading", item)
}

func GetS3ObjectSize(bucket, item string) int64 {
	return *getS3Object(bucket, item).ContentLength
}

func GetS3ObjectEtag(bucket, item string) string {
	return *getS3Object(bucket, item).ETag
}

func getS3Object(bucket, item string) s3.HeadObjectOutput {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(constants.AWS_REGION), Credentials: credentials.AnonymousCredentials},
	)

	svc := s3.New(sess)
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(item),
	}

	result, err := svc.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			ExitWithErrorMessage("Error getting size of file", aerr)
		} else {
			ExitWithErrorMessage("Error getting size of file", err)
		}
	}
	return *result
}
