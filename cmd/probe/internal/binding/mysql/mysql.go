/*
Copyright (C) 2022-2023 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package mysql

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/kit/logger"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	. "github.com/apecloud/kubeblocks/cmd/probe/internal/binding"
	"github.com/apecloud/kubeblocks/cmd/probe/internal/component/configuration_store"
	. "github.com/apecloud/kubeblocks/internal/sqlchannel/util"
)

// MysqlOperations represents MySQL output bindings.
type MysqlOperations struct {
	db *sql.DB
	mu sync.Mutex
	BaseOperations
}

var _ BaseInternalOps = &MysqlOperations{}

const (
	// configurations to connect to MySQL, either a data source name represent by URL.
	connectionURLKey = "url"

	// To connect to MySQL running over SSL you have to download a
	// SSL certificate. If this is provided the driver will connect using
	// SSL. If you have disabled SSL you can leave this empty.
	// When the user provides a pem path their connection string must end with
	// &tls=custom
	// The connection string should be in the following format
	// "%s:%s@tcp(%s:3306)/%s?allowNativePasswords=true&tls=custom",'myadmin@mydemoserver', 'yourpassword', 'mydemoserver.mysql.database.azure.com', 'targetdb'.
	pemPathKey = "pemPath"

	// other general settings for DB connections.
	maxIdleConnsKey    = "maxIdleConns"
	maxOpenConnsKey    = "maxOpenConns"
	connMaxLifetimeKey = "connMaxLifetime"
	connMaxIdleTimeKey = "connMaxIdleTime"
)

const (
	superUserPriv = "SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, SHUTDOWN, PROCESS, FILE, REFERENCES, INDEX, ALTER, SHOW DATABASES, SUPER, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, REPLICATION SLAVE, REPLICATION CLIENT, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, CREATE USER, EVENT, TRIGGER, CREATE TABLESPACE, CREATE ROLE, DROP ROLE ON *.*"
	readWritePriv = "SELECT, INSERT, UPDATE, DELETE ON *.*"
	readOnlyRPriv = "SELECT ON *.*"
	noPriv        = "USAGE ON *.*"

	listUserTpl  = "SELECT user AS userName, CASE password_expired WHEN 'N' THEN 'F' ELSE 'T' END as expired FROM mysql.user WHERE host = '%' and user <> 'root' and user not like 'kb%';"
	showGrantTpl = "SHOW GRANTS FOR '%s'@'%%';"
	getUserTpl   = `
	SELECT user AS userName, CASE password_expired WHEN 'N' THEN 'F' ELSE 'T' END as expired
	FROM mysql.user
	WHERE host = '%%' and user <> 'root' and user not like 'kb%%' and user ='%s';"
	`
	createUserTpl         = "CREATE USER '%s'@'%%' IDENTIFIED BY '%s';"
	deleteUserTpl         = "DROP USER IF EXISTS '%s'@'%%';"
	grantTpl              = "GRANT %s TO '%s'@'%%';"
	revokeTpl             = "REVOKE %s FROM '%s'@'%%';"
	listSystemAccountsTpl = "SELECT user AS userName FROM mysql.user WHERE host = '%' and user like 'kb%';"
)

var (
	defaultDBPort = 3306
	dbUser        = "root"
	dbPasswd      = ""
)

// NewMysql returns a new MySQL output binding.
func NewMysql(logger logger.Logger) bindings.OutputBinding {
	return &MysqlOperations{
		BaseOperations: BaseOperations{
			Logger: logger,
			Cs:     configuration_store.NewConfigurationStore(),
		},
	}
}

// Init initializes the MySQL binding.
func (mysqlOps *MysqlOperations) Init(metadata bindings.Metadata) error {
	mysqlOps.BaseOperations.Init(metadata)
	if viper.IsSet("KB_SERVICE_USER") {
		dbUser = viper.GetString("KB_SERVICE_USER")
	}

	if viper.IsSet("KB_SERVICE_PASSWORD") {
		dbPasswd = viper.GetString("KB_SERVICE_PASSWORD")
	}

	mysqlOps.Logger.Debug("Initializing MySQL binding")
	mysqlOps.DBType = "mysql"
	mysqlOps.InitIfNeed = mysqlOps.initIfNeed
	mysqlOps.BaseOperations.GetRole = mysqlOps.GetRole
	mysqlOps.DBPort = mysqlOps.GetRunningPort()
	mysqlOps.RegisterOperation(GetRoleOperation, mysqlOps.GetRoleOps)
	mysqlOps.RegisterOperation(GetLagOperation, mysqlOps.GetLagOps)
	mysqlOps.RegisterOperation(CheckStatusOperation, mysqlOps.CheckStatusOps)
	mysqlOps.RegisterOperation(ExecOperation, mysqlOps.ExecOps)
	mysqlOps.RegisterOperation(QueryOperation, mysqlOps.QueryOps)
	mysqlOps.RegisterOperation(SwitchoverOperation, mysqlOps.SwitchoverOps)
	mysqlOps.RegisterOperation(FailoverOperation, mysqlOps.FailoverOps)

	mysqlOps.RegisterOperation(GetTestOperation, mysqlOps.GetTest)

	// following are ops for account management
	mysqlOps.RegisterOperation(ListUsersOp, mysqlOps.listUsersOps)
	mysqlOps.RegisterOperation(CreateUserOp, mysqlOps.createUserOps)
	mysqlOps.RegisterOperation(DeleteUserOp, mysqlOps.deleteUserOps)
	mysqlOps.RegisterOperation(DescribeUserOp, mysqlOps.describeUserOps)
	mysqlOps.RegisterOperation(GrantUserRoleOp, mysqlOps.grantUserRoleOps)
	mysqlOps.RegisterOperation(RevokeUserRoleOp, mysqlOps.revokeUserRoleOps)
	mysqlOps.RegisterOperation(ListSystemAccountsOp, mysqlOps.listSystemAccountsOps)
	return nil
}

func (mysqlOps *MysqlOperations) initIfNeed() bool {
	if mysqlOps.db == nil {
		go func() {
			err := mysqlOps.InitDelay()
			if err != nil {
				mysqlOps.Logger.Errorf("MySQL connection init failed: %v", err)
			} else {
				mysqlOps.Logger.Info("MySQL connection init succeeded.")
			}
		}()
		return true
	}
	return false
}

func (mysqlOps *MysqlOperations) InitDelay() error {
	mysqlOps.mu.Lock()
	defer mysqlOps.mu.Unlock()
	if mysqlOps.db != nil {
		return nil
	}

	p := mysqlOps.Metadata.Properties
	url, ok := p[connectionURLKey]
	if !ok || url == "" {
		return fmt.Errorf("missing MySQL connection string")
	}

	db, err := initDB(url, mysqlOps.Metadata.Properties[pemPathKey])
	if err != nil {
		return err
	}

	err = propertyToInt(p, maxIdleConnsKey, db.SetMaxIdleConns)
	if err != nil {
		return err
	}

	err = propertyToInt(p, maxOpenConnsKey, db.SetMaxOpenConns)
	if err != nil {
		return err
	}

	err = propertyToDuration(p, connMaxIdleTimeKey, db.SetConnMaxIdleTime)
	if err != nil {
		return err
	}

	err = propertyToDuration(p, connMaxLifetimeKey, db.SetConnMaxLifetime)
	if err != nil {
		return err
	}

	// test if db is ready to connect or not
	err = db.Ping()
	if err != nil {
		mysqlOps.Logger.Infof("unable to ping the DB")
		return errors.Wrap(err, "unable to ping the DB")
	}
	mysqlOps.db = db

	return nil
}

func (mysqlOps *MysqlOperations) GetRunningPort() int {
	p := mysqlOps.Metadata.Properties
	url, ok := p[connectionURLKey]
	if !ok || url == "" {
		return defaultDBPort
	}

	config, err := mysql.ParseDSN(url)
	if err != nil {
		return defaultDBPort
	}
	index := strings.LastIndex(config.Addr, ":")
	if index < 0 {
		return defaultDBPort
	}
	port, err := strconv.Atoi(config.Addr[index+1:])
	if err != nil {
		return defaultDBPort
	}

	return port
}

/*
func (mysqlOps *MysqlOperations) GetRole(ctx context.Context, request *bindings.InvokeRequest, response *bindings.InvokeResponse) (string, error) {
	sql := "select CURRENT_LEADER, ROLE, SERVER_ID  from information_schema.wesql_cluster_local"

	rows, err := mysqlOps.db.QueryContext(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Infof("error executing %s: %v", sql, err)
		return "", errors.Wrapf(err, "error executing %s", sql)
	}

	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()

	var curLeader string
	var role string
	var serverID string
	var isReady bool
	for rows.Next() {
		if err = rows.Scan(&curLeader, &role, &serverID); err != nil {
			mysqlOps.Logger.Errorf("Role query error: %v", err)
			return role, err
		}
		isReady = true
	}
	if isReady {
		return role, nil
	}
	return "", errors.Errorf("exec sql %s failed: no data returned", sql)
}
*/

