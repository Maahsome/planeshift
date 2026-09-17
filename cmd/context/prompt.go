package context

import "github.com/AlecAivazis/survey/v2"

type surveyPrompter struct{}

func (surveyPrompter) Workspace() (string, error) {
	var answer string
	err := survey.AskOne(&survey.Input{Message: "Workspace slug:"}, &answer)
	return answer, err
}

func (surveyPrompter) Project(options []string) (string, error) {
	var answer string
	err := survey.AskOne(&survey.Select{Message: "Select a project:", Options: options}, &answer)
	return answer, err
}
