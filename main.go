package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/fileutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/go-utils/pathutil"
	"github.com/bitrise-tools/go-steputils/input"
)

func main() {
	apiKey := "643e47c61ee1b9742e67a8b8f403f74fd31705e3"
	buildSecret := "dafed9c300f7223ff43640bd61a5a24240f508cabef7397dbcd17706f548d4b3"
	emailList := ""
	groupAliasesList := ""
	notification := ""
	releaseNotes := ""

	ipaPath := os.Getenv("ipa_path")
	dsymPath := os.Getenv("dsym_path")

	currentDir, err := pathutil.CurrentWorkingDirectoryAbsolutePath()
	if err != nil {
		log.Errorft("Failed to retrieve current working directory, error: %s", err)
		os.Exit(1)
	}

	// Validate parameters
	log.Infoft("Configs:")
	log.Printf("* API key: ***")
	log.Printf("* Build secret: ***")
	log.Printf("* IPA path: %s", ipaPath)
	log.Printf("* DSYM path: %s", dsymPath)
	log.Printf("* Email list: %s", emailList)
	log.Printf("* Group aliases' list: %s", groupAliasesList)
	log.Printf("* Notification: %s", notification)
	log.Printf("* Release notes: %s", releaseNotes)
	log.Printf("\n")

	if err := input.ValidateIfNotEmpty(apiKey); err != nil {
		log.Errorft("API key error: %s", err)
		os.Exit(1)
	}

	if err := input.ValidateIfNotEmpty(buildSecret); err != nil {
		log.Errorft("Build secret error: %s", err)
		os.Exit(1)
	}

	if err := input.ValidateIfNotEmpty(dsymPath); err != nil {
		if err := input.ValidateIfNotEmpty(ipaPath); err != nil {
			log.Errorft("No IPA path nor DSYM path defined")
			os.Exit(1)
		}
	}

	if err := input.ValidateIfNotEmpty(ipaPath); err == nil {
		isPathExists, err := pathutil.IsPathExists(ipaPath)
		if isPathExists == false {
			log.Errorft("IPA path defined, but the file does not exist at path: %s", ipaPath)
		} else if err != nil {
			log.Errorft("Failed to retrieve the contents of IPA path, error: %s", err)
		}

		// - Release Notes: save to file - using a temporary directory
		configReleaseNotesPth, err := pathutil.NormalizedOSTempDirPath("configReleaseNotes")
		if err != nil {
			log.Errorft("Failed to create config release notes path, error: %s", err)
			os.Exit(1)
		}

		configReleaseNotesPth = filepath.Join(configReleaseNotesPth, "app_releaseNotes.txt")
		fileutil.WriteStringToFile(configReleaseNotesPth, releaseNotes)

		// - Optional params
		paramEmails := ""
		if emailList != "" {
			paramEmails = fmt.Sprintf("-emails " + emailList)
		}

		paramGroups := ""
		if groupAliasesList != "" {
			paramGroups = fmt.Sprintf("-groupAliases " + groupAliasesList)
		}

		configIsSendNotifications := "YES"
		if notification == "No" {
			configIsSendNotifications = "NO"
		}

		// - Submit IPA
		log.Infoft("Submitting IPA...")
		submitIPACmd := filepath.Join(currentDir, "Fabric/submit")
		submitIPACmd += fmt.Sprintf("\\%s \\%s", apiKey, buildSecret)
		submitIPACmd += fmt.Sprintf("-ipaPath \\%s -notesPath \\%s", ipaPath, configReleaseNotesPth)
		submitIPACmd += fmt.Sprintf("-notifications \\%s%s%s", configIsSendNotifications, paramEmails, paramGroups)
		log.Printf(submitIPACmd)
		log.Printf("\n")

		//out, err := exec.Command(submitCmd).Output()
		out, err := command.RunCmdAndReturnExitCode(exec.Command(submitIPACmd))
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the IPA, error: %s", err))
			os.Exit(1)
		} else if out == 0 {
			log.Doneft("IPA successfully submitted")
		} else {
			log.Errorft("IPA submit failed")
			os.Exit(1)
		}
	}

	// - Submit DSYM
	if err := input.ValidateIfNotEmpty(dsymPath); err == nil {
		isPathExists, err := pathutil.IsPathExists(dsymPath)
		if isPathExists == false {
			log.Errorft("DSYM path defined, but the file does not exist at path: %s", ipaPath)
		} else if err != nil {
			log.Errorft("Failed to retrieve the contents of DSYM path, error: %s", err)
		}

		log.Infoft("Submitting DSYM...")
		submitDSYMCmd := filepath.Join(currentDir, "Fabric/upload-symbols")
		submitDSYMCmd += fmt.Sprintf("-a \\%s -p ios \\%s", apiKey, dsymPath)
		log.Printf(submitDSYMCmd)
		log.Printf("\n")

		out, err := command.RunCmdAndReturnExitCode(exec.Command(submitDSYMCmd))
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the DSYM, error: %s", err))
			os.Exit(1)
		} else if out == 0 {
			log.Doneft("DSYM successfully submitted")
		} else {
			log.Errorft("DSYM submit failed")
			os.Exit(1)
		}
	}
}
