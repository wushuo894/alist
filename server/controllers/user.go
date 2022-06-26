package controllers

import (
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func ListUsers(c *gin.Context) {
	var req common.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	log.Debugf("%+v", req)
	users, total, err := db.GetUsers(req.PageIndex, req.PageSize)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: users,
		Total:   total,
	})
}

func CreateUser(c *gin.Context) {
	var req model.User
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	if req.IsAdmin() || req.IsGuest() {
		common.ErrorStrResp(c, "admin or guest user can not be created", 400, true)
		return
	}
	if err := db.CreateUser(&req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func UpdateUser(c *gin.Context) {
	var req model.User
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	user, err := db.GetUserById(req.ID)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	if user.Role != req.Role {
		common.ErrorStrResp(c, "role can not be changed", 400, true)
		return
	}
	if err := db.UpdateUser(&req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	if err := db.DeleteUserById(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}
