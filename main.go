package main

import (
	"fmt"
	"os"
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
	frameworkPath := os.Getenv("framework_path")
	workDir := os.Getenv("work_dir")

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

		// - Submit IPA
		log.Infoft("Submitting IPA...")
		submitIPAPth := ""
		if pathutil.IsRelativePath(frameworkPath) {
			submitIPAPth = filepath.Join(workDir, frameworkPath, "submit")
		} else {
			submitIPAPth = filepath.Join(frameworkPath, "submit")
		}

		if pathExist, _ := pathutil.IsPathExists(submitIPAPth); pathExist == false {
			submitIPAPth = filepath.Join(workDir, "Fabric/submit")
		}

		submitIPACmd := []string{}
		submitIPACmd = append(submitIPACmd, submitIPAPth)
		submitIPACmd = append(submitIPACmd, apiKey)
		submitIPACmd = append(submitIPACmd, buildSecret)
		submitIPACmd = append(submitIPACmd, "-ipaPath")
		submitIPACmd = append(submitIPACmd, ipaPath)
		submitIPACmd = append(submitIPACmd, "-notesPath")
		submitIPACmd = append(submitIPACmd, configReleaseNotesPth)

		// - Optional params
		if emailList != "" {
			submitIPACmd = append(submitIPACmd, "-emails")
			submitIPACmd = append(submitIPACmd, emailList)
		}

		if groupAliasesList != "" {
			submitIPACmd = append(submitIPACmd, "-groupAliases")
			submitIPACmd = append(submitIPACmd, groupAliasesList)
		}

		submitIPACmd = append(submitIPACmd, "-notifications")
		if notification == "No" {
			submitIPACmd = append(submitIPACmd, "NO")
		} else {
			submitIPACmd = append(submitIPACmd, "YES")
		}

		log.Printf("%v", submitIPACmd)
		log.Printf("\n")

		cmd, err := command.NewFromSlice(submitIPACmd)
		if err != nil {
			log.Errorft(fmt.Sprintf("Error creating IPA command from slice, error: %s", err))
			os.Exit(1)
		}

		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the IPA, output:\n%s, \nerror: %s", out, err))
			os.Exit(1)
		}
		log.Printf(out)
		log.Printf("\n")
	}

	// - Submit DSYM
	if err := input.ValidateIfNotEmpty(dsymPath); err == nil {
		isPathExists, err := pathutil.IsPathExists(dsymPath)
		if isPathExists == false {
			log.Errorft("DSYM path defined, but the file does not exist at path: %s", dsymPath)
		} else if err != nil {
			log.Errorft("Failed to retrieve the contents of DSYM path, error: %s", err)
		}

		log.Infoft("Submitting DSYM...")
		submitDSYMPth := ""
		if pathutil.IsRelativePath(frameworkPath) {
			submitDSYMPth = filepath.Join(workDir, frameworkPath, "uploadDSYM")
		} else {
			submitDSYMPth = filepath.Join(frameworkPath, "uploadDSYM")
		}

		if pathExist, _ := pathutil.IsPathExists(submitDSYMPth); pathExist == false {
			submitDSYMPth = filepath.Join(workDir, "Fabric/uploadDSYM")
		}

		submitDSYMCmd := []string{}
		submitDSYMCmd = append(submitDSYMCmd, submitDSYMPth)
		submitDSYMCmd = append(submitDSYMCmd, "-a")
		submitDSYMCmd = append(submitDSYMCmd, apiKey)
		submitDSYMCmd = append(submitDSYMCmd, "-p")
		submitDSYMCmd = append(submitDSYMCmd, "ios")
		submitDSYMCmd = append(submitDSYMCmd, dsymPath)

		log.Printf("%v", submitDSYMCmd)
		log.Printf("\n")

		cmd, err := command.NewFromSlice(submitDSYMCmd)
		if err != nil {
			log.Errorft(fmt.Sprintf("Error creating DSYM command from slice, error: %s", err))
			os.Exit(1)
		}

		out, err := cmd.RunAndReturnTrimmedCombinedOutput()
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the DSYM, output:\n%s, \nerror: %s", out, err))
			os.Exit(1)
		}

		log.Printf(out)
	}
}
