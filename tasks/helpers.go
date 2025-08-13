package tasks

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	api "cloud-guardian/api"
	pm "cloud-guardian/linux/packagemanager"
)

func handleAPIError(errorMsg string, statusCode int) {
	// Handle API errors by printing the error message and status code
	// 4xx are user errors, we log them and then quit because the user needs to fix something
	if statusCode == 404 {
		log.Fatal("API URL is incorrect: ", Config.ApiUrl)
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

func updateJobStatus(hostname, jobId, status string, result string) {
	// Update the status of a job for the given hostname
	log.Println("Updating job status for", hostname, "Job ID:", jobId, "Status:", status)

	statusCode, err := api.PutRequest(Config.ApiUrl+"jobs/"+jobId, Config.ApiKey, map[string]interface{}{
		"status": status,
		"result": result,
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error updating job status", statusCode)
		return
	}
	log.Println("Job status updated successfully for", hostname, "Job ID:", jobId, "Status:", status)
}

func checkRebootStatus(job HostJob) (bool, error) {
	// Check the status of a reboot job
	// This function can be used to check if the reboot was successful or not
	if !strings.HasPrefix(job.Result, "initiated reboot, uptime: ") {
		log.Println("Job status:", job.Result)
		log.Println("Error parsing uptime from job status: has not prefix")
		return false, errors.New("status data is not in the expected format")
	}

	// Extract the uptime from the job data
	// The job status should be:
	// initiated reboot, uptime: "+fmt.Sprintf("%d", uptime)
	uptimeBeforeReboot, err := strconv.ParseInt(strings.TrimPrefix(job.Result, "initiated reboot, uptime: "), 10, 64)
	if err != nil {
		log.Println("Job status:", job.Result)
		log.Println("Error parsing uptime from job status:", err.Error())
		return false, errors.New("status data is not in the expected format")
	}

	uptime, err := getUptime()
	if err != nil {
		return false, errors.New("error getting uptime: " + err.Error())
	}
	if uptime > uptimeBeforeReboot && (uptime-uptimeBeforeReboot) > maxRebootDuration {
		return false, errors.New("system is still running after the reboot was initiated")
	}
	if uptime < uptimeBeforeReboot {
		return true, nil // Reboot was successful
	}
	return false, nil
}

type HostJob struct {
	JobId     string `json:"jobId"`
	Signature string `json:"signature"`
	CreatedAt string `json:"createdAt"`
	JobType   string `json:"jobType"`
	JobData   string `json:"jobData"`
	Result    string `json:"result"`
	Status    string `json:"status"`
}

type HostJobPayload struct {
	Command string `json:"command"`
}

type HostJobResponse struct {
	Code    int       `json:"code"`
	Content []HostJob `json:"content"`
	Message string    `json:"message"`
}

func fetchHostJobs(hostname string, status string) (*[]HostJob, error) {
	log.Println("Fetching host jobs from API...")
	statusCode, responseBody, err := api.GetRequest(Config.ApiUrl+"jobs/hosts/"+hostname+"?job_status="+status, Config.ApiKey)
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
