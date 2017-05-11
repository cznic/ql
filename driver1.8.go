// +build go1.8

package ql

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"regexp"
)

const prefix = "$"

func (c *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.Exec(replaceNamed(query, args))
}

func replaceNamed(query string, args []driver.NamedValue) (string, []driver.Value) {
	a := make([]driver.Value, len(args))
	for k, v := range args {
		if v.Name != "" {
			query = strings.Replace(query, prefix+v.Name, fmt.Sprintf("%s%d", prefix, v.Ordinal), -1)
		}
		a[k] = v.Value
	}
	return query, a
}

func (c *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.Query(replaceNamed(query, args))
}

func (c *driverConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return c.Prepare(filterNamedArgs(query))
}

var re = regexp.MustCompile(`^\w+`)

func filterNamedArgs(q string) string {
	c := strings.Count(q, prefix)
	if c == 0 || c == len(q) {
		return q
	}
	pc := strings.Split(q, prefix)
	for k, v := range pc {
		if k == 0 {
			continue
		}
		if v != "" {
			pc[k] = re.ReplaceAllString(v, fmt.Sprint(k))
		}
	}
	return strings.Join(pc, prefix)
}

func (s *driverStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	a := make([]driver.Value, len(args))
	for k, v := range args {
		a[k] = v.Value
	}
	return s.Exec(a)
}

func (s *driverStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	a := make([]driver.Value, len(args))
	for k, v := range args {
		a[k] = v.Value
	}
	return s.Query(a)
}
