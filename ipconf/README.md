ipconf/server.go
- config.Init(path): 初始化配置文件
- source.Init(): 初始化数据层，读取etcd中的数据并监控变化，数据读入到`eventChan`
- domain.Init(): 初始化领域层，消费eventChan中的数据，将结果添加到全局map(candidateTable)中

协程用量：
1. source初始化，用于监控etcd情况，并插入eventChan【1】
2. domain初始化，用于消费eventChan，消费结果插入ed.window.statChan【1】
3. 新endport上线，启动一个协程监控ed.window.statChan，实时更新stat状态【N】

内存泄漏：
1. 设备下线，上述“3”的协程未释放【但是gateway一般不会频繁上下线，仅持有，无频繁迭代】

评价指标：
- 活跃分：gateway 每秒钟收发字节数的 剩余值
- 静态分：gateway 总体持有的长连接数量的 剩余值

运行：
1. etcd
2. go build && ./tiger ipconf --config=./tiger.yaml
3. curl --location --request GET '127.0.0.1:6789/ip/list'

![ipconfig架构设计](https://s3.bmp.ovh/imgs/2023/12/14/3ad4941b5e5069bc.jpg)