package otp

import (
	"fmt"
	"ggpam/pkg/config"
	"net/url"
	"sort"
)

type OTPAuthBuilder struct {
	label  string
	issuer string
	params map[string]string
	mode   config.Mode
}

func NewOTPAuthBuilder(label, issuer string, params map[string]string, mode config.Mode) *OTPAuthBuilder {
	return &OTPAuthBuilder{
		label:  label,
		issuer: issuer,
		params: params,
		mode:   mode,
	}
}

func (b *OTPAuthBuilder) String() string {
	query := url.Values{}
	for k, v := range b.params {
		if v == "" {
			continue
		}
		query.Set(k, v)
	}
	// stable order
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	values := url.Values{}
	for _, k := range keys {
		values[k] = query[k]
	}
	label := url.PathEscape(b.label)
	if b.issuer != "" {
		label = url.PathEscape(fmt.Sprintf("%s:%s", b.issuer, b.label))
	}
	scheme := "totp"
	if b.mode == config.ModeHOTP {
		scheme = "hotp"
	}
	return fmt.Sprintf("otpauth://%s/%s?%s", scheme, label, values.Encode())
}
