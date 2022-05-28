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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return err
	}
	_, err = fs.Create(name)
	if err != nil {
		return err
	}
	return nil
}

func (samba *FSSamba) MkDir(name string) error {
	if samba.Exists(name) {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return err
	}
	return fs.MkdirAll(name, os.ModePerm)
}

func (samba *FSSamba) Remove(name string) error {
	if !samba.Exists(name) {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return err
	}
	return fs.RemoveAll(name)
}

func (samba *FSSamba) Write(name string, data []byte) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return err
	}
	return fs.WriteFile(name, data, os.ModePerm)
}

func (samba *FSSamba) Read(name string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(name)
}

func (samba *FSSamba) Exists(name string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return false
	}
	_, err = fs.Stat(name)
	return !errors.Is(err, os.ErrNotExist)
}

func (samba *FSSamba) IsDir(name string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fs, err := samba.connect(ctx)
	if err != nil {
		return false
	}
	stat, err := fs.Stat(name)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
