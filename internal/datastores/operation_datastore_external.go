package datastores

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"
// 	"reflect"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/AleksandrMac/testfsd/internal/config"
// 	"github.com/AleksandrMac/testfsd/internal/entities"
// 	"github.com/AleksandrMac/testfsd/internal/log"
// 	"github.com/AleksandrMac/testfsd/internal/utils"

// 	"go.opentelemetry.io/otel"
// )

// type operationDatastoreExternal struct {
// 	conf config.Datastore
// }

// // NewOperationDatastoreExternal get operation external api
// func NewOperationDatastoreExternal(cnf config.Datastore) OperationDatastore {
// 	return &operationDatastoreExternal{cnf}
// }

// func (x *operationDatastoreExternal) CreateOperation(ctx context.Context, o *entities.Operation) (string, error) {
// 	_, span := otel.Tracer(utils.GetFnName()).Start(ctx, "datastore")
// 	defer span.End()

// 	span.AddEvent("check the existence of a fact by tag operationId")
// 	factId, err := x.findFact(o)
// 	if err != nil {
// 		return "", err
// 	}

// 	span.AddEvent("get supertags")
// 	tags, err := x.getSupertags()
// 	if err != nil {
// 		return "", err
// 	}

// 	span.AddEvent("make multypart/formdata")
// 	buf, boundary, err := makeOperationMultypartFormData(o, x.conf.IndicatorToMoId, factId, string(makeLabels(o, tags)))
// 	if err != nil {
// 		return "", err
// 	}

// 	span.AddEvent("create a http.request to save the fact")
// 	req, err := http.NewRequest("POST", fmt.Sprintf("%s/_api/facts/save_fact?auth_user_id=%d", x.conf.URL, x.conf.AuthUserId), buf)
// 	if err != nil {
// 		return "", err
// 	}
// 	req.Header.Set("Authorization", "Bearer "+x.conf.Token)
// 	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

// 	span.AddEvent("make a http.request")
// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	span.AddEvent("check response status")
// 	if res.StatusCode >= 400 {
// 		return "", fmt.Errorf("failed connect to DatastoreExternal: %s", res.Status)
// 	}

// 	response := struct {
// 		Data struct {
// 			FactID int64 `json:"indicator_to_mo_fact_id"`
// 		} `json:"DATA"`
// 	}{}

// 	span.AddEvent("read response data")
// 	resBody, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	span.AddEvent("unmarshal response data")
// 	if err := json.Unmarshal(resBody, &response); err != nil {
// 		return "", err
// 	}

// 	return fmt.Sprintf("%d", response.Data.FactID), nil
// }

// type label struct {
// 	Tag   tag    `json:"tag"`
// 	Value string `json:"value"`
// }

// type tag struct {
// 	Id   int64  `json:"id"`
// 	Name string `json:"name"`
// 	Key  string `json:"key"`
// }

// // getSupertags get supertags
// func (x *operationDatastoreExternal) getSupertags() (tags map[int64]tag, err error) {
// 	req, err := http.NewRequest("GET", fmt.Sprintf("%s/_api/supertags?auth_user_id=%d", x.conf.URL, x.conf.AuthUserId), nil)
// 	if err != nil {
// 		return
// 	}
// 	req.Header.Set("Authorization", "Bearer "+x.conf.Token)

// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return
// 	}
// 	if res.StatusCode >= 400 {
// 		return nil, fmt.Errorf("failed connect to DatastoreExternal: %s", res.Status)
// 	}

// 	response := struct {
// 		Data struct {
// 			Rows []tag `json:"rows"`
// 		} `json:"DATA"`
// 	}{}
// 	resBody, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return
// 	}
// 	err = json.Unmarshal(resBody, &response)
// 	if err != nil {
// 		return
// 	}
// 	tags = make(map[int64]tag, len(response.Data.Rows))
// 	for _, tag := range response.Data.Rows {
// 		tags[tag.Id] = tag
// 	}
// 	return
// }

// // isExist check the existence of a fact by tag operationId
// func (x *operationDatastoreExternal) findFact(o *entities.Operation) (int64, error) {
// 	buf, boundary, err := makeGetFactsMultypartFormData(o, x.conf.IndicatorToMoId)
// 	if err != nil {
// 		return 0, err
// 	}

// 	req, err := http.NewRequest("POST", fmt.Sprintf("%s/_api/indicators/get_facts?auth_user_id=%d", x.conf.URL, x.conf.AuthUserId), buf)
// 	if err != nil {
// 		return 0, err
// 	}
// 	req.Header.Set("Authorization", "Bearer "+x.conf.Token)
// 	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return 0, err
// 	}
// 	if res.StatusCode >= 400 {
// 		return 0, fmt.Errorf("failed connect to DatastoreExternal: %s", res.Status)
// 	}

