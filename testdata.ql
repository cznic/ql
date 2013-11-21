// Copyright (c) 2013 Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// vi:filetype=sql

-- 0
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	INSERT INTO t VALUES(11, 22, 33);
COMMIT;
SELECT * FROM t;
|lc1, lc2, lc3
[11 22 33]

-- 1
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int);
	CREATE TABLE t (c1 int);
COMMIT;
||table.*exists

-- 2
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c1 int, c4 int);
COMMIT;
||duplicate column

-- 3
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	ALTER TABLE t ADD c4 string;
	INSERT INTO t VALUES (1, 2, 3, "foo");
COMMIT;
SELECT * FROM t;
|lc1, lc2, lc3, sc4
[1 2 3 foo]

-- 4
BEGIN TRANSACTION;
	ALTER TABLE none ADD c1 int;
COMMIT;
||table .* not exist

-- 5
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	ALTER TABLE t ADD c2 int;
COMMIT;
||column .* exists

-- 6
BEGIN TRANSACTION;
	ALTER TABLE none DROP COLUMN c1;
COMMIT;
||table .* not exist

-- 7
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	ALTER TABLE t DROP COLUMN c4;
COMMIT;
||column .* not exist

-- 8
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	ALTER TABLE t DROP COLUMN c2;
	INSERT INTO t VALUES (1, 2);
COMMIT;
SELECT * FROM t;
|lc1, lc3
[1 2]

-- 9
BEGIN TRANSACTION;
	DROP TABLE none;
COMMIT;
||table .* not exist

-- 10
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int);
	DROP TABLE t;
COMMIT;
SELECT * FROM t;
||table .* not exist

-- 11
BEGIN TRANSACTION;
	INSERT INTO none VALUES (1, 2);
COMMIT;
||table .* not exist

-- 12
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int);
	INSERT INTO t VALUES (1);
COMMIT;
||expect

-- 13
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int);
	INSERT INTO t VALUES (1, 2, 3);
COMMIT;
||expect

-- 14
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int);
	INSERT INTO t VALUES (1, 2/(3*5-15));
COMMIT;
||division by zero

-- 15
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int);
	INSERT INTO t VALUES (2+3*4, 2*3+4);
COMMIT;
SELECT * FROM t;
|lc1, lc2
[14 10]

-- 16
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int, c4 int);
	INSERT INTO t (c2, c4) VALUES (1);
COMMIT;
||expect

-- 17
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int, c4 int);
	INSERT INTO t (c2, c4) VALUES (1, 2, 3);
COMMIT;
||expect

-- 18
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int, c4 int);
	INSERT INTO t (c2, none) VALUES (1, 2);
COMMIT;
||unknown

-- 19
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int, c4 int);
	INSERT INTO t (c2, c3) VALUES (2+3*4, 2*3+4);
	INSERT INTO t VALUES (1, 2, 3, 4, );
COMMIT;
SELECT * FROM t;
|lc1, lc2, lc3, lc4
[1 2 3 4]
[<nil> 14 10 <nil>]

-- 20
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 int, c3 int, c4 int);
	ALTER TABLE t DROP COLUMN c3;
	INSERT INTO t (c1, c4) VALUES (42, 314);
	INSERT INTO t (c1, c2) VALUES (2+3*4, 2*3+4);
	INSERT INTO t VALUES (1, 2, 3);
COMMIT;
SELECT * FROM t;
|lc1, lc2, lc4
[1 2 3]
[14 10 <nil>]
[42 <nil> 314]

-- 21
BEGIN TRANSACTION;
	TRUNCATE TABLE none;
COMMIT;
||table .* not exist

-- 22
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int);
	INSERT INTO t VALUES(278);
	TRUNCATE TABLE t;
	INSERT INTO t VALUES(314);
COMMIT;
SELECT * FROM t;
|lc1
[314]

-- 23
SELECT * FROM none;
||table .* not exist

-- 24
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (1, "a");
COMMIT;
SELECT * FROM t;
|lc1, sc2
[1 a]
[2 b]

-- 25
SELECT c1 FROM none;
||table .* not exist

-- 26
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT none FROM t;
||unknown

-- 27
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT c1, none, c2 FROM t;
||unknown

-- 28
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32, c2 string);
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (1, "a");
COMMIT;
SELECT 3*c1 AS v FROM t;
|kv
[3]
[6]

-- 29
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (1, "a");
COMMIT;
SELECT c2 FROM t;
|sc2
[a]
[b]

-- 30
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (1, "a");
COMMIT;
SELECT c1 AS X, c2 FROM t;
|lX, sc2
[1 a]
[2 b]

-- 31
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (1, "a");
COMMIT;
SELECT c2, c1 AS Y FROM t;
|sc2, lY
[a 1]
[b 2]

-- 32
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT * FROM t WHERE c3 == 1;
||unknown

-- 33
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT * FROM t WHERE c1 == 1;
|lc1, sc2
[1 a]

-- 34
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT * FROM t ORDER BY c3;
||unknown

-- 35
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32, c2 string);
	INSERT INTO t VALUES (22, "bc");
	INSERT INTO t VALUES (11, "ab");
	INSERT INTO t VALUES (33, "cd");
COMMIT;
SELECT * FROM t ORDER BY c1;
|kc1, sc2
[11 ab]
[22 bc]
[33 cd]

-- 36
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT * FROM t ORDER BY c1 ASC;
|lc1, sc2
[1 a]
[2 b]

-- 37
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
COMMIT;
SELECT * FROM t ORDER BY c1 DESC;
|lc1, sc2
[2 b]
[1 a]

-- 38
BEGIN TRANSACTION;
CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "b");
	INSERT INTO t VALUES (3, "c");
	INSERT INTO t VALUES (4, "d");
	INSERT INTO t VALUES (5, "e");
	INSERT INTO t VALUES (6, "f");
	INSERT INTO t VALUES (7, "g");
COMMIT;
SELECT * FROM t
WHERE c1 % 2 == 0
ORDER BY c2 DESC;
|lc1, sc2
[6 f]
[4 d]
[2 b]

-- 39
BEGIN TRANSACTION;
CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (2, "a");
	INSERT INTO t VALUES (3, "b");
	INSERT INTO t VALUES (4, "b");
	INSERT INTO t VALUES (5, "c");
	INSERT INTO t VALUES (6, "c");
	INSERT INTO t VALUES (7, "d");
COMMIT;
SELECT * FROM t
ORDER BY c1, c2;
|lc1, sc2
[1 a]
[2 a]
[3 b]
[4 b]
[5 c]
[6 c]
[7 d]

-- 40
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int, c2 string);
	INSERT INTO t VALUES (1, "d");
	INSERT INTO t VALUES (2, "c");
	INSERT INTO t VALUES (3, "c");
	INSERT INTO t VALUES (4, "b");
	INSERT INTO t VALUES (5, "b");
	INSERT INTO t VALUES (6, "a");
	INSERT INTO t VALUES (7, "a");
COMMIT;
SELECT * FROM t
ORDER BY c2, c1
|lc1, sc2
[6 a]
[7 a]
[4 b]
[5 b]
[2 c]
[3 c]
[1 d]

-- S 41
SELECT * FROM employee, none;
||table .* not exist

-- S 42
SELECT employee.LastName FROM employee, none;
||table .* not exist

-- S 43
SELECT none FROM employee, department;
||unknown

-- S 44
SELECT employee.LastName FROM employee, department;
|semployee.LastName
[John]
[John]
[John]
[John]
[Smith]
[Smith]
[Smith]
[Smith]
[Robinson]
[Robinson]
[Robinson]
[Robinson]
[Heisenberg]
[Heisenberg]
[Heisenberg]
[Heisenberg]
[Jones]
[Jones]
[Jones]
[Jones]
[Rafferty]
[Rafferty]
[Rafferty]
[Rafferty]

-- S 45
SELECT * FROM employee, department
ORDER by employee.LastName;
|semployee.LastName, lemployee.DepartmentID, ldepartment.DepartmentID, sdepartment.DepartmentName
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]

