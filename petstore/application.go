package petstore

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

type Application struct {
	errorLog             *log.Logger
	infoLog              *log.Logger
	pets                 *PetModel
	stores               *StoreModel
	users                *UserModel
	enableAuthValidation *bool
}

// IsAuthValidationEnabled checks whether gateway header validation (X-User-Id,
// X-User-Roles, X-User-Email, X-Gateway-Secret) and ABAC enforcement is enabled.
// Default value is true (enabled).
func (app *Application) IsAuthValidationEnabled() bool {
	if app != nil && app.enableAuthValidation != nil {
		return *app.enableAuthValidation
	}

	val := os.Getenv("ENABLE_AUTH_VALIDATION")
	if val == "" {
		val = os.Getenv("ENABLE_HEADER_VALIDATION")
	}
	if val == "" {
		// Default is enabled
		return true
	}

	lower := strings.ToLower(strings.TrimSpace(val))
	return lower != "false" && lower != "0" && lower != "off" && lower != "no"
}

// SetAuthValidationEnabled sets the auth/header validation state on the application.
func (app *Application) SetAuthValidationEnabled(enabled bool) {
	app.enableAuthValidation = &enabled
}

func (app *Application) serverError(w http.ResponseWriter, err error) {
	output := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	_ = app.errorLog.Output(2, output)
	app.ErrorResponse(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func (app *Application) ErrorResponse(w http.ResponseWriter, status int, message string) {
	app.WriteJSON(w, status, map[string]string{"error": message})
}

func (app *Application) WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	js, err := json.Marshal(data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)
	_, _ = w.Write(js)
}

func NewLog(inLog *log.Logger, errLog *log.Logger,
	pets *PetModel, stores *StoreModel, users *UserModel) *Application {

	// Initialize a new instance of application containing the dependencies.
	app := &Application{errorLog: errLog, infoLog: inLog, pets: pets, stores: stores, users: users}
	return app

}
