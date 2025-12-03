package main

/*
#cgo LDFLAGS: -lpam
#include <security/pam_modules.h>
#include <security/pam_appl.h>
#include <security/pam_ext.h>
#include <syslog.h>
#include <stdlib.h>

typedef const char *pam_const_char;

static int prompt_wrapper(pam_handle_t *pamh, int style, char **resp, const char *text) {
	return pam_prompt(pamh, style, resp, "%s", text);
}

static void error_wrapper(pam_handle_t *pamh, const char *text) {
	pam_error(pamh, "%s", text);
}

static void syslog_wrapper(pam_handle_t *pamh, int priority, const char *text) {
	pam_syslog(pamh, priority, "%s", text);
}
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	"ggpam/pkg/authenticator"
	"ggpam/pkg/config"
	"ggpam/pkg/i18n"
	"ggpam/pkg/logging"
	pamcfg "ggpam/pkg/pam"
)

var (
	fallbackUser = "nobody"
)

func msg(key string, args ...any) string {
	return i18n.Msgf(key, args...)
}

//export pam_sm_authenticate
func pam_sm_authenticate(pamh *C.pam_handle_t, flags C.int, argc C.int, argv *C.pam_const_char) C.int {
	return goPamAuthenticate(pamh, flags, argc, (**C.char)(unsafe.Pointer(argv)))
}

//export pam_sm_setcred
func pam_sm_setcred(pamh *C.pam_handle_t, flags C.int, argc C.int, argv *C.pam_const_char) C.int {
	return goPamSetcred(pamh, flags, argc, (**C.char)(unsafe.Pointer(argv)))
}

func goPamAuthenticate(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	_ = logging.ConfigureDefault("")
	args := parsePamArgs(argc, argv)
	params, err := pamcfg.ParseParams(args)
	if err != nil {
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgInvalidArgs, err))
		return C.PAM_SERVICE_ERR
	}
	return runPamAuth(pamh, params)
}

func goPamSetcred(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	return C.PAM_SUCCESS
}

func runPamAuth(pamh *C.pam_handle_t, params pamcfg.Params) C.int {
	pamUser, rc := getPamUser(pamh)
	if rc != C.PAM_SUCCESS || pamUser == "" {
		return rc
	}
	targetUser := pamUser
	if params.ForcedUser != "" {
		targetUser = params.ForcedUser
	}
	pamDebugf(pamh, params, "start for user %s", targetUser)

	account, err := lookupAccount(targetUser)
	if err != nil {
		pamSyslog(pamh, C.LOG_WARNING, msg(i18n.MsgUserLookupFailed, targetUser, err))
		if fallback, ferr := lookupAccount(fallbackUser); ferr == nil {
			pamSyslog(pamh, C.LOG_INFO, msg(i18n.MsgFallbackUser, fallbackUser))
			account = fallback
			_ = logging.UpdateHome(account.HomeDir)
		} else {
			return C.PAM_SERVICE_ERR
		}
	}
	_ = logging.UpdateHome(account.HomeDir)

	privState, err := dropPrivileges(account)
	if err != nil {
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgDropPrivilegesFailed, account.Username, err))
		return C.PAM_SERVICE_ERR
	}
	defer restorePrivileges(privState)
	_ = logging.UpdateHome(account.HomeDir)

	secretPath, err := pamcfg.ResolveSecretPath(params.SecretSpec, account)
	if err != nil {
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgResolveSecretFailed, err))
		return C.PAM_SERVICE_ERR
	}
	pamDebugf(pamh, params, "using secret file %s", secretPath)

	cfg, state, err := pamcfg.LoadConfig(account, secretPath, params)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && params.NullOK {
			pamSyslog(pamh, C.LOG_INFO, msg(i18n.MsgUserNoSecretNullOK, targetUser))
			return C.PAM_IGNORE
		}
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgReadConfigFailed, secretPath, err))
		pamError(pamh, msg(i18n.MsgReadConfigFailed, secretPath, err))
		return C.PAM_AUTH_ERR
	}

	rhost := getPamRhost(pamh)
	if rhost != "" {
		pamDebugf(pamh, params, "received PAM_RHOST=%s", rhost)
	}
	if params.PromptTemplate != "" {
		rendered, err := preparePromptFromTemplate(pamh, params.PromptTemplate, account, targetUser, rhost)
		if err != nil {
			pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgPromptTemplateFailed, err))
			return C.PAM_SERVICE_ERR
		}
		params.Prompt = rendered
		pamDebugf(pamh, params, "using prompt template %s", params.PromptTemplate)
	}
	if params.GracePeriod > 0 && cfg.WithinGracePeriod(rhost, params.GracePeriod, time.Now()) {
		pamSyslog(pamh, C.LOG_INFO, msg(i18n.MsgGraceSkip, rhost))
		cfg.UpdateLoginRecord(rhost, time.Now())
		pamDebugf(pamh, params, "grace period hit for host %s", rhost)
		if rc := persistConfig(pamh, cfg, secretPath, params, account, state); rc != C.PAM_SUCCESS {
			return rc
		}
		return C.PAM_SUCCESS
	}

	code, remainder, rc := obtainOTP(pamh, params)
	if rc != C.PAM_SUCCESS {
		return rc
	}
	auth := &authenticator.Authenticator{}
	res, err := auth.VerifyCode(cfg, code, authenticator.VerifyOptions{
		DisableSkewAdjustment: params.NoSkewAdjust,
		NoIncrementHOTP:       params.NoIncrementHOTP,
	})
	if err != nil {
		if errors.Is(err, authenticator.ErrInvalidCode) || errors.Is(err, config.ErrRateLimited) {
			pamError(pamh, err.Error())
			pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgUserAuthFailed, targetUser, err))
			return C.PAM_AUTH_ERR
		}
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgAuthFailedGeneric, err))
		pamError(pamh, msg(i18n.MsgInternalError))
		return C.PAM_AUTH_ERR
	}
	if params.ForwardPass && remainder != "" {
		if rc := setPamAuthtok(pamh, remainder); rc != C.PAM_SUCCESS {
			pamSyslog(pamh, C.LOG_WARNING, msg(i18n.MsgUpdateAuthtokFailed))
		}
	}
	if params.GracePeriod > 0 && rhost != "" {
		cfg.UpdateLoginRecord(rhost, time.Now())
	}
	if rc := persistConfig(pamh, cfg, secretPath, params, account, state); rc != C.PAM_SUCCESS {
		return rc
	}
	pamDebugf(pamh, params, "authentication completed for %s", targetUser)
	pamSyslog(pamh, C.LOG_INFO, msg(i18n.MsgUserAuthSuccess, targetUser, res.Type))
	return C.PAM_SUCCESS
}

func parsePamArgs(argc C.int, argv **C.char) []string {
	length := int(argc)
	if length == 0 {
		return nil
	}
	slice := make([]string, length)
	ptrs := unsafe.Slice((**C.char)(unsafe.Pointer(argv)), length)
	for i := 0; i < length; i++ {
		slice[i] = C.GoString(ptrs[i])
	}
	return slice
}

func obtainOTP(pamh *C.pam_handle_t, params pamcfg.Params) (string, string, C.int) {
	switch params.PassMode {
	case pamcfg.ModeUseFirst:
		pw, rc := getPamAuthtok(pamh)
		if rc != C.PAM_SUCCESS {
			return "", "", rc
		}
		logDummyPassword(pamh, params, pw)
		code, rest, ok := extractOTP(pw)
		if !ok {
			return "", "", C.PAM_AUTH_ERR
		}
		return code, rest, C.PAM_SUCCESS
	case pamcfg.ModeTryFirst:
		if pw, rc := getPamAuthtok(pamh); rc == C.PAM_SUCCESS && pw != "" {
			logDummyPassword(pamh, params, pw)
			if code, rest, ok := extractOTP(pw); ok {
				return code, rest, C.PAM_SUCCESS
			}
		}
		return promptCode(pamh, params.Prompt, params.EchoCode)
	default:
		return promptCode(pamh, params.Prompt, params.EchoCode)
	}
}

func promptCode(pamh *C.pam_handle_t, prompt string, echo bool) (string, string, C.int) {
	var resp *C.char
	style := C.PAM_PROMPT_ECHO_OFF
	if echo {
		style = C.PAM_PROMPT_ECHO_ON
	}
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))
	rc := C.prompt_wrapper(pamh, C.int(style), &resp, cPrompt)
	if rc != C.PAM_SUCCESS {
		return "", "", rc
	}
	defer C.free(unsafe.Pointer(resp))
	code := strings.TrimSpace(C.GoString(resp))
	if code == "" {
		return "", "", C.PAM_AUTH_ERR
	}
	return code, "", C.PAM_SUCCESS
}

func extractOTP(raw string) (string, string, bool) {
	if raw == "" {
		return "", "", false
	}
	if code, rest, ok := splitDigits(raw, 6); ok {
		return code, rest, true
	}
	if code, rest, ok := splitDigits(raw, 8); ok {
		return code, rest, true
	}
	return "", "", false
}

func splitDigits(raw string, length int) (string, string, bool) {
	if len(raw) < length {
		return "", "", false
	}
	code := raw[len(raw)-length:]
	if !onlyDigits(code) {
		return "", "", false
	}
	return code, raw[:len(raw)-length], true
}

func onlyDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func logDummyPassword(pamh *C.pam_handle_t, params pamcfg.Params, pw string) {
	if len(pw) > 0 && pw[0] == '\b' {
		pamSyslog(pamh, C.LOG_INFO, msg(i18n.MsgDummyPassword))
	}
}

func getPamUser(pamh *C.pam_handle_t) (string, C.int) {
	var cUser *C.char
	rc := C.pam_get_user(pamh, &cUser, nil)
	if rc != C.PAM_SUCCESS {
		return "", rc
	}
	return C.GoString(cUser), C.PAM_SUCCESS
}

func getPamAuthtok(pamh *C.pam_handle_t) (string, C.int) {
	var item unsafe.Pointer
	rc := C.pam_get_item(pamh, C.PAM_AUTHTOK, &item)
	if rc != C.PAM_SUCCESS {
		return "", rc
	}
	if item == nil {
		return "", C.PAM_AUTHTOK_ERR
	}
	return C.GoString((*C.char)(item)), C.PAM_SUCCESS
}

func setPamAuthtok(pamh *C.pam_handle_t, value string) C.int {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	return C.pam_set_item(pamh, C.PAM_AUTHTOK, unsafe.Pointer(cValue))
}

func getPamRhost(pamh *C.pam_handle_t) string {
	var item unsafe.Pointer
	if C.pam_get_item(pamh, C.PAM_RHOST, &item) != C.PAM_SUCCESS || item == nil {
		return ""
	}
	return C.GoString((*C.char)(item))
}

func pamError(pamh *C.pam_handle_t, text string) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.error_wrapper(pamh, cText)
}

func pamSyslog(pamh *C.pam_handle_t, priority C.int, msg string) {
	logWithPriority(priority, msg)
	cText := C.CString(msg)
	defer C.free(unsafe.Pointer(cText))
	C.syslog_wrapper(pamh, priority, cText)
}

func pamDebugf(pamh *C.pam_handle_t, params pamcfg.Params, format string, args ...interface{}) {
	if !params.Debug {
		return
	}
	message := fmt.Sprintf("debug: "+format, args...)
	pamSyslog(pamh, C.LOG_DEBUG, message)
}

type privilegeState struct {
	origUID int
	origGID int
	dropped bool
}

func dropPrivileges(account *user.User) (*privilegeState, error) {
	state := &privilegeState{
		origUID: unix.Geteuid(),
		origGID: unix.Getegid(),
	}
	uid, err := strconv.Atoi(account.Uid)
	if err != nil {
		return nil, err
	}
	gid, err := strconv.Atoi(account.Gid)
	if err != nil {
		return nil, err
	}
	if state.origUID == uid && state.origGID == gid {
		return state, nil
	}
	if err := unix.Setresgid(-1, gid, -1); err != nil {
		return nil, err
	}
	if err := unix.Setresuid(-1, uid, -1); err != nil {
		unix.Setresgid(-1, state.origGID, -1)
		return nil, err
	}
	state.dropped = true
	return state, nil
}

func restorePrivileges(state *privilegeState) {
	if state == nil || !state.dropped {
		return
	}
	unix.Setresuid(-1, state.origUID, -1)
	unix.Setresgid(-1, state.origGID, -1)
}

func lookupAccount(name string) (*user.User, error) {
	if name == "" {
		return nil, fmt.Errorf("%s", msg(i18n.MsgEmptyUsername))
	}
	u, err := user.Lookup(name)
	if err == nil {
		return u, nil
	}
	if _, convErr := strconv.Atoi(name); convErr == nil {
		if u, err2 := user.LookupId(name); err2 == nil {
			return u, nil
		}
	}
	return nil, err
}

func persistConfig(pamh *C.pam_handle_t, cfg *config.Config, path string, params pamcfg.Params, account *user.User, state pamcfg.FileState) C.int {
	if !cfg.Dirty {
		return C.PAM_SUCCESS
	}
	data, err := cfg.Bytes()
	if err != nil {
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgSerializeConfigFailed, err))
		pamError(pamh, msg(i18n.MsgInternalError))
		return C.PAM_AUTH_ERR
	}
	err = pamcfg.WriteConfig(account, path, data, params.AllowedPerm, state)
	if err != nil {
		if errors.Is(err, pamcfg.ErrSecretModified) {
			pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgSecretChangedDuringProcess))
			pamError(pamh, msg(i18n.MsgSecretChangedRetry))
			return C.PAM_AUTH_ERR
		}
		if params.AllowReadonly && (errors.Is(err, os.ErrPermission) || errors.Is(err, syscall.EROFS) || errors.Is(err, syscall.EPERM)) {
			pamSyslog(pamh, C.LOG_WARNING, msg(i18n.MsgReadonlyWriteIgnored, err))
			return C.PAM_SUCCESS
		}
		pamSyslog(pamh, C.LOG_ERR, msg(i18n.MsgWriteConfigFailed, path, err))
		pamError(pamh, msg(i18n.MsgUpdateConfigFailed))
		return C.PAM_AUTH_ERR
	}
	if err := applySelinuxContext(path); err != nil {
		pamSyslog(pamh, C.LOG_DEBUG, fmt.Sprintf("setting SELinux type \"%s\" on file \"%s\" failed. Okay if SELinux is disabled: %v", selinuxSecretType, path, err))
	}
	return C.PAM_SUCCESS
}

func logWithPriority(priority C.int, msg string) {
	switch priority {
	case C.LOG_DEBUG:
		logging.Debugf("%s", msg)
	case C.LOG_WARNING:
		logging.Warnf("%s", msg)
	case C.LOG_ERR:
		logging.Errorf("%s", msg)
	default:
		logging.Infof("%s", msg)
	}
}

func main() {}
