package tasks

import (
	api "cloud-guardian/api"
	"cloud-guardian/cloudguardian_config"
	"cloud-guardian/cloudguardian_version"
	linux "cloud-guardian/linux"
	linux_container "cloud-guardian/linux/container"
	linux_df "cloud-guardian/linux/df"
	linux_ip "cloud-guardian/linux/ip"
	linux_loggedinusers "cloud-guardian/linux/loggedinusers"
	linux_osrelease "cloud-guardian/linux/osrelease"
	pm "cloud-guardian/linux/packagemanager"
	linux_reboot "cloud-guardian/linux/reboot"
	linux_top "cloud-guardian/linux/top"
	linux_lsblk "cloud-guardian/linux/lsblk"
	linux_mdstat "cloud-guardian/linux/mdstat"
	linux_needrestart "cloud-guardian/linux/needrestart"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

var Config *cloudguardian_config.CloudGuardianConfig // Configuration for the Cloud Gardian client
const maxRebootDuration = 300                        // Maximum allowed reboot duration in seconds

// getUptime is a function variable that can be mocked in tests
var getUptime = linux_top.GetUptime

func ProcessTasks(hostname string, oneShot bool) {

	log.Println("Using API URL:", Config.ApiUrl)

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
	processRunningJobs(hostname)
	processNewJobs(hostname)
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

func processPing(hostname string) {
	// Process ping for the given hostname
	log.Println("Processing ping for", hostname)

	statusCode, err := api.PostRequest(Config.ApiUrl+"hosts/ping/"+hostname, Config.ApiKey, map[string]any{})

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

	networkInterfaces, err := linux_ip.GetIPInterfaces()
	if err != nil {
		log.Println("Error getting network interfaces:", err.Error())
		return
	}

	routes, err := linux_ip.GetRoutes()
	if err != nil {
		log.Println("Error getting IP routes:", err.Error())
		return
	}

	cpuUsage := linux_top.GetCpuUsage()
	cpuInfo := linux_top.GetCpuInfo()
	loadAverage := linux_top.GetLoad()
	memory := linux_top.GetMemory()
	tasks := linux_top.GetTasks()
	blockdevices := linux_lsblk.GetLsBlk()
	mdstat := linux_mdstat.GetMdStat()
	needrestart := linux_needrestart.GetNeedRestart()

	statusCode, err := api.PostRequest(Config.ApiUrl+"hosts/monitoring/"+hostname, Config.ApiKey, map[string]any{
		"Uptime":            uptime,
		"LoadAverage":       loadAverage,
		"LoggedInUsers":     loggedInUsers,
		"CpuUsage":          cpuUsage,
		"CpuInfo":           cpuInfo,
		"Memory":            memory,
		"Tasks":             tasks,
		"DiskFree":          diskFree,
		"NetworkInterfaces": networkInterfaces,
		"Routes":            routes,
		"BlockDevices":      blockdevices,
		"MdStat":            mdstat,
		"NeedRestart":       needrestart,

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
	if Config.Debug {
		log.Println("##########################################")
		log.Println("Name" + linux_osrelease.Release.Name + " " + linux_osrelease.Release.VersionID)
		log.Println("##########################################")
	}
	statusCode, err := api.PostRequest(Config.ApiUrl+"hosts/osinfo/"+hostname, Config.ApiKey, map[string]interface{}{
		"os_name":               linux_osrelease.Release.Name,
		"os_version_id":         linux_osrelease.Release.VersionID,
		"is_container":          linux_container.IsRunningInContainer(),
		"agent_version":         cloudguardian_version.Version,
		"agent_running_as_root": linux.HasRootPrivileges(),
		"accepted_public_keys":  Config.HostSecurityKeys,
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

	if Config.Debug {
		log.Println("##########################################")
		log.Println("Installed packages for", hostname)
		for _, pkg := range packages {
			log.Println(pkg.Name + " - " + pkg.Version + " (" + pkg.Repo + ")")
		}
		log.Println("##########################################")
	}

	statusCode, err := api.PostRequest(Config.ApiUrl+"hosts/packages/"+hostname, Config.ApiKey, map[string]interface{}{
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
	if Config.Debug {
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
		url = Config.ApiUrl + "hosts/updates/" + hostname + "?security=true"
	default:
		url = Config.ApiUrl + "hosts/updates/" + hostname + "?security=false"
	}

	statusCode, err := api.PostRequest(url, Config.ApiKey, map[string]interface{}{
		"updates": formatPackages(updates),
	})
	if err != nil || statusCode != http.StatusOK {
		handleAPIError("Error submitting updates", statusCode)
		return
	}
	log.Println("Updates submitted successfully for", hostname)
}

func processRunningJobs(hostname string) {
	// Process running jobs for the given hostname
	log.Println("Processing running jobs for", hostname)

	runningJobs, err := fetchHostJobs(hostname, "running")
	if err != nil {
		log.Println("Error fetching running jobs:", err.Error())
		return
	}
	if runningJobs == nil {
		log.Println("No running jobs found for host:", hostname)
		return
	}

	for _, job := range *runningJobs {
		log.Println("Running job ID:", job.JobId, "Job Type:", job.JobType)
		switch job.JobType {
		case "reboot":
			log.Println("Processing reboot job for job ID:", job.JobId)

			// Check the status of the reboot job
			rebootSuccessful, err := checkRebootStatus(job)
			if err != nil {
				if err.Error() == "status data is not in the expected format" {
					log.Println("Reboot job: Status data is not in the expected format")
					updateJobStatus(hostname, job.JobId, "failed", "We couldn't check the uptime of the host, just before the reboot")
					return
				}
				if err.Error() == "system is still running after the reboot was initiated" {
					log.Println("Reboot job: System is still running after the reboot was initiated")
					updateJobStatus(hostname, job.JobId, "failed", "System is still running after the reboot was initiated")
					return
				}
				if strings.HasPrefix(err.Error(), "error getting uptime: ") {
					log.Println("Reboot job: Error getting uptime:", err.Error())
					updateJobStatus(hostname, job.JobId, "failed", "We couldn't check the uptime of the host, after the reboot")
					return
				}

			}
			if rebootSuccessful {
				log.Println("Reboot job was successful")
				updateJobStatus(hostname, job.JobId, "completed", "Rebooted successfully")
			}

		}
	}
}

func processNewJobs(hostname string) {
	submittedJobs, err := fetchHostJobs(hostname, "submitted")
	if err != nil {
		log.Fatal("Error fetching host jobs:", err.Error())
		return
	}
	if submittedJobs == nil {
		log.Println("No jobs found for host:", hostname)
		return
	}
	for _, job := range *submittedJobs {

		// {"createdAt":"${job.createdAt}","hostname":"${job.hostname}","jobType":"${job.jobType}","jobData":"${job.jobData}"}
		message := `{"createdAt":"` + job.CreatedAt + `","hostname":"` + hostname + `","jobType":"` + job.JobType + `","jobData":"` + job.JobData + `"}`

		validated, err := tryValidatePayload(Config.HostSecurityKeys, message, job.Signature)
		if err != nil {
			log.Println("Failed to validate job payload:", job.JobId)
			// Report back to the API that the job could not be processed
			updateJobStatus(hostname, job.JobId, "failed", "could not find valid host security key or failed to verify job payload")
			continue
		}
		if !validated {
			log.Println("Invalid job payload signature for job ID:", job.JobId)
			// Report back to the API that the job could not be processed
			updateJobStatus(hostname, job.JobId, "failed", "invalid job payload signature")
			continue
		}

		switch job.JobType {
		case "update":
			processJobUpdate(hostname, job.JobId, job.JobData)
		case "reboot":
			processJobReboot(hostname, job.JobId)
		case "command":
			processJobCommand(hostname, job.JobId, job.JobData)
		case "script":
			// Process script job
			log.Println("Processing script job for job ID:", job.JobId)
		case "update_agent":
			// Process update_agent job
			log.Println("Processing update_agent job for job ID:", job.JobId)
		default:
			log.Println("Unknown job type for job ID:", job.JobId, "Job Type:", job.JobType)
			// Report back to the API that the job could not be processed
			updateJobStatus(hostname, job.JobId, "failed", "unknown job type")
			continue
		}
	}
}

func processJobCommand(hostname string, jobId string, command string) {
	log.Println("Processing command job for job ID:", jobId)
	log.Println("Executing command:", command)
	updateJobStatus(hostname, jobId, "running", "")
	// Execute the command
	cmd := exec.Command("bash", "-c")
	cmd.Args = append(cmd.Args, command)
	stdOut, stdErr, err := linux.RunCommand(cmd)
	if err != nil {
		log.Println("Error executing command:", err.Error())
		updateJobStatus(hostname, jobId, "failed", fmt.Sprintf("failed to execute command: %s", stdErr))
		return
	}
	updateJobStatus(hostname, jobId, "completed", stdOut)
}

func processJobUpdate(hostname string, jobId string, packages string) {
	log.Println("Processing update job for job ID:", jobId)
	log.Println("Updating packages:", packages)
	updateJobStatus(hostname, jobId, "running", "")
	packageList := strings.Split(packages, ",")
	packageManager, err := pm.DetectPackageManager()
	if err != nil {
		log.Println("Error detecting package manager:", err.Error())
		return
	}
	var stdOut, stdErr string
	if packageList[0] == "all" {
		stdOut, stdErr, err = packageManager.UpdateAllPackages()
	} else {
		stdOut, stdErr, err = packageManager.UpdatePackages(packageList)
	}
	if err != nil {
		log.Println("Error updating packages:", err.Error())
		updateJobStatus(hostname, jobId, "failed", fmt.Sprintf("failed to update packages %s", stdErr))
		return
	}
	updateJobStatus(hostname, jobId, "completed", stdOut)
	processUpdates(hostname, pm.AllUpdates, packageManager)
	processUpdates(hostname, pm.SecurityUpdates, packageManager)
}

func processJobReboot(hostname string, jobId string) {
	log.Println("Processing reboot job for job ID:", jobId)
	// For reboot we first update the job status to "running" and then reboot
	// the system. Later we check the running jobs to see if the job was successful
	uptime, err := linux_top.GetUptime()
	if uptime < maxRebootDuration {
		log.Println("Reboot job: Uptime is less than", maxRebootDuration, " seconds. We have to wait until it is safe to reboot. Otherwise it could cause reboot loops.")
		return
	}
	if err != nil {
		log.Println("Reboot job: Error getting uptime:", err.Error())
		updateJobStatus(hostname, jobId, "failed", "Reboot failed, because we couldn't check the uptime of the host")
		return
	}
	updateJobStatus(hostname, jobId, "running", "initiated reboot, uptime: "+fmt.Sprintf("%d", uptime))
	if err := linux_reboot.Reboot(); err != nil {
		log.Println("Reboot job: Error initiating reboot:", err.Error())
		updateJobStatus(hostname, jobId, "failed", "Reboot failed, because we couldn't initiate the reboot")
		return
	}
}