-- S 46
SELECT *
FROM employee, department
WHERE employee.DepartmentID == department.DepartmentID;
|semployee.LastName, lemployee.DepartmentID, ldepartment.DepartmentID, sdepartment.DepartmentName
[Smith 34 34 Clerical]
[Robinson 34 34 Clerical]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Rafferty 31 31 Sales]

-- S 47
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE employee.DepartmentID == department.DepartmentID
ORDER BY department.DepartmentName, employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Clerical 34 Robinson 34]
[Clerical 34 Smith 34]
[Engineering 33 Heisenberg 33]
[Engineering 33 Jones 33]
[Sales 31 Rafferty 31]

-- S 48
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE department.DepartmentName IN ("Sales", "Engineering", "HR", "Clerical")
ORDER BY employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Clerical 34 Heisenberg 33]
[Engineering 33 Heisenberg 33]
[Sales 31 Heisenberg 33]
[Clerical 34 John <nil>]
[Engineering 33 John <nil>]
[Sales 31 John <nil>]
[Clerical 34 Jones 33]
[Engineering 33 Jones 33]
[Sales 31 Jones 33]
[Clerical 34 Rafferty 31]
[Engineering 33 Rafferty 31]
[Sales 31 Rafferty 31]
[Clerical 34 Robinson 34]
[Engineering 33 Robinson 34]
[Sales 31 Robinson 34]
[Clerical 34 Smith 34]
[Engineering 33 Smith 34]
[Sales 31 Smith 34]

-- S 49
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE (department.DepartmentID+1000) IN (1031, 1035, 1036)
ORDER BY employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Marketing 35 Heisenberg 33]
[Sales 31 Heisenberg 33]
[Marketing 35 John <nil>]
[Sales 31 John <nil>]
[Marketing 35 Jones 33]
[Sales 31 Jones 33]
[Marketing 35 Rafferty 31]
[Sales 31 Rafferty 31]
[Marketing 35 Robinson 34]
[Sales 31 Robinson 34]
[Marketing 35 Smith 34]
[Sales 31 Smith 34]

-- S 50
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE department.DepartmentName NOT IN ("Engineering", "HR", "Clerical");
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, ?employee.DepartmentID
[Marketing 35 John <nil>]
[Sales 31 John <nil>]
[Marketing 35 Smith 34]
[Sales 31 Smith 34]
[Marketing 35 Robinson 34]
[Sales 31 Robinson 34]
[Marketing 35 Heisenberg 33]
[Sales 31 Heisenberg 33]
[Marketing 35 Jones 33]
[Sales 31 Jones 33]
[Marketing 35 Rafferty 31]
[Sales 31 Rafferty 31]

-- S 51
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE department.DepartmentID BETWEEN 34 AND 36
ORDER BY employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Marketing 35 Heisenberg 33]
[Clerical 34 Heisenberg 33]
[Marketing 35 John <nil>]
[Clerical 34 John <nil>]
[Marketing 35 Jones 33]
[Clerical 34 Jones 33]
[Marketing 35 Rafferty 31]
[Clerical 34 Rafferty 31]
[Marketing 35 Robinson 34]
[Clerical 34 Robinson 34]
[Marketing 35 Smith 34]
[Clerical 34 Smith 34]

-- S 52
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE department.DepartmentID BETWEEN int64(34) AND int64(36)
ORDER BY employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Marketing 35 Heisenberg 33]
[Clerical 34 Heisenberg 33]
[Marketing 35 John <nil>]
[Clerical 34 John <nil>]
[Marketing 35 Jones 33]
[Clerical 34 Jones 33]
[Marketing 35 Rafferty 31]
[Clerical 34 Rafferty 31]
[Marketing 35 Robinson 34]
[Clerical 34 Robinson 34]
[Marketing 35 Smith 34]
[Clerical 34 Smith 34]

-- S 53
SELECT department.DepartmentName, department.DepartmentID, employee.LastName, employee.DepartmentID
FROM employee, department
WHERE department.DepartmentID NOT BETWEEN 33 AND 34
ORDER BY employee.LastName;
|sdepartment.DepartmentName, ldepartment.DepartmentID, semployee.LastName, lemployee.DepartmentID
[Marketing 35 Heisenberg 33]
[Sales 31 Heisenberg 33]
[Marketing 35 John <nil>]
[Sales 31 John <nil>]
[Marketing 35 Jones 33]
[Sales 31 Jones 33]
[Marketing 35 Rafferty 31]
[Sales 31 Rafferty 31]
[Marketing 35 Robinson 34]
[Sales 31 Robinson 34]
[Marketing 35 Smith 34]
[Sales 31 Smith 34]

-- S 54
SELECT LastName, LastName FROM employee;
||duplicate

-- S 55
SELECT LastName+", " AS a, LastName AS a FROM employee;
||duplicate

-- S 56
SELECT LastName AS a, LastName AS b FROM employee
ORDER by a, b;
|sa, sb
[Heisenberg Heisenberg]
[John John]
[Jones Jones]
[Rafferty Rafferty]
[Robinson Robinson]
[Smith Smith]

-- S 57
SELECT employee.LastName AS name, employee.DepartmentID AS id, department.DepartmentName AS department, department.DepartmentID AS id2
FROM employee, department
WHERE employee.DepartmentID == department.DepartmentID
ORDER BY name, id, department, id2;
|sname, lid, sdepartment, lid2
[Heisenberg 33 Engineering 33]
[Jones 33 Engineering 33]
[Rafferty 31 Sales 31]
[Robinson 34 Clerical 34]
[Smith 34 Clerical 34]

-- S 58
SELECT * FROM;
||syntax

-- S 59
SELECT * FROM employee
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 33]
[John <nil>]
[Jones 33]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- S 60
SELECT * FROM employee AS e
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 33]
[John <nil>]
[Jones 33]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- S 61
SELECT none FROM (
	SELECT * FROM employee;
	SELECT * FROM department;
);
||syntax

-- S 62
SELECT none FROM (
	SELECT * FROM employee;
);
||unknown

-- S 63
SELECT noneCol FROM (
	SELECT * FROM noneTab
);
||not exist

-- S 64
SELECT noneCol FROM (
	SELECT * FROM employee
);
||unknown

-- S 65
SELECT * FROM (
	SELECT * FROM employee
)
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 33]
[John <nil>]
[Jones 33]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- S 66
SELECT * FROM (
	SELECT LastName AS Name FROM employee
)
ORDER BY Name;
|sName
[Heisenberg]
[John]
[Jones]
[Rafferty]
[Robinson]
[Smith]

-- S 67
SELECT Name FROM (
	SELECT LastName AS name FROM employee
);
||unknown

-- S 68
SELECT name AS Name FROM (
	SELECT LastName AS name
	FROM employee AS e
)
ORDER BY Name;
|sName
[Heisenberg]
[John]
[Jones]
[Rafferty]
[Robinson]
[Smith]

-- S 69
SELECT name AS Name FROM (
	SELECT LastName AS name FROM employee
)
ORDER BY Name;
|sName
[Heisenberg]
[John]
[Jones]
[Rafferty]
[Robinson]
[Smith]

-- S 70
SELECT employee.LastName, department.DepartmentName, department.DepartmentID FROM (
	SELECT *
	FROM employee, department
	WHERE employee.DepartmentID == department.DepartmentID
)
ORDER BY department.DepartmentName, employee.LastName
|semployee.LastName, sdepartment.DepartmentName, ldepartment.DepartmentID
[Robinson Clerical 34]
[Smith Clerical 34]
[Heisenberg Engineering 33]
[Jones Engineering 33]
[Rafferty Sales 31]

-- S 71
SELECT e.LastName, d.DepartmentName, d.DepartmentID FROM (
	SELECT *
	FROM employee AS e, department AS d
	WHERE e.DepartmentID == d.DepartmentID
)
ORDER by d.DepartmentName, e.LastName;
|se.LastName, sd.DepartmentName, ld.DepartmentID
[Robinson Clerical 34]
[Smith Clerical 34]
[Heisenberg Engineering 33]
[Jones Engineering 33]
[Rafferty Sales 31]

