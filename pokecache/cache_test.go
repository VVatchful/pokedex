package pokecache

import (
    "testing"
    "time"
    "strconv"
)

func TestCacheAddGet(t *testing.T) {
    cache := NewCache(10 * time.Second)

    cache.Add("test_key", []byte("test_value"))
    cache.Add("another_key", []byte("another_value"))

    val, found := cache.Get("test_key")
    if !found {
        t.Errorf("expected to find 'test_key' in cache, but it wasn't found")
    }

    if string(val) != "test_value" {
        t.Errorf("expected value 'test_value', got '%s'", string(val))
    }

    val, found = cache.Get("another_key")
    if !found {
        t.Errorf("expected to find 'another_key' in cache, but it wasn't found")
    }

    if string(val) != "another_value" {
        t.Errorf("expected value 'another_value', got '%s'", string(val))
    }

    _, found = cache.Get("missing_key")
    if found {
        t.Errorf("expected 'missing_key' to not be found in cache, but it was")
    }
}

func TestCacheExpiration(t *testing.T) {
    cache := NewCache(2 * time.Second)

    cache.Add("temp_key", []byte("temp_value"))

    if _, found := cache.Get("temp_key"); !found {
        t.Errorf("expected to find 'temp_key' in cache, but it wasn't found")
    }

    time.Sleep(3 * time.Second)

    if _, found := cache.Get("temp_key"); found {
        t.Errorf("expected 'temp_key' to be expired, but it was still found in cache")
    }
}

func TestCacheConcurrency(t *testing.T) {
    cache := NewCache(5 * time.Second)

    for i := 0; i < 100; i++ {
        go func(i int) {
            key := strconv.Itoa(i)
            cache.Add(key, []byte("value"+key))
        }(i)
    }

    time.Sleep(1 * time.Second)

    for i := 0; i < 100; i++ {
        key := strconv.Itoa(i)
        val, found := cache.Get(key)
        if !found {
            t.Errorf("expected to find '%s' in cache, but it wasn't found", key)
        }

        expectedValue := "value" + key
        if string(val) != expectedValue {
            t.Errorf("expected value '%s', got '%s'", expectedValue, string(val))
        }
    }
}

func TestCacheReapLoop(t *testing.T) {
    cache := NewCache(1 * time.Second)

    for i := 0; i < 10; i++ {
        key := strconv.Itoa(i)
        cache.Add(key, []byte("data"))
    }

    for i := 0; i < 10; i++ {
        key := strconv.Itoa(i)
        if _, found := cache.Get(key); !found {
            t.Errorf("expected to find '%s' in cache, but it wasn't found", key)
        }
    }

    time.Sleep(2 * time.Second)

    for i := 0; i < 10; i++ {
        key := strconv.Itoa(i)
        if _, found := cache.Get(key); found {
            t.Errorf("expected '%s' to be reaped, but it was still found in cache", key)
        }
    }
}


