package actuator_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ninech/actuator/actuator"
	"github.com/ninech/actuator/github"
	"github.com/ninech/actuator/openshift"
	"github.com/ninech/actuator/test"
)

var ensurePullRequestEventHandlerImplementsEventHandler actuator.EventHandler = &actuator.PullRequestEventHandler{}

func TestPullRequestEventHandler(t *testing.T) {
	t.Run("GetEventResponse", func(t *testing.T) {
		t.Run("the event repository is disabled in the config file", func(t *testing.T) {
			event := test.NewDefaultTestEvent()
			handler := actuator.PullRequestEventHandler{
				RepositoryConfig: actuator.RepositoryConfig{Enabled: false}}

			response := handler.GetEventResponse(event)
			assert.False(t, response.HandleEvent)
			assert.Equal(t, "Repository ninech/actuator is not configured or disabled. Doing nothing.", response.Message)
		})

		t.Run("the action is not supported", func(t *testing.T) {
			event := test.NewTestEvent(1, "yolo", "ninech/actuator")
			handler := actuator.PullRequestEventHandler{}

			response := handler.GetEventResponse(event)
			assert.False(t, response.HandleEvent)
			assert.Equal(t, "Event is not relevant and will be ignored.", response.Message)
		})

		t.Run("the event has the wrong type", func(t *testing.T) {
			event := test.NewDefaultTestEvent()
			event.Type = 999

			handler := actuator.PullRequestEventHandler{}

			response := handler.GetEventResponse(event)
			assert.False(t, response.HandleEvent)
			assert.Equal(t, "Invalid event for this handler.", response.Message)
		})

		t.Run("the event is valid and ready to be handled", func(t *testing.T) {
			event := test.NewDefaultTestEvent()
			handler := actuator.PullRequestEventHandler{
				RepositoryConfig: actuator.RepositoryConfig{Enabled: true}}

			response := handler.GetEventResponse(event)
			assert.True(t, response.HandleEvent)
			assert.Equal(t, "Event for pull request #1 received. Thank you.", response.Message)
		})
	})

	t.Run("HandleEvent", func(t *testing.T) {
		t.Run("EventActionOpened, EventActionReopened", func(t *testing.T) {
			test.DisableLogging()

			for _, action := range [2]string{"opened", "reopened"} {
				event := test.NewTestEvent(1, action, "ninech/actuator")
				githubClient := test.NewMockGithubClient()
				openshiftClient := &test.OpenshiftMock{}
				repositoryConfig := actuator.RepositoryConfig{Template: "actuator-template"}

				actuator.Config = test.NewDefaultConfig()

				handler := actuator.PullRequestEventHandler{
					RepositoryConfig: repositoryConfig,
					GithubClient:     githubClient,
					Openshift:        openshiftClient}

				t.Run("applies the template in openshift", func(t *testing.T) {
					handler.HandleEvent(event)

					assert.Equal(t, repositoryConfig.Template, openshiftClient.AppliedTemplate, "it instantiates the template from the config")

					assert.Equal(t, openshiftClient.AppliedLabels["actuator.nine.ch/create-reason"], "GithubWebhook")
					assert.Equal(t, openshiftClient.AppliedLabels["actuator.nine.ch/branch"], event.HeadRef)
					assert.Equal(t, openshiftClient.AppliedLabels["actuator.nine.ch/pull-request"], strconv.Itoa(event.IssueNumber))

					assert.Equal(t, openshiftClient.AppliedParameters["BRANCH_NAME"], "pr-1")
				})

				t.Run("writes a comment on github", func(t *testing.T) {
					handler.HandleEvent(event)
					githubComment := githubClient.LastComment
					assert.NotNil(t, githubComment, "creates a comment on github")
					assert.Equal(t, "Your environment is being set-up on Openshift. There is no route I can point you to.", githubComment.GetBody())
					assert.Equal(t, "https://github.com/ninech/actuator/issues/1#issuecomment-330230087", githubComment.GetHTMLURL())
				})

				t.Run("posts the url as comment", func(t *testing.T) {
					openshiftClient.NewAppOutputToReturn = openshift.NewAppOutput{Raw: `route "actuator" created`}
					handler.HandleEvent(event)

					githubComment := githubClient.LastComment
					assert.Equal(t, "Your environment is being set-up on Openshift. http://actuator.domain.com", githubComment.GetBody())
				})
			}
		})

		t.Run("EventActionClosed", func(t *testing.T) {
			test.DisableLogging()

			actuator.Config = test.NewDefaultConfig()
			event := test.NewTestEvent(1, github.EventActionClosed, "ninech/actuator")
			openshiftClient := &test.OpenshiftMock{}

			handler := actuator.PullRequestEventHandler{Openshift: openshiftClient}

			t.Run("deletes the objects in openshift", func(t *testing.T) {
				handler.HandleEvent(event)

				assert.Equal(t, openshiftClient.DeletedLabels["actuator.nine.ch/pull-request"], strconv.Itoa(event.IssueNumber))
			})
		})
	})
}
