package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/ini.v1"
)

const (
	EnvBaseURL = "TT_API_BASE_URL"
	EnvToken   = "TT_API_TOKEN"
	EnvProfile = "TT_PROFILE"
)

// Resolved holds the effective API connection settings after resolution.
type Resolved struct {
	Profile string
	BaseURL string
	Token   string
}

// FlagOverrides from CLI global flags (empty means not set).
type FlagOverrides struct {
	Profile string
	BaseURL string
	Token   string
}

// ConfigPath returns the path to ~/.tt/config (or %USERPROFILE%\.tt\config on Windows).
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home directory: %w", err)
	}
	return filepath.Join(home, ".tt", "config"), nil
}

// ConfigDir returns the ~/.tt directory path.
func ConfigDir() (string, error) {
	path, err := ConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}

// Load reads the INI config file. Missing file returns an empty ini.File.
func Load() (*ini.File, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ini.Empty(), nil
	}
	return ini.Load(path)
}

// Save writes the config file with restrictive permissions.
func Save(f *ini.File) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := f.SaveTo(path); err != nil {
			return err
		}
		return os.Chmod(path, 0o600)
	}
	return f.SaveTo(path)
}

// DefaultProfileName reads [default] profile = ... or returns "default".
func DefaultProfileName(f *ini.File) string {
	if f == nil {
		return "default"
	}
	sec, err := f.GetSection("default")
	if err != nil {
		return "default"
	}
	name := strings.TrimSpace(sec.Key("profile").String())
	if name == "" {
		return "default"
	}
	return name
}

// ProfileSectionName returns the INI section name for a profile.
func ProfileSectionName(name string) string {
	return "profile " + name
}

// GetProfile reads base_url and token from [profile NAME].
func GetProfile(f *ini.File, name string) (baseURL, token string, err error) {
	if f == nil {
		return "", "", fmt.Errorf("profile %q not found", name)
	}
	sec, err := f.GetSection(ProfileSectionName(name))
	if err != nil {
		return "", "", fmt.Errorf("profile %q not found", name)
	}
	return strings.TrimSpace(sec.Key("base_url").String()),
		strings.TrimSpace(sec.Key("token").String()), nil
}

// SetProfile creates or updates a profile section.
func SetProfile(f *ini.File, name, baseURL, token string) error {
	if f == nil {
		return fmt.Errorf("nil config")
	}
	secName := ProfileSectionName(name)
	sec, err := f.GetSection(secName)
	if err != nil {
		sec, err = f.NewSection(secName)
		if err != nil {
			return err
		}
	}
	sec.Key("base_url").SetValue(strings.TrimSpace(baseURL))
	sec.Key("token").SetValue(strings.TrimSpace(token))
	return nil
}

// SetDefaultProfile sets [default] profile = name.
func SetDefaultProfile(f *ini.File, name string) error {
	sec, err := f.GetSection("default")
	if err != nil {
		sec, err = f.NewSection("default")
		if err != nil {
			return err
		}
	}
	sec.Key("profile").SetValue(name)
	return nil
}

// ListProfiles returns all profile names defined in the config.
func ListProfiles(f *ini.File) []string {
	if f == nil {
		return nil
	}
	var names []string
	for _, sec := range f.Sections() {
		if strings.HasPrefix(sec.Name(), "profile ") {
			names = append(names, strings.TrimPrefix(sec.Name(), "profile "))
		}
	}
	return names
}

// Resolve determines effective settings using flags > env > config file.
func Resolve(f *ini.File, flags FlagOverrides) (Resolved, error) {
	profile := strings.TrimSpace(flags.Profile)
	if profile == "" {
		profile = strings.TrimSpace(os.Getenv(EnvProfile))
	}
	if profile == "" {
		profile = DefaultProfileName(f)
	}

	baseURL := strings.TrimSpace(flags.BaseURL)
	token := strings.TrimSpace(flags.Token)

	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv(EnvBaseURL))
	}
	if token == "" {
		token = strings.TrimSpace(os.Getenv(EnvToken))
	}

	if baseURL == "" || token == "" {
		pBase, pToken, err := GetProfile(f, profile)
		if err != nil && baseURL == "" && token == "" {
			return Resolved{}, fmt.Errorf("profile %q not configured; run `tt configure`", profile)
		}
		if baseURL == "" {
			baseURL = pBase
		}
		if token == "" {
			token = pToken
		}
	}

	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return Resolved{}, fmt.Errorf("base URL is required (flag, %s, or profile %q)", EnvBaseURL, profile)
	}

	return Resolved{
		Profile: profile,
		BaseURL: baseURL,
		Token:   token,
	}, nil
}

// ResolveOptional is like Resolve but does not require base URL or token (for health).
func ResolveOptional(f *ini.File, flags FlagOverrides) Resolved {
	profile := strings.TrimSpace(flags.Profile)
	if profile == "" {
		profile = strings.TrimSpace(os.Getenv(EnvProfile))
	}
	if profile == "" {
		profile = DefaultProfileName(f)
	}

	baseURL := strings.TrimSpace(flags.BaseURL)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv(EnvBaseURL))
	}
	if baseURL == "" {
		if pBase, _, err := GetProfile(f, profile); err == nil {
			baseURL = pBase
		}
	}
	baseURL = strings.TrimRight(baseURL, "/")

	token := strings.TrimSpace(flags.Token)
	if token == "" {
		token = strings.TrimSpace(os.Getenv(EnvToken))
	}
	if token == "" {
		if _, pToken, err := GetProfile(f, profile); err == nil {
			token = pToken
		}
	}

	return Resolved{
		Profile: profile,
		BaseURL: baseURL,
		Token:   token,
	}
}

// MaskToken returns a masked representation for display.
func MaskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
