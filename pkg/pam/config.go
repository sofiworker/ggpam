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

func validateSecretFile(path string, info os.FileInfo, account *user.User, params Params) error {
	if info.IsDir() {
		return fmt.Errorf("secret file %s is a directory", path)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("secret file %s must not be a symlink", path)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("secret file %s must be a regular file", path)
	}
	mode := info.Mode().Perm()
	if mode > params.AllowedPerm {
		return fmt.Errorf("secret file permissions %04o exceed allowed_perm %04o", mode, params.AllowedPerm)
	}
	if !params.NoStrictOwner {
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("secret file %s: unable to read owner metadata", path)
		}
		wantUID, err := strconv.ParseUint(account.Uid, 10, 32)
		if err != nil {
			return fmt.Errorf("secret file %s: parse user UID %q failed: %w", path, account.Uid, err)
		}
		if stat.Uid != uint32(wantUID) {
			return fmt.Errorf("secret file %s owner=%d, expected %s(%s)", path, stat.Uid, account.Username, account.Uid)
		}
	}
	return nil
}

func LoadConfig(account *user.User, path string, params Params) (*config.Config, FileState, error) {
	f, err := openLocked(path, unix.O_RDONLY, 0)
	if err != nil {
		return nil, FileState{}, fmt.Errorf("open secret file %s: %w", path, err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, FileState{}, fmt.Errorf("stat secret file %s: %w", path, err)
	}
	if err := validateSecretFile(path, info, account, params); err != nil {
		return nil, FileState{}, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, FileState{}, fmt.Errorf("read secret file %s: %w", path, err)
	}
	cfg, err := config.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, FileState{}, fmt.Errorf("parse secret file %s: %w", path, err)
	}
	state := newFileState(info)
	return cfg, state, nil
}

func WriteConfig(account *user.User, path string, data []byte, perm os.FileMode, expected FileState) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("stat secret file %s: %w", path, err)
	}
	if !expected.isZero() && !expected.matches(info) {
		return ErrSecretModified
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ga-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp file %s: %w", tmpName, err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp file %s to %s: %w", tmpName, path, err)
	}
	if dirFile, err := os.Open(dir); err == nil {
		defer dirFile.Close()
		if err := dirFile.Sync(); err != nil {
			return fmt.Errorf("sync dir %s: %w", dir, err)
		}
	} else {
		return fmt.Errorf("open dir %s: %w", dir, err)
	}
	return nil
}

func openLocked(path string, flags int, perm os.FileMode) (*os.File, error) {
	fd, err := unix.Open(path, flags|unix.O_CLOEXEC|unix.O_NOFOLLOW, uint32(perm))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("flock %s: %w", path, err)
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
