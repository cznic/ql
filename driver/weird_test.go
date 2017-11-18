package driver

import (
	"database/sql"
	"testing"
)

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

// Both of the UPDATEs _should_ work but the 2nd one results in a _type missmatch_ error at the time of writing.
func TestArgumentOrder(t *testing.T) {
	db, err := sql.Open("ql-mem", "mem.test")
	if err != nil {
		t.Fatal(err)
	}

	// prepare db
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(dbPositionCreateTable)
	if err != nil {
		t.Fatal(err)
	}
	insStmt, err := tx.Prepare(dbPositionInsert)
	if err != nil {
		t.Fatal(err)
	}
	defer insStmt.Close()
	res, err := insStmt.Exec(0.1, 0.2, 0.3, 0.4, "hello ql")
	if err != nil {
		t.Fatal(err)
	}
	pid, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	// run working
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := tx.Prepare(dbPositionUpdate)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	res, err = stmt.Exec(0.01, 0.02, 0.03, 0.04, "hello QL", pid)
	if err != nil {
		t.Fatal(err)
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Logf("affected: %d\n", cnt)
	}

	// confusing
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	stmt, err = tx.Prepare(dbPositionUpdateTypeMissmatch)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	res, err = stmt.Exec(pid, "HELLO ql", 1.05, 2.05, 3.05, 4.05)
	if err != nil {
		t.Fatal(err)
	}
	cnt, err = res.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}
	if cnt != 1 {
		t.Logf("affected: %d\n", cnt)
	}

	if err != nil {
		t.Fatal(err)
	}
}
