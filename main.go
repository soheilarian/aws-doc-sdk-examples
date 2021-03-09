package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var authFile string = "/Users/soheil/.echobeeAuth.txt"
var apiKey string = "nIREGqvNiBOJoXYoOoMuvnKpe6EefVmO"
var echobeePin string
var authCode string

func main() {
	fmt.Println("------------------------------")
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		log.Println("Could not retrive previus Auths. Auth file does not exist: " + authFile)
		getKey()
		appendFile("AUTH_CODE="+authCode+"\nECHOBEE_PIN="+echobeePin, authFile)
	} else {
		log.Println("Auth File exists")
	}

	//authCode = "lQNgAJsj76e94hDCc47tYLfr"
	//echobeePin = "RHTQ-NGPP "

	//fmt.Println("Application Key: " + apiKey)
	//fmt.Println("Authorization Code is: " + authCode)

	///////getKey()
	//getAuth()
	///////authCode = pinObj.Code
}
func appendFile(text string, file string) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		touchFile(file)
	}

	f, err := os.OpenFile(authFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintln(f, text)
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

//func postJson(url string, target interface{}) error {
//http.Post()
//r, err := myClient.Post(url,"dd",)
//if err != nil {
//	return err
//}
//defer r.Body.Close()
//return json.NewDecoder(r.Body).Decode(target)
//}

type PinObj struct {
	EcobeePin  string
	Code       string
	Interval   int
	Expires_in int
	Scope      string
}

type AuthObj struct {
	Error             string
	Error_description string
	Error_uri         string
}

func getKey() {
	fmt.Println("Getting the key")
	pinUrl := "https://api.ecobee.com/authorize?response_type=ecobeePin&client_id=" + apiKey + "&scope=smartWrite"

	pinObj := new(PinObj)
	getJson(pinUrl, pinObj)

	//fmt.Println("Echobee PIN: " + pinObj.EcobeePin)
	//fmt.Println("Ecobee Auth Code: " + pinObj.Code)
	//fmt.Println(fmt.Sprint("Ecobee Exipery Inteval: ", pinObj.Interval))
	//fmt.Println(fmt.Sprint("Echobee Expiery in Sec: ", pinObj.Expires_in))
	//fmt.Println("Call Scope" + pinObj.Scope)

	echobeePin = pinObj.EcobeePin
	authCode = pinObj.Code
}

func getAuth() {
	fmt.Println("Getting the Auth")

	var data = "grant_type=ecobeePin&code=" + authCode + "&client_id=" + apiKey
	url := "https://api.ecobee.com/token?" + data
	fmt.Println("URL: " + url)

	authObj := new(AuthObj)
	getJson(url, authObj)

	//Proper error catching
	if authObj.Error == "" {
		fmt.Println("Invalid Auth Call: " + url)
		getKey()
	} else {
		//invalid_grant
		fmt.Println("Auth Error: " + authObj.Error)
		fmt.Println("Auth Error Description: " + authObj.Error_description)
		fmt.Println("Auth Error URL: " + authObj.Error_uri)
	}
}
func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

////test
