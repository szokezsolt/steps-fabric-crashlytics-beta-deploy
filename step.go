package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/log"
)

//"github.com/bitrise-tools/go-steputils/input"

// =======================================
// Functions
// =======================================

func validateRequiredInput(key string, value string) {
	if value == "" {
		log.Errorft("Missing required input: %s", key)
		os.Exit(1)
	}
}

func validateRequiredInputWithOptions(key string, value string, options []string) {
	validateRequiredInput(key, value)
	found := false
	for _, option := range options {
		if option == value {
			found := true
		}
	}

	if found == false {
		log.Errorft("Invalid input: %s value: %s, valid options: %s", key, value, strings.Join(options, ", "))
		os.Exit(1)
	}
}

// =======================================
// Main
// =======================================

func main() {
	if thisScriptDir, err := filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		fmt.Printf("Failed to retrieve current directory, error: %s", err)
		os.Exit(1)
	}

	// Validate parameters
	log.Infoft("Configs:")
	log.Printf("* api_key: ***")
	log.Printf("* build_secret: ***")
	log.Printf("* ipa_path: %s", ipa_path)
	log.Printf("* dsym_path: %s", dsym_path)
	log.Printf("* email_list: %s", email_list)
	log.Printf("* group_aliases_list: %s", group_aliases_list)
	log.Printf("* notification: %s", notification)
	log.Printf("* release_notes: %s", release_notes)
	log.Printf("\n")

	// validateRequiredInput("api_key", api_key)
	// validateRequiredInput("build_secret", build_secret)

	if dsym_path == "" && ipa_path == "" {
		log.Errorft("No IPA path nor DSYM path defined")
		os.Exit(1)
	}

	if ipa_path != "" {
		if _, err := os.Stat(ipa_path); os.IsNotExist(err) {
			log.Errorft("IPA path defined, but the file does not exist at path: %s", ipa_path)
		}

		// - Release Notes: save to file
		ConfigReleaseNotesPth := string(os.Getenv("HOME")) + "/app_release_notes.txt"
		ConfigReleaseNotesPth.WriteString(release_notes)

		// - Optional params
		paramEmails := ""
		if email_list != "" {
			paramEmails = fmt.Sprintf("-emails " + email_list)
		}

		paramGroups := ""
		if group_aliases_list != "" {
			paramGroups = fmt.Sprintf("-groupAliases " + group_aliases_list)
		}

		ConfigIsSendNotifications := "YES"
		if notification == "No" {
			ConfigIsSendNotifications = "NO"
		}

		// - Submit IPA
		log.Infoft("Submitting IPA...")
		submitCmd := filepath.Join(thisScriptDir, "Fabric/submit")
		submitCmd += fmt.Sprintf("\\%s \\%s", api_key, build_secret)
		submitCmd += fmt.Sprintf("-ipaPath \\%s -notesPath \\%s", ipa_path, ConfigReleaseNotesPth)
		submitCmd += fmt.Sprintf("-notifications \\%s%s%s", ConfigIsSendNotifications, paramEmails, paramGroups)
		log.Printf(submitCmd)
		log.Printf("\n")

		out, err := exec.Command(submitCmd).Output()
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the IPA, error: %s", err))
			os.Exit(1)
		} else if out == 1 {
			log.Doneft("Success")
		} else {
			log.Errorft("Fail")
			os.Exit(1)
		}
	}

	// - Submit DSYM
	if dsym_path != "" {
		if _, err := os.Stat(dsym_path); os.IsNotExist(err) {
			log.Errorft("DSYM path defined, but the file does not exist at path: %s", dsym_path)
			os.Exit(1)
		}

		log.Infoft("Submitting DSYM...")
		dsymCmd := filepath.Join(ThisScriptDir, "Fabric/upload-symbols")
		dsymCmd += fmt.Sprintf("-a \\%s -p ios \\%s", api_key, dsym_path)
		log.Printf(dsymCmd)
		log.Printf("\n")

		out, err := exec.Command(dsymCmd).Output()
		if err != nil {
			log.Errorft(fmt.Sprintf("Error submitting the DSYM, error: %s", err))
			os.Exit(1)
		} else if out == 1 {
			log.Doneft("Success")
		} else {
			log.Errorft("Fail")
			os.Exit(1)
		}
	}
}
