package main

import (
	"os"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	os.Remove("./policydata.db")
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	policys := policy.NewPolicys()
	db, err := gorm.Open(sqlite.Open("policydata.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		app.Options(err)
		return
	}
	defer os.Remove("./policydata.db")

	{
		// create data
		db.AutoMigrate(&dataUser{})
		db.Create(&dataUser{ID: 1, Name: "user1", Group: "1"})
		db.Create(&dataUser{ID: 2, Name: "user2", Group: "1"})
		db.Create(&dataUser{ID: 3, Name: "user3", Group: "2"})
		db.Create(&dataUser{ID: 4, Name: "user4", Group: "2"})
	}

	{
		policys.AddPolicy(&policy.Policy{
			PolicyID:  1,
			Statement: []byte(`[{"effect":true,"data":[{"kind":"value","name":"id","value":["value:query:id"]},{"kind":"range","name":"id","min":"3"}]}]`),
		})
		policys.AddMember(&policy.Member{
			PolicyID: 1,
			UserID:   0,
		})

	}
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "Policy", "Resource"))
	app.AddMiddleware(policys.HandleHTTP)
	app.AddController(&dataUserController{DB: db})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/data/user/list").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type dataUser struct {
	ID          int
	Name        string
	Group       string
	Description string
}

type dataUserController struct {
	eudore.ControllerAutoRoute
	policy.ControllerAction
	*gorm.DB
}

func (ctl *dataUserController) GetList(ctx eudore.Context) (interface{}, error) {
	var datas []dataUser
	sql, args := policy.CreateExpressions(ctx, "data_users", []string{`name`, `group`, `description`, `id`}, -1)
	err := ctl.DB.Model(&datas).Where(sql, args...).Find(&datas).Error
	return datas, err
}
