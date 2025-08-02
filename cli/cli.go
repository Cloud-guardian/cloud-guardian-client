package cli

import (
	"cloud-guardian/cloudguardian_config"
	cloudguardian_crypto "cloud-guardian/crypto"
	linux_container "cloud-guardian/linux/container"
	linux_df "cloud-guardian/linux/df"
	linux_installer "cloud-guardian/linux/installer"
	linux_loggedinusers "cloud-guardian/linux/loggedinusers"
	linux_osrelease "cloud-guardian/linux/osrelease"
	pm "cloud-guardian/linux/packagemanager"
	linux_top "cloud-guardian/linux/top"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var Version = "fdev"    // Default version, can be overridden at build time with -ldflags "-X main.version=x.x.x"
const apiKeyLength = 16 // Length of the API key, used for validation

var config *cloudguardian_config.CloudGuardianConfig // Configuration for the Cloud Gardian client

func IsValidApiKey(apiKey string) bool {
	// A valid API key is 16 characters long and contains only alphanumeric characters in lowercase
	if len(apiKey) != apiKeyLength {
		return false
	}
	matched, _ := regexp.MatchString("^[a-z0-9]+$", apiKey)
	return matched
}

func Start() {
	// Define command-line flags
	var (
		versionFlag   = flag.Bool("version", false, "Display version information")
		debugFlag     = flag.Bool("debug", false, "Enable debug mode")
		apiUrlFlag    = flag.String("api-url", "", "API URL to submit updates")
		apiKeyFlag    = flag.String("api-key", "", "API key for authentication (required)")
		oneShotFlag   = flag.Bool("one-shot", false, "Run in oneshot mode (process updates and exit)")
		installFlag   = flag.Bool("install", false, "Install the client as a system service (also registers the client)")
		updateFlag    = flag.Bool("update", false, "Update the client to the latest version (if available)")
		uninstallFlag = flag.Bool("uninstall", false, "Uninstall the client service (if installed)")
		registerFlag  = flag.Bool("register", false, "Register the client with the API (register without installing as a service)")
		//jobsFlag      = flag.Bool("jobs", false, "Fetch host jobs from the API")
	)

	var err error

	config, err = cloudguardian_config.FindAndLoadConfig()
	if err != nil {
		if err.Error() == cloudguardian_config.ErrConfigNotFound.Error() {
			// If the config file is not found, we will use the default configuration
			config = cloudguardian_config.DefaultConfig()
		} else {
			// If there is an error loading the configuration, we will print the error and use the default configuration
			log.Fatal(err.Error())
		}
	}

	// Parse the command-line flags
	flag.Parse()
	programName := path.Base(os.Args[0])

	l := len("cloud-guardian-ez-")
	// If programName is in the format cloud-guardian-ez-<apikey>, we can extract the API key
	if strings.HasPrefix(programName, "cloud-guardian-ez") && len(programName) == l+apiKeyLength {
		extractedApiKey := programName[l : l+apiKeyLength] // Extract the API key from the program name
		// Check with regex if the API key is valid. A valid API key is 32 characters long and contains only alphanumeric characters in lowercase:
		if IsValidApiKey(extractedApiKey) {
			config.ApiKey = extractedApiKey
			log.Println("API key extracted from program name:", config.ApiKey)
		}
	}

	if *uninstallFlag {
		// Uninstall the client service
		log.Println("Uninstalling client service...")
		if err := linux_installer.Uninstall(); err != nil {
			if os.IsPermission(err) {
				log.Fatal("Error: You need to run this command with root privileges to uninstall the client service.")
			}
			log.Fatal("Error uninstalling client service:", err.Error())
		}
		log.Println("Client service uninstalled successfully.")
		return
	}

	if *versionFlag {
		printVersion()
		return
	}

	if *debugFlag {
		// Enable debug mode
		log.Println("Debug mode enabled")
		config.Debug = true
	}

	if *apiKeyFlag != "" {
		// Set the API key if provided
		config.ApiKey = *apiKeyFlag
	} else if config.ApiKey == "" {
		log.Fatal("Error: API key is required. Use --api-key to set it.")
		return
	}

	if *apiUrlFlag != "" {
		// Override the default API URL if provided
		apiUrl := *apiUrlFlag
		if !strings.HasSuffix(apiUrl, "/") {
			// Ensure the API URL ends with a slash
			apiUrl += "/"
		}
		config.ApiUrl = apiUrl
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Error getting hostname:", err.Error())
		return
	}

	if *installFlag {
		// Install the client as a system service
		InstallService(hostname)
		return
	}

	if *updateFlag {
		// Update the client to the latest version
		UpdateService()
		return
	}

	if *registerFlag {
		// Register the client with the API
		registerClient(hostname)
		return
	}

	if false { // Replace with *jobsFlag when you want to enable fetching host jobs
		// Fetch host jobs from the API
		hostname = "ubuntu-2404" // For testing purposes, we set a fixed hostname
		processJobs(hostname)

		return
	}

	processTasks(hostname, *oneShotFlag)
}

func InstallService(hostname string) {
	// Install the client as a system service
	log.Println("Installing client as a system service...")

	fetchHostSecurityKey()

	linux_installer.Config = config // Set the configuration for the installer

	if err := linux_installer.Install(); err != nil {
		// check if error is os.ErrPermission, which indicates that the user does not have root privileges
		if os.IsPermission(err) {
			log.Println("Error: You need to run this command with root privileges to install the client as a system service.")
			return
		}
		log.Println("Error installing client as a system service:", err.Error())
		return
	}

	log.Println("Client installed as a system service")

	// Register the client with the API after installing as a service
	registerClient(hostname)
}

func UpdateService() {

	linux_installer.Config = config // Set the configuration for the installer
	if err := linux_installer.Update(); err != nil {
		// check if error is os.ErrPermission, which indicates that the user does not have root privileges
		if os.IsPermission(err) {
			log.Println("Error: You need to run this command with root privileges to update the client service.")
			return
		}
		log.Println("Error updating client service:", err.Error())
		return
	}
	log.Println("Client service updated successfully")
}

func parseErrorResponse(err error) string {
	// The error might be a JSON response with an error message, in that case we try to parse it
	var errorResponse map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(err.Error()), &errorResponse); jsonErr == nil {
		if message, ok := errorResponse["message"].(string); ok {
			return message
		}
	}
	// If we couldn't parse the error, return the error string
	return err.Error()
}