func (mysqlOps *MysqlOperations) GetRole(ctx context.Context, request *bindings.InvokeRequest, response *bindings.InvokeResponse) (string, error) {
	sql := "show slave hosts"
	data, err := mysqlOps.query(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Infof("error executing %s: %v", sql, err)
		return "", errors.Wrapf(err, "error executing %s", sql)
	}

	if string(data) != "null" {
		return PRIMARY, nil
	} else {
		return SECONDARY, nil
	}
}

func (mysqlOps *MysqlOperations) ExecOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	result := OpsResult{}
	sql, ok := req.Metadata["sql"]
	if !ok || sql == "" {
		result["event"] = "ExecFailed"
		result["message"] = ErrNoSQL
		return result, nil
	}
	count, err := mysqlOps.exec(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Infof("exec error: %v", err)
		result["event"] = OperationFailed
		result["message"] = err.Error()
	} else {
		result["event"] = OperationSuccess
		result["count"] = count
	}
	return result, nil
}

func (mysqlOps *MysqlOperations) GetLagOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	result := OpsResult{}
	slaveStatus := make([]SlaveStatus, 0)
	var err error

	if mysqlOps.OriRole == "" {
		mysqlOps.OriRole, err = mysqlOps.GetRole(ctx, req, resp)
		if err != nil {
			result["event"] = OperationFailed
			result["message"] = err.Error()
			return result, nil
		}
	}
	if mysqlOps.OriRole == LEADER {
		result["event"] = OperationSuccess
		result["lag"] = 0
		result["message"] = "This is leader instance, leader has no lag"
		return result, nil
	}

	sql := "show slave status"
	data, err := mysqlOps.query(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Infof("GetLagOps error: %v", err)
		result["event"] = OperationFailed
		result["message"] = err.Error()
	} else {
		err = json.Unmarshal(data, &slaveStatus)
		if err != nil {
			result["event"] = OperationFailed
			result["message"] = err.Error()
		} else {
			result["event"] = OperationSuccess
			result["lag"] = slaveStatus[0].SecondsBehindMaster
		}
	}
	return result, nil
}

