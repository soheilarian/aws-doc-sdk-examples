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

	log.Println("Starting Zone Split")

	initilize()
	room := make(map[string]Room)
	//Loop every 60 sec
	refreshTetmostatStruct(*&room)

	log.Println("TODO: fix the room setting manually. use user input later")
	temRoom := room["Upstairs"]
	temRoom.ControlledRegister = false
	room["Upstairs"] = temRoom

	fmt.Printf("%+v\n", room)

	for key, element := range room {
		if element.ControlledRegister {
			if element.hvacMode == "off" {
				element.RegisterOpen = true
				fmt.Printf("The room %s's vent(s) is %s. Tempreture is %gF with thermostat off\n", key, "open", element.Temprature)
			} else {
				if element.hvacMode == "heat" {
					if element.Temprature < element.DesiredTemrature {
						element.RegisterOpen = true
					} else {
						element.RegisterOpen = false
					}
				} else if element.hvacMode == "cool" {
					if element.Temprature > element.DesiredTemrature {
						element.RegisterOpen = true
					} else {
						element.RegisterOpen = false
					}
				} else {
					fmt.Printf("cold not undrestand hvac mode: %s\n", element.hvacMode)
				}
				oc := "closed"
				if element.ControlledRegister && element.RegisterOpen {
					oc = "open"
				}
				fmt.Printf("The room %s's vent(s) is %s. Tempreture is %gF --> %gF\n", key, oc, element.Temprature, element.DesiredTemrature)
			}
		}
	}
	fmt.Println("---------------------------------------------")

	// End of loop
	//http.HandleFunc("/", index_handler)
	//http.ListenAndServe(":8080", nil)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "you are very goozoo")
}

