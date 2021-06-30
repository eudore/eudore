package main

/*
基于控制器路由组合特性，使用反射创建gorm Model，UserController组合GormController后获得curd相关5个路由规则。

此GormController实现不完善,体现控制器类型，完整使用 github.com/eudore/endpoint/gorm.GormController。
*/

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/eudore/eudore"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	app := eudore.NewApp()
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
	WithDB    func(ctx eudore.Context) *gorm.DB
}

func NewControllerGorm(db *gorm.DB, model interface{}) *GormController {
	db.AutoMigrate(model)
	return &GormController{
		DB:        db,
		ModelType: reflect.Indirect(reflect.ValueOf(model)).Type(),
		WithDB: func(ctx eudore.Context) *gorm.DB {
			return db.WithContext(context.WithValue(ctx.GetContext(), "logger", ctx.Logger()))
		},
	}
}

func (ctl *GormController) ControllerParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}
	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
}

// ControllerRoute 方法返回控制器路由推导修改信息。
func (*GormController) ControllerRoute() map[string]string {
	return map[string]string{
		"Get":  "",
		"Post": "",
	}
}

type GormPaging struct {
	Page   int         `json:"page" alias:"page"`
	Size   int         `json:"size" alias:"size"`
	Order  string      `json:"order" alias:"order"`
	Total  int64       `json:"total" alias:"total"`
	Search string      `json:"search" alias:"search"`
	Data   interface{} `json:"data" alias:"data"`
}

func (ctl *GormController) Get(ctx eudore.Context) (interface{}, error) {
	paging := &GormPaging{Size: 50, Order: "id desc"}
	err := ctx.BindWith(paging, eudore.BindURL)
	if err != nil {
		return nil, err
	}

	paging.Data = reflect.New(reflect.SliceOf(ctl.ModelType)).Interface()
	db := ctl.WithDB(ctx).Model(paging.Data)
	if paging.Search != "" {
		cond, conddata := parseSearchExpression(paging.Search)
		db = db.Where(cond, conddata...)
	}
	err = db.Count(&paging.Total).Error
	if err != nil || paging.Total == 0 {
		return paging, err
	}
	err = db.Limit(paging.Size).Offset(paging.Size * paging.Page).Order(paging.Order).Find(paging.Data).Error
	return paging, err
}

func parseSearchExpression(key string) (string, []interface{}) {
	var sql string
	var data []interface{}
	regs := regexp.MustCompile(`(\S+\'.*\'|\S+)`)
	regc := regexp.MustCompile(`(\w*)(=|>|<|<>|!=|>=|<=|~|!~|:)(.*)`)
	for _, key := range regs.FindAllString(key, -1) {
		exp := regc.FindStringSubmatch(key)
		if len(exp) != 4 {
			if strings.HasSuffix(sql, "AND ") {
				sql = sql[:len(sql)-4] + "OR "
			}
			continue
		}
		if exp[2] == "~" {
			exp[2] = "LIKE"
		} else if exp[2] == "!~" {
			exp[2] = "NOT LIKE"
		} else if exp[2] == ":" {
			exp[2] = "IN"
		}
		sql += fmt.Sprintf("%s %s ? AND ", exp[1], exp[2])
		if exp[2] == "IN" {
			data = append(data, strings.Split(exp[3], ","))
		} else {
			data = append(data, exp[3])
		}
	}
	if strings.HasSuffix(sql, " AND ") || strings.HasSuffix(sql, " OR ") {
		sql = sql[:len(sql)-4]
	}
	return sql, data
}

func (ctl *GormController) GetById(ctx eudore.Context) (interface{}, error) {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctl.WithDB(ctx).Model(data).Find(data, "id=?", ctx.GetParam("id")).Error
	return data, err
}

func (ctl *GormController) Post(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctx.Bind(data)
	if err != nil {
		return err
	}
	return ctl.WithDB(ctx).Create(data).Error
}

func (ctl *GormController) PutById(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctx.Bind(data)
	if err != nil {
		return err
	}
	err = ctl.WithDB(ctx).Model(data).Where("id=?", ctx.GetParam("id")).Updates(data).Error
	return err
}

func (ctl *GormController) DeleteById(ctx eudore.Context) error {
	data := reflect.New(ctl.ModelType).Interface()
	err := ctl.WithDB(ctx).Model(data).Where("id=?", ctx.GetParam("id")).Delete(data).Error
	return err
}
