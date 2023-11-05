# ComboRPC

## 一个基于 TCP + MessagePack 的简易RPC框架

### 特点

* 支持单个请求调用多个方法
* 使用MessagePack序列化消息，Gzip压缩消息
* 支持自定义请求头
* 支持自定义中间件
* 支持自定义负载均衡
* 支持广播调用

# 开始使用

***

## 导入包

### 运行以下Go命令来安装模块包

```
$ go get github.com/dpwgc/comborpc
```

### 在代码中引入模块包

```
import "github.com/dpwgc/comborpc"
```

***

## 服务端 - 简单示例

### 启动一个服务路由：`example.go`

```
package main

import (
    "fmt"
    "github.com/dpwgc/comborpc"
)

// 服务路由示例
func main() {

    // 新建路由，TCP服务端口设为8001
    router := comborpc.NewRouter(comborpc.RouterOptions{
        Endpoint:    "0.0.0.0:8001",
    })

    // 添加路由中间件（一般用于请求头权限信息统一校验）
    router.AddMiddleware(testMiddleware)

    // 添加路由方法（为指定路由设置处理方法）
    router.AddMethod("testMethod", testMethod)
    // router.AddMethod("testMethod1", testMethod1)
    // router.AddMethod("testMethod2", testMethod2)

    // 启动路由
    router.Run()
}

// 路由中间件（示例） 
func testMiddleware(ctx *comborpc.Context) {
    
    // 获取请求头中的token参数值
    token := ctx.GetHeader("token")
    fmt.Println("token:", token)
    
    // 执行下一个方法
    ctx.Next()
}

// 路由方法（示例） 
func testMethod(ctx *comborpc.Context) {
    
    // 将请求数据绑定在request对象上
    request := TestRequest{}
    ctx.Bind(&request)
    
    // 打印请求体
    fmt.Println("A1:", request.A1, "A2:", request.A2, "A3:", request.A3)
    
    // 设置响应数据
    ctx.Write(TestResponse{
        Code: 200,
        Msg:  "ok",
    })
}

// 请求体（示例）
type TestRequest struct {
    A1 string
    A2 int64
    A3 float64
}

// 响应体（示例）
type TestResponse struct {
    Code int
    Msg  string
}
```

### 运行`example.go`程序

```
$ go run example.go
```

### 当终端弹出如下日志时，说明路由已启动

```
2023/11/01 18:00:00 listen and serve on 0.0.0.0:8001
```

***

## 服务端 - 更多设置及用法

* `comborpc.RouterOptions`: 路由设置-结构体
  * `Endpoint`: 服务端地址
  * `QueueLen`: 连接队列长度
  * `MaxGoroutine`: 最大协程数量
  * `Timeout`: 服务端超时时间


* `comborpc.Router`: 服务路由-结构体
  * `AddMethod`: 添加路由方法
  * `AddMiddleware`: 添加路由中间件
  * `AddMiddlewares`: 添加多个路由中间件
  * `Run`: 启动路由
  * `Close`: 关闭路由


* `comborpc.Context`: 路由方法上下文-结构体
  * `Bind`: 将请求数据解析并绑定在指定结构体上
  * `Write`: 设置响应体
  * `GetHeader`: 根据key获取指定请求头参数
  * `GetHeaders`: 获取全部请求头参数
  * `RemoteAddr`: 客户端ip地址
  * `LocalAddr`: 本地ip地址
  * `CallMethod`: 当前被客户端调用的方法名
  * `CustomCache`: 用户自定义的上下文缓存
  * `Next`: 进入下一个方法（路由中间件相关）
  * `Abort`: 中断执行链路并返回上层，不再继续执行下一个方法（路由中间件相关）

***

## 客户端 - 简单示例

### 使用`comborpc.SingleCall`向`0.0.0.0:8001`服务端发起单一调用，一次请求只调用一个方法，返回一个`comborpc.Response`对象

```
// 单一请求示例
func exampleSingleCall() {
  
  // 接收响应结果的结构体
  responseBind := TestResponse{}

  // 构建并发送请求，同时将响应结果绑定到responseBind对象上
  comborpc.NewSingleCall(comborpc.CallOptions{
      Endpoints: []string{"0.0.0.0:8001"},
  }).SetRequest("testMethod", TestRequest{
      A1: "hello world 3",
      A2: 1003,
      A3: 54.1,
  }).DoAndBind(&responseBind)

  // 打印响应结果
  fmt.Println("single response:", "code:", responseBind.Code, "msg:", responseBind.Msg)
}
```

