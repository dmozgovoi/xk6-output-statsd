package statsd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.k6.io/k6/lib/types"
	"go.k6.io/k6/metrics"
	"gopkg.in/guregu/null.v3"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	testCases := map[string]struct {
		jsonRaw json.RawMessage
		env     map[string]string
		arg     string
		config  config
		err     string
	}{
		"default": {
			config: config{
				Addr:         null.NewString("localhost:8125", false),
				BufferSize:   null.NewInt(20, false),
				Namespace:    null.NewString("k6.", false),
				PushInterval: types.NewNullDuration(1*time.Second, false),
				TagBlocklist: metrics.SystemTagSet(metrics.TagVU | metrics.TagIter | metrics.TagURL).Map(),
				EnableTags:   null.NewBool(false, false),
			},
		},
		"overwrite-with-env": {
			env: map[string]string{
				"K6_STATSD_ADDR":          "override:8125",
				"K6_STATSD_BUFFER_SIZE":   "1000",
				"K6_STATSD_NAMESPACE":     "TEST.",
				"K6_STATSD_PUSH_INTERVAL": "100ms",
				"K6_STATSD_TAG_BLOCKLIST": "method,group",
				"K6_STATSD_ENABLE_TAGS":   "true",
			},
			config: config{
				Addr:         null.NewString("override:8125", true),
				BufferSize:   null.NewInt(1000, true),
				Namespace:    null.NewString("TEST.", true),
				PushInterval: types.NewNullDuration(100*time.Millisecond, true),
				TagBlocklist: metrics.SystemTagSet(metrics.TagMethod | metrics.TagGroup).Map(),
				EnableTags:   null.NewBool(true, true),
			},
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			checkConfig, err := getConsolidatedConfig(testCase.jsonRaw, testCase.env, testCase.arg)
			if testCase.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, testCase.config, checkConfig)
		})
	}
}

func TestSanitizeTagName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"tag name/with spaces", "tag_name-with_spaces"},
		{"tag/name/with/slashes", "tag-name-with-slashes"},
		{"tag@name#with!special&chars", "tagnamewithspecialchars"},
		{"tag name with multiple spaces", "tag_name_with_multiple_spaces"},
	}

	for _, test := range testCases {
		require.Equal(t, test.expected, sanitizeTagName(test.input))
	}
}

func TestSanitizeTagValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"value,with:allowed:chars", "value_with_allowed_chars"},
		{"value with spaces", "value_with_spaces"},
		{"value/with/slashes", "value_with_slashes"},
		{"value@name#with!special&chars", "value_name_with_special_chars"},
	}

	for _, test := range testCases {
		require.Equal(t, test.expected, sanitizeTagValue(test.input))
	}
}
