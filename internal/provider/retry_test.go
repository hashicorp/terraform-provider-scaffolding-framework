package provider

import (
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestRetryConfigWith(t *testing.T) {
	base := retryConfig{
		delay:       30 * time.Second,
		interval:    10 * time.Second,
		maxAttempts: 5,
	}

	t.Run("nil override inherits base", func(t *testing.T) {
		assert.Equal(t, base, base.with(nil))
	})

	t.Run("empty block inherits base", func(t *testing.T) {
		override := &RetryModel{
			Delay:       types.NumberNull(),
			Interval:    types.NumberNull(),
			MaxAttempts: types.NumberNull(),
		}
		assert.Equal(t, base, base.with(override))
	})

	t.Run("per-field override keeps unset fields from base", func(t *testing.T) {
		override := &RetryModel{
			Delay:       types.NumberNull(),
			Interval:    types.NumberNull(),
			MaxAttempts: types.NumberValue(big.NewFloat(40)),
		}
		got := base.with(override)
		assert.Equal(t, base.delay, got.delay)
		assert.Equal(t, base.interval, got.interval)
		assert.Equal(t, 40, got.maxAttempts)
	})

	t.Run("full override replaces every field", func(t *testing.T) {
		override := &RetryModel{
			Delay:       types.NumberValue(big.NewFloat(60)),
			Interval:    types.NumberValue(big.NewFloat(15)),
			MaxAttempts: types.NumberValue(big.NewFloat(40)),
		}
		got := base.with(override)
		assert.Equal(t, retryConfig{
			delay:       60 * time.Second,
			interval:    15 * time.Second,
			maxAttempts: 40,
		}, got)
	})

	t.Run("unknown fields are ignored", func(t *testing.T) {
		override := &RetryModel{
			Delay:       types.NumberUnknown(),
			Interval:    types.NumberUnknown(),
			MaxAttempts: types.NumberUnknown(),
		}
		assert.Equal(t, base, base.with(override))
	})
}