### 使用`comborpc.ComboCall`向`0.0.0.0:8001`服务端发起组合调用，一次请求同时调用两个方法，返回一个`comborpc.Response`数组

```
// 组合请求示例
func exampleComboCall() {

  // 构建并发送请求
  responseList, _ := comborpc.NewComboCall(comborpc.CallOptions{
      Endpoints: []string{"0.0.0.0:8001"},
  }).AddRequest("testMethod1", TestRequest{
      A1: "hello world 1",
      A2: 1001,
      A3: 89.2,
  }).AddRequest("testMethod2", TestRequest{
      A1: "hello world 2",
      A2: 1002,
      A3: 67.5,
  }).Do()
	
  // 遍历响应列表
  for _, response := range responseList {
  
      // 将响应列表的每个子项数据绑定到responseBind对象上
      responseBind := TestResponse{}
      response.Bind(&responseBind)
		
      // 打印每个响应结果
      fmt.Println("combo response item:", "code:", responseBind.Code, "msg:", responseBind.Msg)
  }
}
```

### 自定义请求头 `Headers`

#### 请求头是一个`map[string]string`类型对象

* 通过`PutHeader`方法添加请求头参数
* 通过`RemoveHeader`方法删除请求头参数

```
call := comborpc.NewSingleCall(comborpc.CallOptions{
    Endpoints: []string{"0.0.0.0:8001"},
}).PutHeader("token", "12345678").PutHeader("version", "v1.0") // 添加请求头参数

call.RemoveHeader("version") // 删除指定的请求头参数

// 发送请求
call.SetRequest("testMethod", TestRequest{
    A1: "hello world 3",
    A2: 1003,
    A3: 54.1,
}).Do()
```

### 自定义负载均衡器 `comborpc.LoadBalanceFunc`

#### 客户端默认的负载均衡策略是随机负载均衡，当传入`Endpoints`数组长度为1时，直接向第一个服务地址发起请求。

#### 自定义负载均衡方法编写（示例）
```
// 随机负载均衡处理方法
func exampleLoadBalance(endpoints []string) string {
    if len(endpoints) == 0 {
        return ""
    }
    // 随机负载均衡策略
    rand.Seed(time.Now().Unix())
    return endpoints[rand.Intn(len(endpoints))]
}
```

#### 在调用`NewComboCall`/`NewSingleCall`时设置自定义负载均衡方法

```
comborpc.NewComboCall(comborpc.CallOptions{
    Endpoints: []string{"0.0.0.0:8001", "0.0.0.0:8002", "0.0.0.0:8003"}, // 传入三个服务地址
    LoadBalance: exampleLoadBalance, // 将默认的负载均衡方法替换为自定义负载均衡方法
}).AddRequest("testMethod", TestRequest{
    A1: "hello world",
    A2: 1002,
    A3: 67.5,
}).Do()
```

***

## 客户端 - 更多设置及用法

* `comborpc.CallOptions`: 调用参数-结构体
  * `Endpoints`: 服务端地址列表
  * `Timeout`: 客户端超时时间
  * `LoadBalance`: 自定义负载均衡方法


* `comborpc.ComboCall`: 组合调用-结构体
  * `AddRequest`: 添加请求体
  * `Do`: 执行请求
  * `Broadcast`: 广播请求（给所有服务端地址发送请求）
  * `PutHeader`: 设置请求头参数
  * `RemoveHeader`: 删除请求头参数


* `comborpc.SingleCall`: 单一调用-结构体
  * `SetRequest`: 设置请求体
  * `Do`: 执行请求
  * `DoAndBind`: 执行请求，并将响应数据绑定到指定结构体上
  * `Broadcast`: 广播请求（给所有服务端地址发送请求）
  * `PutHeader`: 设置请求头参数
  * `RemoveHeader`: 删除请求头参数


* `comborpc.Response`: 响应-结构体
  * `Bind`: 将响应数据解析并绑定在指定结构体上
  * `Success`: 判断是否响应成功