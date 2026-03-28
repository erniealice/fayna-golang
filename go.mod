module github.com/erniealice/fayna-golang

go 1.25.1

require github.com/erniealice/pyeza-golang v0.0.8-alpha

require (
	github.com/erniealice/espyna-golang v0.0.0-20260324142444-f83ca6989174
	github.com/erniealice/esqyma v0.0.0
	github.com/erniealice/hybra-golang v0.0.0
	github.com/erniealice/lyngua v0.0.0-00010101000000-000000000000
)

require (
	github.com/yuin/goldmark v1.7.17 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251002232023-7c0ddcbb5797 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/erniealice/esqyma => ../esqyma

replace github.com/erniealice/pyeza-golang => ../pyeza-golang

replace github.com/erniealice/hybra-golang => ../hybra-golang

replace github.com/erniealice/lyngua => ../lyngua