func refreshTetmostatStruct(room map[string]Room) {
	var term EcobeeThermostatData = fetchThermostatObj()

	sensorCount := 0
	thermostatCount := len(term.Thermostatlist)
	for i := 0; i < thermostatCount; i++ {
		sensorCount += len(term.Thermostatlist[i].Remotesensors)
	}

	for t := 0; t < len(term.Thermostatlist); t++ {
		for s := 0; s < len(term.Thermostatlist[t].Remotesensors); s++ {
			tempRoom := new(Room)
			tempRoom.hvacMode = term.Thermostatlist[t].Settings.Hvacmode
			tempRoom.Name = term.Thermostatlist[t].Remotesensors[s].Name
			tempRoom.SensorName = term.Thermostatlist[t].Remotesensors[s].Name
			tempRoom.SensorID = term.Thermostatlist[t].Remotesensors[s].ID
			tempRoom.ControlledRegister = true
			if tempRoom.hvacMode == "heat" {
				tempRoom.DesiredTemrature = float64(term.Thermostatlist[t].Runtime.Desiredheat) / 10
			} else if tempRoom.hvacMode == "cool" {
				tempRoom.DesiredTemrature = float64(term.Thermostatlist[t].Runtime.Desiredcool) / 10
			}

			for i := 0; i < len(term.Thermostatlist[t].Remotesensors[s].Capability); i++ {
				if term.Thermostatlist[t].Remotesensors[s].Capability[i].Type == "temperature" {
					tempRoom.Temprature, err = strconv.ParseFloat(term.Thermostatlist[t].Remotesensors[s].Capability[i].Value, 32)
					if err != nil {
						log.Printf("We could not convert %s to a float value", term.Thermostatlist[t].Remotesensors[s].Capability[i].Value)
						tempRoom.Temprature = 0
					} else {
						tempRoom.Temprature = tempRoom.Temprature / 10
					}
				} else if term.Thermostatlist[t].Remotesensors[s].Capability[i].Type == "humidity" {
					tempRoom.Humidity, err = strconv.Atoi(term.Thermostatlist[t].Remotesensors[s].Capability[i].Value)
					if err != nil {
						log.Printf("We could not convert %s to a int value", term.Thermostatlist[t].Remotesensors[s].Capability[i].Value)
						tempRoom.Temprature = 0
					}
				}
			}
			room[tempRoom.Name] = *tempRoom
		}
	}
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
	//req, err := http.NewRequest("GET", "https://api.ecobee.com/1/thermostat?format=json&body={\"selection\":{\"selectionType\":\"registered\",\"selectionMatch\":\"\",\"includeRuntime\":true,\"includeSensors\":true}}", nil)
	req, err := http.NewRequest("GET", "https://api.ecobee.com/1/thermostat?format=json&body={\"selection\":{\"selectionType\":\"registered\",\"selectionMatch\":\"\",\"includeRuntime\":true,\"includeSensors\":true,\"includeSettings\":true}}", nil)

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
	//fmt.Printf("%+v\n", string([]byte(body)))
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

	if authData["error"] != nil {
		log.Println("Error: " + authData["error"].(string))
		log.Println("Error Description: " + authData["error_description"].(string))
		log.Println("Error uri: " + authData["error_uri"].(string))

		if authData["error"] == "invalid_grant" {
			//This is the case that app is registered but token is lost. we need to start over
			log.Println("The Key has expiered. Whta to do?")
			deleteFile(authFile)

		} else if authData["error"] == "invalid_client" {
			log.Println("the token and app wont match")
			deleteFile(authFile)

		} else if authData["error"] == "authorization_pending" {
			fmt.Println("- Please authorize ecobee to use the app")
			fmt.Println("- Please login to https://www.ecobee.com")
			fmt.Println("- Navigate to: MyApps --> Add Apps")
			fmt.Println("- Enter the code: " + ecobeePin)
			fmt.Println("- Click Validate")
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
	//printAuthValues()
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
		Settings       struct {
			Hvacmode                            string `json:"hvacMode"`
			Lastservicedate                     string `json:"lastServiceDate"`
			Serviceremindme                     bool   `json:"serviceRemindMe"`
			Monthsbetweenservice                int    `json:"monthsBetweenService"`
			Remindmedate                        string `json:"remindMeDate"`
			Vent                                string `json:"vent"`
			Ventilatorminontime                 int    `json:"ventilatorMinOnTime"`
			Serviceremindtechnician             bool   `json:"serviceRemindTechnician"`
			Eilocation                          string `json:"eiLocation"`
			Coldtempalert                       int    `json:"coldTempAlert"`
			Coldtempalertenabled                bool   `json:"coldTempAlertEnabled"`
			Hottempalert                        int    `json:"hotTempAlert"`
			Hottempalertenabled                 bool   `json:"hotTempAlertEnabled"`
			Coolstages                          int    `json:"coolStages"`
			Heatstages                          int    `json:"heatStages"`
			Maxsetback                          int    `json:"maxSetBack"`
			Maxsetforward                       int    `json:"maxSetForward"`
			Quicksavesetback                    int    `json:"quickSaveSetBack"`
			Quicksavesetforward                 int    `json:"quickSaveSetForward"`
			Hasheatpump                         bool   `json:"hasHeatPump"`
			Hasforcedair                        bool   `json:"hasForcedAir"`
			Hasboiler                           bool   `json:"hasBoiler"`
			Hashumidifier                       bool   `json:"hasHumidifier"`
			Haserv                              bool   `json:"hasErv"`
			Hashrv                              bool   `json:"hasHrv"`
			Condensationavoid                   bool   `json:"condensationAvoid"`
			Usecelsius                          bool   `json:"useCelsius"`
			Usetimeformat12                     bool   `json:"useTimeFormat12"`
			Locale                              string `json:"locale"`
			Humidity                            string `json:"humidity"`
			Humidifiermode                      string `json:"humidifierMode"`
			Backlightonintensity                int    `json:"backlightOnIntensity"`
			Backlightsleepintensity             int    `json:"backlightSleepIntensity"`
			Backlightofftime                    int    `json:"backlightOffTime"`
			Soundtickvolume                     int    `json:"soundTickVolume"`
			Soundalertvolume                    int    `json:"soundAlertVolume"`
			Compressorprotectionmintime         int    `json:"compressorProtectionMinTime"`
			Compressorprotectionmintemp         int    `json:"compressorProtectionMinTemp"`
			Stage1Heatingdifferentialtemp       int    `json:"stage1HeatingDifferentialTemp"`
			Stage1Coolingdifferentialtemp       int    `json:"stage1CoolingDifferentialTemp"`
			Stage1Heatingdissipationtime        int    `json:"stage1HeatingDissipationTime"`
			Stage1Coolingdissipationtime        int    `json:"stage1CoolingDissipationTime"`
			Heatpumpreversaloncool              bool   `json:"heatPumpReversalOnCool"`
			Fancontrolrequired                  bool   `json:"fanControlRequired"`
			Fanminontime                        int    `json:"fanMinOnTime"`
			Heatcoolmindelta                    int    `json:"heatCoolMinDelta"`
			Tempcorrection                      int    `json:"tempCorrection"`
			Holdaction                          string `json:"holdAction"`
			Heatpumpgroundwater                 bool   `json:"heatPumpGroundWater"`
			Haselectric                         bool   `json:"hasElectric"`
			Hasdehumidifier                     bool   `json:"hasDehumidifier"`
			Dehumidifiermode                    string `json:"dehumidifierMode"`
			Dehumidifierlevel                   int    `json:"dehumidifierLevel"`
			Dehumidifywithac                    bool   `json:"dehumidifyWithAC"`
			Dehumidifyovercooloffset            int    `json:"dehumidifyOvercoolOffset"`
			Autoheatcoolfeatureenabled          bool   `json:"autoHeatCoolFeatureEnabled"`
			Wifiofflinealert                    bool   `json:"wifiOfflineAlert"`
			Heatmintemp                         int    `json:"heatMinTemp"`
			Heatmaxtemp                         int    `json:"heatMaxTemp"`
			Coolmintemp                         int    `json:"coolMinTemp"`
			Coolmaxtemp                         int    `json:"coolMaxTemp"`
			Heatrangehigh                       int    `json:"heatRangeHigh"`
			Heatrangelow                        int    `json:"heatRangeLow"`
			Coolrangehigh                       int    `json:"coolRangeHigh"`
			Coolrangelow                        int    `json:"coolRangeLow"`
			Useraccesscode                      string `json:"userAccessCode"`
			Useraccesssetting                   int    `json:"userAccessSetting"`
			Auxruntimealert                     int    `json:"auxRuntimeAlert"`
			Auxoutdoortempalert                 int    `json:"auxOutdoorTempAlert"`
			Auxmaxoutdoortemp                   int    `json:"auxMaxOutdoorTemp"`
			Auxruntimealertnotify               bool   `json:"auxRuntimeAlertNotify"`
			Auxoutdoortempalertnotify           bool   `json:"auxOutdoorTempAlertNotify"`
			Auxruntimealertnotifytechnician     bool   `json:"auxRuntimeAlertNotifyTechnician"`
			Auxoutdoortempalertnotifytechnician bool   `json:"auxOutdoorTempAlertNotifyTechnician"`
			Disablepreheating                   bool   `json:"disablePreHeating"`
			Disableprecooling                   bool   `json:"disablePreCooling"`
			Installercoderequired               bool   `json:"installerCodeRequired"`
			Draccept                            string `json:"drAccept"`
			Isrentalproperty                    bool   `json:"isRentalProperty"`
			Usezonecontroller                   bool   `json:"useZoneController"`
			Randomstartdelaycool                int    `json:"randomStartDelayCool"`
			Randomstartdelayheat                int    `json:"randomStartDelayHeat"`
			Humidityhighalert                   int    `json:"humidityHighAlert"`
			Humiditylowalert                    int    `json:"humidityLowAlert"`
			Disableheatpumpalerts               bool   `json:"disableHeatPumpAlerts"`
			Disablealertsonidt                  bool   `json:"disableAlertsOnIdt"`
			Humidityalertnotify                 bool   `json:"humidityAlertNotify"`
			Humidityalertnotifytechnician       bool   `json:"humidityAlertNotifyTechnician"`
			Tempalertnotify                     bool   `json:"tempAlertNotify"`
			Tempalertnotifytechnician           bool   `json:"tempAlertNotifyTechnician"`
			Monthlyelectricitybilllimit         int    `json:"monthlyElectricityBillLimit"`
			Enableelectricitybillalert          bool   `json:"enableElectricityBillAlert"`
			Enableprojectedelectricitybillalert bool   `json:"enableProjectedElectricityBillAlert"`
			Electricitybillingdayofmonth        int    `json:"electricityBillingDayOfMonth"`
			Electricitybillcyclemonths          int    `json:"electricityBillCycleMonths"`
			Electricitybillstartmonth           int    `json:"electricityBillStartMonth"`
			Ventilatorminontimehome             int    `json:"ventilatorMinOnTimeHome"`
			Ventilatorminontimeaway             int    `json:"ventilatorMinOnTimeAway"`
			Backlightoffduringsleep             bool   `json:"backlightOffDuringSleep"`
			Autoaway                            bool   `json:"autoAway"`
			Smartcirculation                    bool   `json:"smartCirculation"`
			Followmecomfort                     bool   `json:"followMeComfort"`
			Ventilatortype                      string `json:"ventilatorType"`
			Isventilatortimeron                 bool   `json:"isVentilatorTimerOn"`
			Ventilatoroffdatetime               string `json:"ventilatorOffDateTime"`
			Hasuvfilter                         bool   `json:"hasUVFilter"`
			Coolinglockout                      bool   `json:"coolingLockout"`
			Ventilatorfreecooling               bool   `json:"ventilatorFreeCooling"`
			Dehumidifywhenheating               bool   `json:"dehumidifyWhenHeating"`
			Ventilatordehumidify                bool   `json:"ventilatorDehumidify"`
			Groupref                            string `json:"groupRef"`
			Groupname                           string `json:"groupName"`
			Groupsetting                        int    `json:"groupSetting"`
			Fanspeed                            string `json:"fanSpeed"`
		} `json:"settings"`
		Runtime struct {
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
	Temprature         float64
	Humidity           int
	ControlledRegister bool
	RegisterOpen       bool
	DesiredTemrature   float64
	hvacMode           string
	SensorID           string
	SensorName         string
}
