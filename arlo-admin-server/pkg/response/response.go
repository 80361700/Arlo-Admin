package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code" example:"200"`                // 响应码（200=成功，>0=业务错误）
	Msg  string      `json:"msg" example:"success"`           // 响应消息
	Data interface{} `json:"data" swaggertype:"object"`       // 响应数据（具体类型取决于接口）
}

type PageData struct {
	List     interface{} `json:"list" swaggertype:"object"`     // 数据列表
	Total    int64       `json:"total" example:"100"`          // 总记录数
	Page     int         `json:"page" example:"1"`             // 当前页码
	PageSize int         `json:"pageSize" example:"10"`        // 每页条数
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: data,
	})
}

func SuccessWithMsg(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  msg,
		Data: data,
	})
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}
