package account

import (
	"cane-project/database"
	"cane-project/jwt"
	"cane-project/model"
	"cane-project/util"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	//"github.com/mongodb/mongo-go-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fatih/structs"
	"github.com/go-chi/chi"
	"github.com/mitchellh/mapstructure"
	//"github.com/mongodb/mongo-go-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Login Function
func Login(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount
	var login map[string]interface{}
	var filter primitive.M
	var userName string
	var password string
	var ok bool

	json.NewDecoder(r.Body).Decode(&login)

	userName, ok = login["username"].(string)
	if !ok {
		util.RespondWithError(w, http.StatusBadRequest, "must provide username field")
	} else {
		filter = primitive.M{
			"username": userName,
		}
	}

	findVal, _ := database.FindOne("accounts", "users", filter)
	mapstructure.Decode(findVal, &account)

	password, ok = login["password"].(string)
	if !ok {
		util.RespondWithError(w, http.StatusBadRequest, "must provide password field")
	} else {
		if account.Password == password {
			util.RespondwithJSON(w, http.StatusOK, structs.Map(account))
		} else {
			util.RespondWithError(w, http.StatusBadRequest, "invalid login")
		}
	}
}

// CreateUser Function
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var jBody JSONBody

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	if len(jBody) != 6 {
		util.RespondWithError(w, http.StatusBadRequest, "invalid number of keys in user body")
		return
	}

	user, userErr := jBody.ToUser()

	if userErr != nil {
		fmt.Println(userErr)
		util.RespondWithError(w, http.StatusBadRequest, userErr.Error())
		return
	}

	validErr := user.Valid()

	if validErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, validErr.Error())
		return
	}

	if UserExists(user.UserName) {
		util.RespondWithError(w, http.StatusBadRequest, "username already exists")
		return
	}

	userToken, tokenErr := jwt.GenerateJWT(model.UserAccount(user))

	if tokenErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error generating jwt token")
		return
	}

	user.Token = userToken

	_, saveErr := database.Save("accounts", "users", user)

	if saveErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error saving account to database")
		return
	}

	util.RespondwithString(w, http.StatusCreated, "")
}

// UpdateUser Function
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var currUser User
	var jBody JSONBody
	var value interface{}
	var ok bool

	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	mapErr := mapstructure.Decode(findVal, &currUser)

	if mapErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling user details")
		return
	}

	decodeErr := json.NewDecoder(r.Body).Decode(&jBody)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	user, userErr := jBody.ToUser()

	if userErr != nil {
		fmt.Println(userErr)
		util.RespondWithError(w, http.StatusBadRequest, userErr.Error())
		return
	}

	value, ok = jBody["username"]
	if ok {
		util.RespondWithError(w, http.StatusBadRequest, "cannot modify username")
		return
	}

	if user.FirstName != "" {
		currUser.FirstName = user.FirstName
	}

	if user.LastName != "" {
		currUser.LastName = user.LastName
	}

	if user.Password != "" {
		currUser.Password = user.Password
	}

	if user.Privilege > 0 {
		currUser.Privilege = user.Privilege
	}

	value, ok = jBody["enable"]
	if ok {
		currUser.Enable = value.(bool)
	}

	validErr := currUser.Valid()

	if validErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, validErr.Error())
		return
	}

	_, replaceErr := database.ReplaceOne("accounts", "users", filter, structs.Map(currUser))

	if replaceErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error updating user in database")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// DeleteUser Function
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	deleteErr := database.Delete("accounts", "users", filter)

	if deleteErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	util.RespondwithString(w, http.StatusOK, "")
}

// GetUser Function
func GetUser(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	filter := primitive.M{
		"username": chi.URLParam(r, "username"),
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	mapErr := mapstructure.Decode(findVal, &account)

	if mapErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling user details")
		return
	}

	util.RespondwithJSON(w, http.StatusOK, account)
}

// GetUsers Function
func GetUsers(w http.ResponseWriter, r *http.Request) {
	var opts options.FindOptions
	var accountList []string

	projection := primitive.M{
		"_id":      0,
		"username": 1,
	}

	opts.SetProjection(projection)

	findVals, findErr := database.FindAll("accounts", "users", primitive.M{}, opts)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "no users found")
		return
	}

	for _, user := range findVals {
		accountList = append(accountList, user["username"].(string))
	}

	util.RespondwithJSON(w, http.StatusOK, map[string][]string{"users": accountList})
}

// GetUserFromDB Function
func GetUserFromDB(userName string) (model.UserAccount, error) {
	var user model.UserAccount

	filter := primitive.M{
		"username": userName,
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		return user, errors.New("user no found")
	}

	mapErr := mapstructure.Decode(findVal, &user)

	if mapErr != nil {
		return user, errors.New("error unmarshaling user details")
	}

	return user, nil
}

// UserExists Function
func UserExists(username string) bool {
	filter := primitive.M{
		"username": username,
	}

	_, findErr := database.FindOne("accounts", "users", filter)

	if findErr == nil {
		return true
	}

	return false
}

// ValidateUserToken Function
func ValidateUserToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	decodeErr := json.NewDecoder(r.Body).Decode(&account)

	if decodeErr != nil {
		fmt.Println(decodeErr)
		util.RespondWithError(w, http.StatusBadRequest, "error decoding json body")
		return
	}

	filter := primitive.M{
		"username": account.UserName,
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "invalid username")
		return
	}

	mapErr := mapstructure.Decode(findVal, &account)

	if mapErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling user details")
		return
	}

	jwt.ValidateJWT(account.Token)
}

// RefreshToken Function
func RefreshToken(w http.ResponseWriter, r *http.Request) {
	var account model.UserAccount

	filter := primitive.M{
		"username": chi.URLParam(r, "user"),
	}

	findVal, findErr := database.FindOne("accounts", "users", filter)

	if findErr != nil {
		fmt.Println(findErr)
		util.RespondWithError(w, http.StatusBadRequest, "invalid account")
		return
	}

	mapErr := mapstructure.Decode(findVal, &account)

	if mapErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "error unmarshaling user details")
		return
	}

	newToken, _ := jwt.GenerateJWT(account)

	update := primitive.M{
		"$set": primitive.M{
			"token": newToken,
		},
	}

	updateVal, updateErr := database.FindAndUpdate("accounts", "users", filter, update)

	if updateErr != nil {
		util.RespondWithError(w, http.StatusBadRequest, "token refresh failed")
		return
	}

	util.RespondwithJSON(w, http.StatusCreated, updateVal)
}

// ToUser Function
func (j JSONBody) ToUser() (User, error) {
	var u User

	mapErr := mapstructure.Decode(j, &u)

	if mapErr != nil {
		return u, errors.New("error unmarshaling user details")
	}

	return u, nil
}

// Valid Function
func (u User) Valid() error {
	if u.FirstName == "" {
		return errors.New("fname cannot be empty")
	}

	if u.LastName == "" {
		return errors.New("lname cannot be empty")
	}

	if u.UserName == "" {
		return errors.New("username cannot be empty")
	}

	if u.Password == "" {
		return errors.New("password cannot be empty")
	}

	if !ValidPrivilege(u.Privilege) {
		return errors.New("must use valid privilege level")
	}

	// if u.Enable {
	// 	return false, errors.New("account is missing enable field")
	// }

	return nil
}

// ValidPrivilege Function
func ValidPrivilege(level int) bool {
	for _, privilegeLevel := range PrivilegeLevels {
		if level == privilegeLevel {
			return true
		}
	}
	return false
}
