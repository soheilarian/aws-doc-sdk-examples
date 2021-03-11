package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

var operationSystem string
var echobeePin string
var authCode string
var homePath string
var authFile string

var authFileName string = ".echobeeAuth.txt"
var apiKey string = "nIREGqvNiBOJoXYoOoMuvnKpe6EefVmO"

func main() {
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------=s")

	setHomePath()
	operationSystem = runtime.GOOS
	log.Println("runtime.GOOS (operationSystem): " + operationSystem)
	if operationSystem == "windows" {
		fmt.Println("Hello from Windows")
		authFile = homePath + "\\" + authFileName
	} else {
		fmt.Println("RUNTIME GOOS (runtime.GOOS) is not undrestood")
		authFile = homePath + "/" + authFileName
	}

	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		log.Println("Could not retrive previus Auths. Auth file does not exist: " + authFile)
		getKey()
	} else {
		log.Println("Auth File exists")
		loadAuthData()
	}

	fmt.Println("Application Key: " + apiKey)
	fmt.Println("Authorization Code is: " + authCode)
	fmt.Println("Echoobe PIN: " + echobeePin)

	getAuth()
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
	//deleteFile(authFile)
	appendFile("AUTH_CODE="+authCode+"\nECHOBEE_PIN="+echobeePin, authFile)
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

		if authObj.Error == "authorization_pending" {
			fmt.Println("-------------------------------------------")
			fmt.Println("- Please authorize echobee to use the app")
			fmt.Println("- Please login to: https://www.ecobee.com/consumerportal/index.html#/my-apps/add/newv")
			fmt.Println("- with code: " + echobeePin)
			fmt.Println("-------------------------------------------")
		}
	}
}
func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func loadAuthData() {
	var property string

	file, err := os.Open(authFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		property = strings.Split(scanner.Text(), "=")[0]
		if property == "AUTH_CODE" {
			authCode = strings.Split(scanner.Text(), "=")[1]
		} else if property == "ECHOBEE_PIN" {
			echobeePin = strings.Split(scanner.Text(), "=")[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Println(err)
	}
}

func setHomePath() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homePath = usr.HomeDir
}
