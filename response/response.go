package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/ao9911/go-matrix/ecode"
)

// Resp .
type Resp struct {
	Code int32       `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

// JSONResponse .
func JSONResponse(c *gin.Context, httpStatus int, ecode int32, msg string, data interface{}) {
	resp := &Resp{
		Code: ecode,
		Msg:  msg,
		Data: data,
	}
	c.JSON(httpStatus, resp)
}

// JSONSuccess .
func JSONSuccess(c *gin.Context, data interface{}) {
	JSONResponse(c, http.StatusOK, 0, "success", data)
}

// JSONFail .
func JSONFail(c *gin.Context, err error, data interface{}) {
	var (
		ec ecode.Codes
		ok bool
	)
	err = ecode.FromStatus(err)
	if err != nil {
		ec, ok = errors.Cause(err).(ecode.Codes)
		if ok {
			if tc := ecode.FromGRPCCode(ecode.Int(ec.Code())); tc == ecode.NotGrpcCode {
				JSONResponse(c, http.StatusOK, ec.Code(), ec.Message(), data) //非grpc错误
			} else {
				JSONResponse(c, tc, int32(tc), "Internal Server Error", data)
			}
		} else {
			JSONResponse(c, http.StatusBadRequest, ecode.ServerErr.Code(), err.Error(), data)
		}
	}
}

// AbortWithJSONResponse .
func AbortWithJSONResponse(c *gin.Context, httpStatus int, ecode int32, msg string, data interface{}) {
	resp := &Resp{
		Code: ecode,
		Msg:  msg,
		Data: data,
	}
	c.Abort()
	c.JSON(httpStatus, resp)
}

// AbortWithJSONSuccess .
func AbortWithJSONSuccess(c *gin.Context, data interface{}) {
	AbortWithJSONResponse(c, http.StatusOK, 0, "success", data)
}

// AbortWithJSONFail .
func AbortWithJSONFail(c *gin.Context, ecode int32, msg string) {
	AbortWithJSONResponse(c, http.StatusOK, ecode, msg, nil)
}
