# ComboRPC

## 基于TCP + MessagePack的简易RPC框架，支持自定义中间件、自定义负载均衡，支持单个请求调用多个方法，支持广播服务。

***

### 项目结构

* `test/demo.test.go` 测试程序
* `client.go`: 客户端相关
* `server.go`: 服务端相关
* `net.go`: 基础网络服务（TCP）

***

## 使用方式

### 导入包

```
go get github.com/dpwgc/comborpc
```

```
import "github.com/dpwgc/comborpc"
```

### 服务端路由启动

```
func demoServe() {

    // 新建服务路由
    router := comborpc.NewRouter(comborpc.RouterOptions{
        Endpoint:    "0.0.0.0:8001",
    })

    // 添加中间件（中间件1：testMiddleware1、中间件2：testMiddleware2）
    router.AddMiddlewares(testMiddleware1, testMiddleware2)

    // 往服务路由里添加方法（方法1：testMethod1、方法2：testMethod2）
    router.AddMethod("testMethod1", testMethod1)
    router.AddMethod("testMethod2", testMethod2)

    // 启动路由监听服务
    err := router.Run()
    if err != nil {
        panic(err) 
    }
}
```

### 服务端路由方法列表

* `NewRouter`: 创建Router对象
* `Router`: 服务路由
    * `AddMethod`: 添加方法
    * `AddMiddleware`: 添加中间件
    * `AddMiddlewares`: 添加多个中间件
    * `Run`: 启动路由监听服务
    * `Close`: 关闭路由监听服务
* `RouterOptions`: 路由设置
  * `Endpoint`: 服务端地址
  * `QueueLen`: 连接队列长度
  * `ConsumerNum`: 队列消费者数量
  * `Timeout`: 请求超时时间

### 服务端方法编写示例

```
// 方法1
func testMethod1(ctx *comborpc.Context) {

    // ctx.Read() 直接读取请求体，自行解析
    
    // 将请求数据绑定在TestRequest结构体上
    request := TestRequest{}
    err := ctx.Bind(&request)
    if err != nil {
        panic(err)
    }
    
    // 打印请求体
    fmt.Println("testMethod1 request:", "A1:", request.A1, "A2:", request.A2, "A3:", request.A3)
    
    // 返回数据给客户端
    ctx.Write(TestResponse{
        Code: 200,
        Msg:  "testMethod1 return ok",
    })
}
```

### 样例代码中用到的请求与响应结构体

```
// 请求
type TestRequest struct {
    A1 string
    A2 int64
    A3 float64
}

// 响应
type TestResponse struct {
    Code int
    Msg  string
}
```

### 服务端中间件编写示例

```
// 中间件1
func testMiddleware1(ctx *comborpc.Context) {
    fmt.Println("testMiddleware1 start")
    // 获取请求参数中的A1字段
    fmt.Println("A1: ", ctx.Param("A1"))
    ctx.Next()
    fmt.Println("testMiddleware1 end")
}
```

### 服务端上下文方法列表

* `Context`: 上下文
  * `Bind`: 将请求数据解析并绑定在指定结构体上
  * `Read`: 直接读取请求体，自行解析
  * `Param`: 根据字段名获取某个请求参数的值
  * `Write`: 编写响应体
  * `RemoteAddr`: 客户端ip地址
  * `LocalAddr`: 本地ip地址
  * `CallMethod`: 当前被客户端调用的方法名
  * `CustomCache`: 用户自定义上下文缓存
  * `Next`: 进入下一个方法（中间件相关）
  * `Abort`: 停止继续执行下一个方法（中间件相关）

### 服务端关闭

```
// 关闭路由监听服务
router.Close()
```

### 客户端发送单一请求

#### 一次请求只调用一个方法，返回一个Response对象

