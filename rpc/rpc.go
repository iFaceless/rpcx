package rpc

import (
	"context"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/ifaceless/rpcx/rproxy"
)

type Member struct {
	ID        int64
	Name      string
	Token     string
	Avatar    string
	CreatedAt time.Time
}

func BatchGetMember(_ context.Context, memberIDs []int64) (map[int64]*Member, error) {
	result := make(map[int64]*Member, len(memberIDs))
	for _, mid := range memberIDs {
		// 模拟数据库获取并组装数据
		result[mid] = &Member{
			ID:        mid,
			Name:      randomdata.SillyName(),
			Token:     randomdata.Alphanumeric(32),
			Avatar:    randomdata.Letters(32),
			CreatedAt: time.Now(),
		}
		time.Sleep(100 * time.Microsecond) // 0.1ms
	}

	// 模拟固定网络传输开销
	time.Sleep(10*time.Millisecond + time.Duration(len(memberIDs))*100*time.Microsecond)
	return result, nil
}

func GetMember(ctx context.Context, memberID int64) (*Member, error) {
	res, err := BatchGetMember(ctx, []int64{memberID})
	if err != nil {
		return nil, err
	}
	return res[memberID], nil
}

func init() {
	rproxy.Register("BatchGetMember", func(ctx context.Context, params ...interface{}) (interface{}, error) {
		memberIDs := make([]int64, len(params))
		for i, p := range params {
			memberIDs[i] = p.(int64)
		}
		return BatchGetMember(ctx, memberIDs)
	})
}
