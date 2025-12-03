package config

import (
	"bufio"
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	maxFileSize     = 64 * 1024
	DefaultStepSize = 30
	DefaultWindow   = 3
)

var (
	errInvalidScratch  = errors.New("invalid scratch code line")
	errInvalidOption   = errors.New("unrecognized config option")
	errMissingSecret   = errors.New("missing shared secret")
	errFileTooLarge    = errors.New("config file exceeds 64KB limit")
	errRateLimitFormat = errors.New("RATE_LIMIT option is malformed")
)

type Mode int

const (
	ModeUnknown Mode = iota
	ModeTOTP
	ModeHOTP
)

type RateLimit struct {
	Attempts   int
	Interval   time.Duration
	Timestamps []int64
}

type SkewSample struct {
	Timestamp int64
	Skew      int
}

type LoginRecord struct {
	Host string
	When int64
}

type Options struct {
	TOTPAuth             bool
	HOTPConfigured       bool
	HOTPCounter          int64
	StepSize             int
	WindowSize           int
	DisallowReuse        bool
	DisallowedTimestamps []int64
	RateLimit            *RateLimit
	TimeSkew             int
	ResettingTimeSkew    []SkewSample
	LastLogins           map[int]LoginRecord
	Additional           map[string]string
}

type Config struct {
	Secret       string
	ScratchCodes []int
	Options      Options
	Dirty        bool
}

func Load(path string) (*Config, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat config %s: %w", path, err)
	}
	if fi.Size() > maxFileSize {
		return nil, errFileTooLarge
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config %s: %w", path, err)
	}
	defer f.Close()
	cfg, err := Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func Parse(r io.Reader) (*Config, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if len(data) > maxFileSize {
		return nil, errFileTooLarge
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 2048), maxFileSize)

	var lines []string
	for scanner.Scan() {
		raw := scanner.Text()
		lines = append(lines, strings.TrimRight(raw, "\r"))
		if len(lines) == 1 && strings.TrimSpace(raw) == "" {
			return nil, errMissingSecret
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config: %w", err)
	}
	if len(lines) == 0 {
		return nil, errMissingSecret
	}
	secret := strings.TrimSpace(lines[0])
	if secret == "" {
		return nil, errMissingSecret
	}
	cfg := &Config{
		Secret: secret,
		Options: Options{
			StepSize:   DefaultStepSize,
			WindowSize: DefaultWindow,
			Additional: map[string]string{},
			LastLogins: map[int]LoginRecord{},
		},
	}
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "\" ") {
			if err := cfg.parseOption(line[2:]); err != nil {
				return nil, err
			}
			continue
		}
		if err := cfg.parseScratch(line); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (c *Config) parseScratch(line string) error {
	if len(line) != 8 {
		return errInvalidScratch
	}
	code, err := strconv.Atoi(line)
	if err != nil {
		return errInvalidScratch
	}
	c.ScratchCodes = append(c.ScratchCodes, code)
	return nil
}

func (c *Config) parseOption(payload string) error {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return errInvalidOption
	}
	fields := strings.Fields(payload)
	key := fields[0]
	value := strings.TrimSpace(strings.TrimPrefix(payload, key))
	switch {
	case key == "TOTP_AUTH":
		c.Options.TOTPAuth = true
	case key == "HOTP_COUNTER":
		n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		if err != nil {
			return fmt.Errorf("parse HOTP_COUNTER: %w", err)
		}
		c.Options.HOTPConfigured = true
		c.Options.HOTPCounter = n
	case key == "STEP_SIZE":
		step, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || step < 1 || step > 60 {
			return fmt.Errorf("invalid STEP_SIZE %q (expected 1..60)", value)
		}
		c.Options.StepSize = step
	case key == "WINDOW_SIZE":
		win, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || win < 1 || win > 100 {
			return fmt.Errorf("invalid WINDOW_SIZE %q (expected 1..100)", value)
		}
		c.Options.WindowSize = win
	case key == "RATE_LIMIT":
		rl, err := parseRateLimit(value)
		if err != nil {
			return err
		}
		c.Options.RateLimit = rl
	case key == "DISALLOW_REUSE":
		c.Options.DisallowReuse = true
		if value != "" {
			values := strings.Fields(value)
			for _, v := range values {
				ts, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return fmt.Errorf("invalid DISALLOW_REUSE timestamp %q", v)
				}
				c.Options.DisallowedTimestamps = append(c.Options.DisallowedTimestamps, ts)
			}
		}
	case key == "TIME_SKEW":
		if value == "" {
			break
		}
		skew, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid TIME_SKEW %q", value)
		}
		c.Options.TimeSkew = skew
	case key == "RESETTING_TIME_SKEW":
		samples, err := parseSkewSamples(value)
		if err != nil {
			return err
		}
		c.Options.ResettingTimeSkew = samples
	default:
		if strings.HasPrefix(key, "LAST") && len(key) == 5 {
			rec, idx, err := parseLastLogin(strings.TrimSpace(value), key)
			if err != nil {
				return err
			}
			if c.Options.LastLogins == nil {
				c.Options.LastLogins = map[int]LoginRecord{}
			}
			c.Options.LastLogins[idx] = rec
			return nil
		}
		c.Options.Additional[key] = value
	}
	return nil
}

