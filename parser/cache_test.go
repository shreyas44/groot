package parser

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetCache() {
	cache = map[reflect.Type]Type{}
}

func TestCache(t *testing.T) {
	resetCache()
	stringType := reflect.TypeOf("")
	intType := reflect.TypeOf(0)
	stringScalar := &Scalar{stringType}

	t.Run("CacheExists", func(t *testing.T) {
		cache.set(stringType, stringScalar)

		cacheVal, exists := cache.get(stringType)
		assert.Equal(t, stringScalar, cacheVal)
		assert.True(t, exists)
	})

	t.Run("CacheNotExists", func(t *testing.T) {
		cacheVal, exists := cache.get(intType)
		assert.Nil(t, cacheVal)
		assert.False(t, exists)
	})
}