-- S 72
SELECT e.LastName AS name, d.DepartmentName AS department, d.DepartmentID AS id FROM (
	SELECT *
	FROM employee AS e, department AS d
	WHERE e.DepartmentID == d.DepartmentID
)
ORDER by department, name
|sname, sdepartment, lid
[Robinson Clerical 34]
[Smith Clerical 34]
[Heisenberg Engineering 33]
[Jones Engineering 33]
[Rafferty Sales 31]

-- S 73
SELECT name, department, id FROM (
	SELECT e.LastName AS name, e.DepartmentID AS id, d.DepartmentName AS department, d.DepartmentID AS fid
	FROM employee AS e, department AS d
	WHERE e.DepartmentID == d.DepartmentID
)
ORDER by department, name;
|sname, sdepartment, lid
[Robinson Clerical 34]
[Smith Clerical 34]
[Heisenberg Engineering 33]
[Jones Engineering 33]
[Rafferty Sales 31]

-- S 74
SELECT *
FROM
(
	SELECT *
	FROM employee
),
(
	SELECT *
	FROM department
);
|s, ?, l, s
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]

-- S 75
SELECT *
FROM
(
	SELECT *
	FROM employee
) AS e,
(
	SELECT *
	FROM department
)
ORDER BY e.LastName, e.DepartmentID;
|se.LastName, le.DepartmentID, l, s
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]

-- S 76
SELECT *
FROM
(
	SELECT *
	FROM employee
),
(
	SELECT *
	FROM department
) AS d
ORDER BY d.DepartmentID DESC;
|s, l, ld.DepartmentID, sd.DepartmentName
[Rafferty 31 35 Marketing]
[Jones 33 35 Marketing]
[Heisenberg 33 35 Marketing]
[Robinson 34 35 Marketing]
[Smith 34 35 Marketing]
[John <nil> 35 Marketing]
[Rafferty 31 34 Clerical]
[Jones 33 34 Clerical]
[Heisenberg 33 34 Clerical]
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]
[John <nil> 34 Clerical]
[Rafferty 31 33 Engineering]
[Jones 33 33 Engineering]
[Heisenberg 33 33 Engineering]
[Robinson 34 33 Engineering]
[Smith 34 33 Engineering]
[John <nil> 33 Engineering]
[Rafferty 31 31 Sales]
[Jones 33 31 Sales]
[Heisenberg 33 31 Sales]
[Robinson 34 31 Sales]
[Smith 34 31 Sales]
[John <nil> 31 Sales]

-- S 77
SELECT *
FROM
	employee,
	(
		SELECT *
		FROM department
	)
ORDER BY employee.LastName;
|semployee.LastName, lemployee.DepartmentID, l, s
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]

-- S 78
SELECT *
FROM
(
	SELECT *
	FROM employee
) AS e,
(
	SELECT *
	FROM department
) AS d
WHERE e.DepartmentID == d.DepartmentID
ORDER BY d.DepartmentName, e.LastName;
|se.LastName, le.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Rafferty 31 31 Sales]

-- S 79
SELECT *
FROM
	employee,
	(
		SELECT *
		FROM department
	) AS d
WHERE employee.DepartmentID == d.DepartmentID
ORDER BY d.DepartmentName, employee.LastName;
|semployee.LastName, lemployee.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Rafferty 31 31 Sales]

-- S 80
SELECT *
FROM
	employee AS e,
	(
		SELECT *
		FROM department
	) AS d
WHERE e.DepartmentID == d.DepartmentID
ORDER BY d.DepartmentName, e.LastName;
|se.LastName, le.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Rafferty 31 31 Sales]

-- S 81
SELECT *
FROM
	employee AS e,
	(
		SELECT *
		FROM department
	) AS d
WHERE e.DepartmentID == d.DepartmentID == true
ORDER BY e.DepartmentID, e.LastName;
|se.LastName, le.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Rafferty 31 31 Sales]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]

-- S 82
SELECT *
FROM
	employee AS e,
	(
		SELECT *
		FROM department
	) AS d
WHERE e.DepartmentID != d.DepartmentID == false
ORDER BY e.DepartmentID, e.LastName;
|se.LastName, le.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Rafferty 31 31 Sales]
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]

-- 83
BEGIN TRANSACTION;
	CREATE TABLE t (c1 bool);
	INSERT INTO t VALUES (1);
COMMIT;
||cannot .* int64.*bool .* c1

-- 84
BEGIN TRANSACTION;
	CREATE TABLE t (c1 bool);
	INSERT INTO t VALUES (true);
COMMIT;
SELECT * from t;
|bc1
[true]

-- 85
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int8);
	INSERT INTO t VALUES ("foo");
COMMIT;
||cannot .* string.*int8 .* c1

-- 86
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int8);
	INSERT INTO t VALUES (0x1234);
COMMIT;
SELECT * from t;
|ic1
[52]

-- 87
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int16);
	INSERT INTO t VALUES (87);
COMMIT;
SELECT * from t;
|jc1
[87]

-- 88
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int16);
	INSERT INTO t VALUES (int16(0x12345678));
COMMIT;
SELECT * from t;
|jc1
[22136]

-- 89
BEGIN TRANSACTION;
CREATE TABLE t (c1 int32);
	INSERT INTO t VALUES (uint32(1));
COMMIT;
||cannot .* uint32.*int32 .* c1

-- 90
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32);
	INSERT INTO t VALUES (0xabcd12345678);
COMMIT;
SELECT * from t;
|kc1
[305419896]

-- 91
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int64);
	INSERT INTO t VALUES (int8(1));
COMMIT;
||cannot .* int8.*int64 .* c1

-- 92
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int64);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t;
|lc1
[1]

-- 93
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int);
	INSERT INTO t VALUES (int8(1));
COMMIT;
||cannot .* int8.*int64 .* c1

-- 94
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int);
	INSERT INTO t VALUES (94);
COMMIT;
SELECT * from t;
|lc1
[94]

-- 95
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint8);
	INSERT INTO t VALUES (95);
COMMIT;
SELECT * from t;
|uc1
[95]

-- 96
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint8);
	INSERT INTO t VALUES (uint8(0x1234));
COMMIT;
SELECT * from t;
|uc1
[52]

-- 97
BEGIN TRANSACTION;
	CREATE TABLE t (c1 byte);
	INSERT INTO t VALUES (int8(1));
COMMIT;
||cannot .* int8.*uint8 .* c1

-- 98
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint8);
	INSERT INTO t VALUES (byte(0x1234));
COMMIT;
SELECT * from t;
|uc1
[52]

-- 99
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint16);
	INSERT INTO t VALUES (int(1));
COMMIT;
||cannot .* int64.*uint16 .* c1

-- 100
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint16);
	INSERT INTO t VALUES (0x12345678);
COMMIT;
SELECT * from t;
|vc1
[22136]

-- 101
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint32);
	INSERT INTO t VALUES (int32(1));
COMMIT;
||cannot .* int32.*uint32 .* c1

-- 102
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint32);
	INSERT INTO t VALUES (uint32(0xabcd12345678));
COMMIT;
SELECT * from t;
|wc1
[305419896]

-- 103
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint64);
	INSERT INTO t VALUES (int(1));
COMMIT;
||cannot .* int64.*uint64 .* c1

-- 104
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint64);
	INSERT INTO t VALUES (uint64(1));
COMMIT;
SELECT * from t;
|xc1
[1]

-- 105
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint);
	INSERT INTO t VALUES (int(1));
COMMIT;
||cannot .* int64.*uint64 .* c1

-- 106
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t;
|xc1
[1]

-- 107
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float32);
	INSERT INTO t VALUES (107);
COMMIT;
SELECT * from t;
|fc1
[107]

-- 108
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float32);
	INSERT INTO t VALUES (float64(1));
COMMIT;
SELECT * from t;
||cannot .* float64.*float32 .* c1

-- 109
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float32);
	INSERT INTO t VALUES (1.2);
COMMIT;
SELECT * from t;
|fc1
[1.2]

-- 110
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float64);
	INSERT INTO t VALUES (1.2);
COMMIT;
SELECT * from t;
|gc1
[1.2]

-- 111
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float);
	INSERT INTO t VALUES (111.1);
COMMIT;
SELECT * from t;
|gc1
[111.1]

