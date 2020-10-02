# choo-pos-backend

## Note
- Frontend Repo: https://github.com/choobot/choo-pos-frontend
- Backend Repo: https://github.com/choobot/choo-pos-backend
- API Doc: https://choo-pos-backend.herokuapp.com

## Live Demo
- https://choo-pos-backend.herokuapp.com

## Testing with Postman
- Open Postman App, import `postman/choo-pos.postman_collection.json` to your collections, import `postman/choo-pos-prod.postman_environment.json` to your environments and make sure it's selected
- Open Web Browser to https://choo-pos-backend.herokuapp.com/user/login?callback=/dummy
- Login with your LINE Account
- It will redirect back to https://choo-pos-backend.herokuapp.com/dummy?visa=VISA_VALUE
- Copy `VISA_VALUE` e.g. `c2d5c7eb-e167-4221-8793-77a5d752c94c`
- Open Postman, open `Get Token` request, and then replace `visa` query parameter by `VISA_VALUE`, adn then send request, it will response back with your token
- Now you can call the other requests - `User`, `Get All Product`, `Create Product`, `Get All User Log`, `Update Cart`, and `Logout`

## Prerequisites for Development
- Mac or Linux which can run shell script
- Docker
- ngrok CLI
- Heroku CLI (for Production Deployment only)

## Local Running and Expose to the internet for Development
- Create LINE Login channel in LINE Developers Console
- `$ ./local-dev-server.sh`
- Config environment variables in `.env` (see example in `.env.example`)
- The Callback URL for LINE Login API will be something like https://cebccce9ede4.ngrok.io/auth config it in LINE Developers Console
- `$ ./dev.sh`

## Unit Testing
- Running locally (see above)
- `$ ./test.sh`

## Production Deployment
- Create LINE Login channel in LINE Developers Console
- Create Heroku App with ClearDB MySQL add-on in Heroku Dashboard
- Config environment variables in `.env` (see example in `.env.example`)
- The Callback URL for LINE Login API will be something like https://choo-pos-backend.herokuapp.com/auth config it in LINE Developers Console
- `$ ./prod.sh`

## Tech Stack
- Go
- Echo
- gorilla/sessions
- go-playground/validator
- stretchr/testify
- GoMock
- RESTful API
- OAuth 2.0
- LINE Login API
- MySQL
- Docker
- Heroku
- Postman