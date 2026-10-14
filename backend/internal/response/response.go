// Package response 定义全站统一响应结构：{ code, msg, data }。
package response

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) {
	c.JSON(200, gin.H{"code": 0, "msg": "ok", "data": data})
}

func Fail(c *gin.Context, httpStatus int, code int, msg string) {
	c.JSON(httpStatus, gin.H{"code": code, "msg": msg, "data": nil})
}