func (mysqlOps *MysqlOperations) QueryOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	result := OpsResult{}
	sql, ok := req.Metadata["sql"]
	if !ok || sql == "" {
		result["event"] = OperationFailed
		result["message"] = "no sql provided"
		return result, nil
	}
	data, err := mysqlOps.query(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Infof("Query error: %v", err)
		result["event"] = OperationFailed
		result["message"] = err.Error()
	} else {
		result["event"] = OperationSuccess
		result["message"] = string(data)
	}
	return result, nil
}

func (mysqlOps *MysqlOperations) CheckStatusOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	rwSQL := fmt.Sprintf(`begin;
	create table if not exists kb_health_check(type int, check_ts bigint, primary key(type));
	insert into kb_health_check values(%d, now()) on duplicate key update check_ts = now();
	commit;
	select check_ts from kb_health_check where type=%d limit 1;`, CheckStatusType, CheckStatusType)
	roSQL := fmt.Sprintf(`select check_ts from kb_health_check where type=%d limit 1;`, CheckStatusType)
	var err error
	var data []byte
	switch mysqlOps.DBRoles[strings.ToLower(mysqlOps.OriRole)] {
	case ReadWrite:
		var count int64
		count, err = mysqlOps.exec(ctx, rwSQL)
		data = []byte(strconv.FormatInt(count, 10))
	case Readonly:
		data, err = mysqlOps.query(ctx, roSQL)
	default:
		msg := fmt.Sprintf("unknown access mode for role %s: %v", mysqlOps.OriRole, mysqlOps.DBRoles)
		mysqlOps.Logger.Info(msg)
		data = []byte(msg)
	}

	result := OpsResult{}
	if err != nil {
		mysqlOps.Logger.Infof("CheckStatus error: %v", err)
		result["event"] = OperationFailed
		result["message"] = err.Error()
		if mysqlOps.CheckStatusFailedCount%mysqlOps.FailedEventReportFrequency == 0 {
			mysqlOps.Logger.Infof("status check failed %v times continuously", mysqlOps.CheckStatusFailedCount)
			resp.Metadata[StatusCode] = OperationFailedHTTPCode
		}
		mysqlOps.CheckStatusFailedCount++
	} else {
		result["event"] = OperationSuccess
		result["message"] = string(data)
		mysqlOps.CheckStatusFailedCount = 0
	}
	return result, nil
}

func propertyToInt(props map[string]string, key string, setter func(int)) error {
	if v, ok := props[key]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			setter(i)
		} else {
			return errors.Wrapf(err, "error converting %s:%s to int", key, v)
		}
	}

	return nil
}

func propertyToDuration(props map[string]string, key string, setter func(time.Duration)) error {
	if v, ok := props[key]; ok {
		if d, err := time.ParseDuration(v); err == nil {
			setter(d)
		} else {
			return errors.Wrapf(err, "error converting %s:%s to time duration", key, v)
		}
	}

	return nil
}

