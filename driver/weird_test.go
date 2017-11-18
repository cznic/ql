package driver

import (
	"database/sql"
	"fmt"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		t.Fatal(err)
	}
}

const (
	dbPositionCreateTable = `CREATE TABLE Positions (
	Y float64,
	Z float64,
	Alpha float64,
	Beta float64,
	Comment string
);
CREATE INDEX PositionId on Positions (id());`

	dbPositionUpdate = `
		UPDATE Positions
		SET
			Y = $1, Z = $2,
			Alpha = $3,	Beta = $4,
			Comment = $5
		WHERE id() == $6;`

	dbPositionUpdateTypeMissmatch = `
		UPDATE Positions
		SET
			Comment = $2,
			Y = $3, Z = $4,
			Alpha = $5,	Beta = $6
		WHERE id() == $1;`

	dbPositionInsert = `INSERT INTO Positions (Y,Z,Alpha,Beta,Comment) VALUES($1,$2,$3,$4,$5);`
)

type position struct {
	ID          int64
	Y, Z        float64
	Alpha, Beta float64
	Comment     string
}

func TestArgumentOrder(t *testing.T) {
	db, err := sql.Open("ql-mem", "mem.test")
	check(err, t)

	var pos position
	pos.Y = 0.1
	pos.Z = 0.2
	pos.Alpha = 0.3
	pos.Beta = 0.4
	pos.Comment = "small"

	// prepare db
	tx, err := db.Begin()
	check(err, t)
	_, err = tx.Exec(dbPositionCreateTable)
	check(err, t)
	insStmt, err := tx.Prepare(dbPositionInsert)
	check(err, t)
	defer insStmt.Close()
	res, err := insStmt.Exec(pos.Y, pos.Z, pos.Alpha, pos.Beta, pos.Comment)
	check(err, t)
	pid, err := res.LastInsertId()
	check(err, t)
	pos.ID = pid
	err = tx.Commit()
	check(err, t)

	// run working
	tx, err = db.Begin()
	check(err, t)
	stmt, err := tx.Prepare(dbPositionUpdate)
	check(err, t)
	defer stmt.Close()
	res, err = stmt.Exec(pos.Y, pos.Z, pos.Alpha, pos.Beta, pos.Comment, pos.ID)
	check(err, t)
	cnt, err := res.RowsAffected()
	check(err, t)
	err = tx.Commit()
	check(err, t)
	if cnt != 1 {
		fmt.Printf("affected: %d\n", cnt)
	}

	// confusing
	tx, err = db.Begin()
	check(err, t)
	stmt, err = tx.Prepare(dbPositionUpdateTypeMissmatch)
	check(err, t)
	defer stmt.Close()
	res, err = stmt.Exec(pos.ID, pos.Comment, pos.Y, pos.Z, pos.Alpha, pos.Beta)
	check(err, t)
	cnt, err = res.RowsAffected()
	check(err, t)
	err = tx.Commit()
	check(err, t)
	if cnt != 1 {
		fmt.Printf("affected: %d\n", cnt)
	}

	check(db.Close(), t)
}