```
func demoSingleRequest() {
  
  // 接收响应结果的结构体
  responseBind := TestResponse{}

  // 构建并发送请求，同时将响应结果绑定到TestResponse结构体上
  err := comborpc.NewSingleCall(comborpc.CallOptions{
      Endpoints: []string{"0.0.0.0:8001"},
  }).SetRequest("testMethod3", TestRequest{
      A1: "hello world 3",
      A2: 1003,
      A3: 54.1,
  }).DoAndBind(&responseBind)

  // 抛错
  if err != nil {
      panic(err)
  }

  // 打印响应结果
  fmt.Println("single response:", "code:", responseBind.Code, "msg:", responseBind.Msg)
}
```

### 客户端发送组合请求

#### 一次请求同时调用两个方法（testMethod1、testMethod2），返回一个Response数组

```
func demoComboRequest() {

  // 构建并发送请求
  responseList, err := comborpc.NewComboCall(comborpc.CallOptions{
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
  
  if err != nil {
      panic(err)
  }
	
  // 遍历响应列表
  for _, response := range responseList {
  
      // 可以通过这种方式获取某个响应参数
      // code := response.Param("Code").(uint8) ---> code = 200
  
      // 将每个响应列表子项数据绑定到TestResponse结构体上
      responseBind := TestResponse{}
      err = response.Bind(&responseBind)
      if err != nil {
          panic(err)
      }
		
      // 打印每个响应结果
      fmt.Println("combo response item:", "code:", responseBind.Code, "msg:", responseBind.Msg)
  }
}
```

### 客户端方法列表

* `NewComboCall`: 创建ComboCall对象
* `ComboCall`: 组合调用
  * `AddRequest`: 添加请求体
  * `AddRequests`: 添加多个请求体
  * `Do`: 执行请求
  * `Broadcast`: 广播请求（给所有服务端地址发送请求）
* `NewSingleCall`: 创建SingleCall对象
* `SingleCall`: 单一调用
  * `SetRequest`: 设置请求体
  * `Do`: 执行请求
  * `DoAndBind`: 执行请求，并将响应数据绑定到指定结构体上
  * `Broadcast`: 广播请求（给所有服务端地址发送请求）
* `CallOptions`: 调用参数
  * `Endpoints`: 服务端地址列表
  * `Timeout`: 请求超时时间
  * `LoadBalance`: 自定义负载均衡方法
* `Response`: 响应体
  * `Bind`: 将响应数据解析并绑定在指定结构体上
  * `Param`: 根据字段名获取某个响应参数的值
  * `Success`: 判断是否响应成功

### 自定义负载均衡器

#### 自定义负载均衡方法（样例）
```
// 随机负载均衡处理方法
func deomLoadBalance(endpoints []string) string {
    if len(endpoints) == 0 {
        return ""
    }
    rand.Seed(time.Now().Unix())
    return endpoints[rand.Intn(len(endpoints))]
}
```
#### 设置方法
```
comborpc.NewComboCall(comborpc.CallOptions{
    Endpoints: []string{"0.0.0.0:8001"},
    LoadBalance: deomLoadBalance,
})
```

### 数据传输方式

* 1、建立TCP连接
* 2、客户端发送8位的请求头（内容是请求体长度，int64类型）
```
123
```
* 3、客户端发送请求体（使用MessagePack协议将结构体序列化成字节数组，然后再用gzip压缩。请求结构体：Method为方法名，Data为传入该方法的数据）
```json
[
  {
    "Method": "testMethod1",
    "Data": {
      "A1": "hello world 1",
      "A2": 1001,
      "A3": 89.2
    }
  },
  {
    "Method": "testMethod2",
    "Data": {
      "A1": "hello world 2",
      "A2": 1002,
      "A3": 67.5
    }
  }
]
```
* 4、服务端解析请求体 -> 并发执行方法
* 5、服务端发送响应头（内容是响应体长度，int64类型）
```
123
```
* 6、服务端发送响应体（序列化与压缩方式与请求体相同。响应结构体：Error为报错内容，Data为响应数据，响应体数组排序与请求体数组一致）
```json
[
  {
    "Error": "",
    "Data": {
      "Code": 200,
      "Msg": "testMethod1 return ok"
    }
  },
  {
    "Error": "",
    "Data": {
      "Code": 200,
      "Msg": "testMethod2 return ok"
    }
  }
]
```
* 7、客户端接收响应体
* 8、断开TCP连接