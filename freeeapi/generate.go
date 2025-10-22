package freeapi

//go:generate go tool oapi-codegen -generate types,client -config cfg.yaml -o gen/api_gen.go api-schema.json