func parseLastLogin(value, key string) (LoginRecord, int, error) {
	if len(key) != 5 || key[:4] != "LAST" {
		return LoginRecord{}, 0, fmt.Errorf("unknown field %s", key)
	}
	idx, err := strconv.Atoi(key[4:])
	if err != nil || idx < 0 || idx > 9 {
		return LoginRecord{}, 0, fmt.Errorf("invalid LAST index %s", key)
	}
	chunks := strings.Fields(value)
	if len(chunks) < 2 {
		return LoginRecord{}, 0, fmt.Errorf("invalid LAST line %s", value)
	}
	when, err := strconv.ParseInt(chunks[len(chunks)-1], 10, 64)
	if err != nil {
		return LoginRecord{}, 0, fmt.Errorf("invalid LAST timestamp: %w", err)
	}
	host := strings.Join(chunks[:len(chunks)-1], " ")
	return LoginRecord{Host: host, When: when}, idx, nil
}

func parseSkewSamples(value string) ([]SkewSample, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	var samples []SkewSample
	for _, token := range strings.Fields(value) {
		split := 0
		for split < len(token) && token[split] >= '0' && token[split] <= '9' {
			split++
		}
		if split == 0 || split >= len(token) {
			return nil, fmt.Errorf("parse RESETTING_TIME_SKEW entry %q failed", token)
		}
		ts, err := strconv.ParseInt(token[:split], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid RESETTING_TIME_SKEW timestamp %q", token)
		}
		skewVal, err := strconv.Atoi(token[split:])
		if err != nil {
			return nil, fmt.Errorf("invalid RESETTING_TIME_SKEW skew %q", token)
		}
		samples = append(samples, SkewSample{Timestamp: ts, Skew: skewVal})
	}
	return samples, nil
}

func parseRateLimit(value string) (*RateLimit, error) {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return nil, errRateLimitFormat
	}
	attempts, err := strconv.Atoi(fields[0])
	if err != nil || attempts < 1 || attempts > 100 {
		return nil, errRateLimitFormat
	}
	interval, err := strconv.Atoi(fields[1])
	if err != nil || interval < 1 || interval > 3600 {
		return nil, errRateLimitFormat
	}
	timestamps := make([]int64, 0, len(fields)-2)
	for _, item := range fields[2:] {
		ts, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			return nil, errRateLimitFormat
		}
		timestamps = append(timestamps, ts)
	}
	return &RateLimit{
		Attempts:   attempts,
		Interval:   time.Duration(interval) * time.Second,
		Timestamps: timestamps,
	}, nil
}

func (c *Config) Mode() Mode {
	switch {
	case c.Options.HOTPConfigured:
		return ModeHOTP
	case c.Options.TOTPAuth:
		return ModeTOTP
	default:
		return ModeUnknown
	}
}

func (c *Config) SecretBytes() ([]byte, error) {
	normalized := strings.ToUpper(strings.TrimSpace(c.Secret))
	normalized = strings.ReplaceAll(normalized, " ", "")
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	data, err := enc.DecodeString(normalized)
	if err != nil {
		return nil, fmt.Errorf("base32 decode failed: %w", err)
	}
	return data, nil
}

