package policy

import (
	"database/sql"
	"time"

	"github.com/eudore/eudore"
)

/*
PolicysController 定义Policys api控制器
GET /policys
GET /policys/:id
POST /policys/:id
PUT /policys/:id

PUT /policys/:id/actions
DELTE /policys/:id/actions


GET /policys/:id/members
PUT /policys/:id/members/:memberid
DELETE /policys/:id/members/:memberid

*/
type PolicysController struct {
	eudore.ControllerAutoRoute
	Database *sql.DB
	Policys  *Policys

	//DatabaseSql PolicysSql
}

type policyModel struct {
	Policy
	CreatedAt time.Time `json:"create_at" alias:"create_at" db:"create_at"`
	UpdatedAt time.Time `json:"update_at" alias:"update_at" db:"update_at"`
}

type policyVersion struct {
	PolicyID  int        `json:"policy_id" alias:"policy_id" db:"policy_id" gorm:"primaryKey"`
	Version   string     `json:"version" alias:"version" gorm:"primaryKey"`
	Current   bool       `json:"current" alias:"current"`
	Statement RawMessage `json:"statement" alias:"statement"`
	CreatedAt time.Time  `json:"create_at" alias:"create_at" db:"create_at"`
}

type policyMember struct {
	Member
	CreatedAt time.Time `json:"create_at" alias:"create_at" db:"create_at"`
	UpdatedAt time.Time `json:"update_at" alias:"update_at" db:"update_at"`
}

type policyPermission struct {
	RoleID      int       `json:"role_id" alias:"role_id" db:"role_id" gorm:"primaryKey"`
	Permission  string    `json:"permission" alias:"permission" gorm:"primaryKey"`
	Description string    `json:"description" alias:"description"`
	CreatedAt   time.Time `json:"create_at" alias:"create_at" db:"create_at"`
	UpdatedAt   time.Time `json:"update_at" alias:"update_at" db:"update_at"`
}

type roleMember struct {
	UserID      int       `json:"user_id" alias:"user_id" db:"user_id" gorm:"primaryKey"`
	RoleID      int       `json:"role_id" alias:"role_id" db:"role_id" gorm:"primaryKey"`
	Description string    `json:"description" alias:"description"`
	Expiration  time.Time `json:"expiration" alias:"expiration"`
	CreatedAt   time.Time `json:"create_at" alias:"create_at" db:"create_at"`
	UpdatedAt   time.Time `json:"update_at" alias:"update_at" db:"update_at"`
}

// action={policy,version,member,permission,role}{POST,PUT,DELETE}
type policysLog struct {
	PolicyID  int       `json:"policy_id" alias:"policy_id" db:"policy_id" gorm:"primaryKey"`
	Action    string    `json:"action" alias:"action"`
	Message   string    `json:"message" alias:"message"`
	CreatedAt time.Time `json:"create_at" alias:"create_at" db:"create_at"`
}

// NewPolicysController 方法创建一个Policys相关api的eudore.Controller。
func (ctl *Policys) NewPolicysController(dbtype string, db *sql.DB) eudore.Controller {
	switch dbtype {
	case "postgres":
	case "sqltie", "sqltie3":
		db.Exec("CREATE TABLE `tb_eudore_policys` (`policy_id` integer,`policy_name` text,`description` text,`statement` text,PRIMARY KEY (`policy_id`))")
		db.Exec("CREATE TABLE `tb_eudore_policy_members` (`user_id` integer,`policy_id` integer,`index` integer,`description` text,`expiration` datetime,PRIMARY KEY (`user_id`,`policy_id`))")
	}
	pctl := &PolicysController{
		Database: db,
		Policys:  ctl,
		//DatabaseSql: PolicysSql{},
	}
	if db != nil {
		pctl.GetReloadPolicys()
		pctl.GetReloadMembers()
	}
	return pctl
}

// ControllerRoute 方法设置默认路由。
func (ctl *PolicysController) ControllerRoute() map[string]string {
	return map[string]string{
		"Get":  "",
		"Post": "",
	}
}

// GetRuntime 方法获取Policys运行时全部数据。
func (ctl *PolicysController) GetRuntime(ctx eudore.Context) interface{} {
	return ctl.Policys.HandleRuntime(ctx)
}

// GetReloadPolicys 方法使Policys重新加载策略信息。
func (ctl *PolicysController) GetReloadPolicys() error {
	var policys []Policy
	rows, err := ctl.Database.Query("SELECT policy_id, policy_name,description,statement FROM tb_eudore_policys")
	if err != nil {
		return err
	}
	eudore.ConvertRows(rows, &policys)
	for i := range policys {
		ctl.Policys.AddPolicy(&policys[i])
	}
	return nil
}

// GetReloadMembers 方法使Policys重新加载用户绑定详细。
func (ctl *PolicysController) GetReloadMembers() error {
	var members []Member
	rows, err := ctl.Database.Query("SELECT * FROM tb_eudore_policy_members")
	if err != nil {
		return err
	}
	eudore.ConvertRows(rows, &members)
	for i := range members {
		ctl.Policys.AddMember(&members[i])
	}
	return nil
}

// Get 方法获取全部策略。
func (ctl *PolicysController) Get(ctx eudore.Context) (interface{}, error) {
	var policys []Policy
	rows, err := ctl.Database.Query("SELECT policy_id, policy_name,description,statement FROM tb_eudore_policys")
	if err != nil {
		return nil, err
	}
	return &policys, eudore.ConvertRows(rows, &policys)
}