-- 112
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float);
	INSERT INTO t VALUES (-112.1);
COMMIT;
SELECT * from t;
|gc1
[-112.1]

-- 113
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex64);
	INSERT INTO t VALUES (complex(1, 0.5));
COMMIT;
SELECT * from t;
|cc1
[(1+0.5i)]

-- 114
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex64);
	INSERT INTO t VALUES (complex128(complex(1, 0.5)));
COMMIT;
SELECT * from t;
||cannot .* complex128.*complex64 .* c1

-- 115
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex128);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t;
|dc1
[(1+0i)]

-- 116
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex128);
	INSERT INTO t VALUES (complex(1, 0.5));
COMMIT;
SELECT * from t;
|dc1
[(1+0.5i)]

-- 117
BEGIN TRANSACTION;
	CREATE TABLE t (c1 string);
	INSERT INTO t VALUES (1);
COMMIT;
||cannot .* int64.*string .* c1

-- 118
BEGIN TRANSACTION;
	CREATE TABLE t (c1 string);
	INSERT INTO t VALUES ("a"+"b");
COMMIT;
SELECT * from t;
|sc1
[ab]

-- 119
BEGIN TRANSACTION;
	CREATE TABLE t (c1 bool);
	INSERT INTO t VALUES (true);
COMMIT;
SELECT * from t
WHERE c1 > 3;
||operator .* not defined .* bool

-- 120
BEGIN TRANSACTION;
	CREATE TABLE t (c1 bool);
	INSERT INTO t VALUES (true);
COMMIT;
SELECT * from t
WHERE c1;
|bc1
[true]

-- 121
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int8);
	INSERT INTO t VALUES (float(1));
COMMIT;
SELECT * from t
WHERE c1 == 8;
||cannot .* float64.*int8 .* c1

-- 122
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int8);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == int8(1);
|ic1
[1]

-- 123
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int16);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == int(8);
||mismatched .* int16 .* int64

-- 124
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int16);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|jc1
[1]

-- 125
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == int(8);
||mismatched .* int32 .* int64

-- 126
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|kc1
[1]

-- 127
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int64);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == byte(8);
||mismatched .* int64 .* uint8

-- 128
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int64);
	INSERT INTO t VALUES (int64(1));
COMMIT;
SELECT * from t
WHERE c1 == 1;
|lc1
[1]

-- 129
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|lc1
[1]

-- 130
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint8);
	INSERT INTO t VALUES (byte(1));
COMMIT;
SELECT * from t
WHERE c1 == int8(8);
||mismatched .* uint8 .* int8

-- 131
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint8);
	INSERT INTO t VALUES (byte(1));
COMMIT;
SELECT * from t
WHERE c1 == 1;
|uc1
[1]

-- 132
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint16);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == byte(8);
||mismatched .* uint16 .* uint8

-- 133
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint16);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|vc1
[1]

-- 134
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint32);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|wc1
[1]

-- 135
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint64);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == int(8);
||mismatched .* uint64 .* int64

-- 136
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint64);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|xc1
[1]

-- 137
BEGIN TRANSACTION;
	CREATE TABLE t (c1 uint);
	INSERT INTO t VALUES (1);
COMMIT;
SELECT * from t
WHERE c1 == 1;
|xc1
[1]

-- 138
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float32);
	INSERT INTO t VALUES (8);
COMMIT;
SELECT * from t
WHERE c1 == byte(8);
||mismatched .* float32 .* uint8

-- 139
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float32);
	INSERT INTO t VALUES (8);
COMMIT;
SELECT * from t
WHERE c1 == 8;
|fc1
[8]

-- 140
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float64);
	INSERT INTO t VALUES (2);
COMMIT;
SELECT * from t
WHERE c1 == byte(2);
||mismatched .* float64 .* uint8

-- 141
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float64);
	INSERT INTO t VALUES (2);
COMMIT;
SELECT * from t
WHERE c1 == 2;
|gc1
[2]

-- 142
BEGIN TRANSACTION;
	CREATE TABLE t (c1 float);
	INSERT INTO t VALUES (2.);
COMMIT;
SELECT * from t
WHERE c1 == 2;
|gc1
[2]

-- 143
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex64);
	INSERT INTO t VALUES (complex(2., 5.));
COMMIT;
SELECT * from t
WHERE c1 == "foo";
||mismatched .* complex64 .* string

-- 144
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex64);
	INSERT INTO t VALUES (complex(2, 5.));
COMMIT;
SELECT * from t
WHERE c1 == 2+5i;
|cc1
[(2+5i)]

-- 145
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex128);
	INSERT INTO t VALUES (2+5i);
COMMIT;
SELECT * from t
WHERE c1 == "2";
||mismatched .* complex128 .* string

-- 146
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex128);
	INSERT INTO t VALUES (2+5i);
COMMIT;
SELECT * from t
WHERE c1 == complex(2, 5);
|dc1
[(2+5i)]

-- 147
BEGIN TRANSACTION;
	CREATE TABLE t (c1 string);
	INSERT INTO t VALUES ("foo");
COMMIT;
SELECT * from t
WHERE c1 == 2;
||mismatched .* string .* int64

-- 148
BEGIN TRANSACTION;
	CREATE TABLE t (c1 string);
	INSERT INTO t VALUES ("f"+"oo");
COMMIT;
SELECT * from t
WHERE c1 == "fo"+"o";
|sc1
[foo]

-- 149
SELECT 2/(3*5-15) AS foo FROM bar;
||division by zero

-- 150
SELECT 2.0/(2.0-2.0) AS foo FROM bar;
||division by zero

-- 151
SELECT 2i/(2i-2i) AS foo FROM bar;
||division by zero

-- 152
SELECT 2/(3*5-x) AS foo FROM bar;
||table .* not exist

-- S 153
SELECT 314, 42 AS AUQLUE, DepartmentID, DepartmentID+1000, LastName AS Name
FROM employee
ORDER BY Name;
|l, lAUQLUE, lDepartmentID, l, sName
[314 42 33 1033 Heisenberg]
[314 42 <nil> <nil> John]
[314 42 33 1033 Jones]
[314 42 31 1031 Rafferty]
[314 42 34 1034 Robinson]
[314 42 34 1034 Smith]

-- S 154
SELECT *
FROM
	employee AS e,
	( SELECT * FROM department)
ORDER BY e.LastName;
|se.LastName, le.DepartmentID, l, s
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]

-- S 155
SELECT * FROM employee AS e, ( SELECT * FROM department) AS d
ORDER BY e.LastName;
|se.LastName, le.DepartmentID, ld.DepartmentID, sd.DepartmentName
[Heisenberg 33 35 Marketing]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 31 Sales]
[John <nil> 35 Marketing]
[John <nil> 34 Clerical]
[John <nil> 33 Engineering]
[John <nil> 31 Sales]
[Jones 33 35 Marketing]
[Jones 33 34 Clerical]
[Jones 33 33 Engineering]
[Jones 33 31 Sales]
[Rafferty 31 35 Marketing]
[Rafferty 31 34 Clerical]
[Rafferty 31 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 35 Marketing]
[Robinson 34 34 Clerical]
[Robinson 34 33 Engineering]
[Robinson 34 31 Sales]
[Smith 34 35 Marketing]
[Smith 34 34 Clerical]
[Smith 34 33 Engineering]
[Smith 34 31 Sales]

-- 156
BEGIN TRANSACTION;
	CREATE TABLE t (c1 int32, c2 string);
	INSERT INTO t VALUES (1, "a");
	INSERT INTO t VALUES (int64(2), "b");
COMMIT;
SELECT c2 FROM t;
||cannot .*int64.*int32 .* c1

-- 157
BEGIN TRANSACTION;
	CREATE TABLE t (c1 complex64);
	INSERT INTO t VALUES(1);
COMMIT;
SELECT * FROM t;
|cc1
[(1+0i)]

-- 158
BEGIN TRANSACTION;
	CREATE TABLE p (p bool);
	INSERT INTO p VALUES (NULL), (false), (true);
COMMIT;
SELECT * FROM p;
|bp
[true]
[false]
[<nil>]

