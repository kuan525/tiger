package config

import (
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

func GetStateCmdChannelNum() int {
	return viper.GetInt("state.cmd_channel_num")
}
func GetStateServiceAddr() string {
	return viper.GetString("state.service_addr")
}
func GetStateServiceName() string {
	return viper.GetString("state.service_name")
}
func GetStateServicePort() int {
	return viper.GetInt("state.server_port")
}
func GetStateRPCWeight() int {
	return viper.GetInt("state.weight")
}

var connStateSlotList []int

func GetStateServerLoginSlotRange() []int {
	if len(connStateSlotList) != 0 {
		return connStateSlotList
	}
	slotRangeStr := viper.GetString("state.conn_state_slot_range")
	slotRange := strings.Split(slotRangeStr, ",")
	left, err := strconv.Atoi(slotRange[0])
	if err != nil {
		panic(err)
	}
	right, err := strconv.Atoi(slotRange[1])
	if err != nil {
		panic(err)
	}
	res := make([]int, right-left+1)
	for i := left; i <= right; i++ {
		res[i] = i
	}
	connStateSlotList = res
	return connStateSlotList
}

func GetStateServerGatewayServerEndpoint() string {
	return viper.GetString("state.gateway_server_endpoint")
}
