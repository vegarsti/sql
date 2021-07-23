# sql

<a href="https://github.com/vegarsti/sql/actions"><img src="https://github.com/vegarsti/sql/workflows/test/badge.svg" alt="Build Status"></a>

[![Go Report Card](https://goreportcard.com/badge/github.com/vegarsti/sql)](https://goreportcard.com/report/github.com/vegarsti/sql)

```
$ go run cmd/sql/main.go
>> select 1, 3.14 as pi, '✅' as emoji, 'Vegard' as name
?column? pi       emoji name
1        3.140000 ✅     Vegard
>> 1
ERROR: expected start of statement, got INT token with literal 1
```

Based on Thorsten Ball's excellent [Writing an Interpreter in Go](https://interpreterbook.com/).
