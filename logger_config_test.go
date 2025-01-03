package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLoggerConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		containerDetails ContainerDetails
		want             loggerConfig
		wantErr          string
	}{
		{
			name: "common",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:  "token",
					cfgChatIDKey: "chat_id",
				},
			},
			want: loggerConfig{
				ClientConfig: ClientConfig{
					APIURL:  defaultClientConfig.APIURL,
					Token:   "token",
					ChatID:  "chat_id",
					Retries: defaultClientConfig.Retries,
					Timeout: defaultClientConfig.Timeout,
				},
				Attrs:              make(map[string]string),
				Template:           defaultLoggerConfig.Template,
				MaxBufferSize:      defaultLoggerConfig.MaxBufferSize,
				BatchEnabled:       defaultLoggerConfig.BatchEnabled,
				BatchFlushInterval: defaultLoggerConfig.BatchFlushInterval,
			},
		},
		{
			name: "custom template",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:              "token",
					cfgChatIDKey:             "chat_id",
					cfgTemplateKey:           "{log}",
					cfgBatchFlushIntervalKey: "30s",
				},
			},
			want: loggerConfig{
				ClientConfig: ClientConfig{
					APIURL:  defaultClientConfig.APIURL,
					Token:   "token",
					ChatID:  "chat_id",
					Retries: defaultClientConfig.Retries,
					Timeout: defaultClientConfig.Timeout,
				},
				Attrs:              make(map[string]string),
				Template:           "{log}",
				MaxBufferSize:      defaultLoggerConfig.MaxBufferSize,
				BatchEnabled:       defaultLoggerConfig.BatchEnabled,
				BatchFlushInterval: 30 * time.Second,
			},
		},
		{
			name: "custom filter regex",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:       "token",
					cfgChatIDKey:      "chat_id",
					cfgFilterRegexKey: `"ERROR"`,
				},
			},
			want: loggerConfig{
				ClientConfig: ClientConfig{
					APIURL:  defaultClientConfig.APIURL,
					Token:   "token",
					ChatID:  "chat_id",
					Retries: defaultClientConfig.Retries,
					Timeout: defaultClientConfig.Timeout,
				},
				Attrs:              make(map[string]string),
				Template:           defaultLoggerConfig.Template,
				MaxBufferSize:      defaultLoggerConfig.MaxBufferSize,
				FilterRegex:        regexp.MustCompile(`"ERROR"`),
				BatchEnabled:       defaultLoggerConfig.BatchEnabled,
				BatchFlushInterval: defaultLoggerConfig.BatchFlushInterval,
			},
		},
		{
			name: "custom max buffer size",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:         "token",
					cfgChatIDKey:        "chat_id",
					cfgMaxBufferSizeKey: "100MB",
				},
			},
			want: loggerConfig{
				ClientConfig: ClientConfig{
					APIURL:  defaultClientConfig.APIURL,
					Token:   "token",
					ChatID:  "chat_id",
					Retries: defaultClientConfig.Retries,
					Timeout: defaultClientConfig.Timeout,
				},
				Attrs:              make(map[string]string),
				Template:           defaultLoggerConfig.Template,
				MaxBufferSize:      100 * 1024 * 1024, // 100MB
				BatchEnabled:       defaultLoggerConfig.BatchEnabled,
				BatchFlushInterval: defaultLoggerConfig.BatchFlushInterval,
			},
		},
		{
			name: "custom message_thread_id",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:          "token",
					cfgChatIDKey:         "chat_id",
					cfgMessageThreadIDKey: "message_thread_id", // New test case for message_thread_id
				},
			},
			want: loggerConfig{
				ClientConfig: ClientConfig{
					APIURL:  defaultClientConfig.APIURL,
					Token:   "token",
					ChatID:  "chat_id",
					Retries: defaultClientConfig.Retries,
					Timeout: defaultClientConfig.Timeout,
				},
				Attrs:              make(map[string]string),
				Template:           defaultLoggerConfig.Template,
				MaxBufferSize:      defaultLoggerConfig.MaxBufferSize,
				MessageThreadID:    message_thread_id, // Expecting parsed message_thread_id
				BatchEnabled:       defaultLoggerConfig.BatchEnabled,
				BatchFlushInterval: defaultLoggerConfig.BatchFlushInterval,
			},
		},
		{
			name: "failed to parse client config",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:   "token",
					cfgChatIDKey:  "chat_id",
					cfgRetriesKey: "invalid",
				},
			},
			wantErr: "failed to parse client config",
		},
		{
			name: "failed to parse extra attributes",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:    "token",
					cfgChatIDKey:   "chat_id",
					"labels-regex": `(.*\(`,
				},
			},
			wantErr: "failed to parse extra attributes",
		},
		{
			name: "failed to parse \"filter-regex\"",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:       "token",
					cfgChatIDKey:      "chat_id",
					cfgFilterRegexKey: `(.*\(`,
				},
			},
			wantErr: "failed to parse \"filter-regex\"",
		},
		{
			name: "failed to parse \"batch-flush-interval\"",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:              "token",
					cfgChatIDKey:             "chat_id",
					cfgBatchFlushIntervalKey: "invalid",
				},
			},
			wantErr: "failed to parse \"batch-flush-interval\"",
		},
		{
			name: "invalid \"max-buffer-size\"",
			containerDetails: ContainerDetails{
				Config: map[string]string{
					cfgTokenKey:         "token",
					cfgChatIDKey:        "chat_id",
					cfgMaxBufferSizeKey: "-1",
				},
			},
			wantErr: "failed to parse \"max-buffer-size\" option",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg, err := parseLoggerConfig(&tt.containerDetails)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, *cfg)
		})
	}
}

func TestParseClientConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  map[string]string
		want    ClientConfig
		wantErr string
	}{
		{
			name: "common",
			config: map[string]string{
				cfgTokenKey:  "token",
				cfgChatIDKey: "chat_id",
			},
			want: ClientConfig{
				APIURL:  defaultClientConfig.APIURL,
				Token:   "token",
				ChatID:  "chat_id",
				Retries: defaultClientConfig.Retries,
				Timeout: defaultClientConfig.Timeout,
			},
		},
		{
			name: "custom api url",
			config: map[string]string{
				cfgTokenKey:  "token",
				cfgChatIDKey: "chat_id",
				cfgURLKey:    "https://custom.url",
			},
			want: ClientConfig{
				APIURL:  "https://custom.url",
				Token:   "token",
				ChatID:  "chat_id",
				Retries: defaultClientConfig.Retries,
				Timeout: defaultClientConfig.Timeout,
			},
		},
		{
			name: "custom retries",
			config: map[string]string{
				cfgTokenKey:   "token",
				cfgChatIDKey:  "chat_id",
				cfgRetriesKey: "10",
			},
			want: ClientConfig{
				APIURL:  defaultClientConfig.APIURL,
				Token:   "token",
				ChatID:  "chat_id",
				Retries: 10,
				Timeout: defaultClientConfig.Timeout,
			},
		},
		{
			name: "custom timeout",
			config: map[string]string{
				cfgTokenKey:   "token",
				cfgChatIDKey:  "chat_id",
				cfgTimeoutKey: "20s",
			},
			want: ClientConfig{
				APIURL:  defaultClientConfig.APIURL,
				Token:   "token",
				ChatID:  "chat_id",
				Retries: defaultClientConfig.Retries,
				Timeout: 20 * time.Second,
			},
		},
		{
			name: "failed to parse retries",
			config: map[string]string{
				cfgTokenKey:   "token",
				cfgChatIDKey:  "chat_id",
				cfgRetriesKey: "invalid",
			},
			wantErr: "failed to parse \"retries\" option",
		},
		{
			name: "invalid retries",
			config: map[string]string{
				cfgTokenKey:   "token",
				cfgChatIDKey:  "chat_id",
				cfgRetriesKey: "-1",
			},
			wantErr: "invalid \"retries\" option",
		},
		{
			name: "failed to parse timeout",
			config: map[string]string{
				cfgTokenKey:   "token",
				cfgChatIDKey:  "chat_id",
				cfgTimeoutKey: "invalid",
			},
			wantErr: "failed to parse \"timeout\" option",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			containerDetails := ContainerDetails{Config: tt.config}
			cfg, err := parseClientConfig(&containerDetails)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, cfg)
		})
	}
}
