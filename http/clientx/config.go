package client

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

	TurnOffLogger         bool `json:"turn_off_logger" mapstructure:"turn_off_logger"`
	TurnOffNewrelic       bool `json:"turn_off_newrelic" mapstructure:"turn_off_newrelic"`
	TurnOffJaeger         bool `json:"turn_off_jaeger" mapstructure:"turn_off_jaeger"`
	TurnOffCircuitBreaker bool `json:"turn_off_circuit_breaker" mapstructure:"turn_off_circuit_breaker"`
}

type RetryCfg struct {
	Enabled          bool            `json:"enabled" mapstructure:"enabled"`
	MaxRetryAttempts uint            `json:"max_retry_attempts" json:"enabled"`
	BackOffDurations []time.Duration `json:"back_off_durations" json:"enabled"`
}
