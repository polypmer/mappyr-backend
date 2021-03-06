# Mappyr Backend api

Soo there are a not too many things this can do yet, but here it is.

## Routes

    /

Index doesn't really do anything

    /comment/{id}

returns a specific comment by it's id, the response will look like this:

`{"id":2,"title":"Great food!","description":"Although crap service","latitude":41.353,"longitude":-71.113,"upvotes":-1,"downvotes":1,"date":"2016-10-11T11:27:20.19479779-04:00","user-id":`

    /comments

returns a json list of **all** comments.

    /new

the create route accepts the POST method, it only needs a title, description, and the latitude and longitude. Notice the numbers aren't in quotes, here's an example curl:

`curl -H "Content-Type: application/json" -d '{"title":"This place sux0rs", "description":"Yikes","latitude":41.33894, "longitude":-71.666}' http://localhost:8080/new `

    /upvote/{id}
    //OR
    /downvote/{id}

These accept a GET method, and they update the count of downvotes or upvotes for a comment row. It'll then return the very comment which is being commented (maybe this'll just want to be ignored, but I thought'd be a good idea to return *something*.


    /delete/{id}

this accepts an id, and deletes that post. Uhm, this *needs* some sort of auth.


# API v2, not yet deployed



## User Endpoints

    @POST
    /signup

Respond with the new user in json.

TODO: redirect to login.

The password will be in a hash, to log in, the original password must be entered (you can't just copy and paste the response from signup.

    @POST
    /login

Respond with a Auth Token to be included in Headers

     /all/users

Respond with list of all users.

    /votes/{user_id}

Respond with list of votes made by a user

    /comments/{user_id}

Respond with list of comments posted by a user

## Auth

Future iterations of this application should lock all API calls behind a token provided only to official applications; As it stands, anyone sniffing packets could get lists of users etc.

The Actions (below) require an Auth Token in the Header. Add this field to header:

    Authentication: {TOKEN}

To get this token first sign up:

    @POST
    /signup

Include json with the sign up information (the password then gets hashed before entered into the database, so don't loose it.
>{"username":"david", "password":"supersecret", "email":"david@gmail.com"}

So signing up returns a JSON with the user (and their ID), but not with the authentication token. Signing up (at the moment) does minimal cleaning. No duplicate emails or usernames, but it does not test for valid emails.

After signing up, try logging in to get the authentication token:

    @POST
    /login

> {"username":"david", "password":"supersecret"}

This will return an authentication token with user information signed (with the secret private key and so irreproducible without *this* auth process) in JSON. This is what must be included in the Header requests with the Authentication: option. The JSON will look like this:

> {
>    "Authentication": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyLWlkIjoiMTMiLCJleHAiOjE0NzY5OTkyMzMsImlzcyI6ImxvY2FsaG9zdDo4MDgwIn0.87VgAqtQOHKixe8_TikCFJKGCQiSI5e80cm5zHnA4A8"
>}

This token will tell the database who it is that is performing the actions

    @POST
    /new

Send data of the username

    @GET
    /upvote/{comment_id}

    @GET
    /downvote/{comment_id}

`downvote` and `upvote` will vote on an ID. The authentication token inside the header will indicate who it is that is voting.

## Comment Endpoints

    @GET
    /all/commments

Returns all comments possible. This, in production settings is probably unlikely/not so useful. Use instead the POST `/local`

    @POST
    /Local

Get all comments *within* a certain longitude and latitude. Post with json such as:

> {"lat-max":50.0,"lat-min":0.0, "lon-max":85.0, "lon-min":50.0}

    @GET
    /comment/{id}

Get a single comment by **id**, this will return the posting User within the json, such as:


> {
>	"id": 1,
>	"title": "SUsh1",
>	"description": "THIS Rox0rs",
>	"latitude": 45,
>	"longitude": 88,
>	"upvotes": 0,
>	"downvotes": 0,
>	"date": "2016-10-20T12:13:51.029814Z",
>	"UserId": 2,
>	"user": {
>		"id": 2,
>		"username": "Simone",
>		"date": "2016-10-20T12:13:50.929836Z",
>		"email": "simone@gmail.com"
>	}
}
