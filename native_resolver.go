package coakka_v2_connector

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func RuntimePlatformID(goos string, goarch string) string {
	return normalizeRuntimeOS(goos) + "-" + normalizeRuntimeArch(goarch)
}

func normalizeRuntimeOS(goos string) string {
	switch goos {
	case "darwin", "macos":
		return "macos"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		panic(fmt.Sprintf("unsupported os=%s; supported platforms are macOS, Linux, and Windows", goos))
	}
}

func normalizeRuntimeArch(goarch string) string {
	switch goarch {
	case "arm64", "aarch64":
		return "aarch64"
	case "amd64", "x86_64":
		return "x86_64"
	default:
		panic(fmt.Sprintf("unsupported arch=%s; supported architectures are aarch64 and x86_64", goarch))
	}
}

func runtimeResourceFileNames(goos string) []string {
	versionedBase := "libcoakka_runtime_v2-" + CoakkaV2NativePackageVersion
	switch normalizeRuntimeOS(goos) {
	case "macos":
		return []string{versionedBase + ".dylib"}
	case "linux":
		return []string{versionedBase + ".so"}
	case "windows":
		return []string{versionedBase + ".dll"}
	default:
		panic("unsupported platform")
	}
}

func resolvePackagedRuntimeNative() (string, bool) {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}
	packageDir := filepath.Dir(sourceFile)
	platform := RuntimePlatformID(runtime.GOOS, runtime.GOARCH)
	for _, fileName := range runtimeResourceFileNames(runtime.GOOS) {
		candidate := filepath.Join(packageDir, "native", platform, fileName)
		if _, err := os.Stat(candidate); err == nil {
			abs, absErr := filepath.Abs(candidate)
			if absErr == nil {
				return abs, true
			}
			return candidate, true
		}
	}
	return "", false
}
