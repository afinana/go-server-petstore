package petstore

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

type Application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
	pets     *PetModel
	stores   *StoreModel
	users    *UserModel
}

func (app *Application) serverError(w http.ResponseWriter, err error) {
	output := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, output)
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
	w.Write(js)
}

func NewLog(inLog *log.Logger, errLog *log.Logger,
	pets *PetModel, stores *StoreModel, users *UserModel) *Application {

	// Initialize a new instance of application containing the dependencies.
	app := &Application{errorLog: errLog, infoLog: inLog, pets: pets, stores: stores, users: users}
	return app

}
