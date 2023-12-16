code流程：
1. 启动
2. init_source_date(DataHandler)
3. 操作source，将set和del植入common/discovery（插拔）
4. 网关机的行为称为event（source的概念）
5. 在3中通过set和del使channel监听event事件
6. domain领域层中，获取到上述channel
7. 将event一个个更新到domain的Dispatcher的map中
8. 因为这里涉及到分层，所以做了结构体替换
9. 在hertz中，读取map中数据，通过调度获取top5


- 每个endport都是一个网关机
- 一个网关机中保存了stats和window，为了分层，并且充分保证了异步，利用了go的优势
- 利用window中的channel监听event事件，保存最近的五次，并不断更新stats
- 在访问请求的时候，每次动态的计算最终的静态分和动态分
- 这里的嵌入结构体的方式，更多的是为了分层，架构目的


运行：
1. etcd
2. go build && ./tiger ipconf --config=./tiger.yaml
3. go build && ./tiger ipconf --config=./tiger.yaml

![ipconfig架构设计](https://s3.bmp.ovh/imgs/2023/12/14/3ad4941b5e5069bc.jpg)