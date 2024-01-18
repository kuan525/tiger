# GrpcWrapper-tgrpc
基于grpc进行封装，支持服务发现，服务注册，负载均衡，服务治理等功能。

- go watch: 获取etcd最新状态，更新DownServices，进而更新solver
- run: 监听两个channel，注册/删除 推到etcd上，且做注册/续期