-- 159
BEGIN TRANSACTION;
	CREATE TABLE p (p bool);
	INSERT INTO p VALUES (NULL), (false), (true);
COMMIT;
SELECT p.p AS p, q.p AS q, p.p &oror; q.p AS p_or_q, p.p && q.p aS p_and_q FROM p, p AS q;
|bp, bq, bp_or_q, bp_and_q
[true true true true]
[true false true false]
[true <nil> true <nil>]
[false true true false]
[false false false false]
[false <nil> <nil> false]
[<nil> true true <nil>]
[<nil> false <nil> false]
[<nil> <nil> <nil> <nil>]

-- 160
BEGIN TRANSACTION;
	CREATE TABLE p (p bool);
	INSERT INTO p VALUES (NULL), (false), (true);
COMMIT;
SELECT p, !p AS not_p FROM p;
|bp, bnot_p
[true false]
[false true]
[<nil> <nil>]

-- S 161
SELECT * FROM department WHERE DepartmentID >= 33
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[33 Engineering]
[34 Clerical]
[35 Marketing]

-- S 162
SELECT * FROM department WHERE DepartmentID <= 34
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[31 Sales]
[33 Engineering]
[34 Clerical]

-- S 163
SELECT * FROM department WHERE DepartmentID < 34
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[31 Sales]
[33 Engineering]

-- S 164
SELECT +DepartmentID FROM employee;
|?
[<nil>]
[34]
[34]
[33]
[33]
[31]

-- S 165
SELECT * FROM employee
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 33]
[John <nil>]
[Jones 33]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- S 166
SELECT *
FROM employee
ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Rafferty 31]
[Jones 33]
[John <nil>]
[Heisenberg 33]

-- S 167
SELECT 1023+DepartmentID AS y FROM employee
ORDER BY y DESC;
|ly
[1057]
[1057]
[1056]
[1056]
[1054]
[<nil>]

-- S 168
SELECT +DepartmentID AS y FROM employee
ORDER BY y DESC;
|ly
[34]
[34]
[33]
[33]
[31]
[<nil>]

-- S 169
SELECT * FROM employee ORDER BY DepartmentID, LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Jones 33]
[Heisenberg 33]
[Rafferty 31]
[John <nil>]

-- S 170
SELECT * FROM employee ORDER BY 0+DepartmentID DESC;
|sLastName, lDepartmentID
[Robinson 34]
[Smith 34]
[Jones 33]
[Heisenberg 33]
[Rafferty 31]
[John <nil>]

-- S 171
SELECT * FROM employee ORDER BY +DepartmentID DESC;
|sLastName, lDepartmentID
[Robinson 34]
[Smith 34]
[Jones 33]
[Heisenberg 33]
[Rafferty 31]
[John <nil>]

-- S 172
SELECT ^DepartmentID AS y FROM employee
ORDER BY y DESC;
|ly
[-32]
[-34]
[-34]
[-35]
[-35]
[<nil>]

-- S 173
SELECT ^byte(DepartmentID) AS y FROM employee ORDER BY y DESC;
|uy
[224]
[222]
[222]
[221]
[221]
[<nil>]

-- 174
BEGIN TRANSACTION;
	CREATE TABLE t (r RUNE);
	INSERT INTO t VALUES (1), ('A'), (rune(int(0x21)));
COMMIT;
SELECT * FROM t
ORDER BY r;
|kr
[1]
[33]
[65]

-- 175
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (-2), (-1), (0), (1), (2);
COMMIT;
SELECT i^1 AS y FROM t
ORDER by y;
|ly
[-2]
[-1]
[0]
[1]
[3]

-- 176
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (-2), (-1), (0), (1), (2);
COMMIT;
SELECT i&or;1 AS y FROM t
ORDER BY y;
|ly
[-1]
[-1]
[1]
[1]
[3]

-- 177
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (-2), (-1), (0), (1), (2);
COMMIT;
SELECT i&1 FROM t;
|l
[0]
[1]
[0]
[1]
[0]

-- 178
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (-2), (-1), (0), (1), (2);
COMMIT;
SELECT i&^1 AS y FROM t
ORDER BY y;
|ly
[-2]
[-2]
[0]
[0]
[2]

-- S 179
SELECT * from employee WHERE LastName == "Jones" &oror; DepartmentID IS NULL
ORDER by LastName DESC;
|sLastName, lDepartmentID
[Jones 33]
[John <nil>]

-- S 180
SELECT * from employee WHERE LastName != "Jones" && DepartmentID IS NOT NULL
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 33]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- 181
SELECT 42[0] FROM foo;
||invalid operation.*index of type

-- 182
SELECT "foo"[-1] FROM foo;
||invalid string index.*index .* non.*negative

-- 183
SELECT "foo"[3] FROM foo;
||invalid string index.*out of bounds

-- 184
SELECT "foo"["bar">true] FROM foo;
||mismatched type

-- S 185
SELECT DepartmentID[0] FROM employee;
||run.time error.*invalid operation.*index of type

-- S 186
SELECT "foo"[-DepartmentID] FROM employee;
||run.time error.*invalid string index.*index .* non.*negative

-- S 187
SELECT LastName[100] FROM employee;
||run.time.error.*invalid string index.*out of bounds

-- S 188
SELECT LastName[0], LastName FROM employee ORDER BY LastName;
|u, sLastName
[72 Heisenberg]
[74 John]
[74 Jones]
[82 Rafferty]
[82 Robinson]
[83 Smith]

-- S 189
SELECT LastName, string(LastName[0]), string(LastName[1]), string(LastName[2]), string(LastName[3])
FROM employee
ORDER BY LastName;
|sLastName, s, s, s, s
[Heisenberg H e i s]
[John J o h n]
[Jones J o n e]
[Rafferty R a f f]
[Robinson R o b i]
[Smith S m i t]

-- S 190
SELECT LastName, LastName[:], LastName[:2], LastName[2:], LastName[1:3]
FROM employee
ORDER by LastName;
|sLastName, s, s, s, s
[Heisenberg Heisenberg He isenberg ei]
[John John Jo hn oh]
[Jones Jones Jo nes on]
[Rafferty Rafferty Ra fferty af]
[Robinson Robinson Ro binson ob]
[Smith Smith Sm ith mi]

-- S 191
SELECT LastName
FROM employee
WHERE department IS NULL;
||unknown field department

-- S 192
SELECT
	DepartmentID,
	LastName,
	LastName[:4],
	LastName[:0*DepartmentID],
	LastName[0*DepartmentID:0],
	LastName[0*DepartmentID:0*DepartmentID],
FROM
	employee,
ORDER BY LastName DESC;
|lDepartmentID, sLastName, s, s, s, s
[34 Smith Smit   ]
[34 Robinson Robi   ]
[31 Rafferty Raff   ]
[33 Jones Jone   ]
[<nil> John John <nil> <nil> <nil>]
[33 Heisenberg Heis   ]

-- S 193
SELECT
	DepartmentID AS x,
	DepartmentID<<1 AS a,
	1<<uint(DepartmentID) AS b,
FROM
	employee,
WHERE DepartmentID IS NOT NULL
ORDER BY x;
|lx, la, lb
[31 62 2147483648]
[33 66 8589934592]
[33 66 8589934592]
[34 68 17179869184]
[34 68 17179869184]

-- S 194
SELECT
	DepartmentID AS x,
	DepartmentID>>1 AS a,
	uint(1)<<63>>uint(DepartmentID) AS b,
FROM
	employee,
WHERE DepartmentID IS NOT NULL
ORDER BY x;
|lx, la, xb
[31 15 4294967296]
[33 16 1073741824]
[33 16 1073741824]
[34 17 536870912]
[34 17 536870912]

-- S 195
SELECT DISTINCT DepartmentID
FROM employee
WHERE DepartmentID IS NOT NULL;
|lDepartmentID
[31]
[33]
[34]

-- S 196
SELECT DISTINCT e.DepartmentID, d.DepartmentID, e.LastName
FROM employee AS e, department AS d
WHERE e.DepartmentID == d.DepartmentID;
|le.DepartmentID, ld.DepartmentID, se.LastName
[31 31 Rafferty]
[33 33 Heisenberg]
[33 33 Jones]
[34 34 Robinson]
[34 34 Smith]

