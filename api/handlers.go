package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/polypmer/mappyr-backend/database"
)

func Index(w http.ResponseWriter, r *http.Request) {

	//fmt.Println(r.Header["Authentication"])
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	w.WriteHeader(http.StatusOK)
	//err = json.NewEncoder(w).Encode
	fmt.Fprintln(w, "Index")
}

/* DB Read */

func ShowComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println(err)
	}
	c, err := database.ReadComment(db, id)
	if err != nil {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusNotFound) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			fmt.Fprintf(w, "Error JSON encoding %s", err)
		}
	}
}

func ShowComments(w http.ResponseWriter, r *http.Request) {
	comments, err := database.ReadComments(db)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		fmt.Fprintf(w, "Error JSON encoding %s", err)
	}
}

/* DB Write */
// NewComment Writes a new comment to db
// If no user is specified, use 0 for id
func NewComment(w http.ResponseWriter, r *http.Request) {
	// This is taking a POST method
	// TODO: Take User ID from Auth Token
	var comment database.Comment
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		fmt.Println(err)
	}
	err = r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	// Unmarshal, stick into my struct
	err = json.Unmarshal(body, &comment)
	// Unable to marshal?:
	if err != nil {
		// if shit don't work out
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity) //422
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	}
	// Now I have a comment json object
	comment.Date = time.Now()
	// Check for Authentication, and in that case,
	// Add the user to the comment

	id, err := database.WriteComment(db, comment)
	if err != nil {
		fmt.Println(err)
	}
	comment.Id = id
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusCreated)

	// Feed it back into the response Writer
	err = json.NewEncoder(w).Encode(comment)
	if err != nil {
		fmt.Println(err)
	}
}

/* Update DB, Upvote and Downvote */
// UpVote updates a comment and returns the voted comment
func UpVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	comment_id, err := strconv.Atoi(vars["comment_id"])
	token := r.Header["Authentication"][0]
	id, err := AuthId(token)
	fmt.Println(id, err)
	user_id := 1 // TODO: Take User ID from Auth Token
	if err != nil {
		fmt.Println(err)
	}

	// Upvote
	err = database.VoteComment(db, comment_id, user_id, true)
	if err != nil {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusNotFound) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		c, _ := database.ReadComment(db, comment_id)
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			fmt.Fprintf(w, "Error JSON encoding %s", err)
		}
	}
}

// DownVote updates a vote and returns the voted comment
func DownVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	comment_id, err := strconv.Atoi(vars["comment_id"])
	user_id := 1
	if err != nil {
		fmt.Println(err)
	}

	// Upvote
	err = database.VoteComment(db, comment_id, user_id, false)
	if err != nil {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusNotFound) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		c, _ := database.ReadComment(db, comment_id)
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(c)
		if err != nil {
			fmt.Fprintf(w, "Error JSON encoding %s", err)
		}
	}
}

/* Delete */
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	err = database.DeleteComment(db, id)
	if err != nil {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusNotFound) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		success := "Success"
		err = json.NewEncoder(w).Encode(success)
		if err != nil {
			fmt.Fprintf(w, "Error JSON encoding %s", err)
		}
	}
}

/* Authentication */

var signingKey = []byte("secret key")

type Claims struct {
	UserId string `json:"user-id"`
	jwt.StandardClaims
}

// Login authenticates user using jwt token
func Login(w http.ResponseWriter, r *http.Request) {

}

// Logout deletes the cookie
func Logout(w http.ResponseWriter, r *http.Request) {

}

// Validate is middleware for making sure people can't delete the wrong shit
func Validate(w http.ResponseWriter, r *http.Request) {

}

// ParseToken returns the username from an encrypted token
// DEPRECTATED!!
func ParseToken(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token, ok := vars["token"] // very niave to just pass into route
	// this should be in the headers
	if !ok {
		fmt.Println("wheres your token")
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
			// should this be HMAC instead of HS256?
		}
		return signingKey, nil
	})
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Retrieve claims
	var claims map[string]interface{}
	if claims, ok = parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		fmt.Println(claims) // pass into context
	} else {
		fmt.Println("Not OK claims", err)
		http.NotFound(w, r)
		return
	}
	for v, k := range claims {
		fmt.Println(v, k)

	}
	//fmt.Println(claims["username"])
	str := fmt.Sprint(claims["username"])
	w.Write([]byte(str))

}

// AuthId takes a token and returns the user ID encrypted.
func AuthId(token string) (int, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method")
			// should this be HMAC instead of HS256?
		}
		return signingKey, nil
	})
	if err != nil {
		return 0, nil
	}
	// Retrieve claims
	var claims map[string]interface{}
	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		fmt.Println(claims) // pass into context
	} else {
		fmt.Println("Not OK claims", err)
		return 0, errors.New("Unparsable token")
	}
	id, ok := claims["user-id"].(int)
	if !ok {
		return 0, errors.New("The token couldn't find you're id")
	} else {
		return id, nil
	}
}

// NewToken
func NewToken(w http.ResponseWriter, r *http.Request) {
	// in production I'd authenticate against a database before setting the token

	vars := mux.Vars(r)
	id := vars["id"] // strconv.Atoi(vars["id"]) // this should be ID in production

	expireToken := time.Now().Add(time.Hour * 1).Unix()
	expireCookie := time.Now().Add(time.Hour * 1)

	// Set claims from database:
	claims := Claims{
		id, // This is just name for now, but it should be retreived from db first
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "localhost:8080", // Changes in Production
		},
	}
	// Create the token using your claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Signs the token with the secret
	signedToken, _ := token.SignedString(signingKey)

	cookie := http.Cookie{
		Name:     "Auth",
		Value:    signedToken,
		Expires:  expireCookie,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)

	// Either redirect to profile lol
	// or write a json with the cookie
	w.Write([]byte(signedToken))
}

/* Users */
// ShowUsers collects all users
// TODO: New User
// ShowUsers return a list of all users
func ShowUsers(w http.ResponseWriter, r *http.Request) {
	users, err := database.ReadUsers(db)
	if err != nil {
		fmt.Println(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		fmt.Fprintf(w, "Error JSON encoding %s", err)
	}
}

// TODO: Get ID from USername

func UserVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"]) // this should be ID in production
	if err != nil {
		http.NotFound(w, r)
		return
	}

	votes, err := database.UserVotes(id, db)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(votes)
	if err != nil {
		fmt.Fprintf(w, "Error JSON encoding %s", err)
	}

}

func UserComments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"]) // this should be ID in production
	if err != nil {
		http.NotFound(w, r)
		return
	}

	comments, err := database.UserComments(id, db)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(comments)
	if err != nil {
		fmt.Fprintf(w, "Error JSON encoding %s", err)
	}
}
