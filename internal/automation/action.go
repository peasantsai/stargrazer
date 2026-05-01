package automation

// Action is the kind of operation a step performs.
type Action string

const (
	ActionNavigate       Action = "navigate"
	ActionClick          Action = "click"
	ActionDoubleClick    Action = "doubleClick"
	ActionType           Action = "type"
	ActionKeyDown        Action = "keyDown"
	ActionKeyUp          Action = "keyUp"
	ActionWait           Action = "wait"
	ActionWaitForElement Action = "waitForElement"
	ActionEvaluate       Action = "evaluate"
	ActionScroll         Action = "scroll"
	ActionHover          Action = "hover"
	ActionSetViewport    Action = "setViewport"
	ActionTemplate       Action = "template"
)

// AllActions returns every defined Action constant.
// Used by exhaustiveness checks in tests so the registry can never silently
// drop a new action type.
func AllActions() []Action {
	return []Action{
		ActionNavigate,
		ActionClick,
		ActionDoubleClick,
		ActionType,
		ActionKeyDown,
		ActionKeyUp,
		ActionWait,
		ActionWaitForElement,
		ActionEvaluate,
		ActionScroll,
		ActionHover,
		ActionSetViewport,
		ActionTemplate,
	}
}