// 	response := struct {
// 		Data struct {
// 			Rows []struct {
// 				FactId    int64 `json:"indicator_to_mo_fact_id"`
// 				SuperTags []struct {
// 					Tag   tag    `json:"tag"`
// 					Value string `json:"value"`
// 				} `json:"supertags"`
// 			} `json:"rows"`
// 		} `json:"DATA"`
// 	}{}
// 	resBody, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return 0, err
// 	}
// 	err = json.Unmarshal(resBody, &response)
// 	if err != nil {
// 		return 0, err
// 	}
// 	for _, fact := range response.Data.Rows {
// 		for _, tag := range fact.SuperTags {
// 			if tag.Value == o.OperationId {
// 				return fact.FactId, nil
// 			}
// 		}
// 	}
// 	return 0, nil
// }

// // makeOperationMultypartFormData make multypart/formdata and boundary for facts
// func makeOperationMultypartFormData(o *entities.Operation, itmo_id, fact_id int64, labels string) (*bytes.Buffer, string, error) {
// 	date, _ := time.Parse(time.RFC3339, o.TrxnPostDate)
// 	dates := date.Format("2006-01-02")

// 	return utils.MultypartFormData(map[string]string{
// 		"period_start":            dates,
// 		"period_end":              dates,
// 		"period_key":              "day",
// 		"indicator_to_mo_id":      strconv.Itoa(int(itmo_id)),
// 		"indicator_to_mo_fact_id": strconv.Itoa(int(fact_id)),
// 		"value":                   o.AccountAmount,
// 		"fact_time":               dates,
// 		"is_plan":                 "0",
// 		"comment":                 "Тинькофф: " + o.PayPurpose,
// 		"supertags":               labels,
// 	})
// }

// // makeGetFactsMultypartFormData make multypart/formdata and boundary for facts
// func makeGetFactsMultypartFormData(o *entities.Operation, itmo_id int64) (data *bytes.Buffer, boundary string, err error) {
// 	date, _ := time.Parse(time.RFC3339, o.TrxnPostDate)
// 	dates := date.Format("2006-01-02")

// 	return utils.MultypartFormData(map[string]string{
// 		"period_start":       dates,
// 		"period_end":         dates,
// 		"period_key":         "day",
// 		"indicator_to_mo_id": strconv.Itoa(int(itmo_id)),
// 		"data_type":          "facts",
// 	})
// }

// // makeLabels make labels by config and supertags
// func makeLabels(o *entities.Operation, tags map[int64]tag) []byte {
// 	labels := []label{}

// 	// add default labels
// 	for key, val := range config.Default.Labels.Default {
// 		id, err := strconv.ParseInt(key, 10, 64)
// 		if err != nil {
// 			log.Warn(err.Error())
// 			continue
// 		}
// 		labels = append(labels, label{
// 			Tag:   tags[id],
// 			Value: val,
// 		})
// 	}

// 	// add custom labels
// 	for _, labelParam := range config.Default.Labels.Custom {
// 		// check label param
// 		if labelParam.ID < 1 {
// 			log.Warn(fmt.Sprintf("failed, label.id < 0"))
// 			continue
// 		}
// 		if labelParam.Name == "" {
// 			log.Warn(fmt.Sprintf("failed, label.name is empty"))
// 			continue
// 		}

// 		// read field value
// 		obj := reflect.ValueOf(o)
// 		if obj.Kind() == reflect.Pointer {
// 			obj = obj.Elem()
// 		}
// 		tagParts := strings.Split(labelParam.Name, ".")
// 		var value string
// 		for i, tag := range tagParts {
// 			fieldName := strings.ToUpper(tag[0:1]) + tag[1:]
// 			if i < len(tagParts)-1 {
// 				obj = obj.FieldByName(fieldName)
// 				continue
// 			}
// 			value = obj.FieldByName(fieldName).String()
// 		}

// 		if v, ok := labelParam.Vars[value]; ok {
// 			value = v
// 		}
// 		labels = append(labels, label{
// 			Tag:   tags[labelParam.ID],
// 			Value: value,
// 		})
// 	}

// 	// marshal labels
// 	data, err := json.Marshal(labels)
// 	if err != nil {
// 		log.Warn(fmt.Sprintf("failed marshal labels: %s", err.Error()))
// 	}
// 	return data
// }
