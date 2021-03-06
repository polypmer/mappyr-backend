package api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"crypto/sha1"
	"encoding/base64"
	"errors"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/polypmer/mappyr-backend/database"
)

/* ############################################################
 Index
############################################################ */

func Index(w http.ResponseWriter, r *http.Request) {

	//fmt.Println(r.Header["Authentication"])
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	w.WriteHeader(http.StatusOK)
	//err = json.NewEncoder(w).Encode
	fmt.Fprintln(w, "Index")
}

/* ############################################################
 Comments and Voting
############################################################ */

// ShowComment returns a json of a comment with the id.
// This comment will have the User in the User field.
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

// ShowComments encodes json of ALLL comments
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

type Local struct {
	LatMax float64 `json:"lat-max"`
	LatMin float64 `json:"lat-min"`
	LonMax float64 `json:"lon-max"`
	LonMin float64 `json:"lon-min"`
}

// ShowLocalComments
// A post for finding the comments by local
// TODO: Add users to comments
func ShowLocalComments(w http.ResponseWriter, r *http.Request) {
	local := Local{}
	// Body has the JSON for locality
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		fmt.Println(err)
	}
	err = r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	// Unmarshal, stick into my struct
	err = json.Unmarshal(body, &local)
	if err != nil {
		fmt.Println(err)
	}
	comments, err := database.ReadLocalComments(db,
		[2]float64{local.LatMin, local.LatMax}, [2]float64{local.LonMin, local.LonMax})
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
	var user_id int
	var err error
	// This is taking a POST method
	// Authentication from Token to get User ID
	var comment database.Comment
	if _, ok := r.Header["Authentication"]; ok {
		token := r.Header["Authentication"][0]
		user_id, err = AuthId(token)
		if err != nil {
			w.Header().Set("Content-Type",
				"application/json;charset=UTF-8")
			w.WriteHeader(http.StatusNotFound) // No User
			err = json.NewEncoder(w).Encode(err)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusForbidden) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	}
	// Body has the JSON for comment info (including LAT and LONG)
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
	comment.UserId = user_id
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
	var user_id int
	vars := mux.Vars(r)
	comment_id, err := strconv.Atoi(vars["comment_id"])
	if err != nil {
		fmt.Println(err)
	}
	if _, ok := r.Header["Authentication"]; ok {
		token := r.Header["Authentication"][0]
		user_id, err = AuthId(token)
		if err != nil {
			w.Header().Set("Content-Type",
				"application/json;charset=UTF-8")
			w.WriteHeader(http.StatusNotFound) // Doesn't exist
		}
	} else {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusForbidden) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
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
	var user_id int
	vars := mux.Vars(r)
	comment_id, err := strconv.Atoi(vars["comment_id"])
	if err != nil {
		fmt.Println(err)
	}
	if _, ok := r.Header["Authentication"]; ok {
		token := r.Header["Authentication"][0]
		user_id, err = AuthId(token)
		if err != nil {
			w.Header().Set("Content-Type",
				"application/json;charset=UTF-8")
			w.WriteHeader(http.StatusNotFound) // Doesn't exist
		}
	} else {
		w.Header().Set("Content-Type",
			"application/json;charset=UTF-8")
		w.WriteHeader(http.StatusForbidden) // Doesn't exist
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	}

	// Downvote
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

/* ############################################################
 Authentication
############################################################ */

var signingKey = []byte("secret key")

type Claims struct {
	UserId string `json:"user-id"`

	jwt.StandardClaims
}

// Signup adds a new user to the database.
func Signup(w http.ResponseWriter, r *http.Request) {
	user := database.User{}
	// Hash here?
	// Add password and username
	// Then login lol
	// Get JSON data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		fmt.Println(err)
	}
	err = r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	// Unmarshal, stick into my struct
	err = json.Unmarshal(body, &user)
	if err != nil {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity) //422
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
	}
	user.Date = time.Now()

	h := sha1.New()
	io.WriteString(h, user.Password)
	hashed := base64.URLEncoding.EncodeToString(h.Sum(nil))
	// CONVERT TO STRING, from uint18?
	user.Password = hashed
	id, err := database.SignUp(db, user)
	if err != nil {
		fmt.Println(err)
	}
	user.Id = id
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if id == -1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	// Feed it back into the response Writer
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		fmt.Println(err)
	}
}

type AuthAttempt struct {
	Name     string `json:"username"`
	Password string `json:"password"`
}

// Login authenticates user and returns the token
// this token must be included in the Header Authentication:
// field for Posting and Voting.
func Login(w http.ResponseWriter, r *http.Request) {
	// POST JSON
	// Hash of pass
	// Check against password
	// return token
	attempt := AuthAttempt{}
	// Hash here?
	// Add password and username
	// Then login lol
	// Get JSON data
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		fmt.Println(err)
	}
	err = r.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	// Unmarshal, stick into my struct
	err = json.Unmarshal(body, &attempt)
	h := sha1.New()
	io.WriteString(h, attempt.Password)
	hashed := base64.URLEncoding.EncodeToString(h.Sum(nil))
	attempt.Password = hashed
	if err != nil {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.WriteHeader(http.StatusUnprocessableEntity) //422
		err = json.NewEncoder(w).Encode(err)
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	ok, id := database.LogIn(db, attempt.Name, attempt.Password)
	if ok {
		token, err := AuthToken(id)
		if err == nil {
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.WriteHeader(http.StatusOK) //200
			j := "{\"Authentication\":\"" + token + "\"}"
			w.Write([]byte(j))
			return
		}
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(http.StatusNotFound) //404
	err = json.NewEncoder(w).Encode(err)
	if err != nil {
		fmt.Println(err)
	}
	return
}

// Logout deletes the cookie.
// There is no cookie. So this doesn't make any sense.
func Logout(w http.ResponseWriter, r *http.Request) {
	// Remove Cookie?
}

// Validate is middleware for making sure people can't delete the wrong shit
func Validate(w http.ResponseWriter, r *http.Request) {
	// Wrapper for functions?
	// For admin???
}

// AuthId takes a token and returns the user ID encrypted.
// Now, this can't be faked because I have the secret signing key.
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
	var ok bool
	if claims, ok = parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		//fmt.Println(claims) // pass into context
	} else {
		fmt.Println("Not OK claims", err)
		return 0, errors.New("Unparsable token")
	}
	id, ok := claims["user-id"].(string)
	if !ok {
		return 0, errors.New("The token couldn't find you're id")
	} else {
		uid, err := strconv.Atoi(id)
		if err != nil {
			return 0, errors.New("This isn't a proper ID")
		}
		return uid, nil
	}
}

// AuthToken takes an ID and returns the token.
func AuthToken(id int) (string, error) {
	expireToken := time.Now().Add(time.Hour * 1).Unix()
	//expireCookie := time.Now().Add(time.Hour * 1)
	uid := strconv.Itoa(id)
	claims := Claims{
		uid, // the user_id
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "localhost:8080", // Changes in Production HEROKU
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// TODO: Catch this error?
	signedToken, _ := token.SignedString(signingKey)

	return signedToken, nil
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

/* ############################################################
 Users
############################################################ */
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

// ShowUser writes json of user
func ShowUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		fmt.Println(err)
	}
	u, err := database.ReadUser(db, id)
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
		err = json.NewEncoder(w).Encode(u)
		if err != nil {
			fmt.Fprintf(w, "Error JSON encoding %s", err)
		}
	}
}

// TODO: Get ID from USername

func UserVotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["user_id"]) // this should be ID in production
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
	id, err := strconv.Atoi(vars["user_id"]) // this should be ID in production
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
