# etcdservice-grpc

适用于grpc的etcd组件

## 说明
grpc目前是没有像go-micro那样可插拔的库直接调用etcd服务的，因此本项目实现了一套适用于grpc使用etcd的组件，并使用简单的grpc服务作为使用示例