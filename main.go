package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
)

const authFileName string = ".echobeeAuth.txt"

//const tokenFileName string = ".ecobeeToken.txt"
const apiKey string = "nIREGqvNiBOJoXYoOoMuvnKpe6EefVmO"
const AUTH_URL string = "https://api.ecobee.com/token"

var operationSystem string
var echobeePin string
var authCode string
var homePath string
var authFile string

func main() {
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	log.Println("Starting Zone Split")

	setHomePath()
	operationSystem = runtime.GOOS
	log.Println("Runtime Operation System: " + operationSystem)
	if operationSystem == "windows" { // WINDOWS
		fmt.Println("Hello from Windows")
		authFile = homePath + "\\" + authFileName
	} else if operationSystem == "darwin" { //MAC
		fmt.Println("Hello from Mac")
		authFile = homePath + "/" + authFileName
	} else {
		log.Fatalln("RUNTIME GOOS (runtime.GOOS) is not undrestood")
		os.Exit(0)
	}

	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		log.Println("Could not retrive previus Auths. Auth file does not exist: " + authFile)
		getKey()
	} else {
		log.Println("Auth File exists")
		loadAuthData()
	}

	//fmt.Println("------------------------------")
	//fmt.Println("Application Key: " + apiKey)
	//fmt.Println("Authorization Code is: " + authCode)
	//fmt.Println("Echoobe PIN: " + echobeePin)
	//fmt.Println("------------------------------")

	auth()
	os.Exit(0)

	//getAuth()

}

func auth() {
	fmt.Println("Getting the Auth")
	data := url.Values{
		"grant_type": {"ecobeePin"},
		"code":       {authCode},
		"client_id":  {apiKey},
	}

	authData := postReq(AUTH_URL, data)
	fmt.Println("AAA", authData)
	if authData["error"] != nil {
		log.Println("Error: " + authData["error"].(string))
		log.Println("Error Description: " + authData["error_description"].(string))
		log.Println("Error uri: " + authData["error_uri"].(string))

		if authData["error"] == "invalid_grant" {
			log.Println("The Key has expiered. Refereshing the key...")
			deleteFile(authFile)
			getKey()
		} else if authData["error"] == "authorization_pending" {
			fmt.Println("- Please authorize echobee to use the app")
			fmt.Println("- Please login to https://www.ecobee.com")
			fmt.Println("- Navigate to: MyApps --> Add Apps")
			fmt.Println("- Enter the code: " + echobeePin)
			fmt.Println("- Press Validate")
			//fmt.Println("- Please login to: https://www.ecobee.com/consumerportal/index.html#/my-apps/add/newv")
		}
	}

}

func postReq(url string, data map[string][]string) map[string]interface{} {
	resp, err := http.PostForm("https://api.ecobee.com/token", data)

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	fmt.Println(res["error"])
	fmt.Println(res["error_description"])
	return res
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

type PinObj struct {
	EcobeePin  string
	Code       string
	Interval   int
	Expires_in int
	Scope      string
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
	deleteFile(authFile)
	appendFile("AUTH_CODE="+authCode+"\nECHOBEE_PIN="+echobeePin, authFile)
	printAuthValues()
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
	printAuthValues()
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

func printAuthValues() {
	fmt.Println("------------------------------")
	fmt.Println("Application Key: " + apiKey)
	fmt.Println("Authorization Code is: " + authCode)
	fmt.Println("Echoobe PIN: " + echobeePin)
	fmt.Println("------------------------------")
}
