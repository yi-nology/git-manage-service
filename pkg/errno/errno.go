package errno

import (
	"fmt"
)

type ErrNo struct {
	ErrCode int32
	ErrMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("err_code=%d, err_msg=%s", e.ErrCode, e.ErrMsg)
}

func NewErrNo(code int32, msg string) ErrNo {
	return ErrNo{
		ErrCode: code,
		ErrMsg:  msg,
	}
}

func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrMsg = msg
	return e
}

var (
	Success                = NewErrNo(0, "Success")
	ServiceErr             = NewErrNo(10001, "Service Error")
	ParamErr               = NewErrNo(10002, "Parameter Error")
	AuthorizationFailedErr = NewErrNo(10003, "Authorization Failed")
	RecordNotFound         = NewErrNo(10004, "Record Not Found")
	RepoAlreadyExists      = NewErrNo(10005, "Repo Already Exists")
)

// ConvertErr converts an error to ErrNo
func ConvertErr(err error) ErrNo {
	if err == nil {
		return Success
	}
	if e, ok := err.(ErrNo); ok {
		return e
	}
	return ServiceErr.WithMessage(err.Error())
}
