package main

import (
	"database/sql"
	"os"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	defer os.Remove("rows.db")
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	db, err := gorm.Open(sqlite.Open("rows.db"), &gorm.Config{})
	app.Options(err)
	if err == nil {
		db.AutoMigrate(new(RowsUser))
		db.Create(&RowsUser{ID: 1, Name: "eudore"})
		db.Create(&RowsUser{ID: 2, Name: "middleware"})
		db.Create(&RowsUser{ID: 3, Name: "policy"})
		db.Create(&RowsUser{ID: 4, Name: "gateway"})
		db.Create(&RowsUser{ID: 5, Name: "endpoint"})
		sqlDB, _ := db.DB()
		app.AddController(&RowsController{DB: sqlDB})
	}

	client := httptest.NewClient(app).AddHeaderValue("Accept", eudore.MimeApplicationJSONUtf8)
	client.NewRequest("GET", "/struct/row").Do().Out()
	client.NewRequest("GET", "/struct/rows1").Do().Out()
	client.NewRequest("GET", "/struct/rows2").Do().Out()
	client.NewRequest("GET", "/struct/rows3").Do().Out()
	client.NewRequest("GET", "/struct/rows4").Do().Out()
	client.NewRequest("GET", "/map/row1").Do().Out()
	client.NewRequest("GET", "/map/row2").Do().Out()
	client.NewRequest("GET", "/map/rows").Do().Out()
	client.NewRequest("GET", "/count").Do().Out()
	client.NewRequest("GET", "/slice").Do().Out()
	client.NewRequest("GET", "/err/zero").Do().Out()
	client.NewRequest("GET", "/err/columns").Do().Out()
	client.NewRequest("GET", "/err/struct/notaddr").Do().Out()
	client.NewRequest("GET", "/err/map/key").Do().Out()
	client.NewRequest("GET", "/err/map/nil").Do().Out()
	client.NewRequest("GET", "/err/interface/nil").Do().Out()
	client.NewRequest("GET", "/err/interface/elem").Do().Out()
	client.NewRequest("GET", "/err/ptr/nil").Do().Out()
	client.NewRequest("GET", "/err/type/invalid").Do().Out()
	client.NewRequest("GET", "/err/value/invalid").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type RowsUser struct {
	ID   int `alias:"id"`
	Name string
}

type RowsController struct {
	eudore.ControllerAutoRoute
	*sql.DB
}

func (ctl *RowsController) ControllerGroup(string) string {
	return ""
}

func (ctl *RowsController) GetStructRow(ctx eudore.Context) (interface{}, error) {
	var user RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetStructRows1(ctx eudore.Context) (interface{}, error) {
	var user []RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetStructRows2(ctx eudore.Context) (interface{}, error) {
	var user []*RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetStructRows3(ctx eudore.Context) (interface{}, error) {
	user := make([]*RowsUser, 0, 10)
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetStructRows4(ctx eudore.Context) (interface{}, error) {
	var user [3]*RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetMapRow1(ctx eudore.Context) (interface{}, error) {
	var user map[string]interface{}
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}
func (ctl *RowsController) GetMapRow2(ctx eudore.Context) (interface{}, error) {
	user := make(map[string]interface{})
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetMapRows(ctx eudore.Context) (interface{}, error) {
	var user []map[string]interface{}
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetCount(ctx eudore.Context) (interface{}, error) {
	var user int
	rows, err := ctl.Query("select count(*) FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetSlice(ctx eudore.Context) (interface{}, error) {
	var user []int
	rows, err := ctl.Query("select id FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetErrZero(ctx eudore.Context) (interface{}, error) {
	var user interface{}
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetErrColumns(ctx eudore.Context) (interface{}, error) {
	var user []RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	rows.Close()
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetErrStructNotaddr(ctx eudore.Context) (interface{}, error) {
	var user RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetErrMapKey(ctx eudore.Context) (interface{}, error) {
	var user map[int]interface{}
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetErrMapNil(ctx eudore.Context) (interface{}, error) {
	var user map[string]interface{}
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetErrPtrNil(ctx eudore.Context) (interface{}, error) {
	var user *RowsUser
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, user)
}

func (ctl *RowsController) GetErrInterfaceNil(ctx eudore.Context) (interface{}, error) {
	var user eudore.Config
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetErrInterfaceElem(ctx eudore.Context) (interface{}, error) {
	var user interface{} = make(map[string]interface{})
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetErrValueInvalid(ctx eudore.Context) (interface{}, error) {
	var user []int
	rows, err := ctl.Query("select * FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}

func (ctl *RowsController) GetErrTypeInvalid(ctx eudore.Context) (interface{}, error) {
	var user func()
	rows, err := ctl.Query("select count(*) FROM rows_users")
	if err != nil {
		return nil, err
	}
	return &user, eudore.ConvertRows(rows, &user)
}