type HostJob struct {
	JobId     string         `json:"job_id"`
	Payload   HostJobPayload `json:"payload"`
	Signature string         `json:"signature"`
}

type HostJobPayload struct {
	Command string `json:"command"`
}

type HostJobResponse struct {
	Code    int       `json:"code"`
	Content []HostJob `json:"content"`
	Message string    `json:"message"`
}

func fetchHostJobs(hostname string) (*[]HostJob, error) {
	log.Println("Fetching host jobs from API...")
	statusCode, responseBody, err := getRequest(config.ApiUrl + "jobs/hosts/" + hostname)
	if err != nil {
		log.Println(parseErrorResponse(err))
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, nil // Return nil if no jobs are found
	}

	if statusCode != http.StatusOK {
		handleAPIError("Error retrieving host jobs", statusCode)
		return nil, errors.New("error retrieving host jobs")
	}

	var response HostJobResponse
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		log.Println("Error parsing response body:", err.Error())
		return nil, err
	}
	return &response.Content, nil
}

type SecurityKeyApiResponse struct {
	Code    int               `json:"code"`
	Content map[string]string `json:"content"`
	Message string            `json:"message"`
}

func fetchHostSecurityKey() {
	// Fetch the security key from the API and update the configuration file
	log.Println("Fetching security key from API...")
	statusCode, responseBody, err := getRequest(config.ApiUrl + "hosts/securitykey")
	if err != nil {
		log.Println(parseErrorResponse(err))
		return
	}
	if statusCode == http.StatusNotFound {
		log.Println("Security key not found")
		return
	}

	if statusCode != http.StatusOK {
		handleAPIError("Error retrieving security key", statusCode)
		return
	}

	var response SecurityKeyApiResponse
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		log.Println("Error parsing response body:", err.Error())
		return
	}
	if securityKey, ok := response.Content["hostSecurityKey"]; ok {
		// Save the security key to the configuration
		config.HostSecurityKey = securityKey
		println("Security Key:", securityKey)
	}

}

func registerClient(hostname string) {
	// Register the client with the API
	log.Println("Registering client with hostname:", hostname)

	statusCode, err := postRequest(config.ApiUrl+"hosts/register/"+hostname, map[string]any{})
	if err != nil {
		log.Println(parseErrorResponse(err))
		return
	}
	if statusCode != http.StatusOK {
		handleAPIError("Error registering client", statusCode)
		return
	}
	log.Println("Client registered successfully with hostname:", hostname)
}

func handleAPIError(errorMsg string, statusCode int) {
	// Handle API errors by printing the error message and status code
	// 4xx are user errors, we log them and then quit because the user needs to fix something
	if statusCode == 404 {
		log.Fatal("API URL is incorrect: ", config.ApiUrl)
	}
	if statusCode == 401 {
		log.Fatal("Invalid API key. Please check your API key in the configuration file or command line arguments.")
	}
	if statusCode >= 400 && statusCode < 500 {
		log.Println(errorMsg, "(Client error) - Status code:", statusCode)
		return
	}
	// Everything above 500 is considered a server error, we log it
	if statusCode >= 500 {
		log.Println(errorMsg)
	}
}