-- S 197
SELECT DISTINCT e.DepartmentID, d.DepartmentID, e.LastName
FROM employee AS e, department AS d
WHERE e.DepartmentID == d.DepartmentID
ORDER BY e.LastName;
|le.DepartmentID, ld.DepartmentID, se.LastName
[33 33 Heisenberg]
[33 33 Jones]
[31 31 Rafferty]
[34 34 Robinson]
[34 34 Smith]

-- S 198, http://en.wikipedia.org/wiki/Join_(SQL)#Cross_join
SELECT *
FROM employee, department
ORDER BY employee.LastName, department.DepartmentID;
|semployee.LastName, lemployee.DepartmentID, ldepartment.DepartmentID, sdepartment.DepartmentName
[Heisenberg 33 31 Sales]
[Heisenberg 33 33 Engineering]
[Heisenberg 33 34 Clerical]
[Heisenberg 33 35 Marketing]
[John <nil> 31 Sales]
[John <nil> 33 Engineering]
[John <nil> 34 Clerical]
[John <nil> 35 Marketing]
[Jones 33 31 Sales]
[Jones 33 33 Engineering]
[Jones 33 34 Clerical]
[Jones 33 35 Marketing]
[Rafferty 31 31 Sales]
[Rafferty 31 33 Engineering]
[Rafferty 31 34 Clerical]
[Rafferty 31 35 Marketing]
[Robinson 34 31 Sales]
[Robinson 34 33 Engineering]
[Robinson 34 34 Clerical]
[Robinson 34 35 Marketing]
[Smith 34 31 Sales]
[Smith 34 33 Engineering]
[Smith 34 34 Clerical]
[Smith 34 35 Marketing]

-- S 199, http://en.wikipedia.org/wiki/Join_(SQL)#Inner_join
SELECT *
FROM employee, department
WHERE employee.DepartmentID == department.DepartmentID
ORDER BY employee.LastName, department.DepartmentID;
|semployee.LastName, lemployee.DepartmentID, ldepartment.DepartmentID, sdepartment.DepartmentName
[Heisenberg 33 33 Engineering]
[Jones 33 33 Engineering]
[Rafferty 31 31 Sales]
[Robinson 34 34 Clerical]
[Smith 34 34 Clerical]

-- S 200
BEGIN TRANSACTION;
	INSERT INTO department (DepartmentID, DepartmentName)
	SELECT DepartmentID+1000, DepartmentName+"/headquarters"
	FROM department;
COMMIT;
SELECT * FROM department
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[31 Sales]
[33 Engineering]
[34 Clerical]
[35 Marketing]
[1031 Sales/headquarters]
[1033 Engineering/headquarters]
[1034 Clerical/headquarters]
[1035 Marketing/headquarters]

