package naivefs

import (
	"bytes"
	"context"
	"io"
	"path"
	"sync"

	minioSDK "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FSMinio struct {
	client *minioSDK.Client
	config *MinioConfig
	lock   sync.Mutex
}

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	BaseDir         string
}

func NewFSMinio(config *MinioConfig) *FSMinio {
	fsMinio := &FSMinio{
		config: config,
	}
	return fsMinio
}

func (minio *FSMinio) withClient(f func(*minioSDK.Client) error) error {
	minio.lock.Lock()
	defer minio.lock.Unlock()
	if minio.client == nil {
		var err error
		minio.client, err = minioSDK.New(minio.config.Endpoint, &minioSDK.Options{
			Creds:  credentials.NewStaticV4(minio.config.AccessKeyID, minio.config.SecretAccessKey, ""),
			Secure: minio.config.UseSSL,
		})
		if err != nil {
			return err
		}
	}
	return f(minio.client)
}

func (minio *FSMinio) toAbs(name string) string {
	if path.IsAbs(name) {
		return name
	}
	return path.Join(minio.config.BaseDir, name)
}

func (minio *FSMinio) File(name string) *File {
	return &File{
		fs:   minio,
		name: name,
	}
}

func (minio *FSMinio) Touch(name string) error {
	return minio.Write(name, []byte{})
}

func (minio *FSMinio) MkDir(name string) error {
	// no-op
	return nil
}

func (minio *FSMinio) Remove(name string) (err error) {
	return minio.withClient(func(client *minioSDK.Client) error {
		return client.RemoveObject(context.Background(), minio.config.BucketName, name, minioSDK.RemoveObjectOptions{})
	})
}

func (minio *FSMinio) Write(name string, data []byte) error {
	return minio.withClient(func(client *minioSDK.Client) error {
		reader := bytes.NewReader(data)
		_, err := client.PutObject(context.Background(), minio.config.BucketName, name, reader, int64(len(data)), minioSDK.PutObjectOptions{})
		return err
	})
}

func (minio *FSMinio) Read(name string) (data []byte, err error) {
	minio.withClient(func(client *minioSDK.Client) error {
		var obj *minioSDK.Object
		obj, err = client.GetObject(context.Background(), minio.config.BucketName, name, minioSDK.GetObjectOptions{})
		if err != nil {
			return err
		}
		data, err = io.ReadAll(obj)
		return err
	})
	return
}

func (minio *FSMinio) Exists(name string) (exists bool) {
	minio.withClient(func(client *minioSDK.Client) error {
		_, err := client.StatObject(context.Background(), minio.config.BucketName, name, minioSDK.StatObjectOptions{})
		if err != nil {
			exists = false
			return err
		}
		exists = true
		return nil
	})
	return
}

func (minio *FSMinio) IsDir(name string) (isDir bool) {
	minio.withClient(func(client *minioSDK.Client) error {
		_, err := client.StatObject(context.Background(), minio.config.BucketName, name, minioSDK.StatObjectOptions{})
		if err != nil {
			isDir = false
			return err
		}
		return nil
	})
	return
}

func (minio *FSMinio) SupportDir() bool {
	return false
}
