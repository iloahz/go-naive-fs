package naivefs

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/hirochachacha/go-smb2"
)

type FSSamba struct {
	config *SambaConfig
	lock   sync.Mutex
}

type SambaConfig struct {
	ServerName string
	Username   string
	Password   string
	ShareName  string
}

func NewFSSamba(config *SambaConfig) *FSSamba {
	samba := &FSSamba{
		config: config,
	}
	return samba
}

func ignoreError(f func() error) func() {
	return func() {
		f()
	}
}

func cancelWhenDone(ctx context.Context, cancelFunc func()) {
	go func() {
		<-ctx.Done()
		cancelFunc()
	}()
}

func (samba *FSSamba) withFS(f func(fs *smb2.Share) error) error {
	samba.lock.Lock()
	defer samba.lock.Unlock()
	addr := fmt.Sprintf("%s:445", samba.config.ServerName)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     samba.config.Username,
			Password: samba.config.Password,
		},
	}
	s, err := d.Dial(conn)
	if err != nil {
		return err
	}
	defer s.Logoff()
	fs, err := s.Mount(samba.config.ShareName)
	if err != nil {
		return err
	}
	defer fs.Umount()
	return f(fs)
}

// TODO: perf opt, share conn
func (samba *FSSamba) connect(ctx context.Context) (*smb2.Share, error) {
	samba.lock.Lock()
	cancelWhenDone(ctx, samba.lock.Unlock)
	addr := fmt.Sprintf("%s:445", samba.config.ServerName)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	cancelWhenDone(ctx, ignoreError(conn.Close))
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     samba.config.Username,
			Password: samba.config.Password,
		},
	}
	s, err := d.Dial(conn)
	if err != nil {
		return nil, err
	}
	cancelWhenDone(ctx, ignoreError(s.Logoff))
	fs, err := s.Mount(samba.config.ShareName)
	if err != nil {
		return nil, err
	}
	cancelWhenDone(ctx, ignoreError(fs.Umount))
	return fs, nil
}

func (samba *FSSamba) File(name string) *File {
	return &File{
		fs:   samba,
		name: name,
	}
}

func (samba *FSSamba) Touch(name string) error {
	if samba.Exists(name) {
		return nil
	}
	return samba.withFS(func(fs *smb2.Share) error {
		_, err := fs.Create(name)
		return err
	})
}

func (samba *FSSamba) MkDir(name string) error {
	if samba.Exists(name) {
		return nil
	}
	return samba.withFS(func(fs *smb2.Share) error {
		return fs.MkdirAll(name, os.ModePerm)
	})
}

func (samba *FSSamba) Remove(name string) error {
	if !samba.Exists(name) {
		return nil
	}
	return samba.withFS(func(fs *smb2.Share) error {
		return fs.RemoveAll(name)
	})
}

func (samba *FSSamba) Write(name string, data []byte) error {
	return samba.withFS(func(fs *smb2.Share) error {
		return fs.WriteFile(name, data, os.ModePerm)
	})
}

func (samba *FSSamba) Read(name string) (data []byte, err error) {
	samba.withFS(func(fs *smb2.Share) error {
		data, err = fs.ReadFile(name)
		return nil
	})
	return
}

func (samba *FSSamba) Exists(name string) (exists bool) {
	samba.withFS(func(fs *smb2.Share) error {
		_, err := fs.Stat(name)
		exists = !errors.Is(err, os.ErrNotExist)
		return nil
	})
	return
}

func (samba *FSSamba) IsDir(name string) (isDir bool) {
	samba.withFS(func(fs *smb2.Share) error {
		stat, err := fs.Stat(name)
		if err != nil {
			isDir = false
		} else {
			isDir = stat.IsDir()
		}
		return nil
	})
	return
}

func (samba *FSSamba) SupportDir() bool {
	return true
}
