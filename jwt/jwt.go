package jwt

import (
	"cane-project/model"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"
)

// MySigningKey Variable
var MySigningKey = []byte("secret")

// TokenAuth Variable
var TokenAuth *jwtauth.JWTAuth

func init() {
	TokenAuth = jwtauth.New("HS256", MySigningKey, nil)
}

// GenerateJWT Function
func GenerateJWT(account model.UserAccount) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = account.FirstName + " " + account.LastName
	claims["time"] = time.Now().Unix()
	// claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(MySigningKey)

	if err != nil {
		fmt.Println("Something Went Wrong: ", err.Error())
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT Function
func ValidateJWT(t string) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}
		return MySigningKey, nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	if token.Valid {
		fmt.Println("Valid Token!")
	} else {

		fmt.Println("Not Authorized!")
	}
}
