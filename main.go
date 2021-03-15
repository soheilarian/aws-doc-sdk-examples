package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
)

const authFileName string = ".echobeeAuth.txt"
const tokenFileName string = ".ecobeeToken.txt"
const apiKey string = "nIREGqvNiBOJoXYoOoMuvnKpe6EefVmO"
const AUTH_URL string = "https://api.ecobee.com/token"

var folderSeperator string
var operationSystem string
var echobeePin string
var authCode string
var homePath string
var authFile string
var tokenFile string

func main() {
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	log.Println("Starting Zone Split")

	initilize()

	os.Exit(0)
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		//get info to prapare for registration
		log.Println("Could not retrive previus Auths. Auth file does not exist: " + authFile)
		getAuth()
	} else {
		log.Println("Auth File exists")
		loadAuthFile()
	}

	os.Exit(0)
}

func getToken() {
	fmt.Println("Getting the Token")
	data := url.Values{
		"grant_type": {"ecobeePin"},
		"code":       {authCode},
		"client_id":  {apiKey},
	}
	//return json.NewDecoder(r.Body).Decode(target)

	authData := postReq(AUTH_URL, data)
	fmt.Println("AAA", authData)
	return

	if authData["error"] != nil {
		log.Println("Error: " + authData["error"].(string))
		log.Println("Error Description: " + authData["error_description"].(string))
		log.Println("Error uri: " + authData["error_uri"].(string))

		if authData["error"] == "invalid_grant" {
			log.Println("The Key has expiered. Refereshing the key...")
			deleteFile(authFile)
			getAuth()
		} else if authData["error"] == "authorization_pending" {
			fmt.Println("- Please authorize echobee to use the app")
			fmt.Println("- Please login to https://www.ecobee.com")
			fmt.Println("- Navigate to: MyApps --> Add Apps")
			fmt.Println("- Enter the code: " + echobeePin)
			fmt.Println("- Press Validate")
			//fmt.Println("- Please login to: https://www.ecobee.com/consumerportal/index.html#/my-apps/add/newv")
		} else {
			fmt.Println("AAA", authData)
		}
	}
}

func getReq(url string) string {
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	log.Printf(sb)

	return sb
}

func postReq(url string, data map[string][]string) map[string]interface{} {
	resp, err := http.PostForm("https://api.ecobee.com/token", data)

	if err != nil {
		log.Fatal(err)
	}

	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	//fmt.Println(res["error"])
	//fmt.Println(res["error_description"])
	tokenObj := new(TokenObj)
	json.NewDecoder(resp.Body).Decode(tokenObj)
	log.Print("========================")
	log.Println(tokenObj.Error)
	//authObj := new(AuthObj)
	//getJson(pinUrl, authObj)
	log.Print("========================")

	return res
}

func writeFile(text string, file string) {
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

func deleteFile(path string) {
	var err = os.Remove(path)
	if isError(err) {
		return
	}

	log.Println("File Deleted")
}

func isError(err error) bool {
	if err != nil {
		fmt.Println(err.Error())
	}

	return (err != nil)
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func getAuth() {
	fmt.Println("Getting the key")
	pinUrl := "https://api.ecobee.com/authorize?response_type=ecobeePin&client_id=" + apiKey + "&scope=smartWrite"

	authObj := new(AuthObj)
	getJson(pinUrl, authObj)

	//fmt.Println("Echobee PIN: " + pinObj.EcobeePin)
	//fmt.Println("Ecobee Auth Code: " + pinObj.Code)
	//fmt.Println(fmt.Sprint("Ecobee Exipery Inteval: ", pinObj.Interval))
	//fmt.Println(fmt.Sprint("Echobee Expiery in Sec: ", pinObj.Expires_in))
	//fmt.Println("Call Scope" + pinObj.Scope)

	echobeePin = authObj.EcobeePin
	authCode = authObj.Code
	deleteFile(authFile)
	writeFile("AUTH_CODE="+authCode+"\nECHOBEE_PIN="+echobeePin, authFile)
	printAuthValues()
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func loadAuthFile() {
	log.Println("Loading the Auth file")
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

func initilize() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homePath = usr.HomeDir

	//Setup Home based on OS
	operationSystem = runtime.GOOS
	log.Println("Runtime Operation System: " + operationSystem)
	if operationSystem == "windows" { // WINDOWS
		authFile = homePath + "\\" + authFileName
		tokenFile = homePath + "\\" + tokenFileName
		folderSeperator = "\\"
	} else if operationSystem == "darwin" { //MAC
		authFile = homePath + "/" + authFileName
		tokenFile = homePath + "/" + tokenFileName
		folderSeperator = "/"
	} else {
		log.Fatalln("RUNTIME GOOS (runtime.GOOS) is not undrestood")
		os.Exit(0)
	}

	//Load Token and auth Files. Loop till all is ready
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		log.Println("Auth File does not exist it should be created.")
		getAuth()
	} else {
		loadAuthFile()
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		log.Println("Token File does not exist it should be created.")
		getToken()
	} else {
		log.Println("TODO: process the token file")
	}
	printAuthValues()

	//Load Auth File: to be changed with DB items
	//loadAuthData()
}

func printAuthValues() {
	fmt.Println("------------------------------")
	fmt.Println("Application Key: " + apiKey)
	fmt.Println("Authorization Code is: " + authCode)
	fmt.Println("Echoobe PIN: " + echobeePin)
	fmt.Println("------------------------------")
}

type AuthObj struct {
	EcobeePin  string
	Code       string
	Interval   int
	Expires_in int
	Scope      string
}
type TokenObj struct {
	Error             string
	Error_description string
	Error_uri         int
	Expires_in        int
	Scope             string
}