func (c *Config) Bytes() ([]byte, error) {
	var b strings.Builder
	b.Grow(512)
	b.WriteString(strings.TrimSpace(c.Secret))
	b.WriteByte('\n')

	writeOpt := func(key, value string) {
		if value == "" {
			fmt.Fprintf(&b, "\" %s\n", key)
			return
		}
		fmt.Fprintf(&b, "\" %s %s\n", key, strings.TrimSpace(value))
	}

	if c.Options.TOTPAuth {
		writeOpt("TOTP_AUTH", "")
	}
	if c.Options.HOTPConfigured {
		writeOpt("HOTP_COUNTER", fmt.Sprintf("%d", c.Options.HOTPCounter))
	}
	if c.Options.StepSize != DefaultStepSize {
		writeOpt("STEP_SIZE", fmt.Sprintf("%d", c.Options.StepSize))
	}
	if c.Options.WindowSize != DefaultWindow {
		writeOpt("WINDOW_SIZE", fmt.Sprintf("%d", c.Options.WindowSize))
	}
	if c.Options.RateLimit != nil {
		var parts []string
		parts = append(parts, strconv.Itoa(c.Options.RateLimit.Attempts))
		interval := int(c.Options.RateLimit.Interval / time.Second)
		parts = append(parts, strconv.Itoa(interval))
		for _, ts := range c.Options.RateLimit.Timestamps {
			parts = append(parts, strconv.FormatInt(ts, 10))
		}
		writeOpt("RATE_LIMIT", strings.Join(parts, " "))
	}
	if c.Options.DisallowReuse {
		var parts []string
		for _, ts := range c.Options.DisallowedTimestamps {
			parts = append(parts, strconv.FormatInt(ts, 10))
		}
		writeOpt("DISALLOW_REUSE", strings.Join(parts, " "))
	}
	if c.Options.TimeSkew != 0 {
		writeOpt("TIME_SKEW", strconv.Itoa(c.Options.TimeSkew))
	}
	if len(c.Options.ResettingTimeSkew) > 0 {
		var parts []string
		for _, s := range c.Options.ResettingTimeSkew {
			parts = append(parts, fmt.Sprintf("%d%+d", s.Timestamp, s.Skew))
		}
		writeOpt("RESETTING_TIME_SKEW", strings.Join(parts, " "))
	}
	for i := 0; i < 10; i++ {
		if c.Options.LastLogins == nil {
			break
		}
		rec, ok := c.Options.LastLogins[i]
		if !ok || rec.Host == "" || rec.When == 0 {
			continue
		}
		writeOpt(fmt.Sprintf("LAST%d", i), fmt.Sprintf("%s %d", rec.Host, rec.When))
	}
	if len(c.Options.Additional) > 0 {
		keys := make([]string, 0, len(c.Options.Additional))
		for k := range c.Options.Additional {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			writeOpt(key, c.Options.Additional[key])
		}
	}
	for _, sc := range c.ScratchCodes {
		fmt.Fprintf(&b, "%08d\n", sc)
	}
	if b.Len() > maxFileSize {
		return nil, errFileTooLarge
	}
	return []byte(b.String()), nil
}

func (c *Config) Save(path string, perm os.FileMode) error {
	data, err := c.Bytes()
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return fmt.Errorf("write temp config %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("replace config %s: %w", path, err)
	}
	c.Dirty = false
	return nil
}

func (c *Config) UseScratchCode(code int) bool {
	for idx, value := range c.ScratchCodes {
		if value == code {
			c.ScratchCodes = append(c.ScratchCodes[:idx], c.ScratchCodes[idx+1:]...)
			c.Dirty = true
			return true
		}
	}
	return false
}

func (c *Config) Window() int {
	if c.Options.WindowSize > 0 {
		return c.Options.WindowSize
	}
	return DefaultWindow
}

func (c *Config) Step() int {
	if c.Options.StepSize > 0 {
		return c.Options.StepSize
	}
	return DefaultStepSize
}

func (c *Config) MarkDirty() {
	c.Dirty = true
}

func (c *Config) ResetDisallowList(tm int64, window int) error {
	if !c.Options.DisallowReuse {
		return nil
	}
	filtered := c.Options.DisallowedTimestamps[:0]
	for _, ts := range c.Options.DisallowedTimestamps {
		if ts-tm >= int64(window) || tm-ts >= int64(window) {
			continue
		}
		filtered = append(filtered, ts)
	}
	c.Options.DisallowedTimestamps = filtered
	c.Dirty = true
	return nil
}

