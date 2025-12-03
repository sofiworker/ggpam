package main

import (
	"github.com/opencontainers/selinux/go-selinux"
)

const selinuxSecretType = "auth_home_t"

func applySelinuxContext(path string) error {
	if !selinux.GetEnabled() {
		return nil
	}
	current, err := selinux.FileLabel(path)
	if err != nil {
		return err
	}
	ctx, err := selinux.NewContext(current)
	if err != nil {
		return err
	}
	ctx["type"] = selinuxSecretType
	return selinux.SetFileLabel(path, ctx.Get())
}
