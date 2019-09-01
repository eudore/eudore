package session

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"

	"strings"
	"time"
)

type StoreSql struct {
	*sql.DB
	stmtInsert *sql.Stmt
	stmtDelete *sql.Stmt
	stmtUpdate *sql.Stmt
	stmtSelect *sql.Stmt
	stmtClean  *sql.Stmt
}

func NewSessionDB(db *sql.DB) Session {
	return NewSessionStd(NewStoreDB(db))
}

// NewSessionSql 创建一个使用sql存储的会话对象。
//
// 支持 PostgreSQL MariaDB MySQL
func NewSessionSql(name, config string) Session {
	db, err := sql.Open(name, config)
	if err != nil {
		panic(err)
	}
	return NewSessionStd(NewStoreDB(db))
}

func NewStoreDB(db *sql.DB) SessionStore {
	store := &StoreSql{DB: db}
	err := store.Init()
	if err != nil {
		panic(err)
	}
	return store
}

func (store *StoreSql) Init() error {
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
			`SELECT data,expires FROM tb_eudore_session WHERE id=?`,
			`DELETE FROM tb_eudore_session WHERE expires>?`,
		})
	case "PostgreSQL":
		store.DB.Exec(`CREATE TABLE tb_eudore_session ( "id" VARCHAR (32) PRIMARY KEY , "data" BYTEA, "expires" TIMESTAMP)`)
		err = store.initStmt([]string{
			`INSERT INTO tb_eudore_session("id", "data") VALUES($1,$2)`,
			`DELETE FROM tb_eudore_session WHERE id=$1`,
			`UPDATE tb_eudore_session SET data=$1,expires=now() WHERE id=$2`,
			`SELECT data,expires FROM tb_eudore_session WHERE id=$1`,
			`DELETE FROM tb_eudore_session WHERE expires>$1`,
		})
	default:
		return fmt.Errorf("undinfe sql name: %s", name)
	}

	return err
}

func (store *StoreSql) getServerName() (string, error) {
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

func (store *StoreSql) initStmt(sqls []string) (err error) {
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

func (store *StoreSql) Insert(key string) (err error) {
	_, err = store.stmtInsert.Exec(key, make(map[string]interface{}))
	return
}

func (store *StoreSql) Delete(key string) (err error) {
	_, err = store.stmtDelete.Exec(key)
	return
}

func (store *StoreSql) Update(key string, val map[string]interface{}) error {
	var data bytes.Buffer
	err := gob.NewEncoder(&data).Encode(val)
	if err != nil {
		return err
	}
	_, err = store.stmtUpdate.Exec(data.Bytes(), key)
	return err
}

func (store *StoreSql) Select(key string) (map[string]interface{}, error) {
	var (
		data    []byte
		expires time.Time
	)
	err := store.stmtSelect.QueryRow(key).Scan(&data, &expires)
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

func (store *StoreSql) Clean(expires time.Time) error {
	_, err := store.stmtClean.Exec(expires)
	return err
}
