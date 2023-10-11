package test

import (
	"fmt"
	"github.com/dpwgc/comborpc"
	"testing"
	"time"
)

var router *comborpc.Router

func Test(t *testing.T) {

	fmt.Println("-----\n1. start test")

	endpoint := "0.0.0.0:8001"
	go enableTestRouter(endpoint)
	fmt.Println("-----\n2. enable test router, endpoint:", endpoint)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n3. send combo request")
	responseList, err := comborpc.NewComboRequestBuilder(endpoint).
		AddRequest(comborpc.Request{
			Method: "testMethod1",
			Data:   "test request data 1",
		}).
		AddRequest(comborpc.Request{
			Method: "testMethod2",
			Data:   "test request data 2",
		}).Send()
	if err != nil {
		panic(err)
	}
	fmt.Println("combo response list:", responseList)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n4. send single request")
	response, err := comborpc.NewSingleRequestBuilder(endpoint).
		SetRequest(comborpc.Request{
			Method: "testMethod1",
			Data:   "testData1",
		}).Send()
	if err != nil {
		panic(err)
	}
	fmt.Println("single response:", response)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n5. router close")
	err = router.Close()
	if err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n6. end test")
}

func enableTestRouter(endpoint string) {
	router = comborpc.NewRouter(endpoint, 10000, 100, 30*time.Second).
		AddMethod("testMethod1", testMethod1).
		AddMethod("testMethod2", testMethod2)
	router.ListenAndServe()
}

func testMethod1(data string) string {
	fmt.Println("testMethod1 request:", data)
	return "hello world 1"
}
func testMethod2(data string) string {
	fmt.Println("testMethod2 request:", data)
	return "hello world 2"
}
