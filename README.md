# RPCX

自动聚合单次 RPC 请求调用为若干次批量调用。业务方几乎无感知，侵入性小，保持原有的 GetByID 的写法即可。

![](https://pic3.zhimg.com/v2-700a32e38e1de03ac6997241804b5a6b.png)

## 背景

我们现在的 http api 主要写 model 和 schema。一般和 model 关联的资源，可能是通过 rpc 调用获得的。简单的写法是在 model 层写个 property，rpc.GetX 返回资源，然后 schema 层映射这个字段即可。

这样做的话，marshmallow 默认是逐个 render 的，所以串行调用 N 次会比较慢。一种优化方案是，并发 render，代价，这样对 schema 和 model 没有侵入性。但是对 service 提供方带来较多的请求放大，除非在并发调用时控制并发量，但是这样业务方做起来会比价麻烦，提炼到框架中会更适合。

另一种优化方案，是在 handler 层，BatchGetX 一次到位，然后通过 context 传入到 schema 中，再提取。这样不够优雅，Handler 层需要改动，Schema 层侵入性修改太大。

所以怎么能够保持原有的写法，model 层和 schema 层简单写，底层自动 BatchGetX，来优化。这就是该工具想要解决的问题。

```go
type StudentModel struct {
    MemberID int64
}

func (s *StudentModel) Member(ctx context.Member) *rpc.Member {
    return rpc.GetMemberByID(s.MemberID)
}

type MemberSchema struct { /*...*/ }

type StudentSchema struct {
    Member []MemberSchema `portal: nested`
}

students := StudentModel.GetsBy(limit=20)
var output []StudentSchema
// 串行一个个渲染填充 member，模拟 Python marshmallow 效果
portal.Dump(&output, students, portal.DisableConcurrency()) 

// 默认会并发填充 N 个 StudentSchema，代价是会做最大 N 个并发的 GetMemberByID 请求
// 服务提供方会很辛苦
portal.Dump(&output, students)

// 
// 优化 1：直接使用提供的 BatchGetMemberByID，但是写法上相对恶心
// 具体改造如下
//
type StudentSchema struct {
    Member []MemberSchema `portal: nested; meth: GetMember`
}

func (s *StudentSchema) GetMember(ctx context.Context, student *StudentModel) *rpc.Member {
    // 外层批量调用结果放在 ctx 中传入
    v := ctx.Value("members").(map[int64]*rpc.Member)
    return v[student.MemberID] // 注意不再使用 student.Member 属性了，那样会单次调用
}

// 相当于把原先获取数据源的地方挪到 Handler 层写了，不够优雅
members := BatchGetMemberByID(ctx, []int64{1, 2, 3, 4...})
portal.DumpWithContext(context.WithValue("members", members), &output, students)


// ????
// 那怎么保持最开始的写法，同时又能利用 BatchGetMemberByID 接口在尽可能减少对 member 服务
// 调用次数放大呢？

// 优化 2：
// 使用新的封装版本提供服务，业务上保持不变的写法
// 只是把单次 rpc 调用改掉（使用封装版本）。
// 底层会自动合并成 BatchGetMemberByID，但上层依然表现为单次调用
func (s *StudentModel) Member(ctx context.Member) *rpc.Member {
    return proxy.GetMemberByID(s.MemberID)
}
```

## 说明

仅仅是一个 Demo，演示下这个思路大概怎么实现。具体到业务场景，要走的路还很长。