// GetById 方法获取指定id的策略。
func (ctl *PolicysController) GetById(ctx eudore.Context) (interface{}, error) {
	var policy Policy
	err := ctl.Database.QueryRow("SELECT policy_id,policy_name,description,statement FROM tb_eudore_policys WHERE policy_id=$1",
		ctx.GetParam("id")).Scan(&policy.PolicyID, &policy.PolicyName, &policy.Description, &policy.Statement)
	return policy, err
}

// Post 方法新增策略。
func (ctl *PolicysController) Post(ctx eudore.Context) error {
	var policy Policy
	err := ctx.Bind(&policy)
	if err != nil {
		return err
	}
	result, err := ctl.Database.Exec("INSERT INTO tb_eudore_policys(policy_name,description,statement) VALUES($1,$2,$3)",
		policy.PolicyName, policy.Description, policy.Statement)
	if err == nil {
		id, _ := result.LastInsertId()
		policy.PolicyID = int(id)
		ctl.Policys.AddPolicy(&policy)
	}
	return err
}

// PutById 方法修改指定策略。
func (ctl *PolicysController) PutById(ctx eudore.Context) error {
	var policy Policy
	err := ctx.Bind(&policy)
	if err != nil {
		return err
	}
	_, err = ctl.Database.Exec("UPDATE tb_eudore_policys SET policy_name=$1,description=$2,statement=$3 WHERE policy_id=$4",
		policy.PolicyName, policy.Description, policy.Statement, ctx.GetParam("id"))
	if err == nil {
		policy.PolicyID = eudore.GetStringInt(ctx.GetParam("id"))
		ctl.Policys.AddPolicy(&policy)
	}
	return err
}

// DeleteById 方法删除指定策略
func (ctl *PolicysController) DeleteById(ctx eudore.Context) error {
	_, err := ctl.Database.Exec("DELETE FROM tb_eudore_policys WHERE policy_id=$1", ctx.GetParam("id"))
	if err == nil {
		ctl.Policys.AddPolicy(&Policy{PolicyID: eudore.GetStringInt(ctx.GetParam("id"))})
	}
	return err
}

// GetMembers 方法获取全部绑定用户关系。
func (ctl *PolicysController) GetMembers(ctx eudore.Context) (interface{}, error) {
	var members []Member
	rows, err := ctl.Database.Query("SELECT * FROM tb_eudore_policy_members")
	if err != nil {
		return nil, err
	}
	return &members, eudore.ConvertRows(rows, &members)
}

// GetByIdMembers 方法获取指定策略绑定的用户详细。
func (ctl *PolicysController) GetByIdMembers(ctx eudore.Context) (interface{}, error) {
	var members []Member
	rows, err := ctl.Database.Query(`SELECT * FROM tb_eudore_policy_members WHERE user_id=$1 order by "index" DESC`, ctx.GetParam("id"))
	if err != nil {
		return nil, err
	}
	return &members, eudore.ConvertRows(rows, &members)
}

// PostByIdMembers 方法指定策略新增绑定用户详细。
func (ctl *PolicysController) PostByIdMembers(ctx eudore.Context) error {
	var member Member
	err := ctx.Bind(&member)
	if err != nil {
		return err
	}

	_, err = ctl.Database.Exec(`INSERT INTO tb_eudore_policy_members(policy_id,user_id,"index",description,expiration) VALUES($1,$2,$3,$4,$5)`,
		ctx.GetParam("id"), member.UserID, member.Index, member.Description, member.Expiration)
	if err == nil {
		member.PolicyID = eudore.GetStringInt(ctx.GetParam("id"))
		ctl.Policys.AddMember(&member)
	}
	return err
}

// PutByIdMembersByUserid 方法修改指定策略用户详细。
func (ctl *PolicysController) PutByIdMembersByUserid(ctx eudore.Context) error {
	var member Member
	err := ctx.Bind(&member)
	if err != nil {
		return err
	}
	member.PolicyID = eudore.GetStringInt(ctx.GetParam("id"))
	member.UserID = eudore.GetStringInt(ctx.GetParam("userid"))

	result, err := ctl.Database.Exec("UPDATE tb_eudore_policy_members SET 'index'=$1,description=$2,expiration=$3 WHERE policy_id=$4 AND user_id=$5",
		member.Index, member.Description, member.Expiration, member.PolicyID, member.UserID)
	if err == nil {
		row, _ := result.RowsAffected()
		if row == 1 {
			ctl.Policys.AddMember(&member)
		}
	}
	return err
}

// DeleteByIdMembersByUserid 方法删除指定策略的用户。
func (ctl *PolicysController) DeleteByIdMembersByUserid(ctx eudore.Context) error {
	_, err := ctl.Database.Exec("DELETE FROM tb_eudore_policy_members WHERE policy_id=$1 AND user_id=$2", ctx.GetParam("id"), ctx.GetParam("userid"))
	if err == nil {
		ctl.Policys.AddMember(&Member{
			PolicyID: eudore.GetStringInt(ctx.GetParam("id")),
			UserID:   eudore.GetStringInt(ctx.GetParam("userid")),
			Index:    -1,
		})
	}
	return err
}
