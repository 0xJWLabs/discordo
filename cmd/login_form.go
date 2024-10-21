package cmd

import (
	"errors"

	"github.com/0xJWLabs/discordo/internal/config"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/gdamore/tcell/v2"
	"github.com/0xJWLabs/tview"
	"github.com/zalando/go-keyring"
)

type doneFn func(token string, err error)

type loginForm struct {
	*tview.Form
	done doneFn
}

func newLoginForm(done doneFn, cfg *config.Config) *loginForm {
	if done == nil {
		done = func(_ string, _ error) {}
	}

	lf := &loginForm{
		Form: tview.NewForm(),
		done: done,
	}

	lf.SetBackgroundColor(tcell.GetColor("#1E1E2E"))
	lf.SetFieldTextColor(tcell.GetColor("#CDD6F4"))
	lf.SetFieldBackgroundColor(tcell.GetColor("#181825"))
	lf.SetLabelColor(tcell.GetColor("#CDD6F4"))
	lf.SetCheckboxBackgroundColor(tcell.GetColor("#181825"))
	lf.SetCheckboxTextColor(tcell.GetColor("#F38BA8"))
	lf.SetButtonBackgroundColor(tcell.ColorBlue)
	lf.SetButtonTextColor(tcell.GetColor("#181825"))

	lf.NewInput("Email", "", false)
	lf.NewInput("Password", "", true)
	lf.NewInput("Code (optional)", "", true)
	lf.NewCheckbox("Remember Me", false)


	lf.AddButton("Login", lf.login)

	lf.SetTitle("Login")
	lf.SetTitleColor(tcell.GetColor(cfg.Theme.TitleColor))
	lf.SetFocusTitleColor(tcell.GetColor(cfg.Theme.FocusTitleColor))
	lf.SetTitleAlign(tview.AlignLeft)
	lf.SetTitlePadding(1, 1)

	p := cfg.Theme.BorderPadding
	lf.SetBorder(cfg.Theme.Border)
	lf.SetBorderColor(tcell.GetColor(cfg.Theme.BorderColor))
	lf.SetFocusBorderColor(tcell.GetColor(cfg.Theme.FocusBorderColor))
	lf.SetBorderPadding(p[0], p[1], p[2], p[3])

	return lf
}

func (lf *loginForm) NewInput(label string, defaultValue string, password bool) {
	inputField := tview.NewInputField().
		SetLabel(label).
		SetText("").
		SetFieldWidth(0).
		SetAcceptanceFunc(nil).
		SetChangedFunc(nil)

	if password {
		inputField.SetMaskCharacter('*')
	}

	lf.AddFormItem(inputField)
}

func (lf *loginForm) NewCheckbox(label string, defaultValue bool) {
	checkbox := tview.NewCheckbox().
		SetLabel(label).
		SetChecked(defaultValue)


	lf.AddFormItem(checkbox)
}

func (lf *loginForm) login() {
	email := lf.GetFormItem(0).(*tview.InputField).GetText()
	password := lf.GetFormItem(1).(*tview.InputField).GetText()
	if email == "" || password == "" {
		return
	}

	// Create a new API client without an authentication token.
	apiClient := api.NewClient("")
	// Log in using the provided email and password.
	lr, err := apiClient.Login(email, password)
	if err != nil {
		lf.done("", err)
		return
	}

	// If the account has MFA-enabled, attempt to log in using the provided code.
	if lr.MFA && lr.Token == "" {
		code := lf.GetFormItem(2).(*tview.InputField).GetText()
		if code == "" {
			lf.done("", errors.New("code required"))
			return
		}

		lr, err = apiClient.TOTP(code, lr.Ticket)
		if err != nil {
			lf.done("", err)
			return
		}
	}

	if lr.Token == "" {
		lf.done("", errors.New("missing token"))
		return
	}

	rememberMe := lf.GetFormItem(3).(*tview.Checkbox).IsChecked()
	if rememberMe {
		go func() {
			if err := keyring.Set(config.Name, "token", lr.Token); err != nil {
				lf.done("", err)
			}
		}()
	}

	lf.done(lr.Token, nil)
}
