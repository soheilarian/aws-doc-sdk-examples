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
	"strconv"
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
var err error

func main() {
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	fmt.Println("------------------------------")
	log.Println("Starting Zone Split")

	initilize()

	var term EcobeeThermostatData = fetchThermostatObj()
	fmt.Printf("%+v\n", term.Thermostatlist[0])
	room := new(Room)
	//This must ask for mapping for the first time
	//This must map as pointer
	room.Name = "Aram's Room"
	room.sensorData.Name = term.Thermostatlist[0].Remotesensors[0].Name
	room.sensorData.Temprature, err = strconv.ParseFloat(term.Thermostatlist[0].Remotesensors[0].Capability[0].Value, 32)
	if err != nil {
		log.Panicln(err)
	}

	sensorCount := 0
	thermostatCount := len(term.Thermostatlist)
	for i := 0; i < thermostatCount; i++ {
		sensorCount += len(term.Thermostatlist[i].Remotesensors)
	}
	aa := make(map[string]Room)
	//r := new(Room)
	rr := new(Room)
	rr.Name = "TTTT"
	rr.ControlledRegister = true
	rr.sensorData.Name = "BLA BLA"
	//a["B"] = *new(Room)
	aa["A"] = *rr
	fmt.Printf("%+v\n", aa)

	//How can we use sensorCount?
	a := make(map[string]int)
	a["A"] = 1
	a["B"] = 2
	fmt.Println(a)
	var rooms [4]Room
	rooms[0].Name = "AA"

	fmt.Printf("Termostat Count: %d\n", thermostatCount)
	fmt.Printf("Remote Senror: %d\n", sensorCount)

	for t := 0; t < len(term.Thermostatlist); t++ {
		for s := 0; s < len(term.Thermostatlist[t].Remotesensors); s++ {

		}
	}

	fmt.Println("------------------------------")
	for t := 0; t < len(term.Thermostatlist); t++ {
		fmt.Printf("Termostat Name is: %s\n", term.Thermostatlist[t].Name)
		fmt.Printf("Termostat Temreture is: %.1fF\n", float64(term.Thermostatlist[t].Runtime.Actualtemperature)/10)
		fmt.Printf("Termostat Humidity is: %d%%\n", term.Thermostatlist[t].Runtime.Actualhumidity)
		fmt.Printf("Termostat Remote Sensor Count: %d\n", len(term.Thermostatlist[t].Remotesensors))
		for s := 0; s < len(term.Thermostatlist[t].Remotesensors); s++ {
			fmt.Printf("Sonsor Name is: %s\n", term.Thermostatlist[t].Remotesensors[s].Name)
			fmt.Printf("Sonsor Temprature is: %sF\n", term.Thermostatlist[t].Remotesensors[s].Capability[0].Value)
		}
		fmt.Println("+++++++++++++++++++++++++++++++++++++++++")
	}
	fmt.Printf("%+v\n", room)
}

func fetchThermostatObj() EcobeeThermostatData {
	temostatJson := fetchTemostatJason()
	thermostat := new(EcobeeThermostatData)

	if err := json.Unmarshal(temostatJson, &thermostat); err != nil {
		panic(err)
	}
	return *thermostat
}

func fetchTemostatJason() []byte {
	req, err := http.NewRequest("GET", "https://api.ecobee.com/1/thermostat?format=json&body={\"selection\":{\"selectionType\":\"registered\",\"selectionMatch\":\"\",\"includeRuntime\":true,\"includeSensors\":true}}", nil)
	if err != nil {
		log.Panicln(err)
	}
	bearer := "Bearer " + accessToken

	req.Header.Add("Content-Type", "text/json")
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
	//res := string([]byte(body))
	res := []byte(body)

	return res
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

type EcobeeThermostatData struct {
	Page struct {
		Page       int `json:"page"`
		Totalpages int `json:"totalPages"`
		Pagesize   int `json:"pageSize"`
		Total      int `json:"total"`
	} `json:"page"`
	Thermostatlist []struct {
		Identifier     string `json:"identifier"`
		Name           string `json:"name"`
		Thermostatrev  string `json:"thermostatRev"`
		Isregistered   bool   `json:"isRegistered"`
		Modelnumber    string `json:"modelNumber"`
		Brand          string `json:"brand"`
		Features       string `json:"features"`
		Lastmodified   string `json:"lastModified"`
		Thermostattime string `json:"thermostatTime"`
		Utctime        string `json:"utcTime"`
		Runtime        struct {
			Runtimerev         string `json:"runtimeRev"`
			Connected          bool   `json:"connected"`
			Firstconnected     string `json:"firstConnected"`
			Connectdatetime    string `json:"connectDateTime"`
			Disconnectdatetime string `json:"disconnectDateTime"`
			Lastmodified       string `json:"lastModified"`
			Laststatusmodified string `json:"lastStatusModified"`
			Runtimedate        string `json:"runtimeDate"`
			Runtimeinterval    int    `json:"runtimeInterval"`
			Actualtemperature  int    `json:"actualTemperature"`
			Actualhumidity     int    `json:"actualHumidity"`
			Rawtemperature     int    `json:"rawTemperature"`
			Showiconmode       int    `json:"showIconMode"`
			Desiredheat        int    `json:"desiredHeat"`
			Desiredcool        int    `json:"desiredCool"`
			Desiredhumidity    int    `json:"desiredHumidity"`
			Desireddehumidity  int    `json:"desiredDehumidity"`
			Desiredfanmode     string `json:"desiredFanMode"`
			Desiredheatrange   []int  `json:"desiredHeatRange"`
			Desiredcoolrange   []int  `json:"desiredCoolRange"`
		} `json:"runtime"`
		Remotesensors []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Type       string `json:"type"`
			Code       string `json:"code,omitempty"`
			Inuse      bool   `json:"inUse"`
			Capability []struct {
				ID    string `json:"id"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"capability"`
		} `json:"remoteSensors"`
	} `json:"thermostatList"`
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

type Room struct {
	Name               string
	ControlledRegister bool
	sensorData         struct {
		Name           string
		Temprature     float64
		Humidity       int
		ThermostatName string
	}
}
