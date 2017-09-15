package actuator

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/ninech/actuator/openshift"
)

// SupportedPullRequestActions defines all pull request event actions which are supported by this app.
const (
	ActionOpened   = "opened"
	ActionClosed   = "closed"
	ActionReopened = "reopened"
)

// SupportedPullRequestActions defines all actions which are currently supported to be handled
var SupportedPullRequestActions = [1]string{ActionOpened}

// PullRequestEventHandler handles pull request events
type PullRequestEventHandler struct {
	Event   *github.PullRequestEvent
	Message string
	Config  Configuration
}

// GetMessage returns the end message of this handler to be sent to the client
func (h *PullRequestEventHandler) GetMessage() string {
	if h.Event != nil && h.Message == "" {
		return fmt.Sprintf("Event for pull request #%d received. Thank you.", h.Event.GetNumber())
	}
	return h.Message
}

// HandleEvent handles a pull request event from github
func (h *PullRequestEventHandler) HandleEvent() error {
	if !h.actionIsSupported() {
		h.Message = "Event is not relevant and will be ignored."
		return nil
	}

	repositoryName := h.Event.Repo.GetFullName()
	repositoryConfig := h.Config.GetRepositoryConfig(repositoryName)
	if repositoryConfig == nil {
		h.Message = fmt.Sprintf("Repository %s is not configured. Doing nothing.", repositoryName)
		return nil
	}

	if repositoryConfig.Enabled {
		switch h.Event.GetAction() {
		case ActionOpened:
			output, err := openshift.NewAppFromTemplate(repositoryConfig.Template, openshift.TemplateParameters{})
			if err != nil {
				return err
			} else {
				Logger.Println(output)
			}
			break
		}
	} else {
		h.Message = fmt.Sprintf("Repository %s is disabled. Doing nothing.", repositoryName)
	}

	return nil
}

// actionIsSupported returns true when the provided action is currently supported by the app
func (h *PullRequestEventHandler) actionIsSupported() bool {
	for _, a := range SupportedPullRequestActions {
		if a == h.Event.GetAction() {
			return true
		}
	}
	return false
}
