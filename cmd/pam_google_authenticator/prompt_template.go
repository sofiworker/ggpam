package main

/*
#cgo LDFLAGS: -lpam
#include <security/pam_appl.h>
#include <security/pam_modules.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"text/template"
	"unsafe"

	"golang.org/x/sys/unix"

	pamcfg "gpam/pkg/pam"
)

const maxPromptTemplateSize = 4096

type promptContext struct {
	User    string
	Host    string
	Service string
}

type PromptContextProvider interface {
	Build(pamh *C.pam_handle_t, account *user.User, user, host string) promptContext
}

var promptProvider PromptContextProvider = defaultPromptProvider{}

func RegisterPromptContextProvider(p PromptContextProvider) {
	if p != nil {
		promptProvider = p
	}
}

type defaultPromptProvider struct{}

func (defaultPromptProvider) Build(pamh *C.pam_handle_t, account *user.User, user, host string) promptContext {
	return promptContext{
		User:    user,
		Host:    host,
		Service: getPamService(pamh),
	}
}

func preparePromptFromTemplate(pamh *C.pam_handle_t, pathSpec string, account *user.User, user, host string) (string, error) {
	tmplPath, err := pamcfg.ResolveSecretPath(pathSpec, account)
	if err != nil {
		return "", err
	}
	raw, err := loadPromptTemplate(tmplPath)
	if err != nil {
		return "", err
	}
	return renderPromptTemplate(raw, promptProvider.Build(pamh, account, user, host))
}

func loadPromptTemplate(path string) (string, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return "", err
	}
	file := os.NewFile(uintptr(fd), path)
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxPromptTemplateSize+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxPromptTemplateSize {
		return "", fmt.Errorf("prompt 模板超过 %d 字节", maxPromptTemplateSize)
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}

func renderPromptTemplate(raw string, ctx promptContext) (string, error) {
	tmpl, err := template.New("prompt").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func getPamService(pamh *C.pam_handle_t) string {
	var item unsafe.Pointer
	if C.pam_get_item(pamh, C.PAM_SERVICE, &item) != C.PAM_SUCCESS || item == nil {
		return ""
	}
	return C.GoString((*C.char)(item))
}
