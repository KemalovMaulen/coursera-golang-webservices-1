package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type DataRow struct {
	Id            int       `xml:"id"`
	Guid          string    `xml:"guid"`
	IsActive      bool      `xml:"isActive"`
	Balance       string   `xml:"balance"`
	Picture       string    `xml:"picture"`
	Age           int       `xml:"age"`
	EyeColor      string    `xml:"eyeColor"`
	Name          string    `xml:"first_name"`
	Surname       string    `xml:"last_name"`
	Gender        string    `xml:"gender"`
	Company       string    `xml:"company"`
	Email         string    `xml:"email"`
	Phone         string    `xml:"phone"`
	Address       string    `xml:"address"`
	About         string    `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string    `xml:"favoriteFruit"`
}

func (r DataRow) convert() User {
	return User{
		Id:     r.Id,
		Name:   r.Name,
		Age:    r.Age,
		About:  r.About,
		Gender: r.Gender,
	}
}

type DataStruct struct {
	Data []DataRow `xml:"row"`
}

type TestCase struct {
	Req         SearchRequest
	ErrorMesHas string
	AccessToken string
	Url         string
}

const (
	InvalidToken = "invalidToken"
	TimeoutQuery = "timeoutQuery"
	InternalServerError = "internalServerError"
	BadRequestError = "badRequestError"
	InvalidJson = "invalidJson"
	BadRequestUnknown = "badRequestUnknown"
)

func SearchServer(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("AccessToken") == InvalidToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if r.URL.Query().Get("query") == TimeoutQuery {
		time.Sleep(time.Second * 2)
		w.WriteHeader(http.StatusFound)
		return
	}

	if r.URL.Query().Get("query") == InternalServerError {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if r.URL.Query().Get("query") == InvalidJson {
		w.Write([]byte("invalid_json"))
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Query().Get("query") == BadRequestError {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("query") == BadRequestUnknown {
		resp, _ := json.Marshal(SearchErrorResponse{"UnknownError"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	if  r.URL.Query().Get("order_field") != "Name" && r.URL.Query().Get("order_field") != "Age" +
		"" && r.URL.Query().Get("order_field") != "Id" && r.URL.Query().Get("order_field") != "" {
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	/*switch  r.URL.Query().Get("order_field") {
		case "Name", "", "Age", "Id":
	default:
		resp, err := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
		if err != nil {
			fmt.Println("ErrorBadOrderField->Marshal err: ", err)
			panic("")
			return
		}
		w.Write(resp)
		w.WriteHeader(http.StatusBadRequest)
		return
	}*/

	//if r.URL.Query().Get("query") == BadOrderFieldError {
	//	resp, err := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
	//	if err != nil {
	//		fmt.Println("ErrorBadOrderField->Marshal err: ", err)
	//		panic("")
	//	}
	//	w.Write(resp)
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}

	xmlFile, err := os.Open("dataset.xml")
	if err != nil {
		fmt.Println("os.Open err: ", err)
		panic("")
		return
	}
	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		fmt.Println("ReadAll err: ", err)
		panic("")
		return
	}
	dataStruct := DataStruct{}
	err = xml.Unmarshal(byteValue, &dataStruct)
	if err != nil {
		fmt.Println("Unmarshal err: ", err)
		panic("")
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	dataStruct.Data = dataStruct.Data[offset:limit]
	var users []User
	for _, data := range dataStruct.Data {
		user := data.convert()
		users = append(users, user)
	}
	responce, err := json.Marshal(users)
	if err != nil {
		fmt.Println("Marshal err:", err)
		panic("")
		return
	}
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responce)
}

func TestFindUsersWithErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	cases := []TestCase{

		{
			Req:         SearchRequest{Limit: -1},
			ErrorMesHas: "limit must be > 0",
		},
		{
			Req:         SearchRequest{Offset: -1},
			ErrorMesHas: "offset must be > 0",
		},
		{
			Req:         SearchRequest{Limit: 1},
			Url:         "http://",
			ErrorMesHas: "unknown error",
		},
		{
			Req:         SearchRequest{Query: TimeoutQuery},
			ErrorMesHas: "timeout for",
		},
		{
			AccessToken: InvalidToken,
			ErrorMesHas: "Bad AccessToken",
		},
		{
			Req:         SearchRequest{Query: InternalServerError},
			ErrorMesHas: "SearchServer fatal error",
		},
		{
			Req:         SearchRequest{Query: BadRequestError},
			ErrorMesHas: "cant unpack error json",
		},
		{
			Req:         SearchRequest{Query: BadRequestUnknown},
			ErrorMesHas: "unknown bad request error",
		},
		{
			Req:         SearchRequest{Query: InvalidJson},
			ErrorMesHas: "cant unpack result json",
		},
		{
			Req:         SearchRequest{OrderField: "order_field"},
			ErrorMesHas: "OrderFeld order_field invalid",
		},
	}

	for caseNum, item := range cases {
		url := server.URL
		if item.Url != "" {
			url = item.Url
		}

		client := SearchClient{
			URL:         url,
			AccessToken: item.AccessToken,
		}
		resp, err := client.FindUsers(item.Req)

		if resp != nil || err == nil {
			fmt.Printf("item: %+v\n", item)
			t.Errorf("expected resp: got error in test %d ", caseNum)
		}

		if item.ErrorMesHas == "" && !strings.Contains(err.Error(), item.ErrorMesHas) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.ErrorMesHas, err.Error())
		}
	}

}

func TestFindUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	cases := []TestCase{
		TestCase{
			Req: SearchRequest{Limit: 5},
		},
		TestCase{
			Req: SearchRequest{Limit: 25, Offset: 1},
		},
		TestCase{
			Req: SearchRequest{Limit: 26, Offset: 1},
		},
	}
	for caseNum, item := range cases {

		client := SearchClient{
			URL: server.URL,
		}
		resp, err := client.FindUsers(item.Req)
		if resp == nil || err != nil {
			t.Errorf("expected resp: got error in test %d", caseNum)
		}

	}
}
