package auth

import (
	"context"

	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/cache"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/samber/lo"
)

func InitAuth(appInstance app.Application) {
	appCtx := app.NewContext(context.Background(), appInstance)
	logger.Logger(appCtx).Info("start to init frp user auth token")

	u, err := dao.NewQuery(appCtx).AdminGetAllUsers()
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Fatalf("init frp user auth token failed")
	}

	lo.ForEach(u, func(user *models.UserEntity, _ int) {
		cache.Get().Set([]byte(user.GetUserName()), []byte(user.GetToken()), 0)
	})

	logger.Logger(appCtx).Infof("init frp user auth token success, count: %d", len(u))
}