func initDB(url, pemPath string) (*sql.DB, error) {
	config, err := mysql.ParseDSN(url)
	if err != nil {
		return nil, errors.Wrapf(err, "illegal Data Source Name (DNS) specified by %s", connectionURLKey)
	}
	config.User = dbUser
	config.Passwd = dbPasswd

	if pemPath != "" {
		rootCertPool := x509.NewCertPool()
		pem, err := os.ReadFile(pemPath)
		if err != nil {
			return nil, errors.Wrapf(err, "Error reading PEM file from %s", pemPath)
		}

		ok := rootCertPool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf("failed to append PEM")
		}

		err = mysql.RegisterTLSConfig("custom", &tls.Config{RootCAs: rootCertPool, MinVersion: tls.VersionTLS12})
		if err != nil {
			return nil, errors.Wrap(err, "Error register TLS config")
		}
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, errors.Wrap(err, "error opening DB connection")
	}

	return db, nil
}

func (mysqlOps *MysqlOperations) query(ctx context.Context, sql string) ([]byte, error) {
	mysqlOps.Logger.Debugf("query: %s", sql)
	rows, err := mysqlOps.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, errors.Wrapf(err, "error executing %s", sql)
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	result, err := mysqlOps.jsonify(rows)
	if err != nil {
		return nil, errors.Wrapf(err, "error marshalling query result for %s", sql)
	}
	return result, nil
}

func (mysqlOps *MysqlOperations) exec(ctx context.Context, sql string) (int64, error) {
	mysqlOps.Logger.Debugf("exec: %s", sql)
	res, err := mysqlOps.db.ExecContext(ctx, sql)
	if err != nil {
		return 0, errors.Wrapf(err, "error executing %s", sql)
	}
	return res.RowsAffected()
}

func (mysqlOps *MysqlOperations) jsonify(rows *sql.Rows) ([]byte, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var ret []interface{}
	for rows.Next() {
		values := prepareValues(columnTypes)
		err := rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		r := mysqlOps.convert(columnTypes, values)
		ret = append(ret, r)
	}
	return json.Marshal(ret)
}

func prepareValues(columnTypes []*sql.ColumnType) []interface{} {
	types := make([]reflect.Type, len(columnTypes))
	for i, tp := range columnTypes {
		types[i] = tp.ScanType()
	}
	values := make([]interface{}, len(columnTypes))
	for i := range values {
		switch types[i].Kind() {
		case reflect.String, reflect.Interface:
			values[i] = &sql.NullString{}
		case reflect.Bool:
			values[i] = &sql.NullBool{}
		case reflect.Float64:
			values[i] = &sql.NullFloat64{}
		case reflect.Int16, reflect.Uint16:
			values[i] = &sql.NullInt16{}
		case reflect.Int32, reflect.Uint32:
			values[i] = &sql.NullInt32{}
		case reflect.Int64, reflect.Uint64:
			values[i] = &sql.NullInt64{}
		default:
			values[i] = reflect.New(types[i]).Interface()
		}
	}
	return values
}

func (mysqlOps *MysqlOperations) convert(columnTypes []*sql.ColumnType, values []interface{}) map[string]interface{} {
	r := map[string]interface{}{}
	for i, ct := range columnTypes {
		value := values[i]
		switch v := values[i].(type) {
		case driver.Valuer:
			if vv, err := v.Value(); err == nil {
				value = interface{}(vv)
			} else {
				mysqlOps.Logger.Warnf("error to convert value: %v", err)
			}
		case *sql.RawBytes:
			// special case for sql.RawBytes, see https://github.com/go-sql-driver/mysql/blob/master/fields.go#L178
			switch ct.DatabaseTypeName() {
			case "VARCHAR", "CHAR", "TEXT", "LONGTEXT":
				value = string(*v)
			}
		}
		if value != nil {
			r[ct.Name()] = value
		}
	}
	return r
}

// InternalQuery is used for internal query, implements BaseInternalOps interface
func (mysqlOps *MysqlOperations) InternalQuery(ctx context.Context, sql string) ([]byte, error) {
	return mysqlOps.query(ctx, sql)
}

// InternalExec is used for internal execution, implements BaseInternalOps interface
func (mysqlOps *MysqlOperations) InternalExec(ctx context.Context, sql string) (int64, error) {
	return mysqlOps.exec(ctx, sql)
}

// GetLogger is used for getting logger, implements BaseInternalOps interface
func (mysqlOps *MysqlOperations) GetLogger() logger.Logger {
	return mysqlOps.Logger
}

func (mysqlOps *MysqlOperations) listUsersOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	sqlTplRend := func(user UserInfo) string {
		return listUserTpl
	}

	return QueryObject(ctx, mysqlOps, req, ListUsersOp, sqlTplRend, nil, UserInfo{})
}