func (c *Config) RecordUsedTimestamp(ts int64) {
	if !c.Options.DisallowReuse {
		return
	}
	c.Options.DisallowedTimestamps = append(c.Options.DisallowedTimestamps, ts)
	c.Dirty = true
}

var ErrRateLimited = errors.New("too many login attempts")

func (c *Config) EnforceRateLimit(now time.Time) error {
	if c.Options.RateLimit == nil {
		return nil
	}
	rl := c.Options.RateLimit
	windowStart := now.Add(-rl.Interval).Unix()
	rl.Timestamps = append(rl.Timestamps, now.Unix())
	sort.Slice(rl.Timestamps, func(i, j int) bool { return rl.Timestamps[i] < rl.Timestamps[j] })
	var kept []int64
	for _, ts := range rl.Timestamps {
		if ts < windowStart {
			continue
		}
		if ts > now.Unix() {
			continue
		}
		kept = append(kept, ts)
	}
	exceeded := len(kept) > rl.Attempts
	if exceeded {
		kept = kept[len(kept)-rl.Attempts:]
	}
	rl.Timestamps = kept
	c.Dirty = true
	if exceeded {
		return ErrRateLimited
	}
	return nil
}

func (c *Config) CheckReuse(ts int64) error {
	if !c.Options.DisallowReuse {
		return nil
	}
	for _, blocked := range c.Options.DisallowedTimestamps {
		if blocked == ts {
			return fmt.Errorf("reusing TOTP window %d", ts)
		}
	}
	return nil
}

func (c *Config) RecordSkewObservation(ts int64, skew int) bool {
	if skew == 0 {
		return false
	}
	samples := c.Options.ResettingTimeSkew
	if len(samples) > 0 {
		last := samples[len(samples)-1]
		if last.Timestamp+int64(last.Skew) == ts+int64(skew) {
			// Same sliding window, skip duplicate record.
			return false
		}
	}
	if len(samples) == 3 {
		copy(samples, samples[1:])
		samples = samples[:2]
	}
	samples = append(samples, SkewSample{Timestamp: ts, Skew: skew})
	c.Options.ResettingTimeSkew = samples
	c.Dirty = true
	if len(samples) < 3 {
		return false
	}
	lastTs := samples[0].Timestamp
	lastSkew := samples[0].Skew
	total := lastSkew
	for i := 1; i < len(samples); i++ {
		if samples[i].Timestamp <= lastTs || samples[i].Timestamp > lastTs+2 {
			return false
		}
		diff := lastSkew - samples[i].Skew
		if diff < -1 || diff > 1 {
			return false
		}
		lastTs = samples[i].Timestamp
		lastSkew = samples[i].Skew
		total += samples[i].Skew
	}
	c.Options.TimeSkew = total / len(samples)
	c.Options.ResettingTimeSkew = nil
	c.Dirty = true
	return true
}

func (c *Config) WithinGracePeriod(host string, grace time.Duration, now time.Time) bool {
	if grace <= 0 || host == "" || c.Options.LastLogins == nil {
		return false
	}
	expire := now.Unix()
	window := int64(grace / time.Second)
	for _, rec := range c.Options.LastLogins {
		if rec.Host == host && rec.When+window > expire {
			return true
		}
	}
	return false
}

func (c *Config) UpdateLoginRecord(host string, now time.Time) {
	if host == "" {
		return
	}
	if c.Options.LastLogins == nil {
		c.Options.LastLogins = map[int]LoginRecord{}
	}
	for idx, rec := range c.Options.LastLogins {
		if rec.Host == host {
			c.Options.LastLogins[idx] = LoginRecord{Host: host, When: now.Unix()}
			c.Dirty = true
			return
		}
	}
	for i := 0; i < 10; i++ {
		if _, ok := c.Options.LastLogins[i]; !ok {
			c.Options.LastLogins[i] = LoginRecord{Host: host, When: now.Unix()}
			c.Dirty = true
			return
		}
	}
	oldestIdx := 0
	oldestTime := int64(math.MaxInt64)
	for idx, rec := range c.Options.LastLogins {
		if rec.When < oldestTime {
			oldestTime = rec.When
			oldestIdx = idx
		}
	}
	c.Options.LastLogins[oldestIdx] = LoginRecord{Host: host, When: now.Unix()}
	c.Dirty = true
}
