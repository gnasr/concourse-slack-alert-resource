package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/a8m/envsubst"
	"github.com/gnasr/concourse-slack-alert-resource/concourse"
	"github.com/gnasr/concourse-slack-alert-resource/slack"
)

func parseAttachments(att []concourse.AttachmentMap, src string) []slack.Field {
	var parsedFields []slack.Field
	for _, v := range att {
		attachmentField := slack.Field{
			Title: v.Name,
		}

		// If filepath is not empty set the content as value
		if v.File != "" {
			content, err := ioutil.ReadFile(filepath.Join(src, v.File))
			if err != nil {
				log.Fatalln("error reading file:", err)
			}
			attachmentField.Value = string(content)
			// If filepath is not defined set the value field
		} else if v.Value != "" {
			attachmentField.Value = v.Value
		}

		// Parse environment variables in content
		envSubstValue, err := envsubst.String(attachmentField.Value)
		if err != nil {
			log.Fatalln("error while substituing variable:", err)
		}

		attachmentField.Value = envSubstValue
		parsedFields = append(parsedFields, attachmentField)
	}

	return parsedFields
}

func buildMessage(alert Alert, m concourse.BuildMetadata, src string) *slack.Message {
	fallback := fmt.Sprintf("%s -- %s", fmt.Sprintf("%s: %s/%s/%s", alert.Message, m.PipelineName, m.JobName, m.BuildName), m.URL)
	attachmentFields := parseAttachments(alert.Attachments, src)
	fields := []slack.Field{
		slack.Field{
			Title: "Job",
			Value: fmt.Sprintf("%s/%s", m.PipelineName, m.JobName),
			Short: true,
		},
		slack.Field{
			Title: "Build",
			Value: m.BuildName,
			Short: true,
		},
	}

	attachment := slack.Attachment{
		Fallback:   fallback,
		AuthorName: alert.Message,
		Color:      alert.Color,
		Footer:     m.URL,
		FooterIcon: alert.IconURL,
		Fields:     append(fields, attachmentFields...),
	}

	return &slack.Message{Attachments: []slack.Attachment{attachment}, Channel: alert.Channel}
}

func previousBuildStatus(input *concourse.OutRequest, m concourse.BuildMetadata) (string, error) {
	// Exit early if first build
	if m.BuildName == "1" {
		return "", nil
	}

	c, err := concourse.NewClient(m.Host, m.TeamName, input.Source.Username, input.Source.Password)
	if err != nil {
		return "", fmt.Errorf("error connecting to Concourse: %s", err)
	}

	no, err := strconv.Atoi(m.BuildName)
	if err != nil {
		return "", err
	}

	previous, err := c.JobBuild(m.PipelineName, m.JobName, strconv.Itoa(no-1))
	if err != nil {
		return "", fmt.Errorf("error requesting Concourse build status: %s", err)
	}

	return previous.Status, nil
}

func out(input *concourse.OutRequest, src string) (*concourse.OutResponse, error) {
	if input.Source.URL == "" {
		return nil, errors.New("slack webhook url cannot be blank")
	}

	alert := NewAlert(input)

	metadata := concourse.NewBuildMetadata(input.Source.ConcourseURL)
	send := !alert.Disabled

	if send && (alert.Type == "fixed" || alert.Type == "broke") {
		status, err := previousBuildStatus(input, metadata)
		if err != nil {
			return nil, err
		}
		send = (alert.Type == "fixed" && status != "succeeded") || (alert.Type == "broke" && status == "succeeded")
	}

	if send {
		message := buildMessage(alert, metadata, src)
		err := slack.Send(input.Source.URL, message)
		if err != nil {
			return nil, err
		}
	}

	out := &concourse.OutResponse{
		Version: concourse.Version{"ver": "static"},
		Metadata: []concourse.Metadata{
			concourse.Metadata{Name: "type", Value: alert.Type},
			concourse.Metadata{Name: "channel", Value: alert.Channel},
			concourse.Metadata{Name: "alerted", Value: strconv.FormatBool(send)},
		},
	}
	return out, nil
}

func main() {
	var input *concourse.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&input)
	if err != nil {
		log.Fatalln(err)
	}

	if len(os.Args) < 2 {
		log.Fatalln("destination path not specified:", err)
		os.Exit(1)
		return
	}

	src := os.Args[1]

	o, err := out(input, src)
	if err != nil {
		log.Fatalln(err)
	}

	err = json.NewEncoder(os.Stdout).Encode(o)
	if err != nil {
		log.Fatalln(err)
	}
}
