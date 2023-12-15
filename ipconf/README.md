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


运行：
1. etcd
2. go build && ./tiger ipconf --config=./tiger.yaml
3. go build && ./tiger ipconf --config=./tiger.yaml

![ipconfig架构设计](https://s3.bmp.ovh/imgs/2023/12/14/3ad4941b5e5069bc.jpg)