package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/ini.v1"
)

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".tt", "config")
	if path != want {
		t.Fatalf("got %q want %q", path, want)
	}
}

func TestResolveFlagsOverrideEnvAndFile(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://env.example.com")
	t.Setenv(EnvToken, "env-token")
	t.Setenv(EnvProfile, "envprofile")

	f := newTestConfig(t, "dev", "https://file.example.com", "file-token")

	got, err := Resolve(f, FlagOverrides{
		BaseURL: "https://flag.example.com",
		Token:   "flag-token",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://flag.example.com" {
		t.Errorf("base URL = %q", got.BaseURL)
	}
	if got.Token != "flag-token" {
		t.Errorf("token = %q", got.Token)
	}
}

func TestResolveFromProfile(t *testing.T) {
	f := newTestConfig(t, "dev", "https://api.example.com", "jwt123")

	got, err := Resolve(f, FlagOverrides{Profile: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://api.example.com" {
		t.Errorf("base URL = %q", got.BaseURL)
	}
	if got.Token != "jwt123" {
		t.Errorf("token = %q", got.Token)
	}
}

func TestResolveMissingProfile(t *testing.T) {
	f := newTestConfig(t, "dev", "https://api.example.com", "jwt123")
	_, err := Resolve(f, FlagOverrides{Profile: "missing"})
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
}

func TestResolveTrimsTrailingSlash(t *testing.T) {
	f := newTestConfig(t, "dev", "https://api.example.com/", "jwt")
	got, err := Resolve(f, FlagOverrides{Profile: "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if got.BaseURL != "https://api.example.com" {
		t.Errorf("base URL = %q", got.BaseURL)
	}
}

func TestDefaultProfileName(t *testing.T) {
	f := newTestConfig(t, "dev", "https://x.com", "t")
	if DefaultProfileName(f) != "dev" {
		t.Errorf("default profile = %q", DefaultProfileName(f))
	}
}

func TestListProfiles(t *testing.T) {
	f := newTestConfig(t, "dev", "https://x.com", "t")
	if err := SetProfile(f, "staging", "https://staging.com", "t2"); err != nil {
		t.Fatal(err)
	}
	names := ListProfiles(f)
	if len(names) != 2 {
		t.Fatalf("profiles = %v", names)
	}
}

func TestMaskToken(t *testing.T) {
	if MaskToken("") != "(not set)" {
		t.Error("empty token mask")
	}
	masked := MaskToken("abcdefghijklmnop")
	if masked != "abcd...mnop" {
		t.Errorf("mask = %q", masked)
	}
}

func newTestConfig(t *testing.T, name, baseURL, token string) *ini.File {
	t.Helper()
	f := ini.Empty()
	if err := SetDefaultProfile(f, name); err != nil {
		t.Fatal(err)
	}
	if err := SetProfile(f, name, baseURL, token); err != nil {
		t.Fatal(err)
	}
	return f
}