func (mysqlOps *MysqlOperations) listSystemAccountsOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	sqlTplRend := func(user UserInfo) string {
		return listSystemAccountsTpl
	}
	dataProcessor := func(data interface{}) (interface{}, error) {
		var users []UserInfo
		if err := json.Unmarshal(data.([]byte), &users); err != nil {
			return nil, err
		}
		userNames := make([]string, 0)
		for _, user := range users {
			userNames = append(userNames, user.UserName)
		}
		if jsonData, err := json.Marshal(userNames); err != nil {
			return nil, err
		} else {
			return string(jsonData), nil
		}
	}
	return QueryObject(ctx, mysqlOps, req, ListSystemAccountsOp, sqlTplRend, dataProcessor, UserInfo{})
}

func (mysqlOps *MysqlOperations) describeUserOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	var (
		object = UserInfo{}

		// get user grants
		sqlTplRend = func(user UserInfo) string {
			return fmt.Sprintf(showGrantTpl, user.UserName)
		}

		dataProcessor = func(data interface{}) (interface{}, error) {
			roles := make([]map[string]string, 0)
			err := json.Unmarshal(data.([]byte), &roles)
			if err != nil {
				return nil, err
			}
			user := UserInfo{}
			// only keep one role name of the highest privilege
			userRoles := make([]RoleType, 0)
			for _, roleMap := range roles {
				for k, v := range roleMap {
					if len(user.UserName) == 0 {
						user.UserName = strings.TrimPrefix(strings.TrimSuffix(k, "@%"), "Grants for ")
					}
					mysqlRoleType := mysqlOps.priv2Role(strings.TrimPrefix(v, "GRANT "))
					userRoles = append(userRoles, mysqlRoleType)
				}
			}
			// sort roles by weight
			slices.SortFunc(userRoles, SortRoleByWeight)
			if len(userRoles) > 0 {
				user.RoleName = (string)(userRoles[0])
			}
			if jsonData, err := json.Marshal([]UserInfo{user}); err != nil {
				return nil, err
			} else {
				return string(jsonData), nil
			}
		}
	)

	if err := ParseObjFromRequest(req, DefaultUserInfoParser, UserNameValidator, &object); err != nil {
		result := OpsResult{}
		result[RespTypEve] = RespEveFail
		result[RespTypMsg] = err.Error()
		return result, nil
	}

	return QueryObject(ctx, mysqlOps, req, DescribeUserOp, sqlTplRend, dataProcessor, object)
}

func (mysqlOps *MysqlOperations) createUserOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	var (
		object = UserInfo{}

		sqlTplRend = func(user UserInfo) string {
			return fmt.Sprintf(createUserTpl, user.UserName, user.Password)
		}

		msgTplRend = func(user UserInfo) string {
			return fmt.Sprintf("created user: %s, with password: %s", user.UserName, user.Password)
		}
	)

	if err := ParseObjFromRequest(req, DefaultUserInfoParser, UserNameAndPasswdValidator, &object); err != nil {
		result := OpsResult{}
		result[RespTypEve] = RespEveFail
		result[RespTypMsg] = err.Error()
		return result, nil
	}

	return ExecuteObject(ctx, mysqlOps, req, CreateUserOp, sqlTplRend, msgTplRend, object)
}

func (mysqlOps *MysqlOperations) deleteUserOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	var (
		object  = UserInfo{}
		validFn = func(user UserInfo) error {
			if len(user.UserName) == 0 {
				return ErrNoUserName
			}
			return nil
		}
		sqlTplRend = func(user UserInfo) string {
			return fmt.Sprintf(deleteUserTpl, user.UserName)
		}
		msgTplRend = func(user UserInfo) string {
			return fmt.Sprintf("deleted user: %s", user.UserName)
		}
	)
	if err := ParseObjFromRequest(req, DefaultUserInfoParser, validFn, &object); err != nil {
		result := OpsResult{}
		result[RespTypEve] = RespEveFail
		result[RespTypMsg] = err.Error()
		return result, nil
	}

	return ExecuteObject(ctx, mysqlOps, req, DeleteUserOp, sqlTplRend, msgTplRend, object)
}

func (mysqlOps *MysqlOperations) grantUserRoleOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	var (
		succMsgTpl = "role %s granted to user: %s"
	)
	return mysqlOps.managePrivillege(ctx, req, GrantUserRoleOp, grantTpl, succMsgTpl)
}

func (mysqlOps *MysqlOperations) revokeUserRoleOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	var (
		succMsgTpl = "role %s revoked from user: %s"
	)
	return mysqlOps.managePrivillege(ctx, req, RevokeUserRoleOp, revokeTpl, succMsgTpl)
}