-- S 201`
BEGIN TRANSACTION;
	INSERT INTO department (DepartmentName, DepartmentID)
	SELECT DepartmentName+"/headquarters", DepartmentID+1000
	FROM department;
COMMIT;
SELECT * FROM department
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[31 Sales]
[33 Engineering]
[34 Clerical]
[35 Marketing]
[1031 Sales/headquarters]
[1033 Engineering/headquarters]
[1034 Clerical/headquarters]
[1035 Marketing/headquarters]

-- S 202
BEGIN TRANSACTION;
	DELETE FROM department;
COMMIT;
SELECT * FROM department
|?DepartmentID, ?DepartmentName

-- S 203
BEGIN TRANSACTION;
	DELETE FROM department
	WHERE DepartmentID == 35 &oror; DepartmentName != "" && DepartmentName[0] == 'C';
COMMIT;
SELECT * FROM department
ORDER BY DepartmentID;
|lDepartmentID, sDepartmentName
[31 Sales]
[33 Engineering]

-- S 204
SELECT id(), LastName
FROM employee
ORDER BY id();
|l, sLastName
[5 Rafferty]
[6 Jones]
[7 Heisenberg]
[8 Robinson]
[9 Smith]
[10 John]

-- S 205
BEGIN TRANSACTION;
	DELETE FROM employee
	WHERE LastName == "Jones";
COMMIT;
SELECT id(), LastName
FROM employee
ORDER BY id();
|l, sLastName
[5 Rafferty]
[7 Heisenberg]
[8 Robinson]
[9 Smith]
[10 John]

-- S 206
BEGIN TRANSACTION;
	DELETE FROM employee
	WHERE LastName == "Jones";
	INSERT INTO employee (LastName) VALUES ("Jones");
COMMIT;
SELECT id(), LastName
FROM employee
ORDER BY id();
|l, sLastName
[5 Rafferty]
[7 Heisenberg]
[8 Robinson]
[9 Smith]
[10 John]
[11 Jones]

-- S 207
SELECT id(), e.LastName, e.DepartmentID, d.DepartmentID
FROM
	employee AS e,
	department AS d,
WHERE e.DepartmentID == d.DepartmentID
ORDER BY e.LastName;
|?, se.LastName, le.DepartmentID, ld.DepartmentID
[<nil> Heisenberg 33 33]
[<nil> Jones 33 33]
[<nil> Rafferty 31 31]
[<nil> Robinson 34 34]
[<nil> Smith 34 34]

-- S 208
SELECT e.ID, e.LastName, e.DepartmentID, d.DepartmentID
FROM
	(SELECT id() AS ID, LastName, DepartmentID FROM employee;) AS e,
	department AS d,
WHERE e.DepartmentID == d.DepartmentID
ORDER BY e.ID;
|le.ID, se.LastName, le.DepartmentID, ld.DepartmentID
[5 Rafferty 31 31]
[6 Jones 33 33]
[7 Heisenberg 33 33]
[8 Robinson 34 34]
[9 Smith 34 34]

-- S 209
BEGIN TRANSACTION;
	UPDATE none
		DepartmentID = DepartmentID+1000,
	WHERE DepartmentID == 33;
COMMIT;
SELECT * FROM employee;
||table.*not.*exist

-- S 210
BEGIN TRANSACTION;
	UPDATE employee
		FirstName = "John"
	WHERE DepartmentID == 33;
COMMIT;
SELECT * FROM employee;
||unknown.*FirstName

-- S 211
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentID = DepartmentID+1000,
	WHERE DepartmentID == 33;
COMMIT;
SELECT * FROM employee
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 1033]
[John <nil>]
[Jones 1033]
[Rafferty 31]
[Robinson 34]
[Smith 34]

-- S 212
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentID = DepartmentID+1000,
		LastName = "Mr. "+LastName
	WHERE id() == 7;
COMMIT;
SELECT * FROM employee
ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Rafferty 31]
[Mr. Heisenberg 1033]
[Jones 33]
[John <nil>]

-- S 213
BEGIN TRANSACTION;
	UPDATE employee
		LastName = "Mr. "+LastName,
		DepartmentID = DepartmentID+1000,
	WHERE id() == 7;
COMMIT;
SELECT * FROM employee
ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Rafferty 31]
[Mr. Heisenberg 1033]
[Jones 33]
[John <nil>]

-- S 214
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentID = DepartmentID+1000;
COMMIT;
SELECT * FROM employee
ORDER BY LastName;
|sLastName, lDepartmentID
[Heisenberg 1033]
[John <nil>]
[Jones 1033]
[Rafferty 1031]
[Robinson 1034]
[Smith 1034]

-- S 215
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentId = DepartmentID+1000;
COMMIT;
SELECT * FROM employee;
||unknown

-- S 216
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentID = DepartmentId+1000;
COMMIT;
SELECT * FROM employee;
||unknown

-- S 217
BEGIN TRANSACTION;
	UPDATE employee
		DepartmentID = "foo";
COMMIT;
SELECT * FROM employee;
||cannot .* string.*int64 .* DepartmentID

-- S 218
SELECT foo[len()] FROM bar;
||missing argument

-- S 219
SELECT foo[len(42)] FROM bar;
||invalid argument

-- S 220
SELECT foo[len(42, 24)] FROM bar;
||too many

-- S 221
SELECT foo[len("baz")] FROM bar;
||table

-- S 222
SELECT LastName[len("baz")-4] FROM employee;
||invalid string index

-- S 223
SELECT LastName[:len(LastName)-3] AS y FROM employee
ORDER BY y;
|sy
[Heisenb]
[J]
[Jo]
[Raffe]
[Robin]
[Sm]

-- S 224
SELECT complex(float32(DepartmentID+int(id())), 0) AS x, complex(DepartmentID+int(id()), 0)
FROM employee
ORDER by real(x) DESC;
|cx, d
[(43+0i) (43+0i)]
[(42+0i) (42+0i)]
[(40+0i) (40+0i)]
[(39+0i) (39+0i)]
[(36+0i) (36+0i)]
[<nil> <nil>]

-- S 225
SELECT real(complex(float32(DepartmentID+int(id())), 0)) AS x, real(complex(DepartmentID+int(id()), 0))
FROM employee
ORDER BY x DESC;
|fx, g
[43 43]
[42 42]
[40 40]
[39 39]
[36 36]
[<nil> <nil>]

-- S 226
SELECT imag(complex(0, float32(DepartmentID+int(id())))) AS x, imag(complex(0, DepartmentID+int(id())))
FROM employee
ORDER BY x DESC;
|fx, g
[43 43]
[42 42]
[40 40]
[39 39]
[36 36]
[<nil> <nil>]

-- 227
BEGIN TRANSACTION;
	CREATE TABLE t (c string);
	INSERT INTO t VALUES("foo"), ("bar");
	DELETE FROM t WHERE c == "foo";
COMMIT;
SELECT 100*id(), c FROM t;
|l, sc
[200 bar]

-- 228
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE a;
COMMIT;
SELECT * FROM a;
||table a does not exist

-- 229
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE a;
COMMIT;
SELECT * FROM b;
|?b

-- 230
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE a;
COMMIT;
SELECT * FROM c;
|?c

-- 231
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE b;
COMMIT;
SELECT * FROM a;
|?a

-- 232
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE b;
COMMIT;
SELECT * FROM b;
||table b does not exist

-- 233
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE b;
COMMIT;
SELECT * FROM c;
|?c

-- 234
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE c;
COMMIT;
SELECT * FROM a;
|?a

-- 235
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE c;
COMMIT;
SELECT * FROM b;
|?b

-- 236
BEGIN TRANSACTION;
	CREATE TABLE a (a int);
	CREATE TABLE b (b int);
	CREATE TABLE c (c int);
	DROP TABLE c;
COMMIT;
SELECT * FROM c;
||table c does not exist

-- 237
BEGIN TRANSACTION;
	CREATE TABLE a (c int);
	INSERT INTO a VALUES (10), (11), (12);
	CREATE TABLE b (d int);
	INSERT INTO b VALUES (20), (21), (22), (23);
COMMIT;
SELECT * FROM a, b;
|la.c, lb.d
[12 23]
[12 22]
[12 21]
[12 20]
[11 23]
[11 22]
[11 21]
[11 20]
[10 23]
[10 22]
[10 21]
[10 20]

-- 238
BEGIN TRANSACTION;
	CREATE TABLE a (c int);
	INSERT INTO a VALUES (0), (1), (2);
COMMIT;
SELECT
	9*x2.c AS x2,
	3*x1.c AS x1,
	1*x0.c AS x0,
	9*x2.c + 3*x1.c + x0.c AS y,
FROM
	a AS x2,
	a AS x1,
	a AS x0,
ORDER BY y;
|lx2, lx1, lx0, ly
[0 0 0 0]
[0 0 1 1]
[0 0 2 2]
[0 3 0 3]
[0 3 1 4]
[0 3 2 5]
[0 6 0 6]
[0 6 1 7]
[0 6 2 8]
[9 0 0 9]
[9 0 1 10]
[9 0 2 11]
[9 3 0 12]
[9 3 1 13]
[9 3 2 14]
[9 6 0 15]
[9 6 1 16]
[9 6 2 17]
[18 0 0 18]
[18 0 1 19]
[18 0 2 20]
[18 3 0 21]
[18 3 1 22]
[18 3 2 23]
[18 6 0 24]
[18 6 1 25]
[18 6 2 26]

-- 239
BEGIN TRANSACTION;
	CREATE TABLE t (c int);
	INSERT INTO t VALUES (242);
	DELETE FROM t WHERE c != 0;
COMMIT;
SELECT * FROM t
|?c

-- 240
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (242), (12, 24);
COMMIT;
||expect

-- 241
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (24, 2), (1224);
COMMIT;
||expect

-- 242
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
ROLLBACK;
SELECT * from t;
||does not exist

-- 243
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
COMMIT;
BEGIN TRANSACTION;
	DROP TABLE T;
COMMIT;
SELECT * from t;
||does not exist

-- 244
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
COMMIT;
BEGIN TRANSACTION;
	DROP TABLE t;
ROLLBACK;
SELECT * from t;
|?i

-- 245
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (int(1.2));
COMMIT;
SELECT * FROM t;
||truncated

-- 246
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (string(65.0));
COMMIT;
SELECT * FROM t;
||cannot convert

-- 247
BEGIN TRANSACTION;
	CREATE TABLE t (s string);
	INSERT INTO t VALUES (string(65));
COMMIT;
SELECT * FROM t;
|ss
[A]

-- 248
BEGIN TRANSACTION;
	CREATE TABLE t (i uint32);
	INSERT INTO t VALUES (uint32(int8(uint16(0x10F0))));
COMMIT;
SELECT i == 0xFFFFFFF0 FROM t;
|b
[true]

-- 249
BEGIN TRANSACTION;
	CREATE TABLE t (s string);
	INSERT INTO t VALUES
		(string('a')),		// "a"
		(string(-1)),		// "\ufffd" == "\xef\xbf\xbd"
		(string(0xf8)),		// "\u00f8" == "ø" == "\xc3\xb8"
		(string(0x65e5)),	// "\u65e5" == "日" == "\xe6\x97\xa5"
	;
COMMIT;
SELECT
	id() == 1 && s == "a" &oror;
	id() == 2 && s == "\ufffd" && s == "\xef\xbf\xbd" &oror;
	id() == 3 && s == "\u00f8" && s == "ø" && s == "\xc3\xb8" &oror;
	id() == 4 && s == "\u65e5" && s == "日" && s == "\xe6\x97\xa5"
FROM t;
|b
[true]
[true]
[true]
[true]

-- 250
BEGIN TRANSACTION;
	CREATE TABLE t (i int);
	INSERT INTO t VALUES (0);
COMMIT;
SELECT 2.3+1, 1+2.3 FROM t;
|f, f
[3.3 3.3]

-- 251
BEGIN TRANSACTION;
	CREATE TABLE t (i byte);
	INSERT INTO t VALUES (-1+byte(2));
COMMIT;
SELECT * FROM t;
||mismatched

-- 252
BEGIN TRANSACTION;
	CREATE TABLE t (i byte);
	INSERT INTO t VALUES (1+byte(2));
COMMIT;
SELECT * FROM t;
|ui
[3]

-- 253
BEGIN TRANSACTION;
	CREATE TABLE t (i byte);
	INSERT INTO t VALUES (255+byte(2));
COMMIT;
SELECT * FROM t;
|ui
[1]

-- 254
BEGIN TRANSACTION;
	CREATE TABLE t (i byte);
	INSERT INTO t VALUES (256+byte(2));
COMMIT;
SELECT * FROM t;
||mismatched

-- 255
BEGIN TRANSACTION;
	CREATE TABLE t (i int8);
	INSERT INTO t VALUES (127+int8(2));
COMMIT;
SELECT * FROM t;
|ii
[-127]

-- 256
BEGIN TRANSACTION;
	CREATE TABLE t (i int8);
	INSERT INTO t VALUES (-129+int8(2));
COMMIT;
SELECT * FROM t;
||mismatched

-- 257
BEGIN TRANSACTION;
	CREATE TABLE t (i int8);
	INSERT INTO t VALUES (-128+int8(2));
COMMIT;
SELECT * FROM t;
|ii
[-126]

-- 258
BEGIN TRANSACTION;
	CREATE TABLE t (i int8);
	INSERT INTO t VALUES (128+int8(2));
COMMIT;
SELECT * FROM t;
||mismatched

-- S 259
SELECT count(none) FROM employee;
||unknown

-- S 260
SELECT count() FROM employee;
|l
[6]

-- S 261
SELECT count() AS y FROM employee;
|ly
[6]

-- S 262
SELECT 3*count() AS y FROM employee;
|ly
[18]

-- S 263
SELECT count(LastName) FROM employee;
|l
[6]

-- S 264
SELECT count(DepartmentID) FROM employee;
|l
[5]

-- S 265
SELECT count() - count(DepartmentID) FROM employee;
|l
[1]

-- S 266
SELECT min(LastName), min(DepartmentID) FROM employee;
|s, l
[Heisenberg 31]

-- S 267
SELECT max(LastName), max(DepartmentID) FROM employee;
|s, l
[Smith 34]

-- S 268
SELECT sum(LastName), sum(DepartmentID) FROM employee;
||cannot

-- S 269
SELECT sum(DepartmentID) FROM employee;
|l
[165]

-- S 270
SELECT avg(DepartmentID) FROM employee;
|l
[33]

-- S 271
SELECT DepartmentID FROM employee GROUP BY none;
||unknown

-- S 272
SELECT DepartmentID, sum(DepartmentID) AS s FROM employee GROUP BY DepartmentID ORDER BY s DESC;
|lDepartmentID, ls
[34 68]
[33 66]
[31 31]
[<nil> <nil>]

-- S 273
SELECT DepartmentID, count(LastName+string(DepartmentID)) AS y FROM employee GROUP BY DepartmentID ORDER BY y DESC ;
|lDepartmentID, ly
[34 2]
[33 2]
[31 1]
[<nil> 0]

-- S 274
SELECT DepartmentID, sum(2*DepartmentID) AS s FROM employee GROUP BY DepartmentID ORDER BY s DESC;
|lDepartmentID, ls
[34 136]
[33 132]
[31 62]
[<nil> <nil>]

-- S 275
SELECT min(2*DepartmentID) FROM employee;
|l
[62]

-- S 276
SELECT max(2*DepartmentID) FROM employee;
|l
[68]

-- S 277
SELECT avg(2*DepartmentID) FROM employee;
|l
[66]

-- S 278
SELECT * FROM employee GROUP BY DepartmentID;
|sLastName, ?DepartmentID
[John <nil>]
[Rafferty 31]
[Heisenberg 33]
[Smith 34]

-- S 279
SELECT * FROM employee GROUP BY DepartmentID ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Rafferty 31]
[John <nil>]
[Heisenberg 33]

-- S 280
SELECT * FROM employee GROUP BY DepartmentID, LastName ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Rafferty 31]
[Jones 33]
[John <nil>]
[Heisenberg 33]

-- S 281
SELECT * FROM employee GROUP BY LastName, DepartmentID  ORDER BY LastName DESC;
|sLastName, lDepartmentID
[Smith 34]
[Robinson 34]
[Rafferty 31]
[Jones 33]
[John <nil>]
[Heisenberg 33]

-- 282
BEGIN TRANSACTION;
	CREATE TABLE s (i int);
	CREATE TABLE t (i int);
COMMIT;
BEGIN TRANSACTION;
	DROP TABLE s;
COMMIT;
SELECT * FROM t;
|?i

-- 283
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
COMMIT;
SELECT count() FROM t;
|l
[0]

-- 284
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT count() FROM t;
|l
[2]

-- 285
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT count() FROM t WHERE n < 2;
|l
[2]

-- 286
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT count() FROM t WHERE n < 1;
|l
[1]

-- 287
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT count() FROM t WHERE n < 0;
|l
[0]

-- 288
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT s+10 FROM (SELECT sum(n) AS s FROM t WHERE n < 2);
|l
[11]

-- 289
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT s+10 FROM (SELECT sum(n) AS s FROM t WHERE n < 1);
|l
[10]

-- 290
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT s+10 FROM (SELECT sum(n) AS s FROM t WHERE n < 0);
|?
[<nil>]

-- 291
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT sum(n) AS s FROM t WHERE n < 2;
|ls
[1]

-- 292
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT sum(n) AS s FROM t WHERE n < 1;
|ls
[0]

-- 293
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1);
COMMIT;
SELECT sum(n) AS s FROM t WHERE n < 0;
|?s
[<nil>]

-- 294
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t SELECT count() FROM t;
	INSERT INTO t SELECT count() FROM t;
	INSERT INTO t SELECT count() FROM t;
COMMIT;
SELECT count() FROM t;
|l
[3]

-- 295
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t SELECT count() FROM t;
	INSERT INTO t SELECT count() FROM t;
	INSERT INTO t SELECT count() FROM t;
	INSERT INTO t SELECT * FROM t;
COMMIT;
SELECT count() FROM t;
|l
[6]

-- 296
BEGIN TRANSACTION;
	CREATE TABLE t (n int);
	INSERT INTO t VALUES (0), (1), (2);
	INSERT INTO t SELECT * FROM t;
COMMIT;
SELECT count() FROM t;
|l
[6]

-- 297
BEGIN TRANSACTION;
	CREATE TABLE t(S string);
	INSERT INTO t SELECT "perfect!" FROM (SELECT count() AS cnt FROM t WHERE S == "perfect!") WHERE cnt == 0;
COMMIT;
SELECT count() FROM t;
|l
[1]

-- 298
BEGIN TRANSACTION;
	CREATE TABLE t(S string);
	INSERT INTO t SELECT "perfect!" FROM (SELECT count() AS cnt FROM t WHERE S == "perfect!") WHERE cnt == 0;
	INSERT INTO t SELECT "perfect!" FROM (SELECT count() AS cnt FROM t WHERE S == "perfect!") WHERE cnt == 0;
COMMIT;
SELECT count() FROM t;
|l
[1]

-- 299
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob("a"));
COMMIT;
SELECT * FROM t;
|?c
[[97]]

-- 300
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(`
0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
`));
COMMIT;
SELECT * FROM t;
|?c
[[10 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 10]]

-- 301
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789ABCDEF"));
COMMIT;
SELECT * FROM t;
|?c
[[48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 65 66 67 68 69 70]]

-- 302
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789ABCDEF"+
"!"));
COMMIT;
SELECT * FROM t;
|?c
[[48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 97 98 99 100 101 102 48 49 50 51 52 53 54 55 56 57 65 66 67 68 69 70 33]]

-- 303
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob("hell\xc3\xb8"));
COMMIT;
SELECT string(c) FROM t;
|s
[hellø]

-- 304
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"+
"0123456789abcdef0123456789abcdef0123456789abcdef0123456789ABCDEF"+
"!"));
COMMIT;
SELECT string(c) FROM t;
|s
[0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789ABCDEF!]

-- 305
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(""));
COMMIT;
SELECT string(c) FROM t;
|s
[]

-- 306
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob("hellø"));
COMMIT;
SELECT * FROM t;
|?c
[[104 101 108 108 195 184]]

-- 307
BEGIN TRANSACTION;
	CREATE TABLE t(c blob);
	INSERT INTO t VALUES (blob(""));
COMMIT;
SELECT * FROM t;
|?c
[[]]

-- 308
BEGIN TRANSACTION;
	CREATE TABLE t(i int, b blob);
	INSERT INTO t VALUES
		(0, blob("0")),
	;
COMMIT;
SELECT * FROM t;
|li, ?b
[0 [48]]

-- 309
BEGIN TRANSACTION;
	CREATE TABLE t(i int, b blob);
	INSERT INTO t VALUES
		(0, blob("0")),
		(1, blob("1")),
	;
COMMIT;
SELECT * FROM t;
|li, ?b
[1 [49]]
[0 [48]]

-- 310
BEGIN TRANSACTION;
	CREATE TABLE t(i int, b blob);
	INSERT INTO t VALUES
		(0, blob("0")),
		(1, blob("1")),
		(2, blob("2")),
	;
COMMIT;
SELECT * FROM t;
|li, ?b
[2 [50]]
[1 [49]]
[0 [48]]
