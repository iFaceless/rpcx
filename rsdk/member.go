package rsdk

import (
	"context"

	"github.com/ifaceless/rpcx/rpc"
)

// 尝试使用 codegen 工具生成下面的模板代码
func GetMember(ctx context.Context, memberID int64) (*rpc.Member, error) {
	r, err := proxy.Call(ctx, "BatchGetMember", memberID,
		func(resultMap interface{}, param interface{}) interface{} {
			m := resultMap.(map[int64]*rpc.Member)
			return m[param.(int64)]
		})

	if err != nil {
		return nil, err
	}

	return r.(*rpc.Member), nil
}
