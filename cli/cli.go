package cli

import (
	api "cloud-guardian/api"
	"cloud-guardian/cloudguardian_config"
	linux_installer "cloud-guardian/linux/installer"
	tasks "cloud-guardian/tasks"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"cloud-guardian/cloudguardian_version"
)

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
	tasks.Config = config // Set the configuration for the tasks package
	tasks.ProcessTasks(hostname, *oneShotFlag)
}

type SecurityKeyApiResponse struct {
	Code    int               `json:"code"`
	Content map[string]string `json:"content"`
	Message string            `json:"message"`
}

func fetchHostSecurityKey() {
	// Fetch the security key from the API and update the configuration file
	log.Println("Fetching security key from API...")
	statusCode, responseBody, err := api.GetRequest(config.ApiUrl+"hosts/securitykey", config.ApiKey)
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

func registerClient(hostname string) {
	// Register the client with the API
	log.Println("Registering client with hostname:", hostname)

	statusCode, err := api.PostRequest(config.ApiUrl+"hosts/register/"+hostname, config.ApiKey, map[string]any{})
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

func printVersion() {
	// Print version information
	log.Println("Version:",  cloudguardian_version.Version)
}
