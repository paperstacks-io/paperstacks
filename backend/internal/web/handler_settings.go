package web

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	commonauth "github.com/paperstacks.io/paperstacks/internal/common/server/auth"
	userApp "github.com/paperstacks.io/paperstacks/internal/user/application"
	userDomain "github.com/paperstacks.io/paperstacks/internal/user/domain"
)

const userSettingsSavedMessage = "User settings saved"

type settingsPageData struct {
	pageData
	User userDomain.User
}

func handleSettingsPage(
	logger *slog.Logger,
	tmpl *template.Template,
	hankoAPIURL string,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok || session == nil || !session.IsValid {
			logger.Error("unable to get session from context")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		user, err := userService.CreateIfNotExist(r.Context(), session.UserID, session.Email)
		if err != nil {
			logger.Error("read user", "error", err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		data := settingsPageData{
			pageData: newPageData(r, hankoAPIURL),
			User:     user,
		}

		renderTemplate(w, r, tmpl, data)
	})
}

func handleUserSettingsUpdate(
	logger *slog.Logger,
	tmpl *template.Template,
	userService *userApp.UserService,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, ok := commonauth.SessionFromContext(r.Context())
		if !ok || session == nil || !session.IsValid {
			_ = renderErrorToast(w, tmpl, "Unauthorized. Please log in to update user settings.")
			return
		}

		user, err := userService.CreateIfNotExist(r.Context(), session.UserID, session.Email)
		if err != nil {
			logger.Error("read user before settings update", "userId", session.UserID, "error", err.Error())
			_ = renderErrorToast(w, tmpl, "User settings could not be loaded. Please refresh and try again.")
			return
		}

		user.ORCID = strings.TrimSpace(r.FormValue("orcid"))
		if err := userService.Update(r.Context(), session.UserID, user); err != nil {
			logger.Error("update user settings", "userId", session.UserID, "error", err.Error())

			switch {
			case errors.Is(err, userDomain.ErrInvalidUser):
				_ = renderErrorToast(w, tmpl, "ORCID must use the 0000-0000-0000-0000 format.")
			case errors.Is(err, userDomain.ErrExternalIDMismatch):
				_ = renderErrorToast(w, tmpl, "You are not allowed to update these user settings.")
			default:
				_ = renderErrorToast(w, tmpl, "Failed to save user settings: "+err.Error())
			}
			return
		}

		err = renderSuccessToast(w, tmpl, userSettingsSavedMessage)
		if err != nil {
			logger.Error("render user settings", "error", err.Error())
		}
	})
}
