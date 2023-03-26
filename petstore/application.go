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
	orders   *OrderModel
	users    *UserModel
}

func (app *Application) serverError(w http.ResponseWriter, err error) {

	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	fmt.Printf(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func NewLog(inLog *log.Logger, errLog *log.Logger,
	pets *PetModel, orders *OrderModel, users *UserModel) *Application {

	// Initialize a new instance of application containing the dependencies.
	app := &Application{errorLog: errLog, infoLog: inLog, pets: pets, orders: orders, users: users}
	return app

}
