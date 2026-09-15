package middleware

import (
	"regexp"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/gin-gonic/gin"
)

func RBAC(_ app.Application) func(*gin.Context) {
	return func(c *gin.Context) {
		// appCtx := app.NewContext(c, appInstance)
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			common.ErrUnAuthorized(c, "invalid user")
			c.Abort()
			return
		}
		perms := conf.PermissionsForRole(userInfo.GetRole())
		path := c.Request.URL.Path
		method := c.Request.Method

		if len(perms) == 0 {
			logger.Logger(c).Errorf("user has no permission, userInfo:[%s]", safeUserInfo(userInfo))
			common.ErrResp(c, &pb.CommonResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "user has no permission"}}, "user has no permission")
			c.Abort()
			return
		}

		for _, perm := range perms {
			if ruleMatched(ruleMatchParam{
				RuleMethod:    perm.Method,
				RulePath:      perm.Path,
				RequestPath:   path,
				RequestMethod: method,
			}) {
				logger.Logger(c).Debugf("user has api permission, continue")
				c.Next()
				return
			}
		}

		logger.Logger(c).Errorf("user has no permission, perms: %s, userInfo: [%s]", utils.MarshalForJson(perms), safeUserInfo(userInfo))
		common.ErrResp(c, &pb.CommonResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "user has no permission"}}, "user has no permission")
		c.Abort()
	}
}

type ruleMatchParam struct {
	RuleMethod    string
	RulePath      string
	RequestPath   string
	RequestMethod string
}

func ruleMatched(param ruleMatchParam) bool {
	methodMatch := false
	if param.RuleMethod == param.RequestMethod || param.RuleMethod == "*" {
		methodMatch = true
	}

	if !methodMatch {
		return false
	}

	pathMatch := false
	if param.RulePath == "*" {
		pathMatch = true
	} else {
		rule, err := regexp.Compile(param.RulePath)
		pathMatch = err == nil && rule.MatchString(param.RequestPath)
	}

	return pathMatch
}

func safeUserInfo(userInfo models.UserInfo) string {
	if userInfo == nil {
		return "unknown"
	}
	return utils.MarshalForJson(userInfo.GetSafeUserInfo())
}
