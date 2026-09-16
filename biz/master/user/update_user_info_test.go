package user

import (
	"testing"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/gin-gonic/gin"
)

func TestLegacyProfileEndpointRejectsPasswordChanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ginContext, _ := gin.CreateTestContext(nil)
	ginContext.Set(defs.UserInfoKey, &models.UserEntity{UserID: 1, Status: models.STATUS_NORMAL})
	password := "replacement-password"
	response, err := UpdateUserInfoHander(app.NewContext(ginContext, app.NewApp()), &pb.UpdateUserInfoRequest{
		UserInfo: &pb.User{RawPassword: &password},
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.GetStatus().GetCode() != pb.RespCode_RESP_CODE_INVALID || response.GetStatus().GetMessage() != "use the password endpoint and provide the current password" {
		t.Fatalf("legacy password update was not rejected: %#v", response.GetStatus())
	}
}
