package server

const (
	_ = iota
	ErrorCode1
	ErrorCode2
	ErrorCode3
)

// CommonResult http接口通用返回结果
type CommonResult struct {
	IsSuc bool
	ErrorCode int
	Msg string
	Data interface{}
}

func Suc() *CommonResult {
	return &CommonResult{
		IsSuc:     true,
		ErrorCode: 0,
		Msg:       "执行操作成功",
		Data:      nil,
	}
}

func SucWithMsg(msg string) *CommonResult {
	return &CommonResult{
		IsSuc:     true,
		ErrorCode: 0,
		Msg:       msg,
		Data:      nil,
	}
}

func SucWithData(data interface{}, msg ...string) *CommonResult {
	realMsg := "执行操作成功"
	if len(msg) != 0 {
		realMsg = msg[0]
	}
	return &CommonResult{
		IsSuc:     true,
		ErrorCode: 0,
		Msg:       realMsg,
		Data:      data,
	}
}

func Fail(errCode int, msg string) *CommonResult {
	return &CommonResult{
		IsSuc:     false,
		ErrorCode: errCode,
		Msg:       msg,
		Data:      nil,
	}
}

func FailWithMsg(msg string) *CommonResult {
	return &CommonResult{
		IsSuc:     false,
		ErrorCode: 0,
		Msg:       msg,
		Data:      nil,
	}
}