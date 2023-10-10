package test

import (
	"context"
	"fmt"
	"github.com/dpwgc/comborpc"
	"testing"
	"time"
)

func Test(t *testing.T) {

	fmt.Println("-----\n1. start test")

	endpoint := "0.0.0.0:8001"
	go enableTestRouter(endpoint)
	fmt.Println("-----\n2. enable test router, endpoint:", endpoint)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n3. send combo request")
	responseList, err := comborpc.NewComboRequestBuilder(endpoint).
		Add(comborpc.Request{
			Method: "testMethod1",
			Data:   "test request data 1",
		}).
		Add(comborpc.Request{
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
		Set(comborpc.Request{
			Method: "testMethod1",
			Data:   "testData1",
		}).Send()
	if err != nil {
		panic(err)
	}
	fmt.Println("single response:", response)

	time.Sleep(1 * time.Second)

	fmt.Println("-----\n5. end test")
}

func enableTestRouter(endpoint string) {
	comborpc.NewRouter(endpoint, 30*time.Second).
		Add("testMethod1", testMethod1).
		Add("testMethod2", testMethod2).
		Listen()
}

func testMethod1(ctx context.Context, data string) string {
	fmt.Println("testMethod1 request:", data)
	return "hello world 1"
}
func testMethod2(ctx context.Context, data string) string {
	fmt.Println("testMethod2 request:", data)
	return "hello world 2"
}
