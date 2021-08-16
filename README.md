# sql

<a href="https://github.com/vegarsti/sql/actions"><img src="https://github.com/vegarsti/sql/workflows/test/badge.svg" alt="Build Status"></a>

[![Go Report Card](https://goreportcard.com/badge/github.com/vegarsti/sql)](https://goreportcard.com/report/github.com/vegarsti/sql)

```
$ go run cmd/sql/main.go
>> create table programming_languages (name text, first_appeared integer);
OK
>> insert into programming_languages values ('C', 1972)
OK
>> insert into programming_languages values ('Python', 1990)
OK
>> insert into programming_languages values ('Lisp', 1958)
OK
>> insert into programming_languages values ('Go', 2009)
OK
>> select name, first_appeared from programming_languages where first_appeared < 2000 order by first_appeared
name     first_appeared
'Lisp'   1958
'C'      1972
'Python' 1990
>> select 1, 3.14 as pi, '✅' as emoji, 'Vegard' as name
1 pi       emoji name
1 3.140000 '✅'   'Vegard'
>> select 1 + '3.14'
ERROR: unknown operator: INTEGER + STRING
>> 1
ERROR: expected start of statement, got INT token with literal 1
>> create table a (b int);
OK
>> create table b (b int);
OK
>> insert into a values (1);
OK
>> insert into a values (2);
OK
>> insert into b values (1);
OK
>> select a.a from a join b on a.b = b.b
ERROR: no such column: a.a
>> select a.b from a join b on a.b = b.b
b
1
```

Based on Thorsten Ball's excellent [Writing an Interpreter in Go](https://interpreterbook.com/).
