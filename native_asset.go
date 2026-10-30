package coakka_v2_connector

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// A sealed package generates one build-selected init with the matching asset.
// An unstaged development tree requires an explicit Core path. This is artifact
// loading, not Core capability/configuration policy; no request uses this state.
type runtimeAsset struct {
	image  string
	sha256 string
	name   string
}

var bundledRuntimeAsset runtimeAsset
var runtimeAssetMu sync.Mutex

const maxRuntimeAssetBytes = 32 << 20

// The two cold filesystem seams permit write/publication fault tests without
// process-global hooks or a simulated runtime. The files remain store-owned.
type runtimeAssetIO struct {
	createTemp func(string, string) (*os.File, error)
	link       func(string, string) error
}

func resolvePackagedRuntimeNative() (string, error) {
	runtimeAssetMu.Lock()
	defer runtimeAssetMu.Unlock()
	cache := os.Getenv("COAKKA_RUNTIME_CACHE_DIR")
	if cache == "" {
		var err error
		cache, err = os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("runtime artifact cache: %w", err)
		}
	}
	return materializeRuntimeAsset(bundledRuntimeAsset, cache, runtimeAssetIO{os.CreateTemp, os.Link})
}

func verifyRuntimeAsset(path string, size int64, digest string) (bool, error) {
	before, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !before.Mode().IsRegular() || before.Size() != size {
		return true, fmt.Errorf("runtime artifact is not the expected regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return true, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return true, err
	}
	if !os.SameFile(before, opened) {
		return true, fmt.Errorf("runtime artifact changed while opening")
	}
	hash := sha256.New()
	count, err := io.Copy(hash, io.LimitReader(file, size+1))
	if err != nil {
		return true, err
	}
	after, err := os.Lstat(path)
	if err != nil {
		return true, err
	}
	if count != size || !after.Mode().IsRegular() || !os.SameFile(opened, after) ||
		after.Size() != size || !after.ModTime().Equal(before.ModTime()) || hex.EncodeToString(hash.Sum(nil)) != digest {
		return true, fmt.Errorf("runtime artifact identity mismatch")
	}
	return true, nil
}

func materializeRuntimeAsset(asset runtimeAsset, cache string, fs runtimeAssetIO) (result string, resultErr error) {
	if len(asset.image) == 0 {
		return "", fmt.Errorf("this source tree has no packaged runtime asset; supply an explicit Core library")
	}
	if len(asset.image) > maxRuntimeAssetBytes || asset.name == "" || asset.name == "." ||
		filepath.Base(asset.name) != asset.name || strings.ContainsAny(asset.name, "/\\\x00") {
		return "", fmt.Errorf("runtime asset representation refused")
	}
	digest, err := hex.DecodeString(asset.sha256)
	if err != nil || len(digest) != sha256.Size || hex.EncodeToString(digest) != asset.sha256 {
		return "", fmt.Errorf("runtime asset digest representation refused")
	}
	hash := sha256.New()
	// LimitReader hides the string reader's WriteTo, avoiding a whole-image copy.
	if _, err := io.Copy(hash, io.LimitReader(strings.NewReader(asset.image), int64(len(asset.image)))); err != nil {
		return "", err
	}
	if hex.EncodeToString(hash.Sum(nil)) != asset.sha256 {
		return "", fmt.Errorf("bundled runtime asset identity mismatch")
	}
	if cache == "" {
		return "", fmt.Errorf("runtime artifact cache directory is empty")
	}
	directory, err := filepath.Abs(filepath.Join(cache, "coakka", "runtime", asset.sha256))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("runtime artifact directory is not a regular directory")
	}
	destination := filepath.Join(directory, asset.name)
	if exists, err := verifyRuntimeAsset(destination, int64(len(asset.image)), asset.sha256); exists || err != nil {
		if err != nil {
			return "", err
		}
		return destination, nil
	}
	file, err := fs.createTemp(directory, ".runtime-staging-")
	if err != nil {
		return "", err
	}
	staging := file.Name()
	closed := false
	defer func() {
		if !closed {
			resultErr = errors.Join(resultErr, file.Close())
		}
		if err := os.Remove(staging); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("runtime staging cleanup: %w", err))
		}
		if resultErr != nil {
			result = ""
		}
	}()
	written, err := file.WriteString(asset.image)
	if err != nil {
		return "", err
	}
	if written != len(asset.image) {
		return "", io.ErrShortWrite
	}
	if err := file.Sync(); err != nil {
		return "", err
	}
	err = file.Close()
	closed = true
	if err != nil {
		return "", err
	}
	// Same-directory hard-link publication cannot overwrite an existing entry.
	// A concurrent publisher wins only if its complete bytes verify identically.
	if publishErr := fs.link(staging, destination); publishErr != nil {
		exists, verifyErr := verifyRuntimeAsset(destination, int64(len(asset.image)), asset.sha256)
		if verifyErr != nil {
			return "", errors.Join(publishErr, verifyErr)
		}
		if !exists {
			return "", publishErr
		}
	}
	if exists, err := verifyRuntimeAsset(destination, int64(len(asset.image)), asset.sha256); !exists || err != nil {
		if err == nil {
			err = fmt.Errorf("runtime artifact disappeared after publication")
		}
		return "", err
	}
	return destination, nil
}
