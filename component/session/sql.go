package session

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"

	"strings"
	"time"
)

// StoreSQL 定义sql数据库存储，支持mysql、mariadb、pgsql
type StoreSQL struct {
	*sql.DB
	stmtInsert *sql.Stmt
	stmtDelete *sql.Stmt
	stmtUpdate *sql.Stmt
	stmtSelect *sql.Stmt
	stmtClean  *sql.Stmt
}

// NewSessionDB 使用sql.DB连接创建会话。
func NewSessionDB(db *sql.DB) Session {
	return NewsessionStd(NewStoreDB(db))
}

// NewSessionSQL 创建一个使用sql存储的会话对象。
//
// 支持 PostgreSQL MariaDB MySQL
func NewSessionSQL(name, config string) Session {
	db, err := sql.Open(name, config)
	if err != nil {
		return nil
	}
	return NewsessionStd(NewStoreDB(db))
}

// NewStoreDB 使用sql.DB连接创建存储。
func NewStoreDB(db *sql.DB) Store {
	store := &StoreSQL{DB: db}
	err := store.Init()
	if err != nil {
		return nil
	}
	return store
}

// Init 方法执行存储的初始化。
func (store *StoreSQL) Init() error {
	name, err := store.getServerName()
	if err != nil {
		return err
	}
	switch name {
	case "MariaDB", "MySQL":
		store.DB.Exec(`CREATE TABLE IF NOT EXISTS tb_eudore_session ( id VARCHAR (32) PRIMARY KEY , data BLOB, expires TIMESTAMP)`)
		err = store.initStmt([]string{
			`INSERT INTO tb_eudore_session(id,data) VALUES(?,?)`,
			`DELETE FROM tb_eudore_session WHERE id=?`,
			`UPDATE tb_eudore_session SET data=?,expires=now() WHERE id=?`,
			`SELECT data FROM tb_eudore_session WHERE id=?`,
			`DELETE FROM tb_eudore_session WHERE expires>?`,
		})
	case "PostgreSQL":
		store.DB.Exec(`CREATE TABLE tb_eudore_session ( "id" VARCHAR (32) PRIMARY KEY , "data" BYTEA, "expires" TIMESTAMP)`)
		err = store.initStmt([]string{
			`INSERT INTO tb_eudore_session("id", "data") VALUES($1,$2)`,
			`DELETE FROM tb_eudore_session WHERE id=$1`,
			`UPDATE tb_eudore_session SET data=$1,expires=now() WHERE id=$2`,
			`SELECT data FROM tb_eudore_session WHERE id=$1`,
			`DELETE FROM tb_eudore_session WHERE expires>$1`,
		})
	default:
		return fmt.Errorf("undinfe sql name: %s", name)
	}

	return err
}

// getServerName 获得sql.DB的数据库名称。
func (store *StoreSQL) getServerName() (string, error) {
	var version string
	err := store.DB.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return "", err
	}
	version = strings.ToLower(version)
	switch {
	case strings.Contains(version, "postgre"):
		return "PostgreSQL", nil
	case strings.Contains(version, "maria"):
		return "MariaDB", nil
	}

	var name string
	err = store.DB.QueryRow("show variables like '%version_comment%'").Scan(&name, &version)
	if err != nil {
		return "", err
	}
	version = strings.ToLower(version)
	switch {
	case strings.Contains(version, "maria"):
		return "MariaDB", nil
	case strings.Contains(version, "mysql"):
		return "MySQL", nil
	}

	return "", err
}

// initStmt 初始化sql.Stmt
func (store *StoreSQL) initStmt(sqls []string) (err error) {
	store.stmtInsert, err = store.DB.Prepare(sqls[0])
	if err != nil {
		return err
	}
	store.stmtDelete, err = store.DB.Prepare(sqls[1])
	if err != nil {
		return err
	}
	store.stmtUpdate, err = store.DB.Prepare(sqls[2])
	if err != nil {
		return err
	}
	store.stmtSelect, err = store.DB.Prepare(sqls[3])
	if err != nil {
		return err
	}
	store.stmtClean, err = store.DB.Prepare(sqls[4])
	if err != nil {
		return err
	}
	return nil
}

// Insert 方法执行插入
func (store *StoreSQL) Insert(key string) (err error) {
	var data bytes.Buffer
	err = gob.NewEncoder(&data).Encode(make(map[string]interface{}))
	if err != nil {
		return nil
	}
	_, err = store.stmtInsert.Exec(key, data.Bytes())
	return
}

// Delete 方法执行删除数据
func (store *StoreSQL) Delete(key string) (err error) {
	_, err = store.stmtDelete.Exec(key)
	return
}

// Update 方法执行更新数据。
func (store *StoreSQL) Update(key string, val map[string]interface{}) error {
	var data bytes.Buffer
	err := gob.NewEncoder(&data).Encode(val)
	if err != nil {
		return err
	}
	_, err = store.stmtUpdate.Exec(data.Bytes(), key)
	return err
}

// Select 方法查询加载一个会话数据。
func (store *StoreSQL) Select(key string) (map[string]interface{}, error) {
	var data []byte
	err := store.stmtSelect.QueryRow(key).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrDataNotFound
		}
		return nil, err
	}

	var i map[string]interface{}
	err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&i)
	return i, err
}

// Clean 方法清理过期数据。
func (store *StoreSQL) Clean(expires time.Time) error {
	_, err := store.stmtClean.Exec(expires)
	return err
}
