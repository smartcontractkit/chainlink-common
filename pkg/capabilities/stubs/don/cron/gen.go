package cron

//go:generate protoc -I. -I../../../protoc/pkg/pb --go_out=. --go_opt=paths=source_relative "--cre_out=mode=don,id=cron@1.0.0:." cron.proto
