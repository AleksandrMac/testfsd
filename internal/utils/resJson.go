package utils

import (
	"encoding/json"
	"net/http"

	"github.com/AleksandrMac/testfsd/internal/log"
)

func RJSON(rw http.ResponseWriter, statusCode int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error(err.Error())
	}
	rw.WriteHeader(statusCode)
	rw.Write(data)
}