func processTasks(hostname string, oneShot bool) {

	log.Println("Using API URL:", config.ApiUrl)

	var minuteCounter int = 0

	for {

		if minuteCounter%5 == 0 {
			// Process tasks that need to run every 5 minutes
			processFiveMinuteTasks(hostname)
		}

		if minuteCounter%60 == 0 {
			// Process tasks that need to run every hour
			processHourlyTasks(hostname)
		}
		if minuteCounter%1440 == 0 {
			// Process tasks that need to run every day
			processDailyTasks(hostname)
		}

		if oneShot {
			// If in oneshot mode, exit after processing tasks
			log.Println("Exiting after oneshot execution.")
			return
		}

		// Sleep for 1 minute before the next iteration
		time.Sleep(1 * time.Minute)
		minuteCounter++

		if minuteCounter > 1440 {
			// Reset the minute counter after 24 hours
			minuteCounter = 0
		}
	}
}

func processFiveMinuteTasks(hostname string) {
	log.Println("Processing 5-minute tasks...")
	processPing(hostname)
	processBasicMonitoring(hostname)
}

func processDailyTasks(hostname string) {
	log.Println("Processing daily tasks...")

	// Detect package manager
	packageManager, err := pm.DetectPackageManager()
	if err != nil {
		log.Println("Error detecting package manager:", err.Error())
		return
	}
	processSystemInfo(hostname)
	processUpdates(hostname, pm.AllUpdates, packageManager)
	processUpdates(hostname, pm.SecurityUpdates, packageManager)
	processInstalledPackages(hostname, packageManager)
}

func processHourlyTasks(hostname string) {
	// This function can be used to process hourly tasks if needed
	log.Println("Processing hourly tasks...")
	// For example, you can call processPing or other functions here
}

func formatPackages(packages []pm.Package) []map[string]string {
	formatted := []map[string]string{}
	for _, update := range packages {
		formatted = append(formatted, map[string]string{
			"name":    strings.ToLower(update.Name),
			"version": strings.ToLower(update.Version),
			"repo":    strings.ToLower(update.Repo),
		})
	}
	return formatted
}

func postRequest(url string, data interface{}) (int, error) {

	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling system info to JSON:", err.Error())
		return 500, err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Println("Error creating request:", err.Error())
		return 500, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", config.ApiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err.Error())
		return 500, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("%s", body)
	}
	return resp.StatusCode, nil
}

func getRequest(url string) (int, string, error) {
	// Send a GET request to the specified URL with the API key
	// Returns the status code and response body as a string

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err.Error())
		return 500, "", err
	}
	req.Header.Set("x-api-key", config.ApiKey)
	println("API Key:", config.ApiKey)
	println("Get request to URL:", url)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err.Error())
		return 500, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, "", nil
	}
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}

func processPing(hostname string) {
	// Process ping for the given hostname
	log.Println("Processing ping for", hostname)

	statusCode, err := postRequest(config.ApiUrl+"hosts/ping/"+hostname, map[string]any{})

	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting ping", statusCode)
		return
	}
	log.Println("Ping submitted successfully for", hostname)
}

