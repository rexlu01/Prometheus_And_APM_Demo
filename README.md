###  prometheus and APM  接入 gin 的实例

### 代码结构
```
 |-apm
   |-ginapm.go -- 定义接口，初始化操作
 |-prometheus
   |-ginprometheus.go -- prometheus注册加入中间件等操作
 |-config
   |-server_config.go -- 读取本地文件，存服务信息，用于注册consul
 |-log
   |-zaplog_config.go -- 利用zaplog模块，定义日志模版: 初始化，定义log级别等
   |-zaplog_gin.go -- gin接入zaplog，加入到中间件中:r.Use(...)
 |-util
   |-consul_util.go -- consul的操作
   |-prometheus_util.go -- prometheus中的公共方法，只是初步，具体参照例子 https://github.com/zsais/go-gin-prometheus/blob/master/middleware.go
   |-comm_util.go -- 公共方法，获取本机IP等
   |-apm_util.go --apm的具体方法，创建span等
 |-server
   |-写具体服务的逻辑
 |-main.go ---主函数
```

