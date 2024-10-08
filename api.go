package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type ApiError struct {
	Error string `json:"error"`
}

type APIServer struct {
	listenAddr  string
	accountRepo AccountRepositoryInterface
}

// NewAPIServer is a constructor function that creates and returns a pointer to a new ApiServer instance.
// It takes the server's listening address as a string parameter and assigns it to the listenAddr field of the ApiServer.
func NewAPIServer(listenAddr string, repository AccountRepositoryInterface) *APIServer {

	return &APIServer{
		listenAddr:  listenAddr,
		accountRepo: repository,
	}
}

func (server *APIServer) Start() error {
	router := mux.NewRouter()

	router.HandleFunc("/account", convertToHttpHandleFunc(server.handleAccount))

	router.HandleFunc("/account/{id}", convertToHttpHandleFunc(server.handleGetAccountById))

	router.HandleFunc("/account/deleteMulti", convertToHttpHandleFunc(server.handleMutipleAccountRemoval))

	router.HandleFunc("/transfer", convertToHttpHandleFunc(server.handleTransfer))

	log.Printf("Listening on %s", server.listenAddr)

	err := http.ListenAndServe(server.listenAddr, router)

	if err != nil {
		return err
	}

	return nil
}

// Account controller
func (server *APIServer) handleAccount(writer http.ResponseWriter, request *http.Request) error {
	switch request.Method {
	case "GET":
		return server.handleGetAccountsWithGoldMemberShip(writer, request)

	case "POST":

		return server.handleCreateAccount(writer, request)

	default:

		return fmt.Errorf("Unsupported method: %s", request.Method)
	}

}

// Account by ID controller
func (server *APIServer) handleAccountById(writer http.ResponseWriter, request *http.Request) error {

	switch request.Method {

	case "GET":
		return server.handleGetAccountById(writer, request)

	case "DELETE":
		return server.handleDeleteAccount(writer, request)

	default:
		return fmt.Errorf("usupported method call: %s", request.Method)
	}
}

// WriteJson write values to json so we can return values
func WriteJson(writer http.ResponseWriter, status int, value any) error {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)

	return json.NewEncoder(writer).Encode(value)
}

func (server *APIServer) handleGetAccountsWithGoldMemberShip(writer http.ResponseWriter, request *http.Request) error {
	accounts, err := server.accountRepo.GetAccountsWithGoldMemberShip()
	//SQL response check, if no content is return we return a 204
	emptyContentCheck, jsonConvertErr := noResponseContentCheck(writer, nil, err)
	if jsonConvertErr != nil {
		return err
	}

	if emptyContentCheck {
		return errors.New("Content not found")
	}
	if err != nil {
		return err
	}

	if accounts == nil {
		return errors.New("Accounts with gold membership returned nil")
	}

	return WriteJson(writer, http.StatusOK, accounts)
}

func (server *APIServer) handleGetAccountById(writer http.ResponseWriter, request *http.Request) error {
	id, err := getId(request)
	if err != nil {
		return err
	}

	account, err := server.accountRepo.GetAccountById(id)
	if err != nil {

		//SQL response check, if no content is return we return a 204
		emptyContentCheck, jsonConvertErr := noResponseContentCheck(writer, id, err)
		if jsonConvertErr != nil {
			return err
		}

		if emptyContentCheck {
			return errors.New("Content not found")
		}
	}

	if err != nil {
		return err
	}

	return WriteJson(writer, http.StatusOK, account)

}

func (server *APIServer) handleCreateAccount(writer http.ResponseWriter, req *http.Request) error {
	accountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(req.Body).Decode(accountRequest); err != nil {
		return err
	}

	account := NewAccount(accountRequest.FirstName, accountRequest.LastName)
	if err := server.accountRepo.CreateAccount(account); err != nil {
		return err
	}

	return WriteJson(writer, http.StatusOK, accountRequest)
}

func (server *APIServer) handleDeleteAccount(writer http.ResponseWriter, request *http.Request) error {
	id, err := getId(request)
	if err != nil {
		return err
	}

	responseMessage, err := server.accountRepo.DeleteAccount(id)
	if err != nil {
		return err
	}

	return WriteJson(writer, http.StatusOK, responseMessage)
}

func (server *APIServer) handleMutipleAccountRemoval(writer http.ResponseWriter, request *http.Request) error {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}
	defer request.Body.Close()
	var ids []*int
	err = nil

	if err = json.Unmarshal(body, &ids); err != nil {
		return err
	}
	err = nil

	if err = server.accountRepo.DeleteMutipleAccounts(ids); err != nil {
		return err
	}
	responseMessage := fmt.Sprintf("Accounts %s have been deleted", ids)

	return WriteJson(writer, http.StatusOK, responseMessage)
}

func (server *APIServer) handleTransfer(writer http.ResponseWriter, request *http.Request) error {
	transferReq := new(TranferRequest)

	if err := json.NewDecoder(request.Body).Decode(transferReq); err != nil {
		return err
	}
	defer request.Body.Close()

	return WriteJson(writer, http.StatusOK, transferReq)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func convertToHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if err := f(writer, request); err != nil {
			WriteJson(writer, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getId(request *http.Request) (int, error) {
	idStr := mux.Vars(request)["id"]
	id, err := strconv.Atoi(idStr)

	if err != nil {
		return 0, fmt.Errorf("Error invalid id given %s", idStr)
	}

	return id, nil
}

func noResponseContentCheck(writer http.ResponseWriter, T any, err error) (bool, error) {
	if errors.Is(err, sql.ErrNoRows) {
		var message string
		if T == nil {
			message = fmt.Sprint("Connection Succesful but Content Cannot be fount")
		} else {
			message = fmt.Sprintf("Connection Succesful: %s cannot be found", T)
		}

		return true, WriteJson(writer, http.StatusNoContent, message)
	}

	return false, nil
}
