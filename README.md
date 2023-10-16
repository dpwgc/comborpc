# ComboRPC

## 基于TCP的简易RPC框架，支持自定义中间件、自定义负载均衡，支持单个请求调用多个方法，支持广播服务。

***

### 项目结构

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
    fmt.Println("testMethod1 request:", ctx.ReadString())
    ctx.WriteString("hello world 1")
}

// 方法2
func testMethod2(ctx *comborpc.Context) {
    fmt.Println("testMethod2 request:", ctx.ReadString())
    ctx.WriteString("hello world 2")
}
```

### 服务端中间件编写示例

```
// 中间件1
func testMiddleware1(ctx *comborpc.Context) {
    fmt.Println("testMiddleware1 start")
    ctx.Next()
    fmt.Println("testMiddleware1 end")
}

// 中间件2
func testMiddleware2(ctx *comborpc.Context) {
    fmt.Println("testMiddleware2 start")
    ctx.Next()
    fmt.Println("testMiddleware2 end")
}
```

### 服务端上下文方法列表

* `Context`: 上下文
  * `ReadString`: 从请求体中读取字符串
  * `ReadJson`: 从请求体中读取Json格式字符串，将其解析为对象
  * `ReadYaml`: 从请求体中读取Yaml格式字符串，将其解析为对象
  * `ReadXml`: 从请求体中读取Xml格式字符串，将其解析为对象
  * `WriteString`: 将字符串写入响应体
  * `WriteJson`: 将对象序列化为Json格式字符串，并写入响应体
  * `WriteYaml`: 将对象序列化为Yaml格式字符串，并写入响应体
  * `WriteXml`: 将对象序列化为Xml格式字符串，并写入响应体
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

### 客户端发送组合请求

#### 一次请求同时调用两个方法（testMethod1、testMethod2），返回一个Response数组

```
func demoComboRequest() {

    // 构建并发送请求
    responseList, err := comborpc.NewComboCall(comborpc.CallOptions{
        Endpoints: []string{"0.0.0.0:8001"},
    }).AddRequest(comborpc.Request{
        Method: "testMethod1",
        Data:   "test request data 1",
    }).AddRequest(comborpc.Request{
        Method: "testMethod2",
        Data:   "test request data 2",
    }).Do()

    // 抛错
    if err != nil {
        panic(err)
    }

    // 响应结果打印
    fmt.Println("combo response list:", responseList)
}
```

### 客户端发送单一请求

#### 一次请求只调用一个方法，返回一个Response对象

```
func demoSingleRequest() {

    // 构建并发送请求
    response, err := comborpc.NewSingleCall(comborpc.CallOptions{
        Endpoints: []string{"0.0.0.0:8001"},
    }).SetRequest(comborpc.Request{
        Method: "testMethod1",
        Data:   "testData1",
    }).Do()

    // 抛错
    if err != nil {
        panic(err)
    }

    // 响应结果打印
    fmt.Println("single response:", response)
}
```

### 客户端方法列表

* `NewComboCall`: 创建ComboCall对象
* `ComboCall`: 组合调用
  * `AddRequest`: 添加请求体
  * `AddRequests`: 添加多个请求体
  * `AddStringRequest`: 添加请求体（传入普通字符串）
  * `AddJsonRequest`: 添加请求体（将传入对象序列化成Json字符串）
  * `AddYamlRequest`: 添加请求体（将传入对象序列化成Yaml字符串）
  * `AddXmlRequest`: 添加请求体（将传入对象序列化成Xml字符串）
  * `Do`: 执行请求
  * `Broadcast`: 广播请求
* `NewSingleCall`: 创建SingleCall对象
* `SingleCall`: 单一调用
  * `SetRequest`: 设置请求体
  * `SetStringRequest`: 设置请求体（传入普通字符串）
  * `SetJsonRequest`: 设置请求体（将传入对象序列化成Json字符串）
  * `SetYamlRequest`: 设置请求体（将传入对象序列化成Yaml字符串）
  * `SetXmlRequest`: 设置请求体（将传入对象序列化成Xml字符串）
  * `Do`: 执行请求
  * `Broadcast`: 广播请求
* `CallOptions`: 调用参数
  * `Endpoints`: 服务端地址列表
  * `Timeout`: 请求超时时间
  * `LoadBalance`: 自定义负载均衡方法
* `Response`: 响应体
  * `ParseJson`: 将Json字符串格式的响应数据解析为对象
  * `ParseYaml`: 将Yaml字符串格式的响应数据解析为对象
  * `ParseXml`: 将Xml字符串格式的响应数据解析为对象

### 自定义负载均衡器

#### 自定义负载均衡方法
```
// 随机负载均衡处理方法
func deomLoadBalance(endpoints []string) string {
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
95
```
* 3、客户端发送请求体（yaml格式数组字符串，gzip压缩，m为调用方法名，d为请求数据）
```yaml
- m: testMethod1
  d: '{"A":"test","B":"test"}'
- m: testMethod2
  d: '{"A":"test","B":"test"}'
```
* 4、服务端并发执行方法
* 5、服务端发送响应头（内容是响应体长度，int64类型）
```
73
```
* 6、服务端发送响应体（yaml格式数组字符串，gzip压缩，e为报错内容，d为响应数据，响应体数组排序与请求体数组一致）
```yaml
- e: 
  d: '{"A":"test","B":"test"}'
- e: 
  d: '{"A":"test","B":"test"}'
```
* 7、断开TCP连接