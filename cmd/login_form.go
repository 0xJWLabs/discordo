package cmd

import (
	"errors"

	"github.com/0xJWLabs/discordo/internal/config"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
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

	inputBackgroundColor := tcell.GetColor("#181825")
	buttonBackgroundColor := tcell.GetColor("#89B4FA")
	inputTextColor := tcell.GetColor("#CDD6F4")
	checkboxTextColor := tcell.GetColor("#F38BA8")
	buttonTextColor := tcell.GetColor("#1E1E2E")

	emailInput := lf.AddInputField("Email", "", 0, nil, nil)
	passwordInput := lf.AddPasswordField("Password", "", 0, 0, nil)
	codeInput := lf.AddPasswordField("Code (optional)", "", 0, 0, nil)
	checkbox := lf.AddCheckbox("Remember Me", true, nil)
	loginButton := lf.AddButton("Login", lf.login)

	emailInput.SetFieldBackgroundColor(inputBackgroundColor)
	passwordInput.SetFieldBackgroundColor(inputBackgroundColor)
	codeInput.SetFieldBackgroundColor(inputBackgroundColor)
	checkbox.SetFieldBackgroundColor(inputBackgroundColor)
	loginButton.SetButtonBackgroundColor(buttonBackgroundColor)

	emailInput.SetFieldTextColor(inputTextColor)
	passwordInput.SetFieldTextColor(inputTextColor)
	codeInput.SetFieldTextColor(inputTextColor)
	checkbox.SetFieldTextColor(checkboxTextColor)
	loginButton.SetButtonTextColor(buttonTextColor)

	lf.SetTitle("Login")
	lf.SetTitleColor(tcell.GetColor(cfg.Theme.TitleColor))
	lf.SetTitleAlign(tview.AlignLeft)

	p := cfg.Theme.BorderPadding
	lf.SetBorder(cfg.Theme.Border)
	lf.SetBorderColor(tcell.GetColor(cfg.Theme.BorderColor))
	lf.SetBorderPadding(p[0], p[1], p[2], p[3])

	return lf
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
