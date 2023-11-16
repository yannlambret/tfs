package tfs

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func TestCacheLoad(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	v := "1.9.0"
	writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	if _, ok := cache.releases[v]; !ok {
		t.Errorf("Expected release %s to be in cache", v)
	}
}

func TestCacheIsEmpty(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	if !cache.IsEmpty() {
		t.Errorf("Expected cache to be empty")
	}
}

func TestCacheSize(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	v := "1.9.0"

	content := []byte("dummy content")
	writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), content)

	cache := NewLocalCache(cacheDir)
	size, err := cache.Size()

	if err != nil {
		t.Fatalf("Cache.Size() failed: %v", err)
	}

	if size != uint64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size)
	}
}

func TestCachePrune(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	versions := []string{"1.9.0", "1.10.0"}

	for _, v := range versions {
		writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))
	}

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	if err := cache.Prune(); err != nil {
		t.Fatalf("Cache.Prune() failed: %v", err)
	}

	for _, v := range versions {
		filePath := filepath.Join(cacheDir, getFilePrefix()+v)
		if exists, _ := afero.Exists(AppFs, filePath); exists {
			t.Errorf("Expected file %s to be removed", filePath)
		}
	}
}

func TestCachePruneUntil(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	versions := []string{"1.9.0", "1.10.0", "1.11.0"}

	for _, v := range versions {
		writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))
	}

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	// Prune until version 1.10.0
	v110, _ := version.NewVersion("1.10.0")

	if err := cache.PruneUntil(v110); err != nil {
		t.Fatalf("Cache.PruneUntil() failed: %v", err)
	}

	if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+"1.9.0")); exists {
		t.Errorf("Expected 1.9.0 to be pruned")
	}
	if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+"1.10.0")); !exists {
		t.Errorf("Expected 1.10.0 to remain")
	}
	if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+"1.11.0")); !exists {
		t.Errorf("Expected 1.11.0 to remain")
	}
}

func TestCacheAutoClean_DefaultConfig(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	viper.Set("cache_auto_clean", true)
	viper.Set("cache_history", 2)

	versions := []string{"1.9.0", "1.10.0", "1.11.0"}

	for _, v := range versions {
		writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))
	}

	cache := NewLocalCache(cacheDir)
	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	cache.AutoClean()

	if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+"1.9.0")); exists {
		t.Errorf("Expected 1.9.0 to be removed")
	}
}

func TestCacheAutoClean_MinorVersionLimit(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	viper.Set("cache_auto_clean", true)
	viper.Set("cache_minor_version_nb", 2)
	viper.Set("cache_patch_version_nb", 99) // disable patch pruning for this test

	releases := []string{"1.6.5", "1.6.6", "1.8.1", "1.9.2", "1.9.8", "1.10.1", "1.10.2"}

	for _, v := range releases {
		writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))
	}

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	cache.AutoClean()

	// Should keep only two most recent minor versions (1.9 and 1.10)
	shouldRemain := []string{"1.9.2", "1.9.8", "1.10.1", "1.10.2"}
	shouldBeRemoved := []string{"1.6.5", "1.6.6", "1.8.1"}

	t.Run("should remain", func(t *testing.T) {
		for _, v := range shouldRemain {
			if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+v)); !exists {
				t.Errorf("Expected version %s to remain", v)
			}
		}
	})

	t.Run("should be removed", func(t *testing.T) {
		for _, v := range shouldBeRemoved {
			if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+v)); exists {
				t.Errorf("Expected version %s to be removed", v)
			}
		}
	})
}

func TestCacheAutoClean_PatchVersionLimit(t *testing.T) {
	cacheDir, cleanup := initTestFS(t)
	defer cleanup()

	viper.Set("cache_auto_clean", true)
	viper.Set("cache_minor_version_nb", 99) // disable minor pruning
	viper.Set("cache_patch_version_nb", 2)

	releases := []string{"1.10.0", "1.10.1", "1.10.2", "1.10.3", "1.10.4"}
	for _, v := range releases {
		writeTestFile(t, filepath.Join(cacheDir, getFilePrefix()+v), []byte("dummy content"))
	}

	cache := NewLocalCache(cacheDir)

	if err := cache.Load(); err != nil {
		t.Fatalf("Cache.Load() failed: %v", err)
	}

	cache.AutoClean()

	// Should keep only the two latest 1.10.x releases
	shouldRemain := []string{"1.10.3", "1.10.4"}
	shouldBeRemoved := []string{"1.10.0", "1.10.1", "1.10.2"}

	t.Run("should remain", func(t *testing.T) {
		for _, v := range shouldRemain {
			if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+v)); !exists {
				t.Errorf("Expected version %s to remain", v)
			}
		}
	})

	t.Run("should be removed", func(t *testing.T) {
		for _, v := range shouldBeRemoved {
			if exists, _ := afero.Exists(AppFs, filepath.Join(cacheDir, getFilePrefix()+v)); exists {
				t.Errorf("Expected version %s to be removed", v)
			}
		}
	})
}
