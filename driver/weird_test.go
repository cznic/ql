package driver

import (
	"database/sql"
	"testing"
)

// Both of the UPDATEs _should_ work but the 2nd one results in a _type missmatch_ error at the time of writing.
// see https://github.com/cznic/ql/issues/190
func TestIssue190(t *testing.T) {
	db, err := sql.Open("ql-mem", "mem.test")
	if err != nil {
		t.Fatal(err)
	}

	// prepare db
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	const createStmt = `CREATE TABLE issue190 (
	Number float64,
	Comment string
); `
	_, err = tx.Exec(createStmt)
	if err != nil {
		t.Fatal(err)
	}
	const insertStmt = `INSERT INTO issue190 (Number,Comment) VALUES($1,$2);`
	insStmt, err := tx.Prepare(insertStmt)
	if err != nil {
		t.Fatal(err)
	}
	defer insStmt.Close()
	res, err := insStmt.Exec(0.1, "hello ql")
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
	const updateWorks = `
		UPDATE issue190
		SET
			Number = $1,
			Comment = $2
		WHERE id() == $3;`
	stmt, err := tx.Prepare(updateWorks)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	res, err = stmt.Exec(0.01, "hello QL", pid)
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
		t.Errorf("affected: %d\n", cnt)
	}

	// confusing
	tx, err = db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	const updateWithTypeMissmatch = `
		UPDATE issue190
		SET
			Comment = $2,
			Number = $3
		WHERE id() == $1;`
	stmt, err = tx.Prepare(updateWithTypeMissmatch)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	res, err = stmt.Exec(pid, "HELLO ql", 4.05)
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
		t.Errorf("affected: %d\n", cnt)
	}

	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}
