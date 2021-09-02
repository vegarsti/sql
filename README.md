# sql

<a href="https://github.com/vegarsti/sql/actions"><img src="https://github.com/vegarsti/sql/workflows/test/badge.svg" alt="Build Status"></a>

[![Go Report Card](https://goreportcard.com/badge/github.com/vegarsti/sql)](https://goreportcard.com/report/github.com/vegarsti/sql)

The interpreter can be run interactively.

```
$ go run cmd/sql/main.go
>> create table programming_languages (name text, first_appeared integer);
OK
>> insert into programming_languages values ('C', 1972), ('Python', 1990), ('Lisp', 1958), ('Go', 2009)
OK
>> select name, first_appeared from programming_languages where first_appeared < 2000 order by first_appeared
name     first_appeared
'Lisp'   1958
'C'      1972
'Python' 1990
>> select 1, 3.14 as pi, '✅' as emoji, 'Vegard' as name
1 pi       emoji name
1 3.140000 '✅'   'Vegard'
>> create table squares (number int, square int)
OK
>> insert into squares values (1, 1), (2, 2^2), (3, 3^2)
OK
>> create table cubes (number int, cube int)
OK
>> insert into cubes values (1, 1), (2, 2^3), (3, 3^3)
OK
>> select s.number, square, cube from squares s join cubes c on s.number = c.number
number square cube
1      1      1
2      4      8
3      9      27
```

The interpreter also supports running against standard input.

```
$ cat example.sql | go run cmd/sql/main.go
name     first_appeared
'C'      1972
'Python' 1990
'Lisp'   1958
'Go'     2009
```

Based on Thorsten Ball's excellent [Writing an Interpreter in Go](https://interpreterbook.com/).