func (mysqlOps *MysqlOperations) managePrivillege(ctx context.Context, req *bindings.InvokeRequest, op bindings.OperationKind, sqlTpl string, succMsgTpl string) (OpsResult, error) {
	var (
		object     = UserInfo{}
		sqlTplRend = func(user UserInfo) string {
			// render sql stmts
			roleDesc, _ := mysqlOps.role2Priv(user.RoleName)
			// update privilege
			sql := fmt.Sprintf(sqlTpl, roleDesc, user.UserName)
			return sql
		}
		msgTplRend = func(user UserInfo) string {
			return fmt.Sprintf(succMsgTpl, user.RoleName, user.UserName)
		}
	)
	if err := ParseObjFromRequest(req, DefaultUserInfoParser, UserNameAndRoleValidator, &object); err != nil {
		result := OpsResult{}
		result[RespTypEve] = RespEveFail
		result[RespTypMsg] = err.Error()
		return result, nil
	}
	return ExecuteObject(ctx, mysqlOps, req, op, sqlTplRend, msgTplRend, object)
}

func (mysqlOps *MysqlOperations) role2Priv(roleName string) (string, error) {
	roleType := String2RoleType(roleName)
	switch roleType {
	case SuperUserRole:
		return superUserPriv, nil
	case ReadWriteRole:
		return readWritePriv, nil
	case ReadOnlyRole:
		return readOnlyRPriv, nil
	}
	return "", fmt.Errorf("role name: %s is not supported", roleName)
}

func (mysqlOps *MysqlOperations) priv2Role(priv string) RoleType {
	if strings.HasPrefix(priv, readOnlyRPriv) {
		return ReadOnlyRole
	}
	if strings.HasPrefix(priv, readWritePriv) {
		return ReadWriteRole
	}
	if strings.HasPrefix(priv, superUserPriv) {
		return SuperUserRole
	}
	if strings.HasPrefix(priv, noPriv) {
		return NoPrivileges
	}
	return CustomizedRole
}

func (mysqlOps *MysqlOperations) Promote(ctx context.Context, podName string) error {
	stopReadOnly := `set global read_only=off;`
	stopSlave := `stop slave;`
	resp, err := mysqlOps.exec(ctx, stopReadOnly+stopSlave)
	if err != nil {
		mysqlOps.Logger.Errorf("promote err: %v", err)
		return err
	}
	mysqlOps.Logger.Infof("promote success, resp:%v", resp)

	return nil
}

func (mysqlOps *MysqlOperations) Demote(ctx context.Context, podName string) error {
	setReadOnly := `set global read_only=on;`

	_, err := mysqlOps.exec(ctx, setReadOnly)
	if err != nil {
		mysqlOps.Logger.Errorf("promote err: %v", err)
		return err
	}

	opTime, _ := mysqlOps.GetOpTime(ctx)
	err = mysqlOps.Cs.DeleteLeader(opTime)

	// Give a time to somebody to take the leader lock
	time.Sleep(time.Second * 5)
	//TODO:Follow which
	for i := 0; i < 3; i++ {
		err = mysqlOps.Cs.GetClusterFromKubernetes()
		if err == nil && mysqlOps.Cs.GetCluster().Leader != nil {
			break
		}
		time.Sleep(time.Second * 2)

		if i == 2 {
			return errors.Errorf("get no leader now")
		}
	}

	return mysqlOps.follow(ctx, podName, mysqlOps.Cs.GetCluster().Leader.GetMember().GetName())
}

func (mysqlOps *MysqlOperations) GetExtra(ctx context.Context) (map[string]string, error) {
	return nil, nil
}

func (mysqlOps *MysqlOperations) GetOpTime(ctx context.Context) (int64, error) {
	sql := `show status like '%Innodb_redo_log_current_lsn%';`
	data, err := mysqlOps.query(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Errorf("get engine innodb status failed, err:%v", err)
		return 0, err
	}

	result, err := ParseQuery(string(data))
	if err != nil {
		return 0, err
	}

	currentLsn, err := strconv.ParseInt(result["Value"].(string), 10, 64)
	if err != nil {
		return 0, err
	}

	return currentLsn, nil
}

func (mysqlOps *MysqlOperations) IsRunning(ctx context.Context, podName string) bool {
	// TODO:后期直接用进程判断
	return mysqlOps.db.Ping() == nil
}

func (mysqlOps *MysqlOperations) IsHealthiest(ctx context.Context, podName string) bool {
	err := mysqlOps.Cs.GetClusterFromKubernetes()
	if err != nil {
		mysqlOps.Logger.Errorf("get cluster from k8s failed, err:%v", err)
	}

	//这部分引擎相关可以扩充
	lastLsn, _ := mysqlOps.GetOpTime(ctx)
	if mysqlOps.isLagging(lastLsn) {
		mysqlOps.Logger.Infof("my lag exceeds max lag")
		return false
	}

	/*
		requestBody := `{"operation":"getRole"}`
		for _, m := range mysqlOps.Cs.GetCluster().Members {
			resp, err := mysqlOps.FetchOtherStatus(m.GetData().GetUrl(), requestBody)
			if err != nil {
				mysqlOps.Logger.Errorf("fetch other status err:%v", err)
				return false
			}
			if resp["role"] == PRIMARY {
				mysqlOps.Logger.Errorf("Primary %s is still alive")
				return false
			}
		}
	*/

	return true
}

