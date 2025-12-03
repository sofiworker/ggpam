package pam

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"ggpam/pkg/config"
)

type FileState struct {
	Dev       uint64
	Ino       uint64
	Size      int64
	MtimeNano int64
}

var ErrSecretModified = errors.New("secret file modified")

func ResolveSecretPath(spec string, account *user.User) (string, error) {
	username := account.Username
	home := account.HomeDir
	if spec == "" {
		return filepath.Join(home, ".google_authenticator"), nil
	}
	path := strings.ReplaceAll(spec, "%u", username)
	path = strings.ReplaceAll(path, "%h", home)
	if strings.HasPrefix(path, "~") {
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return os.ExpandEnv(path), nil
}

func validateSecretFile(info os.FileInfo, account *user.User, params Params) error {
	if info.IsDir() {
		return fmt.Errorf("不是普通文件")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("不允许符号链接")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("必须是普通文件")
	}
	mode := info.Mode().Perm()
	if mode > params.AllowedPerm {
		return fmt.Errorf("权限 %04o 超过 allowed_perm %04o", mode, params.AllowedPerm)
	}
	if !params.NoStrictOwner {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("无法获取文件所有者")
		}
		wantUID, err := strconv.ParseUint(account.Uid, 10, 32)
		if err != nil {
			return fmt.Errorf("无法解析用户 UID: %v", err)
		}
		if stat.Uid != uint32(wantUID) {
			return fmt.Errorf("文件 owner=%d, 需要 %s(%s)", stat.Uid, account.Username, account.Uid)
		}
	}
	return nil
}

func LoadConfig(account *user.User, path string, params Params) (*config.Config, FileState, error) {
	f, err := openLocked(path, unix.O_RDONLY, 0)
	if err != nil {
		return nil, FileState{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, FileState{}, err
	}
	if err := validateSecretFile(info, account, params); err != nil {
		return nil, FileState{}, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, FileState{}, err
	}
	cfg, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, FileState{}, err
	}
	state := newFileState(info)
	return cfg, state, nil
}

func WriteConfig(account *user.User, path string, data []byte, perm os.FileMode, expected FileState) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !expected.isZero() && !expected.matches(info) {
		return ErrSecretModified
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ga-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	if dirFile, err := os.Open(dir); err == nil {
		dirFile.Sync()
		dirFile.Close()
	}
	return nil
}

func openLocked(path string, flags int, perm os.FileMode) (*os.File, error) {
	fd, err := unix.Open(path, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW, uint32(perm))
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		unix.Close(fd)
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func newFileState(info os.FileInfo) FileState {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return FileState{}
	}
	return FileState{
		Dev:       uint64(stat.Dev),
		Ino:       stat.Ino,
		Size:      info.Size(),
		MtimeNano: info.ModTime().UnixNano(),
	}
}

func (s FileState) isZero() bool {
	return s.Dev == 0 && s.Ino == 0
}

func (s FileState) matches(info os.FileInfo) bool {
	if s.isZero() {
		return false
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	return s.Dev == uint64(stat.Dev) &&
		s.Ino == stat.Ino &&
		s.Size == info.Size() &&
		s.MtimeNano == info.ModTime().UnixNano()
}
