package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/anishmgoyal/calagora/constants"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadFileToPublic uploads a file to the public AWS S3 bucket
func UploadFileToPublic(name, mimeType string, file *os.File) bool {
	file.Seek(0, os.SEEK_SET)

	if constants.DoUploadAWS {
		svc := s3.New(session.New(), &aws.Config{
			Region: aws.String(constants.S3RegionString),
		})

		_, err := svc.PutObject(&s3.PutObjectInput{
			Bucket:      aws.String(constants.S3Bucket),
			ContentType: aws.String(mimeType),
			Key:         aws.String(constants.S3ObjectPrefix + name),
			Body:        file,
		})
		if err != nil {
			fmt.Println("ERROR!")
			fmt.Println(err.Error())
			return false
		}
		return true
	}

	// Defaults to uploading in place
	outputFile, err := os.Create(constants.FileSaveDir +
		string(os.PathSeparator) + name)
	if err != nil || outputFile == nil {
		return false
	}
	defer outputFile.Close()
	if _, err = io.Copy(outputFile, file); err != nil {
		return false
	}
	err = outputFile.Sync()
	if err != nil {
	}
	return err == nil
}

// DeleteFileFromPublic deletes a file uploaded to public from the AWS S3 Bucket
func DeleteFileFromPublic(name string) bool {
	if constants.DoUploadAWS {
		svc := s3.New(session.New(),
			&aws.Config{Region: aws.String(constants.S3RegionString)})
		_, err := svc.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(constants.S3Bucket),
			Key:    aws.String(constants.S3ObjectPrefix + name),
		})
		if err != nil {
			fmt.Println("ERROR!")
			fmt.Println(err.Error())
			return false
		}
		return true
	}
	err := os.Remove(constants.FileSaveDir + string(os.PathSeparator) + name)
	return err == nil
}
