package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ifaceless/rpcx/rpc"
	"github.com/ifaceless/rpcx/rsdk"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.PanicLevel)

	ctx := context.Background()
	defer func() {
		rsdk.Release()
		time.Sleep(2 * time.Second)
	}()

	rsdk.Init(10, 1*time.Millisecond)

	timeit("getMembersOneByOne", func() {
		getMembersOneByOne(ctx, 20)
	})

	timeit("getMembersConcurrently", func() {
		getMembersConcurrently(ctx, 20)
	})

	timeit("getMembersAutoAgg", func() {
		getMembersAutoAgg(ctx, 20)
	})
}

func timeit(name string, fn func()) {
	begin := time.Now()
	func() {
		defer func() {
			if e := recover(); e != nil {
				fmt.Println(e)
			}
		}()

		fn()
	}()
	fmt.Printf("%s: %s\n", name, time.Now().Sub(begin).String())
}

func getMembersOneByOne(ctx context.Context, max int) {
	for i := 0; i <= max; i++ {
		r, err := rpc.GetMember(ctx, int64(i+1))
		_, _ = r, err
	}
}

func getMembersConcurrently(ctx context.Context, max int) {
	var wg sync.WaitGroup
	for i := 0; i <= max; i++ {
		wg.Add(1)
		go func(n int) {
			r, err := rpc.GetMember(ctx, int64(n)+1)
			_, _ = r, err
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func getMembersAutoAgg(ctx context.Context, max int) {
	var wg sync.WaitGroup
	for i := 0; i <= max; i++ {
		wg.Add(1)
		// 底层自动将并发单次调用转换成批量调用
		go func(n int) {
			r, err := rsdk.GetMember(ctx, int64(n)+1)
			_, _ = r, err
			wg.Done()
		}(i)
	}
	wg.Wait()
}