func (mysqlOps *MysqlOperations) isLagging(lastLsn int64) bool {
	lag := mysqlOps.Cs.GetCluster().GetOpTime() - lastLsn
	return lag > mysqlOps.Cs.GetCluster().Config.GetData().GetMaxLagOnSwitchover()
}

func (mysqlOps *MysqlOperations) HandleFollow(ctx context.Context, leader *configuration_store.Leader, podName string) error {
	needChange := mysqlOps.checkRecoveryConf(ctx, leader.GetMember().GetName())
	if needChange {
		return mysqlOps.follow(ctx, podName, leader.GetMember().GetName())
	}

	mysqlOps.Logger.Infof("no action coz i am still follow the leader")
	return nil
}

func (mysqlOps *MysqlOperations) checkRecoveryConf(ctx context.Context, leader string) bool {
	sql := "show slave status"
	data, err := mysqlOps.query(ctx, sql)
	if err != nil {
		mysqlOps.Logger.Errorf("error executing %s: %v", sql, err)
		return true
	}

	result, err := ParseQuery(string(data))
	if err != nil {
		mysqlOps.Logger.Errorf("parse query err:%v", err)
		return true
	}

	if result == nil || strings.Split(result["Master_Host"].(string), ".")[0] != leader {
		return true
	}

	return false
}

func (mysqlOps *MysqlOperations) follow(ctx context.Context, podName string, leader string) error {
	stopSlave := `stop slave;`
	changeMaster := fmt.Sprintf(`change master to master_host='%s.%s-headless',master_user='%s',master_password='%s',master_port=%s,master_auto_position=1;`,
		leader, mysqlOps.Cs.GetClusterCompName(), os.Getenv("KB_SERVICE_USER"), os.Getenv("KB_SERVICE_PASSWORD"), os.Getenv("KB_SERVICE_PORT"))
	startSlave := `start slave;`

	_, err := mysqlOps.exec(ctx, stopSlave+changeMaster+startSlave)
	if err != nil {
		mysqlOps.Logger.Errorf("sql query failed, err:%v", err)
	}

	return nil
}

func (mysqlOps *MysqlOperations) EnforcePrimaryRole(ctx context.Context, podName string) error {
	if isLeader, err := mysqlOps.IsLeader(ctx); isLeader && err == nil {
		return nil
	} else {
		return mysqlOps.Promote(ctx, podName)
	}
}

func (mysqlOps *MysqlOperations) ProcessManualSwitchoverFromLeader(ctx context.Context, podName string) (bool, error) {
	err := mysqlOps.Cs.GetClusterFromKubernetes()
	if err != nil {
		mysqlOps.Logger.Errorf("get cluster from k8s failed, err:%v", err)
		return false, err
	}

	switchover := mysqlOps.Cs.GetCluster().Switchover
	if switchover == nil {
		return false, nil
	}

	leader := switchover.GetLeader()
	candidate := switchover.GetCandidate()
	if leader == "" || leader == podName {
		if candidate == "" || candidate != podName {
			var members []string
			for _, m := range mysqlOps.Cs.GetCluster().Members {
				if switchover.GetCandidate() == "" || m.GetName() == candidate {
					members = append(members, m.GetName())
				}
			}

			if mysqlOps.isFailoverPossible(members) {
				return true, mysqlOps.Demote(ctx, podName)
			}
		} else {
			mysqlOps.Logger.Warnf("manual failover: I am already the leader, no need to failover")
		}
	} else {
		mysqlOps.Logger.Warnf("manual switchover, leader name does not match, %s != %s", switchover.GetLeader(), podName)
	}

	return false, mysqlOps.Cs.DeleteConfigMap(mysqlOps.Cs.GetClusterCompName() + configuration_store.SwitchoverSuffix)
}

func (mysqlOps *MysqlOperations) isFailoverPossible(members []string) bool {
	return true
}

func (mysqlOps *MysqlOperations) ProcessManualSwitchoverFromNoLeader(ctx context.Context, podName string) bool {
	err := mysqlOps.Cs.GetClusterFromKubernetes()
	if err != nil {
		mysqlOps.Logger.Errorf("get cluster from k8s err:%v", err)
		return false
	}

	switchover := mysqlOps.Cs.GetCluster().Switchover
	if switchover != nil && switchover.GetCandidate() != "" {
		if switchover.GetCandidate() == podName {
			return true
		}
		member := mysqlOps.Cs.GetCluster().GetMemberWithName(switchover.GetCandidate())
		requestBody := `{"operation":"checkStatus"}`
		resp, err := mysqlOps.FetchOtherStatus(member.GetData().GetUrl(), requestBody)
		if err != nil {
			mysqlOps.Logger.Errorf("fetch other status err:%v", err)
		}
		if resp["event"] == OperationSuccess {
			return false
		}
	}

	return mysqlOps.IsHealthiest(ctx, podName)
}

