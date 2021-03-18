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
	"reflect"
	"runtime"
	"strings"
)

const authFileName string = ".ecobeeAuth.txt"
const tokenFileName string = ".ecobeeToken.txt"
const apiKey string = "nIREGqvNiBOJoXYoOoMuvnKpe6EefVmO"
const AUTH_URL string = "https://api.ecobee.com/token"

var folderSeperator string
var operationSystem string
var ecobeePin string
var authCode string
var accessToken string
var refreshToken string
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

	req, err := http.NewRequest("GET", "https://api.ecobee.com/1/thermostat?format=json&body={\"selection\":{\"selectionType\":\"registered\",\"selectionMatch\":\"\",\"includeRuntime\":true,\"includeSensors\":true}}", nil)
	if err != nil {
		// handle err
	}
	bearer := "Bearer " + accessToken

	req.Header.Add("Content-Type", "text/json")
	//req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Add("Authorization", bearer)

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error on response.\n[ERROR] -", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error while reading the response bytes:", err)
	}
	log.Println(string([]byte(body)))
	log.Println("------------------------")
	//log.Println(string([]byte(body."page")))
	getFieldName("thermostatList")
}

func getFieldName(tag, key string, s interface{}) (fieldname string) {
	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		panic("bad type")
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(f.Tag.Get(key), ",")[0] // use split to ignore tag "options" like omitempty, etc.
		if v == tag {
			return f.Name
		}
	}
	return ""
}

func refreshAccessToken() {
	log.Println("Refreshing the Token")

	data := url.Values{
		"grant_type": {"refresh_token"},
		"code":       {refreshToken},
		"client_id":  {apiKey},
	}

	authData := postReq(AUTH_URL, data)
	//fmt.Println("fullpayload", authData)

	if authData["error"] != nil {
		log.Println("Error: " + authData["error"].(string))
		log.Println("Error Description: " + authData["error_description"].(string))
		log.Println("Error uri: " + authData["error_uri"].(string))

		if authData["error"] == "invalid_grant" {
			//This is the case that app is registered but token is lost. we need to start over
			log.Println("The Key has expiered. or the value passed are wrong")
		}
	} else {
		accessToken = authData["access_token"].(string)
		deleteFile(tokenFile)
		writeFile(tokenFile, "ACCESS_TOKEN="+accessToken+"\nREFRESH_TOKEN="+refreshToken)
	}
}

func getToken() {
	fmt.Println("Getting the Token")
	data := url.Values{
		"grant_type": {"ecobeePin"},
		"code":       {authCode},
		"client_id":  {apiKey},
	}

	authData := postReq(AUTH_URL, data)
	fmt.Println("AAA", authData)
	fmt.Println("AAA", authData["error"])

	if authData["error"] != nil {
		log.Println("Error: " + authData["error"].(string))
		log.Println("Error Description: " + authData["error_description"].(string))
		log.Println("Error uri: " + authData["error_uri"].(string))

		if authData["error"] == "invalid_grant" {
			//This is the case that app is registered but token is lost. we need to start over
			log.Println("The Key has expiered. Whta to do?")
			deleteFile(authFile)
			//getAuth()

		} else if authData["error"] == "invalid_client" {
			log.Println("the token and app wont match")
			deleteFile(authFile)

		} else if authData["error"] == "authorization_pending" {
			fmt.Println("- Please authorize ecobee to use the app")
			fmt.Println("- Please login to https://www.ecobee.com")
			fmt.Println("- Navigate to: MyApps --> Add Apps")
			fmt.Println("- Enter the code: " + ecobeePin)
			fmt.Println("- Click Validate")
			//fmt.Println("- Please login to: https://www.ecobee.com/consumerportal/index.html#/my-apps/add/newv")
		}
	} else {
		accessToken = authData["access_token"].(string)
		refreshToken = authData["refresh_token"].(string)
		writeFile(tokenFile, "ACCESS_TOKEN="+accessToken+"\nREFRESH_TOKEN="+refreshToken)
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
	return res
}

func writeFile(path string, text string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		log.Println(err)
	}
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

	fmt.Println("ecobee PIN: " + authObj.EcobeePin)
	fmt.Println("Ecobee Auth Code: " + authObj.Code)
	//fmt.Println(fmt.Sprint("Ecobee Exipery Inteval: ", pinObj.Interval))
	//fmt.Println(fmt.Sprint("Ecobee Expiery in Sec: ", pinObj.Expires_in))
	//fmt.Println("Call Scope" + pinObj.Scope)
	ecobeePin = authObj.EcobeePin
	authCode = authObj.Code
	writeFile(authFile, "AUTH_CODE="+authCode+"\nECOBEE_PIN="+ecobeePin)
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
		} else if property == "ECOBEE_PIN" {
			ecobeePin = strings.Split(scanner.Text(), "=")[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func loadTokenFile() {
	log.Println("Loading the Token file")
	var property string

	file, err := os.Open(tokenFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		property = strings.Split(scanner.Text(), "=")[0]
		if property == "ACCESS_TOKEN" {
			accessToken = strings.Split(scanner.Text(), "=")[1]
		} else if property == "REFRESH_TOKEN" {
			refreshToken = strings.Split(scanner.Text(), "=")[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func initilize() {
	setHomePath()

	setupAuth()
	setupToken()
	refreshAccessToken()
	printAuthValues()

}

func setupAuth() {
	if _, err := os.Stat(authFile); os.IsNotExist(err) {
		log.Println("Auth File does not exist it should be created.")
		getAuth()
	} else {
		loadAuthFile()
	}
}

func setupToken() {
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		log.Println("Token File does not exist it should be created.")
		getToken()
	} else {
		log.Println("TODO: process the token file")
		loadTokenFile()
	}
}

func setHomePath() {
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
}

func printAuthValues() {
	fmt.Println("------------------------------")
	fmt.Println("Application Key: " + apiKey)
	fmt.Println("Authorization Code is: " + authCode)
	fmt.Println("Ecoobe PIN: " + ecobeePin)
	fmt.Println("ACCESS TOKEN: " + accessToken)
	fmt.Println("REFRESH TOKEN: " + refreshToken)
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
