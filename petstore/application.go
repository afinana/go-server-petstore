package petstore

import (
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
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func NewLog(inLog *log.Logger, errLog *log.Logger,
	pets *PetModel, stores *StoreModel, users *UserModel) *Application {

	// Initialize a new instance of application containing the dependencies.
	app := &Application{errorLog: errLog, infoLog: inLog, pets: pets, stores: stores, users: users}
	return app

}

func (app *Application) enableCors(w *http.ResponseWriter, r *http.Request) {

	allowedHeaders := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token"
	if origin := r.Header.Get("Origin"); origin != "" {
		(*w).Header().Set("Access-Control-Allow-Origin", origin)
		(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		(*w).Header().Set("Access-Control-Allow-Headers", allowedHeaders)
		(*w).Header().Set("Access-Control-Expose-Headers", "Authorization")
		(*w).Header().Set("Access-Control-Allow-Credentials", "true")
	}

}
