# ComboRPC

## 基于TCP的简易RPC框架，支持单个请求调用多个方法

***

## 使用方式

### 服务端启动

```
// 新建服务路由（地址、缓冲队列长度、队列消费者数量、请求处理超时时间）
router := comborpc.NewRouter("0.0.0.0:8001", 10000, 100, 30*time.Second)
// 往服务路由里添加处理方法（处理方法1：testMethod1、处理方法2：testMethod2）
router.AddMethod("testMethod1", testMethod1)
router.AddMethod("testMethod2", testMethod2)
// 启动路由监听服务
router.ListenAndServe()
```

### 服务端处理方法编写示例

#### 入参：context.Context、string，返回：string

```
// 处理方法1
func testMethod1(ctx *comborpc.Context) {
    fmt.Println("testMethod1 request:", ctx.ReadString())
    ctx.WriteString("hello world 1")
}
// 处理方法2
func testMethod2(ctx *comborpc.Context) {
    fmt.Println("testMethod2 request:", ctx.ReadString())
    ctx.WriteString("hello world 2")
}
```

### 服务端关闭

```
// 关闭路由监听服务
router.Close()
```

### 客户端发送组合请求

#### 一次请求同时调用两个方法（testMethod1、testMethod2），返回一个Response数组

```
// 构建并发送请求
responseList, err := comborpc.NewComboRequestBuilder("0.0.0.0:8001", 1*time.Minute).AddRequest(comborpc.Request{
    Method: "testMethod1",
    Data:   "test request data 1",
}).AddRequest(comborpc.Request{
    Method: "testMethod2",
    Data:   "test request data 2",
}).Send()
if err != nil {
    panic(err)
}
// 响应结果打印
fmt.Println("combo response list:", responseList)
```

### 客户端发送单一请求

#### 一次请求只调用一个方法，返回一个Response对象

```
// 构建并发送请求
response, err := comborpc.NewSingleRequestBuilder("0.0.0.0:8001", 1*time.Minute).SetRequest(comborpc.Request{
    Method: "testMethod1",
    Data:   "testData1",
}).Send()
if err != nil {
    panic(err)
}
// 响应结果打印
fmt.Println("single response:", response)
```