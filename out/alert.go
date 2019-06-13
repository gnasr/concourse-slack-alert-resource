package main

import (
	"github.com/gnasr/concourse-slack-alert-resource/concourse"
)

// An Alert defines the notification that will be sent to Slack.
type Alert struct {
	Type        string
	Channel     string
	Color       string
	IconURL     string
	Message     string
	Disabled    bool
	Attachments []concourse.AttachmentMap
	File        string
}

// NewAlert constructs and returns an Alert.
func NewAlert(input *concourse.OutRequest) Alert {
	var alert Alert
	attachments := input.Params.Attachments

	switch input.Params.AlertType {
	case "success":
		alert = Alert{
			Type:        "success",
			Color:       "#32cd32",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-succeeded.png",
			Message:     "Success",
			Attachments: attachments,
		}
	case "failed":
		alert = Alert{
			Type:        "failed",
			Color:       "#d00000",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-failed.png",
			Message:     "Failed",
			Attachments: attachments,
		}
	case "started":
		alert = Alert{
			Type:        "started",
			Color:       "#f7cd42",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-started.png",
			Message:     "Started",
			Attachments: attachments,
		}
	case "aborted":
		alert = Alert{
			Type:        "aborted",
			Color:       "#8d4b32",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-aborted.png",
			Message:     "Aborted",
			Attachments: attachments,
		}
	case "fixed":
		alert = Alert{
			Type:        "fixed",
			Color:       "#32cd32",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-succeeded.png",
			Message:     "Fixed",
			Attachments: attachments,
		}
	case "broke":
		alert = Alert{
			Type:        "broke",
			Color:       "#d00000",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-failed.png",
			Message:     "Broke",
			Attachments: attachments,
		}
	default:
		alert = Alert{
			Type:        "default",
			Color:       "#35495c",
			IconURL:     "https://ci.concourse-ci.org/public/images/favicon-pending.png",
			Message:     "",
			Attachments: attachments,
		}
	}

	alert.Disabled = input.Params.Disable
	if alert.Disabled == false {
		alert.Disabled = input.Source.Disable
	}
	alert.Channel = input.Params.Channel
	if alert.Channel == "" {
		alert.Channel = input.Source.Channel
	}
	if input.Params.Message != "" {
		alert.Message = input.Params.Message
	}
	if input.Params.Color != "" {
		alert.Color = input.Params.Color
	}

	return alert
}
