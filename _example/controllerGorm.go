package main

/*
基于控制器路由组合特性，使用反射创建gorm Model，UserController组合GormController后获得curd相关5个路由规则。

此GormController实现不完善。
*/

import (
	"os"
	"reflect"
	"time"

	"github.com/eudore/eudore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	app := eudore.NewApp()
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	app.Options(err)
	if err == nil {
		app.AddController(NewUserController(db))
		defer os.Remove("gorm.db")
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UserController struct {
	eudore.ControllerAutoRoute
	*GormController
}

func NewUserController(db *gorm.DB) eudore.Controller {
	return &UserController{
		GormController: NewControllerGorm(db, new(User)),
	}
}

type GormController struct {
	*gorm.DB
	ModelType reflect.Type
}

func NewControllerGorm(db *gorm.DB, model interface{}) *GormController {
	db.AutoMigrate(model)
	return &GormController{
		DB:        db,
		ModelType: reflect.Indirect(reflect.ValueOf(model)).Type(),
	}
}

func (ctl *GormController) GetList(ctx eudore.Context) (interface{}, error) {
	size := eudore.GetStringInt(ctx.GetParam("size"), 20)
	page := eudore.GetStringInt(ctx.GetParam("page")) * size
	order := eudore.GetString(ctx.GetParam("order"), "id desc")
	datas := reflect.New(reflect.SliceOf(ctl.ModelType)).Interface()
	err := ctl.Model(datas).Limit(size).Offset(page).Order(order).Find(datas).Error
	return datas, err
}

func (ctl *GormController) GetById(ctx eudore.Context) (interface{}, error) {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctl.Model(data).Find(data, "id=?", ctx.GetParam("id")).Error
	return data, err
}

func (ctl *GormController) PostNew(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctx.Bind(data)
	if err != nil {
		return err
	}
	return ctl.Create(data).Error
}

func (ctl *GormController) PutById(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctx.Bind(data)
	if err != nil {
		return err
	}
	err = ctl.Model(data).Where("id=?", ctx.GetParam("id")).Updates(data).Error
	return err
}

func (ctl *GormController) DeleteById(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctl.Model(data).Where("id=?", ctx.GetParam("id")).Delete(data).Error
	return err
}
