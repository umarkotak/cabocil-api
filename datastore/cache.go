package datastore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// WrapCache is a generic cache wrapper that fetches data from cache or executes
// the provided function if cache miss, then stores the result in cache.
// T is the type of data being cached.
func WrapCache[T any](ctx context.Context, key string, expiration time.Duration, fetchFunc func() (T, error)) (T, error) {
	var result T

	// Try to get from cache
	cached, err := Get().Redis.Get(ctx, key).Result()
	if err == nil {
		// Cache hit - unmarshal and return
		if err := json.Unmarshal([]byte(cached), &result); err == nil {
			logrus.Infof("[cache] hit for key: %s", key)
			return result, nil
		}
		logrus.Warnf("[cache] failed to unmarshal cached value for key: %s, fetching fresh data", key)
	}

	// Cache miss - execute the fetch function
	logrus.Infof("[cache] miss for key: %s, fetching fresh data", key)
	result, err = fetchFunc()
	if err != nil {
		return result, err
	}

	// Store in cache
	data, err := json.Marshal(result)
	if err != nil {
		logrus.Warnf("[cache] failed to marshal data for key: %s, error: %v", key, err)
		return result, nil // Return result anyway, just don't cache
	}

	if err := Get().Redis.Set(ctx, key, data, expiration).Err(); err != nil {
		logrus.Warnf("[cache] failed to set cache for key: %s, error: %v", key, err)
	} else {
		logrus.Infof("[cache] stored key: %s with expiration: %v", key, expiration)
	}

	return result, nil
}