func (mysqlOps *MysqlOperations) Start(ctx context.Context, podName string) error {
	start := `./entrypoint.sh mysqld --defaults-file=/opt/my.cnf &`

	_, err := mysqlOps.Cs.ExecCmdWithPod(ctx, podName, start, mysqlOps.DBType)
	if err != nil {
		mysqlOps.Logger.Errorf("start err: %v", err)
		return err
	}

	mysqlOps.OriRole = SECONDARY
	_ = mysqlOps.InitDelay()

	return nil
}

func (mysqlOps *MysqlOperations) GetSysID(ctx context.Context) (string, error) {
	return "", nil
}

func (mysqlOps *MysqlOperations) GetTest(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	result := OpsResult{}
	_, _ = mysqlOps.GetRole(ctx, req, resp)

	return result, nil
}

func (mysqlOps *MysqlOperations) Stop(ctx context.Context) error {
	return nil
}

func (mysqlOps *MysqlOperations) SwitchoverOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	result := OpsResult{}
	primary, _ := req.Metadata[PRIMARY]
	candidate, _ := req.Metadata[CANDIDATE]
	operation := req.Operation

	err := mysqlOps.Cs.GetClusterFromKubernetes()
	if err != nil {
		mysqlOps.Logger.Errorf("get cluster err:%v", err)
		result["event"] = OperationFailed
		result["message"] = err.Error()
		return result, nil
	}

	if can, err := mysqlOps.isSwitchoverPossible(ctx, primary, candidate, operation); !can || err != nil {
		result["event"] = OperationFailed
		result["message"] = err.Error()
		return result, nil
	}

	if err := mysqlOps.ManualSwitchover(primary, candidate); err != nil {
		result["event"] = OperationFailed
		result["message"] = err.Error()
		return result, nil
	}

	result["event"] = OperationSuccess
	if candidate == "" {
		result["message"] = "Successfully switchover to a healthiest node"
	} else {
		result["message"] = fmt.Sprintf("Successfully %s to: %s", string(req.Operation), candidate)
	}

	return result, nil
}

func (mysqlOps *MysqlOperations) isSwitchoverPossible(ctx context.Context, primary string, candidate string, operation bindings.OperationKind) (bool, error) {
	if operation == FailoverOperation && candidate == "" {
		return false, errors.Errorf("failover need a candidate")
	} else if operation == SwitchoverOperation && primary == "" {
		return false, errors.Errorf("switchover need a primary")
	}
	if primary != "" && (mysqlOps.Cs.GetCluster().Leader == nil || mysqlOps.Cs.GetCluster().Leader.GetMember().GetName() != primary) {
		return false, errors.Errorf("leader name does not match")
	}

	if candidate != "" {
		if mysqlOps.Cs.GetCluster().Leader != nil && mysqlOps.Cs.GetCluster().Leader.GetMember().GetName() != primary {
			return false, errors.Errorf("leader name does not match ,leader name: %s, primary: %s", mysqlOps.Cs.GetCluster().Leader.GetMember().GetName(), primary)
		}

		if !mysqlOps.Cs.GetCluster().HasMember(candidate) {
			return false, errors.Errorf("candidate does not exist")
		}
	} else {
		hasMemberExceptLeader := false
		for _, member := range mysqlOps.Cs.GetCluster().Members {
			if member.GetName() != mysqlOps.Cs.GetCluster().Leader.GetMember().GetName() {
				hasMemberExceptLeader = true
			}
		}
		if !hasMemberExceptLeader {
			return false, errors.Errorf("cluster does not have member except leader")
		}
	}

	runningMembers := 0
	requestBody := `{"operation":"checkStatus"}`
	for _, m := range mysqlOps.Cs.GetCluster().Members {
		resp, err := mysqlOps.FetchOtherStatus(m.GetData().GetUrl(), requestBody)
		if err != nil {
			mysqlOps.Logger.Errorf("fetch other member failed, err:%v", err)
		}
		mysqlOps.Logger.Infof("other status:%v", resp)
		if resp["event"] == OperationSuccess {
			runningMembers++
		}
	}
	if runningMembers == 0 {
		return false, errors.Errorf("no running candidates have been found")
	}

	return true, nil
}

func (mysqlOps *MysqlOperations) FailoverOps(ctx context.Context, req *bindings.InvokeRequest, resp *bindings.InvokeResponse) (OpsResult, error) {
	return mysqlOps.SwitchoverOps(ctx, req, resp)
}
