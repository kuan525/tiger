### 流程
1. 监听端口
2. 初始化协程池（ants）
3. 初始化epoll池
   1. 打开进程打开文件限制（fd：1048576）
   2. 初始化epoll池（多个epoll）
   3. 启动（cpu核数）协程，负责accept事件，NewConnect插入eChan
   4. 启动eSize个epoll，每个分配一个协程处理eChan，以及一个协程处理msg：runProc
4. 初始化rpc，tgrpc注册到etcd
5. 启动rpc-client（这里启动state的client）
6. 启动rpc-server：通过cmdChannel异步处理

### 协程用量
1. 监听accept事件【cpu数量】
2. eSize个epoll，对应一个处理eChan（accept），一个处理sendMsg【eSize*2】
3. ants/pool【1024】
4. tgrpc内部【TODO】

### 上报etcd

### 运行
1. etcd
1. go build && ./tiger gateway --config=./tiger.yaml