func processBasicMonitoring(hostname string) {
	// Process simple monitoring metrics for the given hostname
	log.Println("Processing basic monitoring for", hostname)

	uptime, err := linux_top.GetUptime()
	if err != nil {
		log.Println("Error getting uptime:", err.Error())
		return
	}

	// Get logged in users
	loggedInUsers, err := linux_loggedinusers.GetLoggedInUsers()
	if err != nil {
		log.Println("Error getting logged in users:", err.Error())
		return
	}

	diskFree, err := linux_df.GetDf()
	if err != nil {
		log.Println("Error getting disk usage:", err.Error())
		return
	}

	cpuUsage := linux_top.GetCpuUsage()
	cpuInfo := linux_top.GetCpuInfo()
	loadAverage := linux_top.GetLoad()
	memory := linux_top.GetMemory()
	tasks := linux_top.GetTasks()

	statusCode, err := postRequest(config.ApiUrl+"hosts/monitoring/"+hostname, map[string]any{
		"Uptime":        uptime,
		"LoadAverage":   loadAverage,
		"LoggedInUsers": loggedInUsers,
		"CpuUsage":      cpuUsage,
		"CpuInfo":       cpuInfo,
		"Memory":        memory,
		"Tasks":         tasks,
		"DiskFree":      diskFree,
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting basic monitoring data", statusCode)
		return
	}

	log.Println("Basic monitoring submitted successfully for", hostname)
}

func processSystemInfo(hostname string) {
	// Process system information for the given hostname

	linux_osrelease.GetOsReleaseInfo()
	// The operating system:
	if config.Debug {
		log.Println("##########################################")
		log.Println("Name" + linux_osrelease.Release.Name + " " + linux_osrelease.Release.VersionID)
		log.Println("##########################################")
	}
	statusCode, err := postRequest(config.ApiUrl+"hosts/osinfo/"+hostname, map[string]interface{}{
		"os_name":               linux_osrelease.Release.Name,
		"os_version_id":         linux_osrelease.Release.VersionID,
		"is_container":          linux_container.IsRunningInContainer(),
		"agent_version":         Version,
		"agent_running_as_root": linux_installer.HasRootPrivileges(),
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting system info", statusCode)
		return
	}

	log.Println("System information submitted successfully for", hostname)
}

func processInstalledPackages(hostname string, packageManager pm.PackageManager) {
	// Process installed packages for the given hostname
	packages, err := packageManager.GetInstalledPackages()
	if err != nil {
		log.Println("Error getting installed packages:", err.Error())
		return
	}

	if config.Debug {
		log.Println("##########################################")
		log.Println("Installed packages for", hostname)
		for _, pkg := range packages {
			log.Println(pkg.Name + " - " + pkg.Version + " (" + pkg.Repo + ")")
		}
		log.Println("##########################################")
	}

	statusCode, err := postRequest(config.ApiUrl+"hosts/packages/"+hostname, map[string]interface{}{
		"packages": formatPackages(packages),
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting installed packages", statusCode)
		return
	}
	log.Println("Installed packages submitted successfully for", hostname)
}

func processUpdates(hostname string, updateType pm.UpdateType, packageManager pm.PackageManager) {
	// Process updates for the given hostname
	updates, obsolete, err := packageManager.CheckUpdates(updateType)
	if err != nil {
		log.Println("Error checking updates:", err.Error())
		return
	}
	if config.Debug {
		log.Println("##########################################")
		switch updateType {
		case pm.SecurityUpdates:
			log.Println("Security updates available for", hostname)
		default:
			log.Println("Updates available for", hostname)
		}
		for _, update := range updates {
			log.Println(update.Name + " - " + update.Version + " (" + update.Repo + ")")
		}
		log.Println("Obsolete packages for", hostname)
		for _, obso := range obsolete {
			log.Println(obso.Name + " - " + obso.Version + " (" + obso.Repo + ")")
		}
		log.Println("##########################################")
	}

	// Submit updates to the API
	var url string
	switch updateType {
	case pm.SecurityUpdates:
		url = config.ApiUrl + "hosts/updates/" + hostname + "?security=true"
	default:
		url = config.ApiUrl + "hosts/updates/" + hostname + "?security=false"
	}

	statusCode, err := postRequest(url, map[string]interface{}{
		"updates": formatPackages(updates),
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting updates", statusCode)
		return
	}
	log.Println("Updates submitted successfully for", hostname)
}

func processJobs(hostname string) {
	jobs, err := fetchHostJobs(hostname)
	if err != nil {
		log.Fatal("Error fetching host jobs:", err.Error())
		return
	}
	if jobs == nil {
		log.Println("No jobs found for host:", hostname)
		return
	}
	for _, job := range *jobs {
		log.Println("Processing job:", job.JobId)
		log.Println("Command:", job.Payload.Command)
		log.Println("Signature:", job.Signature)

		// Here you would typically execute the command in the payload
		// For demonstration, we will just log it
		log.Println("Executing command:", job.Payload.Command)

		// After processing the job, you might want to send a response back to the API
		// statusCode, err := postRequest(config.ApiUrl+"jobs/hosts/"+hostname+"/"+job.JobId+"/response", map[string]interface{}{
		// 	"status": "success",
		// 	"message": "Job processed successfully",
		// })
		// if err != nil || statusCode != http.StatusOK {
		// 	handleAPIError("Error submitting job response", statusCode)
		// 	return
		// }
		println("Public key:", config.HostSecurityKey)
		validated, err := cloudguardian_crypto.ValidatePayload(config.HostSecurityKey, job.Payload.Command, job.Signature)
		if err != nil {
			log.Println("Failed to validate job payload:", job.JobId)
			// Report back to the API that the job could not be processed
			continue
		}
		if !validated {
			log.Println("Invalid job payload signature for job ID:", job.JobId)
			// Report back to the API that the job could not be processed
			continue
		}
		log.Println("Job response submitted successfully for job ID:", job.JobId)
	}
}

func printVersion() {
	// Print version information
	log.Println("Version:", Version)
}
