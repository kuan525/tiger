# tiger
亿级通信的IM系统

方案汇总：[个人blog](https://blog.kuan525.com/categories/)

Client:
- 使用cobra组件，作为命令行解析层，众多知名开源golang项目的首选，k8s等，扩展性好。
- 使用gocui组件，用于绘制ui交互层，简单，代码好读，符合DDD策略，可以最小化开发成本。