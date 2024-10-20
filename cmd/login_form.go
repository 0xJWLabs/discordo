package cmd

import (
	"errors"

	"github.com/0xJWLabs/discordo/internal/config"
	"github.com/diamondburned/arikawa/v3/api"
	"log/slog"
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

	lf.NewInput("Email", "", "#CDD6F4", "#181825", false)
	lf.NewInput("Password", "", "#CDD6F4", "#181825", true)
	lf.NewInput("Code (optional)", "", "#CDD6F4", "#181825", true)
	lf.NewCheckbox("Remember Me", false, "#F38BA8", "#181825")

	// lf.AddInputField("Email", "", 0, nil, nil)
	// lf.AddPasswordField("Password", "", 0, 0, nil)
	// lf.AddPasswordField("Code (optional)", "", 0, 0, nil)
	lf.AddCheckbox("Remember Me", false, nil)
	lf.AddButton("Login", lf.login)

	// lf.SetFieldBackgroundColor(tcell.GetColor("#181825"))
	// lf.SetFieldTextColor(tcell.GetColor("#CDD6F4"))
	lf.SetButtonBackgroundColor(tcell.ColorBlue)
	lf.SetButtonTextColor(tcell.GetColor("#1E1E2E"))

	lf.SetTitle("Login")
	lf.SetTitleColor(tcell.GetColor(cfg.Theme.TitleColor))
	lf.SetTitleAlign(tview.AlignLeft)

	p := cfg.Theme.BorderPadding
	lf.SetBorder(cfg.Theme.Border)
	lf.SetBorderColor(tcell.GetColor(cfg.Theme.BorderColor))
	lf.SetBorderPadding(p[0], p[1], p[2], p[3])

	lf.updateColor()

	return lf
}

func (lf *loginForm) NewInput(label string, defaultValue string, textColor string, backgroundColor string, password bool) {
	inputField := tview.NewInputField().
		SetLabel(label).
		SetText("").
		SetFieldWidth(0).
		SetAcceptanceFunc(nil).
		SetChangedFunc(nil).
		SetFieldTextColor(tcell.GetColor(textColor)).
		SetFieldBackgroundColor(tcell.GetColor(backgroundColor))

	if password {
		inputField.SetMaskCharacter('*')
	}

	lf.AddFormItem(inputField)
}

func (lf *loginForm) NewCheckbox(label string, defaultValue bool, textColor string, backgroundColor string) {
	checkbox := tview.NewCheckbox().
		SetLabel(label).
		SetChecked(defaultValue).
		SetFieldBackgroundColor(tcell.GetColor(backgroundColor)).
		SetFieldTextColor(tcell.GetColor(textColor))


	lf.AddFormItem(checkbox)
}

func (lf *loginForm) updateColor() {
	emailInput := lf.GetFormItem(0).(*tview.InputField)
	passwordInput := lf.GetFormItem(1).(*tview.InputField)
	codeInput := lf.GetFormItem(2).(*tview.InputField)
	checkbox := lf.GetFormItem(3).(*tview.Checkbox)

	// Log checks for the form items
	backgroundColor := tcell.ColorBlack
	textColor := tcell.ColorWhite
	checkboxColor := tcell.ColorRed

	// Check if form items exist before updating
	if emailInput != nil {
		slog.Info("Email Input: exist")
		emailInput.SetFieldBackgroundColor(backgroundColor)
		emailInput.SetFieldTextColor(textColor)
	}
	if passwordInput != nil {
		slog.Info("Password: exist")
		passwordInput.SetFieldBackgroundColor(backgroundColor)
		passwordInput.SetFieldTextColor(textColor)
	}
	if codeInput != nil {
		slog.Info("Code Input: exist")
		codeInput.SetFieldBackgroundColor(backgroundColor)
		codeInput.SetFieldTextColor(textColor)
	}
	if checkbox != nil {
		slog.Info("Checkbox: exist")
		checkbox.SetFieldBackgroundColor(backgroundColor)
		checkbox.SetFieldTextColor(checkboxColor)
	}
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
