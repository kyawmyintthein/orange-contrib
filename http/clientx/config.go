package clientx

import "time"

type HttpClientCfg struct {
	DefaultContentType    string        `json:"default_content_type" mapstructure:"default_content_type"` // appication/json
	DefaultRetrySetting   RetryCfg      `json:"default_retry_setting" mapstructure:"default_retry_setting"`
	DefaultRequestTimeout time.Duration `json:"default_request_timeout" mapstructure:"default_request_timeout"` // milisecond
	/*
		CustomRetrySetting - is map type which can be used to specify custom value for each of the API.
							 The configuration key is to identify the API and it should follow the following format:
							 "[GET]::/users/profile": retry setting

	*/

	APISpecificRetrySetting map[string]RetryCfg `json:"api_specific_retry_setting" mapstructure:"api_specific_retry_setting"`

	HytrixSetting   HytrixCfg `json:"hytrix_setting" mapstructure:"hytrix_setting"`
	TurnOffLogger   bool      `json:"turn_off_logger" mapstructure:"turn_off_logger"`
	TurnOffNewrelic bool      `json:"turn_off_newrelic" mapstructure:"turn_off_newrelic"`
	TurnOffJaeger   bool      `json:"turn_off_jaeger" mapstructure:"turn_off_jaeger"`
}

type HytrixCfg struct {
	Enabled        bool                        `json:"enabled" mapstructure:"enabled"`
	CommandSetting map[string]HytrixCommandCfg `json:"command_setting" mapstructure:"command_setting"`
}

type HytrixCommandCfg struct {
	Enabled                bool `json:"enabled" mapstructure:"enabled"`
	RetryCount             int  `json:"retry_count" mapstructure:"retry_count"`
	Timeout                int  `json:"timeout" mapstructure:"timeout"`
	MaxConcurrentRequest   int  `json:"max_concurrent_request" mapstructure:"max_concurrent_request"`
	ErrorPercentThreshold  int  `json:"error_percent_threshold" mapstructure:"error_percent_threshold"`
	SleepWindow            int  `json:"sleep_window" mapstructure:"sleep_window"`
	RequestVolumeThreshold int  `json:"request_volume_threshold" mapstructure:"request_volume_threshold"`
}

type RetryCfg struct {
	Enabled          bool            `json:"enabled" mapstructure:"enabled"`
	MaxRetryAttempts uint            `json:"max_retry_attempts" json:"enabled"`
	BackOffDurations []time.Duration `json:"back_off_durations" json:"enabled"`
}
