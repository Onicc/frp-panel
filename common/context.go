package common

import (
	"context"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
)

func GetUserInfo(c context.Context) models.UserInfo {
	val := c.Value(defs.UserInfoKey)
	if val == nil {
		return nil
	}

	u, ok := val.(*models.UserEntity)
	if !ok {
		return nil
	}

	return u
}

func GetTokenString(c context.Context) string {
	value, _ := c.Value(defs.TokenKey).(string)
	return value
}
