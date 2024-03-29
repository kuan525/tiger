package config

import "github.com/spf13/viper"

// 获取discov用那种方式实现
func GetDiscovName() string {
	return viper.GetString("tgrpc.discov.name")
}

// 获取discov的endpoints
func GetDiscovEndpoints() []string {
	return viper.GetStringSlice("discovery.endpoints")
}

// 是否开启trace
func GetTraceEnable() bool {
	return viper.GetBool("tgrpc.trace.enable")
}

// 获取trace collection url
func GetTraceCollectionUrl() string {
	return viper.GetString("tgrpc.trace.url")
}

// 获取服务名
func GetTraceServiceName() string {
	return viper.GetString("tgrpc.trace.service_name")
}

// 获取trace采样率
func GetTraceSampler() float64 {
	return viper.GetFloat64("tgrpc.trace.sampler")
}
