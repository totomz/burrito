module github.com/totomz/template-burrito/services/kargo

go 1.19

replace github.com/totomz/template-burrito/common/httpserver => ../../common/httpserver

require github.com/totomz/template-burrito/common/httpserver v0.0.0-00010101000000-000000000000

replace github.com/totomz/template-burrito/common/burrito-common => ../../common/burrito-common

require github.com/totomz/template-burrito/common/burrito-common v0.0.0-00010101000000-000000000000

require github.com/form3tech-oss/jwt-go v3.2.5+incompatible
