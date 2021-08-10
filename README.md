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
>> select name, first_appeared from programming_languages
name     first_appeared
'C'      1972
'Python' 1990
'Lisp'   1958
'Go'     2009
>> select name, 2021 - first_appeared as years_since_introduction from programming_languages
name     years_since_introduction
'C'      49
'Python' 31
'Lisp'   63
'Go'     12
>> select 1, 3.14 as pi, '✅' as emoji, 'Vegard' as name
1 pi       emoji name
1 3.140000 '✅'   'Vegard'
>> select 1 + '3.14'
ERROR: unknown operator: INTEGER + STRING
>> 1
ERROR: expected start of statement, got INT token with literal 1
```

Based on Thorsten Ball's excellent [Writing an Interpreter in Go](https://interpreterbook.com/).
