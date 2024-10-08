package api

import (
	"database/sql"
	"fmt"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/errors"
	. "github.com/RobsonDevCode/BankApi/api/Setting"
	_ "github.com/microsoft/go-mssqldb"
	"log"
	"strings"
)

type AccountRepositoryInterface interface {
	CreateAccount(request *Account) error
	DeleteAccount(int) (string, error)
	DeleteMutipleAccounts(ids []*int) error
	UpdateAccount(*Account) error
	GetAccountById(int) (*Account, error)
	GetAccountsWithGoldMemberShip() ([]*Account, error)
}

type SQLStore struct {
	db *sql.DB
}

func NewAccountRepository() (*SQLStore, error) {
	db, err := sql.Open("sqlserver", Config.ConnectionStrings.BankDb)

	if err != nil {
		return nil, err
	}

	//ping db to test connection is valid
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &SQLStore{
		db: db,
	}, nil

}

func (repo *SQLStore) Init() error {
	return repo.createAccountTable()
}

func (repo *SQLStore) createAccountTable() error {
	query := `IF NOT EXISTS (SELECT * FROM sys.objects WHERE name = 'account_info') CREATE TABLE account_info(
 			  id int IDENTITY(1,1) PRIMARY KEY,
              firstName varchar(150),
              lastName varchar(150), 
              number bigint  , 
	          balance float,
              createdAt DateTime)`

	_, err := repo.db.Exec(query)

	return err

}

func (repo *SQLStore) CreateAccount(account *Account) error {

	userExists, err := repo.doesUserExist(account.FirstName, account.LastName)
	if err != nil {
		return err
	}
	if userExists {
		return errors.New("Account already exists")
	}

	query := `INSERT INTO account_info VALUES (
              @firstName,
              @lastName,
              @number,
              @balance,
              @createdAt)`

	result, err := repo.db.Exec(query,
		sql.Named("firstName", account.FirstName),
		sql.Named("lastName", account.LastName),
		sql.Named("number", account.Number),
		sql.Named("balance", account.Balance),
		sql.Named("createdAt", account.CreatedAt))

	if err != nil {
		return err
	}
	rowsAff, err := result.RowsAffected()

	if err != nil {
		return err
	}
	if rowsAff != 1 {
		return errors.New("Something went wrong, please contact IT support")
	}

	return nil
}

func (repo *SQLStore) UpdateAccount(*Account) error {
	return nil
}

func (repo *SQLStore) DeleteAccount(id int) (string, error) {

	query := `DELETE FROM account_info WHERE id = @id`

	result, err := repo.db.Exec(query, sql.Named("id", id))
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return "", err
	}

	if rowsAffected != 1 {
		return "", errors.New("Something went wrong when deleting account, Please try again or contact support")
	}

	return fmt.Sprintf("Account %s has been deleted", id), err
}

func (repo *SQLStore) DeleteMutipleAccounts(ids []*int) error {

	//create placeholder foreach id like "@p1, @p2, @p3,...
	idsToQuery := make([]string, len(ids))
	for i := range ids {
		idsToQuery[i] = fmt.Sprintf("@p%d", i+1)
	}

	placeholder := strings.Join(idsToQuery, ", ")
	query := fmt.Sprintf(`DELETE FROM account_info WHERE id In (%s)`, placeholder)

	args := make([]interface{}, len(ids))

	for i, id := range ids {
		args[i] = sql.Named(fmt.Sprintf("p%d", i+1), *id)
	}

	result, err := repo.db.Exec(query, args...)

	if err != nil {
		return err
	}

	res, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if res > 0 {
		return errors.New("No Accounts where deleted")
	}

	return nil
}

func (repo *SQLStore) GetAccountById(id int) (*Account, error) {

	query := `SELECT * FROM account_info WHERE id = @id`

	rows, err := repo.db.Query(query, sql.Named("id", id))

	//Check to see if the error is a no data error
	if errors.Is(err, sql.ErrNoRows) {

	}
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		account, scanerr := scanIntoAccount(rows)
		if scanerr != nil {
			return nil, err
		}

		return account, nil
	}

	return nil, fmt.Errorf(`Account not found with id %d`, id)
}

func (repo *SQLStore) GetAccountsWithGoldMemberShip() ([]*Account, error) {
	query := `SELECT * FROM account_info WHERE gold_member = true`

	rows, err := repo.db.Query(query)

	if err != nil {
		return nil, err
	}

	var accounts []*Account

	for rows.Next() {
		account, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil

}

func (repo *SQLStore) doesUserExist(firstName, lastName string) (bool, error) {
	query := `SELECT COUNT(*) FROM [dbo].[user_list]WHERE user_firstname = @firstName AND user_surname = @lastName`

	var count int16

	err := repo.db.QueryRow(query,
		sql.Named("firstName", firstName),
		sql.Named("lastname", lastName)).Scan(&count)

	if err != nil {
		return false, err
	}
	defer func() {
		if err = repo.db.Close(); err != nil {
			log.Fatalf("Error closing DB: %v", err)
		}
	}()

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}
	var err error

	if err = rows.Scan(&account.Id, &account.FirstName, &account.LastName,
		&account.Number, &account.Balance, &account.GoldMemeber, &account.CreatedAt); err != nil {
		return nil, err
	}

	return account, err
}
